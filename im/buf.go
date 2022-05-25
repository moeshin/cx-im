package im

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Buf struct {
	Buf []byte
	Pos int
	Len int
}

func NewBuf(data []byte) *Buf {
	return &Buf{
		data,
		0,
		len(data),
	}
}

func (b *Buf) testRange(length int) error {
	if length > b.Len {
		return errors.New(fmt.Sprintf("数组越界，%d > %d", length, b.Len))
	}
	return nil
}

func (b *Buf) Read() byte {
	v := b.Buf[b.Pos]
	b.Pos++
	return v
}

func (b *Buf) ReadE() (byte, error) {
	err := b.testRange(1)
	if err != nil {
		return 0, err
	}
	return b.Read(), nil
}

func (b *Buf) ReadLength2() (int, error) {
	err := b.testRange(2)
	if err != nil {
		return -1, err
	}
	h := int(b.Read())
	l := int(b.Read())
	return h + (l-1)*0x80, nil
}

func (b *Buf) ReadEnd2() (int, error) {
	i, err := b.ReadLength2()
	if err != nil {
		return -1, err
	}
	return i + b.Pos, nil
}

func (b *Buf) ReadString2() (string, error) {
	l, err := b.ReadLength2()
	if err != nil {
		return "", err
	}
	s := b.Pos
	e := s + l
	err = b.testRange(e)
	if err != nil {
		return "", err
	}
	b.Pos = e
	return string(b.Buf[s:e]), nil
}

func (b *Buf) ReadAttachment() (map[string]any, error) {
	i := indexSlice(b.Buf, MsgPartAttachment)
	if i == -1 {
		return nil, errors.New("未找到 MsgPartAttachment")
	}
	b.Pos = i + len(MsgPartAttachment)
	data, err := b.ReadString2()
	if err != nil {
		return nil, err
	}
	var v map[string]any
	err = json.Unmarshal([]byte(data), &v)
	return v, err
}

func (b *Buf) ReadLongBits() (*LongBits, error) {
	n := NewLongBits(0, 0)
	isLow := b.Len-b.Pos <= 4
	for i := 0; i < 4; i++ {
		r, err := b.ReadE()
		if err != nil {
			return nil, err
		}
		n.L = stu(n.L | (int64(127&r) << (7 * i)))
		if r < 128 || (isLow && i == 3) {
			return n, nil
		}
	}
	r, err := b.ReadE()
	if err != nil {
		return nil, err
	}
	d := int64(127 & r)
	n.L = stu(n.L | d<<28)
	n.H = stu(n.H | d>>4)
	if r < 128 {
		return n, nil
	}
	for i := 0; i < 5; i++ {
		r, err = b.ReadE()
		if err != nil {
			return nil, err
		}
		n.H = stu(n.H | (int64(127&r) << (7*i + 3)))
		if r < 128 {
			return n, nil
		}
	}
	return nil, errors.New("解析失败：ReadLongBits")
}
