package cmd

import (
	"cx-im/config"
	"cx-im/core"
	"github.com/moeshin/go-errs"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var workCmd = &cobra.Command{
	Use:   "work [账号]",
	Short: "工作模式，监听 IM 即时通讯",
	Run: func(cmd *cobra.Command, args []string) {
		appConfig := config.GetAppConfig()
		def := appConfig.GetDefaultUsername()
		var username string
		if len(args) == 0 {
			if def == "" {
				cmd.Println("参数错误：没有设置『默认账号』，需要指定『账号』")
				errs.Panic(cmd.Help())
				os.Exit(1)
			} else {
				username = def
			}
		} else {
			username = args[0]
		}
		userConfig := appConfig.GetUserConfig(username)
		if userConfig.New {
			log.Println("无该账号配置，请初始化")
			errs.Panic(initCmd.Help())
			os.Exit(1)
			return
		}
		if def == "" {
			appConfig.SetDefaultUsername(username)
			errs.Print(appConfig.Save())
		}
		work := core.NewWork(userConfig, nil)
		errs.Panic(work.Connect())
	},
}

func init() {
	rootCmd.AddCommand(workCmd)
}
