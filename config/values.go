package config

type ValueType = int

const (
	ValueHide ValueType = 1 << iota
	ValueBool
	ValueNumber
	ValueString
	ValuePassword
	ValueJson
	ValueArray
	//ValueAny
)

var KeyValues = map[string]int{
	"DefaultUsername": ValueString,

	"WebHost": ValueString,
	"WebPort": ValueNumber,

	"NotifyEmail":  ValueString,
	"SmtpHost":     ValueString,
	"SmtpPort":     ValueNumber,
	"SmtpUsername": ValueString,
	"SmtpPassword": ValueString | ValuePassword,
	"SmtpSSL":      ValueBool,

	"NotifyPushPlusToken":     ValueString | ValuePassword,
	"NotifyBarkApi":           ValueString | ValuePassword,
	"NotifyTelegramBotToken":  ValueString | ValuePassword,
	"NotifyTelegramBotChatId": ValueString,

	"NotifySign":   ValueBool,
	"NotifyActive": ValueBool,

	"SignAddress":   ValueString,
	"SignLongitude": ValueNumber,
	"SignLatitude":  ValueNumber,
	"SignIp":        ValueString,

	"SignDelay":    ValueNumber,
	"SignEnable":   ValueBool,
	"SignNormal":   ValueBool,
	"SignPhoto":    ValueBool | ValueString | ValueArray | ValueJson,
	"SignGesture":  ValueBool,
	"SignLocation": ValueBool,
	"SignCode":     ValueBool,

	"Username": ValueString,
	"Password": ValueString | ValuePassword,
	"Fid":      ValueString,

	"Courses":    ValueJson | ValueHide,
	"ChatId":     ValueString,
	"CourseId":   ValueString,
	"ClassId":    ValueString,
	"CourseName": ValueString,
	"ClassName":  ValueString,
}

func ValidKeyValue(k string, v any) bool {
	typ, ok := KeyValues[k]
	if !ok {
		return false
	}
	if typ == typ|ValueHide {
		return false
	}
	if typ == typ|ValueBool {
		_, ok = v.(bool)
		if ok {
			return true
		}
	}
	if typ == typ|ValueNumber {
		_, ok = v.(float64)
		if ok {
			return true
		}
	}
	if typ == typ|ValueString {
		_, ok = v.(string)
		if ok {
			return true
		}
	}
	if typ == typ|ValueJson {
		_, ok = v.(map[string]any)
		if ok {
			return true
		}
	}
	if typ == typ|ValueArray {
		_, ok = v.([]any)
		if ok {
			return true
		}
	}
	return false
}
