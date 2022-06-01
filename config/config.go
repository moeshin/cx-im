package config

import (
	"cx-im/model"
	"encoding/json"
	"github.com/moeshin/go-errs"
	"github.com/moeshin/go-orderedmap"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Object = orderedmap.OrderedMap

func NewObject() *Object {
	v := orderedmap.New()
	v.SetEscapeHTML(false)
	return v
}

type Value interface {
	string | float64 | bool | []any | map[string]any
}

type User struct {
	Running bool
	Mutex   *sync.RWMutex
}

type Config struct {
	Path   string
	Data   *Object
	Parent *Config
	User   *User
	New    bool
	Mutex  *sync.RWMutex
}

func Load(path string, parent *Config) *Config {
	var v *Object
	file, err := os.Open(path)
	if err == nil {
		defer errs.Close(file)
		var data []byte
		data, err = ioutil.ReadAll(file)
		if err == nil {
			err = json.Unmarshal(data, &v)
		}
	}
	if !os.IsNotExist(err) {
		errs.Print(err)
	}
	New := v == nil
	if New {
		v = NewObject()
	}
	return &Config{
		path,
		v,
		parent,
		nil,
		New,
		newMutex(),
	}
}

func (c *Config) Get(k string, r bool) any {
	v, ok := c.GetC(k)
	if r && (!ok || v == nil) {
		if c.Parent == nil {
			v, _ = Default.Get(k)
		} else {
			v = c.Parent.Get(k, true)
		}
	}
	return v
}

func (c *Config) Set(k string, v any) {
	c.Lock()
	defer c.UnLock()
	c.Data.Set(k, v)
}

func (c *Config) GetR(k string) any {
	return c.Get(k, true)
}

func (c *Config) GetC(k string) (any, bool) {
	c.RLock()
	defer c.RUnlock()
	return c.Data.Get(k)
}

func (c *Config) JsonMarshal() ([]byte, error) {
	c.RLock()
	defer c.RUnlock()
	return json.MarshalIndent(c.Data, "", "  ")
}

func (c *Config) Save() error {
	if c.Path == "" {
		if c.Parent == nil {
			return nil
		}
		return c.Parent.Save()
	}
	dir := filepath.Dir(c.Path)
	data, err := c.JsonMarshal()
	if err != nil {
		return err
	}
	err = os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		return err
	}
	log.Println("保存配置：", c.Path)
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
	s := GodCI(c, DefaultUsername, "")
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
	log.Println("设置默认用户：" + username)
	c.Set(DefaultUsername, username)
}

func (c *Config) GetCourses() *Object {
	c.RLock()
	defer c.RUnlock()
	return GocObjI(c.Data, Courses)
}

func (c *Config) GetCourseConfig(chatId string) *Config {
	course, ok := GocObj(c.GetCourses(), chatId)
	return &Config{
		"",
		course,
		c,
		nil,
		!ok,
		newMutex(),
	}
}

func (c *Config) GetSignOptions(signTypeKey string) *model.SignOptions {
	if !GodRI(c, signTypeKey, false) {
		return nil
	}
	return &model.SignOptions{
		Address:   GodRI(c, SignAddress, DefaultSignAddress),
		Longitude: FloatToString(GodRI(c, SignLongitude, DefaultSignLongitude)),
		Latitude:  FloatToString(GodRI(c, SignLatitude, DefaultSignLatitude)),
		Ip:        GodRI(c, SignIp, DefaultSignIp),
	}
}

func (c *Config) Keys() []string {
	c.RLock()
	defer c.RUnlock()
	return c.Data.Keys()
}

var HasMutex = false

func newMutex() *sync.RWMutex {
	if HasMutex {
		return &sync.RWMutex{}
	}
	return nil
}

func (c *Config) Lock() {
	if c.Mutex == nil {
		return
	}
	c.Mutex.Lock()
}

func (c *Config) UnLock() {
	if c.Mutex == nil {
		return
	}
	c.Mutex.Unlock()
}

func (c *Config) RLock() {
	if c.Mutex == nil {
		return
	}
	c.Mutex.RLock()
}

func (c *Config) RUnlock() {
	if c.Mutex == nil {
		return
	}
	c.Mutex.RUnlock()
}

var Default = NewObject()

var AppConfig *Config
var UsersConfig = map[string]*Config{}

const UserDir = "users"

func GetUserDir(user string) string {
	return filepath.Join(UserDir, user)
}

func GetUserConfigPath(user string) string {
	return filepath.Join(GetUserDir(user), "config.json")
}

func GetAppConfig() *Config {
	if AppConfig == nil {
		v := Load("./settings.json", nil)
		if v.New {
			v.Data = Default
		}
		AppConfig = v
	}
	return AppConfig
}

func (c *Config) GetUserConfig(user string) *Config {
	v, ok := UsersConfig[user]
	if !ok {
		v = Load(GetUserConfigPath(user), c)
		UsersConfig[user] = v
	}
	return v
}

const (
	DefaultWebHost = ""
	DefaultWebPort = 5743.

	DefaultSignAddress   = "中国"
	DefaultSignLongitude = -1.
	DefaultSignLatitude  = -1.
	DefaultSignIp        = "1.1.1.1"
)

func init() {
	set := func(k string, v any) {
		Default.Set(k, v)
	}
	set(WebHost, DefaultWebHost)
	set(WebPort, DefaultWebPort)

	set(NotifyEmail, "")
	set(SmtpHost, "")
	set(SmtpPort, 465.)
	set(SmtpUsername, "")
	set(SmtpPassword, "")
	set(SmtpSSL, true)

	set(NotifyPushPlusToken, "")
	set(NotifyBarkApi, "")
	set(NotifyTelegramBotToken, "")
	set(NotifyTelegramBotChatId, "")

	set(NotifyActive, true)
	set(NotifySign, true)

	set(SignAddress, DefaultSignAddress)
	set(SignLongitude, DefaultSignLongitude)
	set(SignLatitude, DefaultSignLatitude)
	set(SignIp, DefaultSignIp)

	set(SignDelay, 0.)
	set(SignEnable, false)
	set(SignNormal, true)
	set(SignPhoto, true)
	set(SignGesture, true)
	set(SignLocation, true)
	set(SignCode, true)
}

func GocObj(data *Object, key string) (*Object, bool) {
	v, ok := data.Get(key)
	if ok {
		v, ok := v.(*Object)
		if ok {
			return v, true
		}
	}
	obj := NewObject()
	data.Set(key, obj)
	return obj, false
}

func GocObjI(data *Object, key string) *Object {
	v, _ := GocObj(data, key)
	return v
}

func God[T Value](config *Config, key string, def T, r bool) (T, bool) {
	v := config.Get(key, r)
	if v != nil {
		v, ok := v.(T)
		if ok {
			return v, true
		}
	}
	return def, false
}

func GodCI[T Value](config *Config, key string, def T) T {
	v, _ := God(config, key, def, false)
	return v
}

func GodRI[T Value](config *Config, key string, def T) T {
	v, _ := God(config, key, def, true)
	return v
}

func FloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
