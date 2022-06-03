package cmd

import (
	"cx-im/src/config"
	"cx-im/src/core"
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
		def, err := cmd.Flags().GetBool("default")
		errs.Print(err)
		errs.Panic(initUser(username, password, fid, def))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("default", "d", false, "设置为默认账号")
}

func initUser(username string, password string, fid string, def bool) error {
	appConfig := config.GetAppConfig()
	userConfig := appConfig.GetUserConfig(username)
	client, err := core.NewClient(username, password, fid, nil)
	if err != nil {
		return err
	}
	err = client.Login()
	if err != nil {
		return err
	}

	userConfig.Set(config.Username, username)
	userConfig.Set(config.Password, password)
	userConfig.Set(config.Fid, fid)

	courses := userConfig.GetCourses()
	err = client.GetCourses(courses)
	if err != nil {
		return err
	}
	err = userConfig.Save()
	if err != nil {
		return err
	}

	save := appConfig.New
	if def || !appConfig.HasDefaultUsername() {
		appConfig.SetDefaultUsername(username)
		save = true
	}
	if save {
		err = appConfig.Save()
		if err != nil {
			return err
		}
	}
	return nil
}
