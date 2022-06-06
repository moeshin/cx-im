package core

import (
	"cx-im/src/config"
	"github.com/moeshin/go-errs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const UsersDir = "users"

type UserDir struct {
	Path string
}

func GetUserDirPath(username string) string {
	return filepath.Join(UsersDir, username)
}

func GetUserConfigPath(username string) string {
	return filepath.Join(UsersDir, username)
}

func GetDefaultUsername() string {
	username := config.GodCI(config.GetAppConfig(), config.DefaultUsername, "")
	if username != "" {
		_, err := os.Stat(GetUserConfigPath(username))
		if err != nil {
			return ""
		}
	}
	return username
}

func HasDefaultUsername() bool {
	return GetDefaultUsername() != ""
}

func SetDefaultUsername(username string) {
	log.Println("设置默认用户：" + username)
	config.GetAppConfig().Set(config.DefaultUsername, username)
}

func NewUserDir(username string) (*UserDir, error) {
	dir := GetUserDirPath(username)
	err := os.MkdirAll(dir, 0666)
	if err != nil {
		return nil, err
	}
	return &UserDir{
		Path: dir,
	}, nil
}

func (u *UserDir) Join(name string) string {
	return filepath.Join(u.Path, name)
}

func (u *UserDir) GetConfigPath() string {
	return u.Join("config.json")
}

func (u *UserDir) GetLogPath() string {
	return u.Join("log.txt")
}

func (u *UserDir) GetCookiesPath() string {
	return u.Join("cookies.json")
}

func (u *UserDir) GetImageTokenPath() string {
	return u.Join("image-token.txt")
}

type User struct {
	Username   string
	Password   string
	Fid        string
	Dir        *UserDir
	Config     *config.Config
	Client     *CxClient
	ImageToken string
	Log        *LogE
	LogFile    *os.File
}

func NewUser(username string) (*User, error) {
	dir, err := NewUserDir(username)
	if err != nil {
		return nil, err
	}
	var _ok bool
	var file *os.File
	cfg := config.Load(dir.GetConfigPath(), config.GetAppConfig())
	var logger *log.Logger
	if Users == nil {
		logger = log.Default()
	} else {
		file, err = os.OpenFile(dir.GetLogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
		defer func() {
			if !_ok {
				errs.Close(file)
			}
		}()
		logger = NewLogger(file)
		cfg.User = &config.User{
			Running: false,
			Mutex:   &sync.RWMutex{},
		}
	}
	user := &User{
		Username: config.GodCI(cfg, config.Username, ""),
		Password: config.GodCI(cfg, config.Password, ""),
		Fid:      config.GodCI(cfg, config.Fid, ""),
		Dir:      dir,
		Config:   cfg,
		Log:      &LogE{logger},
		LogFile:  file,
	}
	user.Client, err = NewClient(user)
	if err != nil {
		return nil, err
	}
	if !ClientNormalLogin {
		err = user.LoadImageToken()
		if err != nil {
			return nil, err
		}
	}
	_ok = true
	return user, nil
}

func (u *User) Close() error {
	if u.LogFile == nil {
		return nil
	}
	return u.LogFile.Close()
}

func (u *User) LoadImageToken() error {
	filename := u.Dir.GetImageTokenPath()
	file, err := os.Open(filename)
	if err != nil {
		return u.SaveImageToken(nil)
	}
	u.Log.Println("加载图床凭证缓存：" + filename)
	defer u.Log.ErrClose(file)
	data, err := ioutil.ReadAll(file)
	if err == nil {
		token := strings.TrimSpace(string(data))
		if token != "" {
			u.ImageToken = token
			return nil
		}
	}
	return u.SaveImageToken(nil)
}

func (u *User) SaveImageToken(client *CxClient) error {
	if client == nil {
		client = u.Client
	}
	token, err := client.GetImageToken()
	if err != nil {
		return err
	}
	u.ImageToken = token
	data, err := JsonMarshal(u.ImageToken)
	if err != nil {
		return err
	}
	filename := u.Dir.GetImageTokenPath()
	log.Println("保存图床凭证缓存：" + filename)
	err = os.WriteFile(filename, data, 0666)
	u.Log.ErrPrint(err)
	return nil
}

type users struct {
	Mutex *sync.RWMutex
	Map   map[string]*User
}

var Users *users

func InitUsers() {
	Users = &users{
		Mutex: &sync.RWMutex{},
		Map:   map[string]*User{},
	}
}

func (u *users) Get(user string) (*User, bool) {
	u.Mutex.RLock()
	defer u.Mutex.RUnlock()
	v, ok := u.Map[user]
	return v, ok
}

func (u *users) Set(user string, cfg *User) {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()
	v, ok := u.Map[user]
	if ok {
		errs.Close(v)
	}
	u.Map[user] = cfg
}

func (u *users) Close() error {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()
	for _, user := range u.Map {
		errs.Close(user)
	}
	return nil
}

func GetUser(username string) (*User, error) {
	if Users != nil {
		Users.Mutex.Lock()
		defer Users.Mutex.Unlock()
		user, ok := Users.Map[username]
		if ok {
			return user, nil
		}
	}
	user, err := NewUser(username)
	if err != nil {
		return nil, err
	}
	if Users != nil {
		Users.Map[username] = user
	}
	return user, nil
}
