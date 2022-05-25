package im

import (
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

func (b *Buf) Read() (byte, error) {
	err := b.testRange(1)
	if err != nil {
		return 0, err
	}
	v := b.Buf[b.Pos]
	b.Pos++
	return v, nil
}

func (b *Buf) ReadLength2() (int, error) {
	err := b.testRange(2)
	if err != nil {
		return -1, err
	}
	h := int(b.Buf[b.Pos])
	b.Pos++
	l := int(b.Buf[b.Pos])
	b.Pos++
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
