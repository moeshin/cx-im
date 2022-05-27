package core

import (
	"cx-im/config"
	"cx-im/model"
	"log"
)

type WorkSign struct {
	Type SignType
	Cfg  *config.Config
	Opts *model.SignOptions
}

func NewWorkSign(cfg *config.Config) *WorkSign {
	return &WorkSign{
		SignUnknown,
		cfg,
		nil,
	}
}

func (w *WorkSign) SetSignType(signType SignType, active JObject) bool {
	w.Type = signType
	getSignCode := func() string {
		return GodJObjectI(active, "signCode", "")
	}
	switch signType {
	case SignGesture:
		log.Println("手势：", getSignCode())
	case SignCode:
		log.Println("签到码：", getSignCode())
	case SignQR:
		log.Println("目前无法二维码签到")
		return true
	}
	return false
}

func (w *WorkSign) IsSkip() bool {
	if config.GodRI(w.Cfg, config.SignEnable, false) {
		log.Println("因用户配置", config.SignEnable, "跳过签到")
		return true
	}
	w.Opts = w.Cfg.GetSignOptions(GetSignTypeKey(w.Type))
	if w.Opts == nil {
		log.Println("因用户配置，跳过签到")
		return true
	}
	return false
}
