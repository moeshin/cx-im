package config

import (
	"encoding/json"
	"github.com/iancoleman/orderedmap"
	"github.com/moeshin/go-errs"
	"io/fs"
	"io/ioutil"
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

var Default = orderedmap.New()

var AppConfig *Config
var UsersConfig = map[string]*Config{}

func GetUserDir(user string) string {
	return "users/" + user
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
		v, _ = Load(GetUserDir(user)+"/config.json", c)
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
