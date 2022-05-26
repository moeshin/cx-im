package core

import "cx-im/config"

type SignType = int8

const (
	SignUnknown SignType = iota - 1
	SignNormal
	SignPhoto
	SignQR
	SignGesture
	SignLocation
	SignCode
	SignTypeLength
)

func GetSignTypeName(signType SignType) string {
	var n string
	switch signType {
	case SignUnknown:
		n = "未知签到类型"
	case SignNormal:
		n = "普通签到"
	case SignPhoto:
		n = "图片签到"
	case SignQR:
		n = "二维码签到"
	case SignGesture:
		n = "手势签到"
	case SignLocation:
		n = "位置签到"
	case SignCode:
		n = "签到码签到"
	}
	return n
}

func GetSignTypeKey(signType SignType) string {
	var k string
	switch signType {
	case SignNormal:
		k = config.SignNormal
	case SignPhoto:
		k = config.SignPhoto
	case SignGesture:
		k = config.SignGesture
	case SignLocation:
		k = config.SignLocation
	case SignCode:
		k = config.SignCode
	}
	return k
}

func GetSignType(typ int8) SignType {
	if typ >= 0 && typ < SignTypeLength {
		return typ
	}
	return SignUnknown
}
