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

	ValueLevelApp
	ValueLevelUser
	ValueLevelCourse
	ValueLevelAU  = ValueLevelApp | ValueLevelUser
	ValueLevelAUC = ValueLevelAU | ValueLevelCourse
)

var KeyValues = map[string]int{
	"DefaultUsername": ValueLevelApp | ValueString,

	"WebHost": ValueLevelApp | ValueString,
	"WebPort": ValueLevelApp | ValueNumber,

	"NotifyEmail":  ValueLevelAUC | ValueString,
	"SmtpHost":     ValueLevelAUC | ValueString,
	"SmtpPort":     ValueLevelAUC | ValueNumber,
	"SmtpUsername": ValueLevelAUC | ValueString,
	"SmtpPassword": ValueLevelAUC | ValueString | ValuePassword,
	"SmtpSSL":      ValueBool,

	"NotifyPushPlusToken":     ValueLevelAUC | ValueString | ValuePassword,
	"NotifyBarkApi":           ValueLevelAUC | ValueString | ValuePassword,
	"NotifyTelegramBotToken":  ValueLevelAUC | ValueString | ValuePassword,
	"NotifyTelegramBotChatId": ValueLevelAUC | ValueString,

	"NotifySign":   ValueLevelAUC | ValueBool,
	"NotifyActive": ValueLevelAUC | ValueBool,

	"SignAddress":   ValueLevelAUC | ValueString,
	"SignLongitude": ValueLevelAUC | ValueNumber,
	"SignLatitude":  ValueLevelAUC | ValueNumber,
	"SignIp":        ValueLevelAUC | ValueString,

	"SignDelay":    ValueLevelAUC | ValueNumber,
	"SignEnable":   ValueLevelAUC | ValueBool,
	"SignNormal":   ValueLevelAUC | ValueBool,
	"SignPhoto":    ValueLevelAUC | ValueBool | ValueString | ValueArray | ValueJson,
	"SignGesture":  ValueLevelAUC | ValueBool,
	"SignLocation": ValueLevelAUC | ValueBool,
	"SignCode":     ValueLevelAUC | ValueBool,

	"Username": ValueLevelUser | ValueString,
	"Password": ValueLevelUser | ValueString | ValuePassword,
	"Fid":      ValueLevelUser | ValueString,

	"Courses":    ValueLevelUser | ValueJson | ValueHide,
	"ChatId":     ValueLevelCourse | ValueString | ValueHide,
	"CourseId":   ValueLevelCourse | ValueString,
	"ClassId":    ValueLevelCourse | ValueString,
	"CourseName": ValueLevelCourse | ValueString,
	"ClassName":  ValueLevelCourse | ValueString,
}

func ValidKeyValue(lv int, k string, v any) bool {
	typ, ok := KeyValues[k]
	if !ok {
		return false
	}
	if typ != typ|lv || typ == typ|ValueHide {
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
