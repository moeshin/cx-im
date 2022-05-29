package cmd

import (
	"cx-im/config"
	"cx-im/core"
	"github.com/moeshin/go-errs"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var updateCmd = &cobra.Command{
	Use:   "update [账号]",
	Short: "更新课程",
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		if all {
			updateAll()
			return
		}
		userConfig := getUserConfig(cmd, args)
		client, err := core.NewClientFromConfig(userConfig, nil)
		errs.Panic(err)
		errs.Panic(client.Login())
		courses := userConfig.GetCourses()
		errs.Panic(client.GetCourses(courses))
		errs.Panic(userConfig.Save())
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolP("all", "a", false, "更新全部账号")
}

func updateAll() {
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
		userConfig := config.Load(config.GetUserConfigPath(name), nil)
		client, err := core.NewClientFromConfig(userConfig, nil)
		if errs.Print(err) {
			continue
		}
		if errs.Print(client.Login()) {
			continue
		}
		courses := userConfig.GetCourses()
		if errs.Print(client.GetCourses(courses)) {
			continue
		}
		errs.Print(userConfig.Save())
	}
}
