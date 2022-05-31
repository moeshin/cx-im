package config

type ValueType = int

const (
	ValueHide ValueType = 1 << iota
	ValueNumber
	ValueString
	ValueBool
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
	"SignPhoto":    ValueBool | ValueArray | ValueJson,
	"SignGesture":  ValueBool,
	"SignLocation": ValueBool,
	"SignCode":     ValueBool,

	"Username": ValueString,
	"Password": ValueString | ValueHide,
	"Fid":      ValueString,

	"Courses":    ValueJson,
	"ChatId":     ValueString,
	"CourseId":   ValueString,
	"ClassId":    ValueString,
	"CourseName": ValueString,
	"ClassName":  ValueString,
}
