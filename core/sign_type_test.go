package core

import "testing"

func TestGetSignTypeName(t *testing.T) {
	for i := int8(0); i < SignTypeLength; i++ {
		name := GetSignTypeName(i)
		t.Log(i, name)
		if name == "" {
			t.Fail()
		}
	}
}

func TestGetSignTypeKey(t *testing.T) {
	for i := int8(0); i < SignTypeLength; i++ {
		if i == SignQR {
			continue
		}
		key := GetSignTypeKey(i)
		t.Log(i, key)
		if key == "" {
			t.Fail()
		}
	}
}
