package core

import (
	"crypto/rand"
	"cx-im/config"
	"cx-im/model"
	"log"
	"math/big"
	"time"
)

const ImageIdNone = "041ed4756ca9fdf1f9b6dde7a83f8794"

type WorkSign struct {
	Type SignType
	Cfg  *config.Config
	Opts *model.SignOptions
}

func NewWorkSign(cfg *config.Config) *WorkSign {
	return &WorkSign{
		SignTypeUnknown,
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
	case SignTypeGesture:
		log.Println("手势：", getSignCode())
	case SignTypeCode:
		log.Println("签到码：", getSignCode())
	case SignTypeQR:
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

func (w *WorkSign) GetImagePath(tm time.Time) string {
	images := w.Cfg.GetImages(tm)
	l := len(images)
	if l == 0 {
		return ""
	}
	var path string
	if l == 1 {
		path = images[0]
	} else {
		log.Println("将从这些图片中随机选择一张进行图片签到：", images)
		i := 0
		b, err := rand.Int(rand.Reader, big.NewInt(int64(l)))
		if err != nil {
			log.Println("随机失败", err)
		} else {
			i = int(b.Int64())
		}
		path = images[i]
	}
	if path != "" {
		log.Println("将使用这张照片进行图片签到：", path)
	}
	return path
}

func (w *WorkSign) GetImageId(tm time.Time, client *CxClient) string {
	path := w.GetImagePath(tm)
	var err error
	if client == nil {
		client, err = NewClientFromConfig(w.Cfg.Parent)
	}
	if err == nil {
		id, err := client.UploadImage(path)
		if err == nil {
			return id
		}
	}
	if err != nil {
		log.Println("上传图片失败", err)
	}
	log.Println("将使用一张黑图进行图片签到")
	return ImageIdNone
}
