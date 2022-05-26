package core

import (
	"cx-im/config"
	"cx-im/im"
	"cx-im/im/cmd_course_chat_feedback"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/moeshin/go-errs"
	"log"
	"reflect"
	"strconv"
)

type Work struct {
	Config *config.Config
	Client *CxClient
	Conn   *websocket.Conn
	Done   chan struct{}
}

func (w *Work) Connect() error {
	client, err := NewClientFromConfig(w.Config)
	if err != nil {
		return err
	}
	w.Client = client
	err = client.Login()
	if err != nil {
		return err
	}
	url := im.GetUrl()
	log.Println("IM 连接：" + url)
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
			if errs.Print(err) {
				return
			}
			w.OnMsg(typ, msg)
		}
	}()

	for {
		select {
		case <-w.Done:
			return err
		default:
			user := w.Config.User
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
	log.Println("IM 发送消息", len(data), ":", string(data))
	return w.Conn.WriteMessage(websocket.TextMessage, data)
}

func (w *Work) OnMsg(typ int, msg []byte) {
	length := len(msg)
	if typ == websocket.TextMessage && length == 1 && msg[0] == 'h' {
		// TODO 心跳包
	} else {
		log.Println("IM 接收到消息", typ, length, ":", string(msg))
	}
	if length == 1 && msg[0] == 'o' {
		log.Println("IM 登录")
		uid, token, err := w.Client.GetImToken()
		if errs.Print(err) {
			return
		}
		errs.Print(w.Send(im.BuildLoginMsg(uid, token)))
		return
	}
	if length == 0 || msg[0] != 'a' {
		return
	}
	msg = msg[1:]
	var messages []string
	err := json.Unmarshal(msg, &messages)
	if errs.Print(err) {
		return
	}
	for _, message := range messages {
		msg, err = base64.StdEncoding.DecodeString(message)
		if errs.Print(err) {
			continue
		}
		w.OnMessage(msg)
	}
}

func (w *Work) OnMessage(msg []byte) {
	length := len(msg)
	if length < 6 {
		return
	}

	header := msg[0:5]
	if reflect.DeepEqual(header, im.MsgHeaderCourse) {
		chatId := im.GetChatId(msg)
		if chatId == "" {
			log.Println("IM 不是课程消息")
			return
		}
		log.Println("IM 接收到课程消息，并请求获取活动信息：" + chatId)
		msg[3] = 0x00
		msg[6] = 0x1a
		msg = append(msg, 0x58, 0x00)
		errs.Print(w.Send(im.BuildMsg(msg)))
		return
	}
	if !reflect.DeepEqual(header, im.MsgHeaderActive) {
		return
	}
	log.Println("IM 接收到活动信息")
	chatId := im.GetChatId(msg)
	if chatId == "" {
		log.Println("IM 解析失败，无法获取 chatId")
		return
	}
	log.Println("IM chatId:", chatId)

	sessionEnd := 11
	buf := im.NewBuf(msg)
	for {
		end := sessionEnd
		buf.Pos = sessionEnd
		errs.Print(w.onSession(buf, &sessionEnd, chatId))
		if sessionEnd == end {
			break
		}
	}
}

func (w *Work) onSession(buf *im.Buf, sessionEnd *int, chatId string) error {
	b, err := buf.ReadE()
	if err != nil {
		return err
	}
	if b != 0x22 {
		return nil
	}
	i, err := buf.ReadEnd2()
	if err != nil {
		return err
	}
	*sessionEnd = i
	exit := false
	if i == 0 {
		exit = true
	} else {
		i, err := buf.ReadE()
		if err != nil {
			return err
		}
		if i != 0x08 {
			exit = true
		}
	}
	if exit {
		log.Println("IM 解析 Session 失败")
		return nil
	}
	log.Println("IM 释放 Session")
	end := buf.Pos + 9
	errs.Print(w.Send(im.BuildReleaseSessionMsg(
		chatId,
		buf.Buf[buf.Pos:end],
	)))

	buf = im.NewBuf(buf.Buf[buf.Pos+1:])
	att, err := buf.ReadAttachment()
	if att == nil {
		i = im.IndexSlice(buf.Buf, cmd_course_chat_feedback.BytesCmd)
		if i == -1 {
			log.Println("IM 解析失败，无法获取 attachment")
			return err
		}
		state, err := cmd_course_chat_feedback.GetState(buf)
		var s string
		if errs.Print(err) {
			s = "未知状态"
		} else if state {
			s = "开启"
		} else {
			s = "关闭"
		}
		courseConfig := w.Config.GetCourseConfig(chatId)
		courseName := config.GodCI(courseConfig, config.CourseName, "")
		log.Printf("IM 收到来自《%s》的群聊：%s\n", courseName, s)
		activeId, err := cmd_course_chat_feedback.GetActiveId(buf)
		if err != nil {
			return err
		}
		log.Println("IM activeId:", activeId)
		return nil
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
			courseConfig := w.Config.GetCourseConfig(chatId)
			courseName = config.GodCI(courseConfig, config.CourseName, "")
		}
		log.Printf("IM 收到来自《%s》的主题讨论：%s\n", courseName, title)
		return nil
	}
	if attType != 15 {
		log.Println("IM 解析失败，attachmentType != 15")
		log.Printf("attr: %#v\n", att)
		return nil
	}

	attCourse, ok := GodJObject(att, "att_chat_course", map[string]any{})
	if !ok {
		log.Println("IM 解析失败，无法获取 att_chat_course")
		log.Printf("attr: %#v\n", att)
		return nil
	}

	activeId := strconv.FormatInt(int64(GodJObjectI(attCourse, "aid", 0.)), 10)
	if activeId == "0" {
		log.Println("IM 解析失败，无法获取 aid")
		log.Printf("attr: %#v\n", att)
		return nil
	}
	log.Println("IM activeId:", activeId)

	courseInfo, ok := GodJObject(attCourse, "courseInfo", map[string]any{})
	if !ok {
		log.Println("IM 解析失败，无法获取 courseInfo")
		log.Printf("attr: %#v\n", att)
		return nil
	}

	aType := GodJObjectI(attCourse, "atype", -1.)
	log.Println("IM aType:", aType)
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
		log.Printf("IM 收到来自《%s》的%s：%s\n", courseName, name, GodJObjectI(attCourse, "title", ""))
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
		log.Println("IM 接收到的不是签到活动")
		log.Printf("attr: %#v\n", att)
		return nil
	}

	courseConfig := w.Config.GetCourseConfig(chatId)
	if courseConfig.New {
		log.Println("该课程不在配置列表")
		courseConfig.Data.Set(config.ChatId, chatId)
		courseConfig.Data.Set(config.CourseName, courseName)
		courseConfig.Data.Set(config.CourseId, GodJObjectI(courseInfo, "courseid", ""))
		courseConfig.Data.Set(config.ClassId, GodJObjectI(courseInfo, "classid", ""))
		errs.Print(courseConfig.Save())
	}

	work := NewWorkSign(courseConfig)
	active, err := w.Client.GetActiveDetail(activeId)
	if err != nil {
		return err
	}

	activeType := GodJObjectI(active, "activeType", -1.)
	if activeType != 2 {
		log.Println("不是签到活动，activeType:", activeType)
		return nil
	}

	signType := GetSignType(int8(GodJObjectI(active, "otherId", -1.)))
	log.Println(signType, GetSignTypeName(signType))
	if signType == SignNormal {
		photo := GodJObjectI(active, "otherId", 0.)
		if photo != 0 {
			signType = SignPhoto
		}
	}

	if work.SetSignType(signType, active) || work.IsSkip() {
		return nil
	}

	signOptions := work.Opts
	switch signType {
	case SignPhoto:
		// TODO Upload image
		break
	case SignLocation:
		if GodJObjectI(active, "ifopenAddress", 0.) != 0 {
			signOptions.Address = GodJObjectI(active, "locationText", config.DefaultSignAddress)
			signOptions.Longitude = GodJObjectI(active, "locationLongitude", config.DefaultSignLongitude)
			signOptions.Latitude = GodJObjectI(active, "locationLatitude", config.DefaultSignLatitude)
			log.Printf(
				"教师指定签到地点：%s (%s, %s) ~%s 米\n",
				signOptions.Address,
				signOptions.Longitude,
				signOptions.Latitude,
				GodJObjectI(active, "locationRange", "0"),
			)
		}
	}
	// TODO
	return nil
}
