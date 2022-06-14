class ConfigValues {
    static ValueHide = 1;
    static ValueBool = 2;
    static ValueNumber = 4;
    static ValueString = 8;
    static ValuePassword = 16;
    static ValueJson = 32;
    static ValueArray = 64;
    static ValueLevelApp = 128;
    static ValueLevelUser = 256;
    static ValueLevelCourse = 512;
    static ValueLevelAU = this.ValueLevelApp | this.ValueLevelUser;
    static ValueLevelAUC = this.ValueLevelAU | this.ValueLevelCourse;
}
window.ConfigKey = {
    SmtpPort: "SmtpPort",
    SmtpUsername: "SmtpUsername",
    SignLatitude: "SignLatitude",
    SignLocation: "SignLocation",
    ClassId: "ClassId",
    SmtpHost: "SmtpHost",
    NotifyActive: "NotifyActive",
    SignDelay: "SignDelay",
    SignEnable: "SignEnable",
    SignGesture: "SignGesture",
    NotifySign: "NotifySign",
    SignPhoto: "SignPhoto",
    SignCode: "SignCode",
    ChatId: "ChatId",
    SignAddress: "SignAddress",
    Courses: "Courses",
    ClassName: "ClassName",
    SignIp: "SignIp",
    Password: "Password",
    CourseName: "CourseName",
    WebPort: "WebPort",
    NotifyBarkApi: "NotifyBarkApi",
    NotifyTelegramBotChatId: "NotifyTelegramBotChatId",
    SignLongitude: "SignLongitude",
    Username: "Username",
    SmtpSSL: "SmtpSSL",
    NotifyTelegramBotToken: "NotifyTelegramBotToken",
    Fid: "Fid",
    NotifyEmail: "NotifyEmail",
    WebHost: "WebHost",
    SmtpPassword: "SmtpPassword",
    NotifyPushPlusToken: "NotifyPushPlusToken",
    SignNormal: "SignNormal",
    CourseId: "CourseId",
    DefaultUsername: "DefaultUsername",
};
window.ConfigKeyValues = {
    Fid: 264,
    NotifyEmail: 904,
    NotifyTelegramBotToken: 920,
    SmtpPassword: 920,
    NotifyPushPlusToken: 920,
    SignNormal: 898,
    CourseId: 520,
    DefaultUsername: 136,
    WebHost: 136,
    SmtpUsername: 904,
    SignLatitude: 900,
    SignLocation: 898,
    ClassId: 520,
    SmtpHost: 904,
    SmtpPort: 900,
    SignDelay: 900,
    SignEnable: 898,
    SignGesture: 898,
    NotifySign: 898,
    NotifyActive: 898,
    SignCode: 898,
    ChatId: 521,
    SignAddress: 904,
    SignPhoto: 1002,
    ClassName: 520,
    SignIp: 904,
    Courses: 289,
    CourseName: 520,
    WebPort: 132,
    Password: 280,
    NotifyTelegramBotChatId: 904,
    SignLongitude: 900,
    Username: 264,
    SmtpSSL: 898,
    NotifyBarkApi: 920,
};
