package cmd

import (
	"cx-im/src/config"
	"cx-im/src/core"
	"github.com/moeshin/go-errs"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var updateCmd = &cobra.Command{
	Use:   "update [账号]",
	Short: "更新课程",
	Run: func(cmd *cobra.Command, args []string) {
		core.ClientNormalLogin = true
		all, _ := cmd.Flags().GetBool("all")
		if all {
			updateAll()
			return
		}
		user, err := getUser(cmd, args)
		errs.Panic(err)
		client := user.Client
		userConfig := user.Config
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
	dirs, err := os.ReadDir(config.DirUser)
	errs.Panic(err)
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		name := dir.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		user, err := core.GetUser(name)
		if errs.Print(err) {
			continue
		}
		client := user.Client
		if errs.Print(client.Login()) {
			continue
		}
		userConfig := user.Config
		courses := userConfig.GetCourses()
		if errs.Print(client.GetCourses(courses)) {
			continue
		}
		errs.Print(userConfig.Save())
	}
}
