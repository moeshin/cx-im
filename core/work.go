package core

import (
	"cx-im/config"
	"cx-im/im"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/moeshin/go-errs"
	"log"
	"reflect"
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
	log.Println("chatId:", chatId)

	// TODO
}
