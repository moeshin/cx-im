package config

import (
	"testing"
	"time"
)

func TestMatchSignPhotoKey(t *testing.T) {
	var tm time.Time
	test := func(expected bool, key string) {
		t.Log(expected, tm.Format(TimeLayout), "|", key)
		actual := MatchSignPhotoKey(tm, key)
		eq := expected == actual
		t.Log(eq, actual)
		if !eq {
			t.Fail()
		}
	}
	tm = time.Date(2022, 4, 7, 8, 30, 0, 0, time.UTC)
	test(true, "1-4|am")
	test(true, "1-4|am")
	test(true, "4|08:00-11:40")
	test(true, "4|pm,08:00-11:40")
	test(true, "4|")
	test(true, "|am")
	test(false, "|pm")
	test(false, "1-3|")
	tm = time.Date(2022, 4, 7, 22, 33, 0, 0, time.UTC)
	test(true, "4|pm")
}
