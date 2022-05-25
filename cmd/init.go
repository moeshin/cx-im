package cmd

import (
	"cx-im/config"
	"cx-im/core"
	"github.com/moeshin/go-errs"
	"github.com/spf13/cobra"
	"os"
)

var initCmd = &cobra.Command{
	Use:   "init <账号> <密码> [fid 学校编码]",
	Short: "账号初始化",
	Run: func(cmd *cobra.Command, args []string) {
		argc := len(args)
		if argc < 2 {
			cmd.Println("参数错误：至少需要『账号』和『密码』")
			errs.Panic(cmd.Help())
			os.Exit(1)
		}
		username := args[0]
		password := args[1]
		fid := ""
		if argc > 2 {
			fid = args[2]
		}

		appConfig := config.GetAppConfig()
		userConfig, err := appConfig.GetUserConfig(username)
		errs.Panic(err)
		client, err := core.NewClient(username, password, fid)
		errs.Panic(err)
		errs.Panic(client.Login())

		data := userConfig.Data
		data.Set(config.Username, username)
		data.Set(config.Password, password)
		data.Set(config.Fid, fid)

		courses := config.GocObj(data, config.Courses)
		errs.Panic(client.GetCourses(courses))
		errs.Panic(userConfig.Save())

		isSetDefault, err := cmd.Flags().GetBool("default")
		errs.Print(err)
		if isSetDefault || !appConfig.HasDefaultUsername() {
			appConfig.SetDefaultUsername(username)
		}
		errs.Panic(appConfig.Save())
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("default", "d", false, "设置为默认账号")
}
