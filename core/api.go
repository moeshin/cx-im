package core

import (
	"encoding/json"
	"github.com/moeshin/go-errs"
	"net/http"
)

type Api struct {
	Ok   bool   `json:"ok"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func NewApi() *Api {
	return &Api{
		Ok:   false,
		Msg:  "",
		Data: nil,
	}
}

func (a *Api) Response(w http.ResponseWriter) {
	w.Header().Set("Content-Type", MimeJson)
	data, err := json.MarshalIndent(a, "", "  ")
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
