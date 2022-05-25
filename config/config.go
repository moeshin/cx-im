package config

import (
	"encoding/json"
	"github.com/iancoleman/orderedmap"
	"github.com/moeshin/go-errs"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
)

type Object = orderedmap.OrderedMap

type User struct {
	Running bool
	Mutex   *sync.RWMutex
}

type Config struct {
	Path   string
	Data   *Object
	Parent *Config
	User   *User
}

func Load(path string, parent *Config) (*Config, bool) {
	var v *Object
	file, err := os.Open(path)
	if err == nil {
		defer errs.Close(file)
		var data []byte
		data, err = ioutil.ReadAll(file)
		if err == nil {
			err = json.Unmarshal(data, &v)
		}
		errs.Print(err)
	}
	if v == nil {
		v = orderedmap.New()
	}
	return &Config{
		path,
		v,
		parent,
		nil,
	}, err != nil
}

func (c *Config) Get(key string) any {
	if v, ok := c.Data.Get(key); ok {
		return v
	}
	return c.Parent.Get(key)
}

func (c *Config) JsonMarshal() ([]byte, error) {
	return json.MarshalIndent(c.Data, "", "  ")
}

func (c *Config) Save() error {
	dir := path.Dir(c.Path)
	data, err := c.JsonMarshal()
	if err != nil {
		return err
	}
	err = os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.Path, data, fs.ModePerm)
}

func (c *Config) String() string {
	data, err := c.JsonMarshal()
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (c *Config) GetDefaultUsername() string {
	var s string
	v, ok := c.Data.Get(DefaultUsername)
	if ok {
		s, _ = v.(string)
	}
	if s != "" {
		_, err := os.Stat(GetUserConfigPath(s))
		if err != nil {
			return ""
		}
	}
	return s
}

func (c *Config) HasDefaultUsername() bool {
	return c.GetDefaultUsername() != ""
}

func (c *Config) SetDefaultUsername(username string) {
	log.Println("设置默认用户：", username)
	c.Data.Set(DefaultUsername, username)
}

var Default = orderedmap.New()

var AppConfig *Config
var UsersConfig = map[string]*Config{}

func GetUserDir(user string) string {
	return "users/" + user
}

func GetUserConfigPath(user string) string {
	return GetUserDir(user) + "/config.json"
}

func GetAppConfig() *Config {
	if AppConfig == nil {
		v, err := Load("./settings.json", nil)
		if err {
			v.Data = Default
		}
		AppConfig = v
	}
	return AppConfig
}

func (c *Config) GetUserConfig(user string) (*Config, error) {
	v, ok := UsersConfig[user]
	if !ok {
		v, _ = Load(GetUserConfigPath(user), c)
		UsersConfig[user] = v
	}
	return v, nil
}

func init() {
	set := func(k string, v any) {
		Default.Set(k, v)
	}
	set(Email, "")
	set(SmtpHost, "")
	set(SmtpPort, 465)
	set(SmtpUsername, "")
	set(SmtpPassword, "")
	set(SmtpSSL, true)

	set(PushPlusToken, "")
	set(TelegramBotToken, "")
	set(TelegramBotChatId, "")

	set(SignAddress, "中国")
	set(SignLongitude, -1)
	set(SignLatitude, -1)
	set(SignIp, "1.1.1.1")

	set(SignDelay, 0)
	set(SignEnable, false)
	set(SignNormal, true)
	set(SignPhoto, true)
	set(SignGesture, true)
	set(SignLocation, true)
	set(SignCode, true)
}

func GocObj(data *Object, key string) *Object {
	r := orderedmap.New()
	v, ok := data.Get(key)
	if ok {
		vv, _ := v.(Object)
		r = &vv
	} else {
		data.Set(key, r)
	}
	return r
}
