package im

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"time"
)

var (
	MsgHeaderCourse = []byte{0x08, 0x00, 0x40, 0x02, 0x4a}
	MsgHeaderActive = []byte{0x08, 0x00, 0x40, 0x00, 0x4a}
	MsgFooter       = []byte{
		0x1A, 0x16, 0x63, 0x6F, 0x6E, 0x66, 0x65, 0x72, 0x65, 0x6E, 0x63, 0x65, 0x2E, 0x65, 0x61, 0x73, 0x65, 0x6D,
		0x6F, 0x62, 0x2E, 0x63, 0x6F, 0x6D,
	}
	MsgPartAttachment = []byte{
		0x0a, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6D, 0x65, 0x6E, 0x74, 0x10, 0x08, 0x32,
	}
)

func indexSlice[T any](data []T, slice []T) int {
	dataLen := len(data)
	matchLen := len(slice)
	if dataLen >= matchLen {
		length := dataLen - matchLen
		for i := 0; i <= length; i++ {
			if reflect.DeepEqual(slice, data[i:i+matchLen]) {
				return i
			}
		}
	}
	return -1
}

func lastIndexSlice[T any](data []T, slice []T) int {
	dataLen := len(data)
	matchLen := len(slice)
	if dataLen >= matchLen {
		for i := dataLen - matchLen; i >= 0; i-- {
			if reflect.DeepEqual(slice, data[i:i+matchLen]) {
				return i
			}
		}
	}
	return -1
}

func GetUrl() string {
	const (
		Text = "abcdefghijklmnopqrstuvwxyz012345"
		Len  = int64(len(Text))
	)
	s := ""
	for i := 0; i < 8; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(Len))
		s += string(Text[n.Int64()])
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(1000))
	return fmt.Sprintf(
		"wss://im-api-vip6-v2.easemob.com/ws/%03d/%s/websocket",
		n,
		s,
	)
}

func BuildMsg(data []byte) []byte {
	encoding := base64.StdEncoding
	size := encoding.EncodedLen(len(data))
	r := make([]byte, size+4)
	r[0] = '['
	r[1] = '"'
	r[size+2] = '"'
	r[size+3] = ']'
	d := r[2 : size+2]
	encoding.Encode(d, data)
	return r
}

func BuildLoginMsg(uid string, token string) []byte {
	bTime := []byte(strconv.FormatInt(time.Now().UnixMilli(), 10))
	bToken := []byte(token)
	size := byte(len(uid))
	var d []byte
	d = append(d, 0x08, 0x00, 0x12, 0x34+size, 0x0a, 0x0e)
	d = append(d, []byte("cx-dev#cxstudy")...)
	d = append(d, 0x12, size)
	d = append(d, []byte(uid)...)
	d = append(d, 0x1a, 0x0b)
	d = append(d, []byte("easemob.com")...)
	d = append(d, 0x22, 0x13)
	d = append(d, []byte("webim_")...)
	d = append(d, bTime...)
	d = append(d, 0x1a, 0x85, 0x01, 0x24, 0x74, 0x24)
	d = append(d, bToken...)
	d = append(d, 0x40, 0x03, 0x4a, 0xc0, 0x01, 0x08, 0x10, 0x12, 0x05, 0x33, 0x2e, 0x30, 0x2e, 0x30, 0x28,
		0x00, 0x30, 0x00, 0x4a, 0x0d)
	d = append(d, bTime...)
	d = append(d, 0x62, 0x05, 0x77, 0x65, 0x62, 0x69, 0x6d, 0x6a, 0x13, 0x77, 0x65, 0x62, 0x69, 0x6d, 0x5f)
	d = append(d, bTime...)
	d = append(d, 0x72, 0x85, 0x01, 0x24, 0x74, 0x24)
	d = append(d, bToken...)
	d = append(d, 0x50, 0x00, 0x58, 0x00)
	return BuildMsg(d)
}

func GetChatId(data []byte) string {
	index := lastIndexSlice(data, MsgFooter)
	if index != -1 {
		data = data[:index]
		i := bytes.LastIndexByte(data, 0x12)
		if i != -1 {
			i++
			size := int(data[i])
			i++
			if i+size == index {
				return string(data[i:])
			}
		}
	}
	return ""
}
