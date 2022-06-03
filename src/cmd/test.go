package cmd

import (
	"cx-im/src/config"
	"cx-im/src/core"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var regexpInt = regexp.MustCompile(`^\d+$`)

var testCmd = &cobra.Command{
	Use:   "test [账号]",
	Short: "测试签到配置",
	Run: func(cmd *cobra.Command, args []string) {
		signType := core.GetSignType(testArgs.typ)
		now := time.Now()
		if signType == core.SignTypePhoto {
			s := strings.TrimSpace(testArgs.now)
			if s != "" {
				if regexpInt.MatchString(s) {
					n, err := strconv.ParseInt(s, 10, 64)
					if err != nil {
						log.Println("模拟时间失败，无法转换时间戳", err)
						os.Exit(1)
					}
					if n < 1e12 {
						n *= 1000
					}
					now = time.UnixMilli(n)
				} else {
					t, err := time.Parse(config.TimeLayout, s)
					if err != nil {
						log.Println("模拟时间失败，无法解析时间", err)
						os.Exit(1)
					}
					now = t
				}
			}
		}
		userConfig := getUserConfig(cmd, args)
		logE := &core.LogE{Logger: log.Default()}
		logN := logE.NewLogN(userConfig)
		logN.State = core.NotifySign
		logN.Println("这是一个测试")
		defer logE.ErrClose(logN)
		courseConfig := userConfig.GetCourseConfig(testArgs.id)
		if courseConfig.New {
			logN.Println("该课程不在配置列表")
		}
		logN.Printf(
			"收到来自《%s》的%s",
			config.GodCI(courseConfig, config.CourseName, "测试课程"),
			core.GetSignTypeName(signType),
		)
		logN.Cfg = courseConfig
		work := core.NewWorkSign(courseConfig, logN.LogE)
		if work.SetSignType(signType, core.JObject{}) || work.IsSkip() {
			return
		}
		if signType == core.SignTypePhoto {
			logN.Println("模拟时间：" + now.Format(config.TimeLayout))
			imageId := work.GetImageId(now, nil)
			logN.Println("预览（略缩图）：" + config.GetSignPhotoImageUrl(imageId, false))
			logN.Println("预览（原图）：" + config.GetSignPhotoImageUrl(imageId, true))
		}
		logN.State = core.NotifySignOk
		opts := work.Opts
		logN.Printf(`签到信息
地址：%s
经纬度：%s, %s
IP: %s`,
			opts.Address,
			opts.Longitude, opts.Latitude,
			opts.Ip,
		)
	},
}

var testArgs = struct {
	id  string
	typ int8
	now string
}{}

func init() {
	signTypes := ""
	for i := int8(0); i < core.SignTypeLength; i++ {
		if i != 0 {
			signTypes += "，"
		}
		if i%3 == 0 {
			signTypes += "\n"
		}
		signTypes += fmt.Sprintf("%d %s", i, core.GetSignTypeName(i))
	}
	rootCmd.AddCommand(testCmd)
	flags := testCmd.Flags()
	flags.StringVarP(&testArgs.id, "id", "i", "", "配置中的 ChatId")
	flags.Int8VarP(&testArgs.typ, "type", "t", 0, "签到类型，默认 0："+signTypes)
	flags.StringVarP(&testArgs.now, "now", "n", "",
		"模拟当前时间，用于拍照签到。\n可为时间戳或如：\n"+config.TimeLayout)
}
