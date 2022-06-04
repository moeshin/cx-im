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
	core.InitUsers()
	defer errs.Close(core.Users)
	appConfig := config.GetAppConfig()

	// 载入所有用户
	dirs, err := os.ReadDir(core.UsersDir)
	errs.Panic(err)
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		username := dir.Name()
		if strings.HasPrefix(username, ".") {
			continue
		}
		log.Println("载入用户配置：" + username)
		_, err = core.GetUser(username)
		errs.Print(err)
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
			core.Users.Mutex.RLock()
			for k, v := range core.Users.Map {
				v.Config.User.Mutex.RLock()
				data[k] = v.Config.User.Running
				v.Config.User.Mutex.RUnlock()
			}
			core.Users.Mutex.RUnlock()
			api.O(data)
			return
		case "users/start":
			fallthrough
		case "users/stop":
			run := urlPath[6:] == "start"
			core.Users.Mutex.RLock()
			for _, user := range core.Users.Map {
				user.Config.User.Mutex.RLock()
				ok := user.Config.User.Running != run
				if ok {
					user.Config.User.Running = run
				}
				user.Config.User.Mutex.RUnlock()
				if ok && run {
					go core.StartWork(user)
				}
			}
			core.Users.Mutex.RUnlock()
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
						core.Users.Mutex.Lock()
						user, ok := core.Users.Map[username]
						if ok {
							user.Config.User.Mutex.Lock()
							user.Config.User.Running = false
							user.Config.User.Mutex.Unlock()
							errs.Close(user)
							delete(core.Users.Map, username)
							api.Ok = true
							api.Err(os.RemoveAll(user.Dir.Path))
						}
						core.Users.Mutex.Unlock()
						return
					}
				}
				user, ok := core.Users.Get(username)
				if !ok {
					api.OE("用户不存在：" + username)
					return
				}
				if root {
					user.Config.User.Mutex.RLock()
					api.O(user.Config.User.Running)
					user.Config.User.Mutex.RUnlock()
				} else {
					run := urlPath[5:] == "start"
					user.Config.User.Mutex.Lock()
					ok = user.Config.User.Running != run
					if ok {
						user.Config.User.Running = run
					}
					api.O(ok)
					user.Config.User.Mutex.Unlock()
					if ok && run {
						go core.StartWork(user)
					}
				}
				return
			}
		case "images":
			if r.Method == http.MethodGet {
				var ids []string
				core.CacheImage.Mutex.Lock()
				for _, id := range core.CacheImage.Map {
					ids = append(ids, id)
				}
				core.CacheImage.Mutex.Unlock()
				api.O(ids)
				return
			}
		case "image":
			if r.Method == http.MethodPost {
				username := r.URL.Query().Get("username")
				if username == "" {
					username = core.GetDefaultUsername()
					if username == "" {
						api.OE("没有指定账号和默认账号")
						return
					}
				}
				file, header, err := r.FormFile("image")
				if api.Err(err) {
					return
				}
				defer errs.Close(file)
				user, err := core.GetUser(username)
				if api.Err(err) {
					return
				}
				id, err := user.Client.GetImageId(header.Filename, file, header.Size)
				if api.Err(err) {
					return
				}
				api.O(id)
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
							user, err := core.GetUser(username)
							if api.Err(err) {
								return
							}
							api.SetConfigValues(user.Config.GetCourseConfig(chatId))
							return
						}
					}
				}
			}
		}
		api.Bad()
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}
