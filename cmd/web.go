package cmd

import (
	"cx-im/config"
	"cx-im/core"
	"fmt"
	"github.com/moeshin/go-errs"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "网页模式",
	Run: func(cmd *cobra.Command, args []string) {
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
		log.Println("载入用户配置：", name)
		userConfig := appConfig.GetUserConfig(name)
		userConfig.User = &config.User{
			Running: false,
			Mutex:   &sync.RWMutex{},
		}
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
	log.Println("网页监听：", webAddress)

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
		if urlPath == "users" {
			data := map[string]bool{}
			for k, v := range config.UsersConfig {
				v.User.Mutex.RLock()
				data[k] = v.User.Running
				v.User.Mutex.RUnlock()
			}
			api.O(data)
			return
		}
		if strings.HasPrefix(urlPath, "cfg/") {
			urlPath = urlPath[4:]
			if urlPath == "app" {
				api.HandleConfig("")
				return
			} else if strings.HasPrefix(urlPath, "user/") {
				username := urlPath[5:]
				if username != "" {
					api.HandleConfig(username)
					return
				}
			}
		}
	}
	w.WriteHeader(http.StatusBadRequest)
}
