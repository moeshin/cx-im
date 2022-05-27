package core

import (
	"bytes"
	"crypto/rand"
	"cx-im/config"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/moeshin/go-errs"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const NotifyTitle = "cx-im 通知："

type NotifyState = int

const (
	NotifyActive NotifyState = iota
	NotifySign
	NotifySignOk
)

type LogN struct {
	*LogE
	Writer *logWriter
	Cfg    *config.Config
	State  NotifyState
	title  string
}

func (l *LogE) NewLogN(cfg *config.Config) *LogN {
	var tag string
	var buf [4]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if l.ErrPrint(err) {
		tag = fmt.Sprintf("[%d] ", time.Now().UnixMilli())
	} else {
		tag = fmt.Sprintf("[%X] ", buf[:])
	}
	writer := &logWriter{
		Buffer: &bytes.Buffer{},
		Log:    l,
	}
	return &LogN{
		LogE: &LogE{
			Logger: NewLogger(writer, tag),
		},
		Writer: writer,
		Cfg:    cfg,
	}
}

func NotifyPushPlus(token string, title string, content string) error {
	data, err := json.Marshal(map[string]string{
		"token":    token,
		"title":    title,
		"content":  content,
		"template": "txt",
	})
	if err != nil {
		return err
	}
	resp, err := http.Post("https://www.pushplus.plus/send", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer errs.CloseResponse(resp)
	err = testResponseStatus(resp)
	if err != nil {
		return err
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var v JObject
	err = json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	if GodJObjectI(v, "code", 0.) != 200 {
		return errors.New(GodJObjectI(v, "msg", ""))
	}
	return nil
}

func (l *LogN) NotifyPushPlus(content string) {
	token := config.GodRI(l.Cfg, config.PushPlusToken, "")
	if token == "" {
		l.Printf("由于 %s 为空，没有发送 PushPlus 通知\n", config.PushPlusToken)
		return
	}
	l.Println("正在发送 PushPlus 通知")
	err := NotifyPushPlus(token, l.title, content)
	if err == nil {
		l.Println("已发送 PushPlus 通知")
	} else {
		l.Println("发送 PushPlus 通知失败！", err)
	}
}

func (l *LogN) Notify() error {
	l.Writer.Skip = true
	{
		var s string
		if l.State == NotifyActive {
			s = "活动"
		} else {
			s = "签到"
			switch l.State {
			case NotifySignOk:
				s += "✔"
			case NotifySign:
				s += "✖"
			}
		}
		l.title = NotifyTitle + s
	}
	data, err := ioutil.ReadAll(l.Writer.Buffer)
	if err != nil {
		return err
	}
	content := string(data)

	l.NotifyPushPlus(content)
	return nil
}

func (l *LogN) Close() error {
	return l.Notify()
}
