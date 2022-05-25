package im

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

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
