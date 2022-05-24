package core

import (
	"fmt"
)

type JObject map[string]any
type JArray []any

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
