package core

import (
	"cx-im/src/config"
	"encoding/json"
	"fmt"
	"github.com/moeshin/go-errs"
	"io/ioutil"
	"log"
	"net/http"
)

type Api struct {
	Ok   bool   `json:"ok"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
	req  *http.Request
	code int
}

func NewApi(req *http.Request) *Api {
	return &Api{
		Ok:   false,
		Msg:  "",
		Data: nil,
		req:  req,
		code: http.StatusOK,
	}
}

func (a *Api) Response(w http.ResponseWriter) {
	w.Header().Set("Content-Type", MimeJson)
	if a.code != http.StatusOK {
		w.WriteHeader(a.code)
		return
	}
	data, err := JsonMarshal(a)
	if errs.Print(err) {
		return
	}
	_, err = w.Write(data)
	errs.Print(err)
}

func (a *Api) O(data any) {
	a.Data = data
	a.Ok = true
}

func (a *Api) ParseJson(v any) error {
	data, err := ioutil.ReadAll(a.req.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (a *Api) OE(msg string) {
	a.Ok = false
	a.Msg = msg
	log.Println(a.req.Method, a.req.URL.String(), msg)
}

func (a *Api) Err(err error) bool {
	b := err != nil
	if b {
		a.OE(err.Error())
		errs.Print(err)
	}
	return b
}

func (a *Api) Bad() {
	a.code = http.StatusBadRequest
}

func (a *Api) AddMsg(msg string) {
	if a.Msg != "" {
		a.Msg += "\n"
	}
	a.Msg += msg
}

func (a *Api) HandleConfig(username string) {
	isApp := username == ""
	var lv int
	var cfg *config.Config
	if isApp {
		lv = config.ValueLevelApp
		cfg = config.GetAppConfig()
	} else {
		lv = config.ValueLevelUser
		user, err := GetUser(username)
		if a.Err(err) {
			return
		}
		cfg = user.Config
	}
	switch a.req.Method {
	case http.MethodGet:
		data := map[string]any{}
		a.O(data)
		var fun func(*config.Config, map[string]any, int)
		fun = func(cfg *config.Config, data map[string]any, lv int) {
			for _, k := range cfg.Keys() {
				if !isApp && lv != config.ValueLevelCourse && k == config.Courses {
					courses := cfg.GetCourses()
					m := map[string]map[string]any{}
					data[k] = m
					cfg.Mutex.RLock()
					for _, chatId := range courses.Keys() {
						cfg := cfg.GetCourseConfig(chatId)
						data := map[string]any{}
						m[chatId] = data
						fun(cfg, data, config.ValueLevelCourse)
					}
					cfg.Mutex.RUnlock()
					continue
				}
				typ, ok := config.KeyValues[k]
				if !ok || typ != typ|lv || typ == typ|config.ValueHide {
					continue
				}
				v, ok := cfg.GetC(k)
				if !ok {
					continue
				}
				if typ == typ|config.ValuePassword {
					v = "*"
				}
				data[k] = v
			}
		}
		fun(cfg, data, lv)
	case http.MethodPost:
		a.SetConfigValues(cfg)
	default:
		a.Bad()
		return
	}
}

func (a *Api) SetConfigValues(cfg *config.Config) {
	var data JObject
	err := a.ParseJson(&data)
	if a.Err(err) {
		return
	}
	save := false
	for k, v := range data {
		if !config.ValidKeyValue(config.ValueLevelCourse, k, v) {
			a.AddMsg(fmt.Sprintf("无效键值：%s", k))
			continue
		}
		save = true
		cfg.Set(k, v)
	}
	if save && a.Err(cfg.Save()) {
		return
	}
	a.Ok = a.Msg == ""
}
