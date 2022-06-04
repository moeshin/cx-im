package cmd

import (
	"cx-im/src/config"
	"cx-im/src/core"
	"github.com/moeshin/go-errs"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:     "cx-im",
	Short:   "超星学习通 IM 即时通讯",
	Version: "1.0.0",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func getUser(cmd *cobra.Command, args []string) (*core.User, error) {
	appConfig := config.GetAppConfig()
	def := core.GetDefaultUsername()
	var username string
	if len(args) == 0 {
		if def == "" {
			log.Println("参数错误：没有设置『默认账号』，需要指定『账号』")
			log.Println(cmd.Help())
			os.Exit(1)
		} else {
			username = def
		}
	} else {
		username = args[0]
	}
	user, err := core.GetUser(username)
	if err != nil {
		return nil, err
	}
	userConfig := user.Config
	if userConfig.New {
		log.Println("无该账号配置，请初始化")
		log.Println(initCmd.Help())
		os.Exit(1)
	}
	if def == "" {
		core.SetDefaultUsername(username)
		errs.Print(appConfig.Save())
	}
	return user, nil
}
