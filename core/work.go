package core

import (
	"cx-im/config"
	"cx-im/im"
	"github.com/gorilla/websocket"
	"github.com/moeshin/go-errs"
	"log"
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
	conn, _, err := websocket.DefaultDialer.Dial(im.GetUrl(), nil)
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
					continue
				}
				errs.Print(w.Send(im.BuildLoginMsg(uid, token)))
				continue
			}
			// TODO
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
