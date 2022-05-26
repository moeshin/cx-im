package cmd_course_chat_feedback

import (
	"cx-im/im"
	"errors"
	"strconv"
)

var (
	BytesCmd = (func() []byte {
		d := []byte{0x52, 0x18}
		d = append(d, []byte("CMD_COURSE_CHAT_FEEDBACK")...)
		return d
	})()
	BytesAid = (func() []byte {
		d := []byte{0x0A, 0x03}
		d = append(d, []byte("aid")...)
		return d
	})()
	BytesState = (func() []byte {
		d := []byte{0x0A, 0x0B}
		d = append(d, []byte("stuFeedback")...)
		return d
	})()
)

func GetState(buf *im.Buf) (bool, error) {
	i := im.IndexSlice(buf.Buf, BytesState)
	if i == -1 {
		return false, errors.New("未找到 BytesState")
	}
	buf.Pos = i + len(BytesState) + 3
	longBits, err := buf.ReadLongBits()
	if err != nil {
		return false, err
	}
	return !longBits.ToLong(false).IsZero(), nil
}

func GetActiveId(buf *im.Buf) (string, error) {
	i := im.IndexSlice(buf.Buf, BytesAid)
	if i == -1 {
		return "", errors.New("未找到 BYTES_AID")
	}
	buf.Pos = i + len(BytesAid) + 3
	longBits, err := buf.ReadLongBits()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(longBits.ToLong(false).ToNumber(), 10), nil
}
