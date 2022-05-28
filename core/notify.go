package core

import (
	"bytes"
	"crypto/rand"
	"cx-im/config"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/moeshin/go-errs"
	mail "github.com/xhit/go-simple-mail/v2"
	"io"
	"io/ioutil"
	"net/url"
	"time"
)

const NotifyTitle = "cx-im 通知"

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
	header string
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

func (l *LogN) SetHeader(header string) {
	l.Println(header)
	l.header = header
}

func (l *LogN) getCfgString(name string, key string) (string, bool) {
	s := config.GodRI(l.Cfg, key, "")
	b := s == ""
	if b {
		l.Printf("由于 %s 为空，没有发送 %s 通知", key, name)
	}
	return s, b
}

func (l *LogN) getCfgStrings(name string, keys []string, vales []*string) bool {
	kl := len(keys)
	vl := len(vales)
	if kl != vl {
		l.Printf("数组长度不一致，%d != %d", kl, vl)
		return true
	}
	for i, key := range keys {
		v, b := l.getCfgString(name, key)
		if b {
			return true
		}
		*vales[i] = v
	}
	return false
}

func (l *LogN) Notifying(name string, fn func() error) {
	l.Printf("正在发送 %s 通知", name)
	err := fn()
	if err == nil {
		l.Printf("已发送 %s 通知", name)
	} else {
		l.Printf("发送 %s 通知失败！", name)
		l.ErrPrint(err)
	}
}

func NotifyEmail(title string, content string,
	email string, host string, port int, username string, password string, ssl bool) error {
	m := mail.NewMSG()
	m.SetFrom(fmt.Sprintf("%s <%s>", NotifyTitle, username)).
		AddTo(email).
		SetSubject(title).
		SetBody(mail.TextPlain, content)
	if m.Error != nil {
		return m.Error
	}
	server := mail.NewSMTPClient()
	server.Host = host
	server.Port = port
	server.Username = username
	server.Password = password
	if ssl {
		server.Encryption = mail.EncryptionSSLTLS
		//server.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client, err := server.Connect()
	if err != nil {
		return err
	}
	err = m.Send(client)
	return err
}

func (l *LogN) NotifyEmail(content string) {
	const name = "邮件"
	var email, host, username, password string
	if l.getCfgStrings(name, []string{
		config.NotifyEmail,
		config.SmtpHost,
		config.SmtpUsername,
		config.SmtpPassword,
	}, []*string{
		&email, &host, &username, &password,
	}) {
		return
	}
	port := int(config.GodRI(l.Cfg, config.SmtpPort, 465.))
	ssl := config.GodRI(l.Cfg, config.SmtpSSL, true)
	l.Notifying(name, func() error {
		return NotifyEmail(l.title, content, email, host, port, username, password, ssl)
	})
}

func NotifyPushPlus(title string, content string, token string) error {
	data, err := json.Marshal(map[string]string{
		"token":    token,
		"title":    title,
		"content":  content,
		"template": "txt",
	})
	if err != nil {
		return err
	}
	resp, err := HttpClient.Post("https://www.pushplus.plus/send", MimeJson, bytes.NewReader(data))
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
		msg := GodJObjectI(v, "msg", "")
		if msg == "" {
			msg = string(data)
		}
		return errors.New(msg)
	}
	return nil
}

func (l *LogN) NotifyPushPlus(content string) {
	const name = "PushPlus"
	token, b := l.getCfgString(name, config.NotifyPushPlusToken)
	if b {
		return
	}
	l.Notifying(name, func() error {
		return NotifyPushPlus(l.title, content, token)
	})
}

func NotifyTelegramBot(title string, content string, token string, chatId string) error {
	text := title + "\n" + content
	data, err := json.Marshal(map[string]string{
		"chat_id": chatId,
		"text":    text,
	})
	if err != nil {
		return err
	}
	resp, err := HttpClient.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token),
		MimeJson,
		bytes.NewReader(data),
	)
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
	if !GodJObjectI(v, "ok", false) {
		return errors.New(string(data))
	}
	return nil
}

func (l *LogN) NotifyTelegramBot(content string) {
	const name = "Telegram Bot"
	var token, chatId string
	if l.getCfgStrings(name, []string{
		config.NotifyTelegramBotToken,
		config.NotifyTelegramBotChatId,
	}, []*string{
		&token, &chatId,
	}) {
		return
	}
	l.Notifying(name, func() error {
		return NotifyTelegramBot(l.title, content, token, chatId)
	})
}

func NotifyBark(title string, content string, api string) error {
	resp, err := HttpClient.PostForm(api, url.Values{
		"title": []string{title},
		"body":  []string{content},
		"group": []string{NotifyTitle},
		"level": []string{"timeSensitive"},
	})
	if err != nil {
		return err
	}
	defer errs.CloseResponse(resp)
	err = testResponseStatus(resp)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var v JObject
	err = json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	if GodJObjectI(v, "code", 0.) != 200 {
		msg := GodJObjectI(v, "message", "")
		if msg == "" {
			msg = string(data)
		}
		return errors.New(msg)
	}
	return nil
}

func (l *LogN) NotifyBark(content string) {
	const name = "Bark"
	api, b := l.getCfgString(name, config.NotifyBarkApi)
	if b {
		return
	}
	l.Notifying(name, func() error {
		return NotifyBark(l.title, content, api)
	})
}

func (l *LogN) IsSkip(key string) bool {
	b := !config.GodRI(l.Cfg, key, true)
	if b {
		l.Printf("由于用户配置 %s 跳过通知", key)
	}
	return b
}

func (l *LogN) Notify() error {
	l.Writer.Skip = true
	var n string
	switch l.State {
	case NotifyActive:
		if l.IsSkip(config.NotifyActive) {
			return nil
		}
		n = "活动"
	case NotifySign:
		fallthrough
	case NotifySignOk:
		if l.IsSkip(config.NotifySign) {
			return nil
		}
		n = "签到"
		if l.State == NotifySignOk {
			n += "✔"
		} else {
			n += "✖"
		}
	}
	if n != "" {
		n = NotifyTitle + "：" + n
	}
	l.title = n
	data, err := ioutil.ReadAll(l.Writer.Buffer)
	if err != nil {
		return err
	}
	content := string(data)
	if l.header != "" {
		content = l.header + "\n" + content
	}

	l.NotifyPushPlus(content)
	l.NotifyEmail(content)
	l.NotifyTelegramBot(content)
	l.NotifyBark(content)
	return nil
}

func (l *LogN) Close() error {
	return l.Notify()
}
