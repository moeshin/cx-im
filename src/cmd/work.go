package cmd

import (
	"cx-im/src/core"
	"github.com/moeshin/go-errs"
	"github.com/spf13/cobra"
)

var workCmd = &cobra.Command{
	Use:   "work [账号]",
	Short: "工作模式，监听 IM 即时通讯",
	Run: func(cmd *cobra.Command, args []string) {
		userConfig := getUserConfig(cmd, args)
		work := core.NewWork(userConfig, nil)
		errs.Panic(work.Connect())
	},
}

func init() {
	rootCmd.AddCommand(workCmd)
}
