package core

import "fmt"

type JObject map[string]interface{}

func AnyToString(data any) string {
	switch v := data.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
