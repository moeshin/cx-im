package core

import (
	"cx-im/src/config"
	"cx-im/src/im"
	"cx-im/src/im/cmd_course_chat_feedback"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/moeshin/go-errs"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HeartTimeout = 30 * time.Second

type Work struct {
	User        *User
	Conn        *websocket.Conn
	Done        chan struct{}
	CacheActive *CacheActive
	HeartTicker *time.Ticker
}

func NewWork(user *User) *Work {
	return &Work{
		User: user,
		Done: make(chan struct{}),
		CacheActive: &CacheActive{
			Mutex: &sync.Mutex{},
			Map:   map[string]int64{},
		},
	}
}

func StartWork(user *User) {
	_, err := user.LogFile.WriteString(strings.Repeat("-", 19) + "\n")
	errs.Print(err)
	work := NewWork(user)
	errs.Print(work.Connect())
}

func (w *Work) Connect() error {
	url := im.GetUrl()
	w.User.Log.Println("IM 连接：" + url)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	w.Done = make(chan struct{})
	w.Conn = conn
	go func() {
		defer close(w.Done)
		for {
			typ, msg, err := w.Conn.ReadMessage()
			if w.User.Log.ErrPrint(err) {
				return
			}
			w.onMsg(typ, msg)
		}
	}()

	w.HeartTicker = time.NewTicker(HeartTimeout)
	defer w.HeartTicker.Stop()
	go func() {
		for {
			<-w.HeartTicker.C
			w.User.Log.Println("IM 30 秒内，未收到心跳消息")
		}
	}()

	for {
		select {
		case <-w.Done:
			user := w.User.Config.User
			if user != nil {
				user.Mutex.Lock()
				user.Running = false
				user.Mutex.Unlock()
			}
			return err
		default:
			user := w.User.Config.User
			if user != nil {
				user.Mutex.RLock()
				running := user.Running
				user.Mutex.RUnlock()
				if !running {
					err = conn.Close()
				}
			}
		}
	}
}

func (w *Work) Send(data []byte) error {
	w.User.Log.Println(fmt.Sprintf("IM 发送消息 %d：", len(data)) + string(data))
	return w.Conn.WriteMessage(websocket.TextMessage, data)
}

func (w *Work) onMsg(typ int, msg []byte) {
	startTime := time.Now().UnixMilli()
	length := len(msg)
	if typ == websocket.TextMessage && length == 1 && msg[0] == 'h' {
		w.HeartTicker.Reset(HeartTimeout)
	} else {
		w.User.Log.Println(fmt.Sprintf("IM 接收到消息 %d %d：", typ, len(msg)) + string(msg))
	}
	if length == 1 && msg[0] == 'o' {
		w.User.Log.Println("IM 登录")
		uid, token, err := w.User.Client.GetImToken()
		if w.User.Log.ErrPrint(err) {
			return
		}
		w.User.Log.ErrPrint(w.Send(im.BuildLoginMsg(uid, token)))
		return
	}
	if length == 0 || msg[0] != 'a' {
		return
	}
	msg = msg[1:]
	var messages []string
	err := json.Unmarshal(msg, &messages)
	if w.User.Log.ErrPrint(err) {
		return
	}
	for _, message := range messages {
		msg, err = base64.StdEncoding.DecodeString(message)
		if w.User.Log.ErrPrint(err) {
			continue
		}
		w.onMessage(msg, startTime)
	}
}

func (w *Work) onMessage(msg []byte, startTime int64) {
	length := len(msg)
	if length < 6 {
		return
	}

	header := msg[0:5]
	if reflect.DeepEqual(header, im.MsgHeaderCourse) {
		chatId := im.GetChatId(msg)
		if chatId == "" {
			w.User.Log.Println("IM 不是课程消息")
			return
		}
		w.User.Log.Println("IM 接收到课程消息，并请求获取活动信息：" + chatId)
		msg[3] = 0x00
		msg[6] = 0x1a
		msg = append(msg, 0x58, 0x00)
		w.User.Log.ErrPrint(w.Send(im.BuildMsg(msg)))
		return
	}
	if !reflect.DeepEqual(header, im.MsgHeaderActive) {
		return
	}
	w.User.Log.Println("IM 接收到活动信息")
	chatId := im.GetChatId(msg)
	if chatId == "" {
		w.User.Log.Println("IM 解析失败，无法获取 chatId")
		return
	}
	w.User.Log.Println("chatId:", chatId)

	sessionEnd := 11
	buf := im.NewBuf(msg)
	for {
		end := sessionEnd
		buf.Pos = sessionEnd
		w.onSession(buf, &sessionEnd, chatId, startTime)
		if sessionEnd == end {
			break
		}
	}
}

func (w *Work) onSession(buf *im.Buf, sessionEnd *int, chatId string, startTime int64) {
	b, err := buf.ReadE()
	if w.User.Log.ErrPrint(err) {
		return
	}
	if b != 0x22 {
		return
	}
	i, err := buf.ReadEnd2()
	if w.User.Log.ErrPrint(err) {
		return
	}
	*sessionEnd = i
	exit := false
	if i == 0 {
		exit = true
	} else {
		i, err := buf.ReadE()
		if w.User.Log.ErrPrint(err) {
			return
		}
		if i != 0x08 {
			exit = true
		}
	}
	if exit {
		w.User.Log.Println("IM 解析 Session 失败")
		return
	}
	w.User.Log.Println("IM 释放 Session")
	end := buf.Pos + 9
	w.User.Log.ErrPrint(w.Send(im.BuildReleaseSessionMsg(
		chatId,
		buf.Buf[buf.Pos:end],
	)))

	logN := w.User.Log.NewLogN(w.User.Config)
	defer w.User.Log.ErrClose(logN)
	logN.Println("chatId:", chatId)

	buf = im.NewBuf(buf.Buf[buf.Pos+1:])
	att, err := buf.ReadAttachment()
	if att == nil {
		i = im.IndexSlice(buf.Buf, cmd_course_chat_feedback.BytesCmd)
		if i == -1 {
			logN.Println("IM 解析失败，无法获取 attachment")
			logN.ErrPrint(err)
		}
		state, err := cmd_course_chat_feedback.GetState(buf)
		var s string
		if logN.ErrPrint(err) {
			s = "未知状态"
		} else if state {
			s = "开启"
		} else {
			s = "关闭"
		}
		courseConfig := w.User.Config.GetCourseConfig(chatId)
		courseName := config.GodCI(courseConfig, config.CourseName, "")
		logN.SetHeader(fmt.Sprintf("收到来自《%s》的群聊：%s", courseName, s))
		activeId, err := cmd_course_chat_feedback.GetActiveId(buf)
		if w.User.Log.ErrPrint(err) {
			return
		}
		logN.Println("activeId:", activeId)
		return
	}
	attType := GodJObjectI(att, "attachmentType", 0.)
	if attType == 1 {
		topic := GodJObjectI(att, "att_topic", map[string]any{})
		title, ok := GodJObject(topic, "content", "")
		if !ok {
			title = GodJObjectI(topic, "title", "")
		}
		courseName := GodJObjectI(GodJObjectI(topic, "att_group", map[string]any{}), "name", "")
		if courseName == "" {
			courseConfig := w.User.Config.GetCourseConfig(chatId)
			courseName = config.GodCI(courseConfig, config.CourseName, "")
		}
		logN.SetHeader(fmt.Sprintf("收到来自《%s》的主题讨论：%s", courseName, title))
		return
	}
	if attType != 15 {
		logN.Println("IM 解析失败，attachmentType != 15")
		logN.Printf("attr: %#v", att)
		return
	}

	attCourse, ok := GodJObject(att, "att_chat_course", map[string]any{})
	if !ok {
		logN.Println("IM 解析失败，无法获取 att_chat_course")
		logN.Printf("attr: %#v", att)
		return
	}

	activeId := strconv.FormatInt(int64(GodJObjectI(attCourse, "aid", 0.)), 10)
	if activeId == "0" {
		logN.Println("IM 解析失败，无法获取 aid")
		logN.Printf("attr: %#v", att)
		return
	}
	logN.Println("activeId:", activeId)
	if w.CacheActive.Add(activeId) {
		logN.Skip()
		logN.Println("该活动已处理")
		return
	}

	courseInfo, ok := GodJObject(attCourse, "courseInfo", map[string]any{})
	if !ok {
		logN.Println("IM 解析失败，无法获取 courseInfo")
		logN.Printf("attr: %#v", att)
		return
	}

	aType := GodJObjectI(attCourse, "atype", -1.)
	logN.Println("aType:", aType)
	courseName := GodJObjectI(courseInfo, "coursename", "")

	{
		var name string
		if aType == -1 && GodJObjectI(attCourse, "type", 0.) == 4 {
			name = "直播"
		} else {
			name = GodJObjectI(attCourse, "atypeName", "")
			if aType != 17 && aType != 35 {
				name += "活动"
			}
		}
		logN.SetHeader(fmt.Sprintf(
			"收到来自《%s》的%s：%s", courseName, name, GodJObjectI(attCourse, "title", ""),
		))
	}

	attActiveType := GodJObjectI(attCourse, "activeType", 0.)
	if (attActiveType != 0 && attActiveType != 2) || (aType != 0 && aType != 2) {
		/**
		aType:
		0: 签到
		2: 签到
		4: 抢答
		11: 选人
		14: 问卷
		17: 直播
		23: 评分
		35: 分组任务
		42: 随堂练习
		43: 投票
		49: 白板

		没有通知：计时器 47
		没有测试：腾讯会议

		type: 4: 直播
		*/
		logN.Println("接收到的不是签到活动")
		logN.Printf("attr: %#v", att)
		return
	}

	courseConfig := w.User.Config.GetCourseConfig(chatId)
	if courseConfig.New {
		logN.Println("该课程不在配置列表")
		courseConfig.Set(config.ChatId, chatId)
		courseConfig.Set(config.CourseName, courseName)
		courseConfig.Set(config.CourseId, GodJObjectI(courseInfo, "courseid", ""))
		courseConfig.Set(config.ClassId, GodJObjectI(courseInfo, "classid", ""))
		logN.ErrPrint(courseConfig.Save())
	}
	logN.Cfg = courseConfig

	work := NewWorkSign(courseConfig, logN.LogE)
	active, err := w.User.Client.GetActiveDetail(activeId)
	if logN.ErrPrint(err) {
		return
	}

	activeType := GodJObjectI(active, "activeType", -1.)
	if activeType != 2 {
		logN.Println("不是签到活动，activeType:", activeType)
		return
	}

	logN.State = NotifySign
	signType := GetSignType(int8(GodJObjectI(active, "otherId", -1.)))
	logN.Println(signType, GetSignTypeName(signType))
	if signType == SignTypeNormal {
		photo := GodJObjectI(active, "otherId", 0.)
		if photo != 0 {
			signType = SignTypePhoto
		}
	}

	if work.SetSignType(signType, active) || work.IsSkip() {
		return
	}

	taskTime := int64(GodJObjectI(active, "starttime", 0.))
	logN.Printf("任务时间戳：%d", taskTime)

	signOptions := work.Opts
	switch signType {
	case SignTypePhoto:
		imageId := work.GetImageId(time.UnixMilli(taskTime), w.User.Client)
		signOptions.ImageId = imageId
		logN.Println("预览：" + config.GetSignPhotoImageUrl(imageId, false))
		break
	case SignTypeLocation:
		if GodJObjectI(active, "ifopenAddress", 0.) != 0 {
			signOptions.Address = GodJObjectI(active, "locationText", "")
			signOptions.Longitude = GodJObjectI(active, "locationLongitude", "")
			signOptions.Latitude = GodJObjectI(active, "locationLatitude", "")
			logN.Printf(
				"教师指定签到地点：%s (%s, %s) ~%s 米",
				signOptions.Address,
				signOptions.Longitude,
				signOptions.Latitude,
				GodJObjectI(active, "locationRange", "0"),
			)
		}
	}

	logN.Println("准备签到中……")
	err = w.User.Client.PreSign(activeId)
	if err != nil {
		logN.Println(err)
	}

	logN.Printf("签到准备完毕，耗时：%dms", time.Now().UnixMilli()-startTime)
	takeTime := time.Now().UnixMilli() - taskTime
	logN.Printf("签到已发布：%dms", takeTime)
	delayTime := int64(config.GodRI(courseConfig, config.SignDelay, 0.))
	logN.Printf("用户配置延迟签到：%d", delayTime)
	if delayTime > 0 {
		delay := delayTime*1000 - takeTime
		if delay > 0 {
			logN.Printf("将等待：%dms", delay)
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}

	logN.Println("开始签到")
	content, err := w.User.Client.Sign(activeId, signOptions)
	if logN.ErrPrint(err) {
		return
	}
	switch content {
	case "success":
		content = "签到完成"
	case "您已签到过了":
	default:
		logN.Println("签到失败：" + content)
		return
	}
	logN.State = NotifySignOk
	logN.Println(content)
	return
}
