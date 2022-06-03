package core

import (
	"cx-im/src/config"
)

type SignType = int8

const (
	SignTypeUnknown SignType = iota - 1
	SignTypeNormal
	SignTypePhoto
	SignTypeQR
	SignTypeGesture
	SignTypeLocation
	SignTypeCode
	SignTypeLength
)

func GetSignTypeName(signType SignType) string {
	var n string
	switch signType {
	case SignTypeUnknown:
		n = "未知签到类型"
	case SignTypeNormal:
		n = "普通签到"
	case SignTypePhoto:
		n = "图片签到"
	case SignTypeQR:
		n = "二维码签到"
	case SignTypeGesture:
		n = "手势签到"
	case SignTypeLocation:
		n = "位置签到"
	case SignTypeCode:
		n = "签到码签到"
	}
	return n
}

func GetSignTypeKey(signType SignType) string {
	var k string
	switch signType {
	case SignTypeNormal:
		k = config.SignNormal
	case SignTypePhoto:
		k = config.SignPhoto
	case SignTypeGesture:
		k = config.SignGesture
	case SignTypeLocation:
		k = config.SignLocation
	case SignTypeCode:
		k = config.SignCode
	}
	return k
}

func GetSignType(typ int8) SignType {
	if typ >= 0 && typ < SignTypeLength {
		return typ
	}
	return SignTypeUnknown
}
