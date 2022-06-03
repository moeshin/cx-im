package im

import "testing"

func TestLong_ToNumber(t *testing.T) {
	test := func(expected int64, h int64, l int64) {
		t.Log(expected, h, l)
		long := NewLong(h, l, false)
		actual := long.ToNumber()
		eq := expected == actual
		t.Log(eq, actual)
		if !eq {
			t.Fail()
		}
	}

	test(2000020909381, 465, 2861116741)
	test(2000020832716, 465, 2861040076)
	test(2000020829804, 465, 2861037164)
	test(2000020909381, 465, -1433850555)
	test(2000020832716, 465, -1433927220)
	test(2000020829804, 465, -1433930132)
}

func TestBuf_ReadLongBits(t *testing.T) {
	test := func(expected int64, data []byte) {
		t.Logf("%d, % X", expected, data)
		buf := NewBuf(data)
		longBits, err := buf.ReadLongBits()
		if err != nil {
			t.Error(err)
			return
		}
		long := longBits.ToLong(false)
		actual := long.ToNumber()
		eq := expected == actual
		t.Log(eq, actual)
		if !eq {
			t.Log(long)
			t.Fail()
		}
	}
	test(2000020909381, []byte{0xC5, 0xDA, 0xA4, 0xD4, 0x9A, 0x3A})
	test(2000020832716, []byte{0xCC, 0x83, 0xA0, 0xD4, 0x9A, 0x3A})
	test(2000020829804, []byte{0xEC, 0xEC, 0x9F, 0xD4, 0x9A, 0x3A})
}
