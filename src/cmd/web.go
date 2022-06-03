package cmd

import (
	"cx-im/src/config"
	"cx-im/src/core"
	"fmt"
	"github.com/moeshin/go-errs"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "网页模式",
	Run: func(cmd *cobra.Command, args []string) {
		config.HasMutex = true
		webRun()
	},
}

var webArgs = struct {
	host string
	port int
	work bool
}{}

func init() {
	rootCmd.AddCommand(webCmd)
	flags := webCmd.Flags()
	flags.StringVarP(&webArgs.host, "host", "a", "", "主机")
	flags.IntVarP(&webArgs.port, "port", "p", 0, "端口")
	//flags.BoolVarP(&webArgs.work, "work", "w", false, "立即运行监听")
}

func webRun() {
	config.InitUsersConfig()
	appConfig := config.GetAppConfig()

	// 载入所有用户
	dirs, err := os.ReadDir(config.UserDir)
	errs.Panic(err)
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		name := dir.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		log.Println("载入用户配置：" + name)
		appConfig.GetUserConfig(name)
	}

	webHost := webArgs.host
	if webHost == "" {
		webHost = config.GodCI(appConfig, config.WebHost, config.DefaultWebHost)
	}
	webPort := webArgs.port
	if webPort == 0 {
		webPort = int(config.GodCI(appConfig, config.WebPort, config.DefaultWebPort))
	}
	webAddress := fmt.Sprintf("%s:%d", webHost, webPort)
	{
		s := webAddress
		if webHost == "" {
			s = "*" + s
		}
		log.Println("网页监听：" + s)
	}

	webHandler := &WebHandler{}
	errs.Panic(http.ListenAndServe(webAddress, webHandler))
}

type WebHandler struct {
}

func (h *WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer errs.Close(r.Body)
	urlPath := r.URL.Path
	if strings.HasPrefix(urlPath, "/api/") {
		api := core.NewApi(r)
		defer api.Response(w)
		urlPath = urlPath[5:]
		log.Println(r.Method, urlPath)
		switch urlPath {
		case "users":
			data := map[string]bool{}
			config.UsersConfig.Mutex.RLock()
			for k, v := range config.UsersConfig.Map {
				v.User.Mutex.RLock()
				data[k] = v.User.Running
				v.User.Mutex.RUnlock()
			}
			config.UsersConfig.Mutex.RUnlock()
			api.O(data)
			return
		case "users/start":
			fallthrough
		case "users/stop":
			run := urlPath[6:] == "start"
			config.UsersConfig.Mutex.RLock()
			for _, v := range config.UsersConfig.Map {
				v.User.Mutex.RLock()
				ok := v.User.Running != run
				if ok {
					v.User.Running = run
				}
				v.User.Mutex.RUnlock()
				if ok && run {
					go core.StartWork(v)
				}
			}
			config.UsersConfig.Mutex.RUnlock()
			api.Ok = true
			return
		case "user/start":
			fallthrough
		case "user/stop":
			fallthrough
		case "user":
			username := r.URL.Query().Get("username")
			if username != "" {
				root := urlPath == "user"
				if root {
					switch r.Method {
					case http.MethodPost:
						data, err := ioutil.ReadAll(r.Body)
						if api.Err(err) {
							return
						}
						query := r.URL.Query()
						password := string(data)
						fid := query.Get("fid")
						def := false
						{
							s := query.Get("default")
							if s != "" {
								def, err = strconv.ParseBool(s)
								if api.Err(err) {
									return
								}
							}
						}
						log.Println("创建用户：" + username)
						err = initUser(username, password, fid, def)
						api.Ok = true
						api.Err(err)
						return
					case http.MethodDelete:
						config.UsersConfig.Mutex.Lock()
						delete(config.UsersConfig.Map, username)
						config.UsersConfig.Mutex.Unlock()
						api.Ok = true
						api.Err(os.RemoveAll(config.GetUserDir(username)))
						return
					}
				}
				v, ok := config.UsersConfig.Get(username)
				if !ok {
					api.OE("用户不存在：" + username)
					return
				}
				if root {
					v.User.Mutex.RLock()
					api.O(v.User.Running)
					v.User.Mutex.RUnlock()
				} else {
					run := urlPath[5:] == "start"
					v.User.Mutex.Lock()
					ok = v.User.Running != run
					if ok {
						v.User.Running = run
					}
					api.O(ok)
					v.User.Mutex.Unlock()
					if ok && run {
						go core.StartWork(v)
					}
				}
				return
			}
		default:
			if strings.HasPrefix(urlPath, "cfg/") {
				urlPath = urlPath[4:]
				if urlPath == "app" {
					api.HandleConfig("")
					return
				} else {
					username := r.URL.Query().Get("username")
					if urlPath == "user" {
						if username != "" {
							api.HandleConfig(username)
							return
						}
					} else if urlPath == "user/course" && r.Method == http.MethodPost {
						chatId := r.URL.Query().Get("chatId")
						if chatId != "" {
							cfg := config.GetAppConfig().GetUserConfig(username).GetCourseConfig(chatId)
							api.SetConfigValues(cfg)
							return
						}
					}
				}
			}
		}
		api.Bad()
	}
	w.WriteHeader(http.StatusBadRequest)
}
