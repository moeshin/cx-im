package core

import (
	"fmt"
)

type JObject map[string]any
type JArray []any
type JValue interface {
	string | float64 | bool | []any | map[string]any
}

func AnyToString(data any) string {
	switch v := data.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func AnyToJObject(data any) JObject {
	v, _ := data.(map[string]any)
	return v
}

func AnyToJArray(data any) JArray {
	v, _ := data.([]any)
	return v
}

func (j JArray) Get(i int) any {
	if i < len(j) {
		return j[i]
	}
	return nil
}

func GodJObject[T JValue](obj JObject, key string, def T) (T, bool) {
	v, ok := obj[key]
	if ok && v != nil {
		v, ok := v.(T)
		if ok {
			return v, true
		}
	}
	return def, false
}

func GodJObjectI[T JValue](obj JObject, key string, def T) T {
	v, _ := GodJObject(obj, key, def)
	return v
}

func SliceRemove[T any](slice []T, i int) []T {
	return append(slice[:i], slice[i+1:]...)
}
