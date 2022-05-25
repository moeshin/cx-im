package im

const (
	L1E8H int64 = 0x100000000
)

func stu[T byte | int | int8 | int32 | int64](n T) int64 {
	return int64(uint32(n))
}

type Long struct {
	H        int64
	L        int64
	Unsigned bool
}

func NewLong(h int64, l int64, unsigned bool) *Long {
	return &Long{h, l, unsigned}
}

func (l *Long) ToNumber() int64 {
	h := l.H
	if l.Unsigned {
		h = stu(h)
	}
	return L1E8H*h + stu(l.L)
}

func (l *Long) IsZero() bool {
	return l.H == 0 && l.L == 0
}

type LongBits struct {
	H int64
	L int64
}

func NewLongBits(h int64, l int64) *LongBits {
	return &LongBits{h, l}
}

func (l *LongBits) ToLong(unsigned bool) *Long {
	return NewLong(l.H, l.L, unsigned)
}

//func (l *LongBits) ToNumber(e bool) int64 {
//	if !e && stu(l.H)>>31 != 0 {
//		H := stu(^l.H)
//		L := 1 + stu(^l.L)
//		if L == 0 {
//			H = stu(L + 1)
//		}
//		return -(L1E8H*H + L)
//	}
//	return L1E8H*l.H + l.L
//}
