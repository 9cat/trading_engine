package main

import (
	"context"
	"os"

	"github.com/sevlyar/go-daemon"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"github.com/yzimhao/trading_engine/haoquote"
	"github.com/yzimhao/trading_engine/utils/app"
)

//	@title			Haoquote交易行情管理系统
//	@version		1.0
//	@description	根据成交记录，快速统计出各个时间周期的行情数据。

//	@contact.name	yzimhao
//	@contact.url	https://github.com/yzimhao

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// host		www.demo.com
// @BasePath	/api/v1
func main() {
	appm := &cli.App{
		Name:      "haoquote",
		UsageText: "Issues: https://github.com/yzimhao/trading_engine/issues",
		Usage:     "交易行情管理系统",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: "./config.toml", Aliases: []string{"c"}},
			&cli.BoolFlag{Name: "deamon", Value: false, Aliases: []string{"d"}},
		},

		Before: func(ctx *cli.Context) error {
			return nil
		},

		Commands: []*cli.Command{
			{
				Name: "version",
				Action: func(ctx *cli.Context) error {
					app.ShowVersion()
					return nil
				},
			},
		},
		Action: func(ctx *cli.Context) error {
			app.ConfigInit(ctx.String("config"))
			app.LogsInit("haoquote.run", ctx.Bool("deamon"))

			if viper.GetString("main.mode") != "prod" {
				logrus.Infof("当前运行在%s模式下，生产环境时main.mode请务必成prod", viper.GetString("main.mode"))
			}

			if ctx.Bool("deamon") {
				logrus.Info("开始守护进程")
				context, d, err := app.Deamon("haoquote.pid", "")
				if err != nil {
					logrus.Fatal("创建守护进程失败, err:", err.Error())
				}
				if d != nil {
					logrus.Printf("这是在父进程的标志")
					return nil
				}

				defer func(context *daemon.Context) {
					err := context.Release()
					if err != nil {
						logrus.Printf("释放失败:%s", err.Error())
					}
					logrus.Printf("释放成功!!!")
				}(context)

			}

			db := app.DatabaseInit()

			rc := app.RedisInit()
			ctext := context.Background()
			haoquote.Start(&ctext, rc, db)
			return nil
		},
	}

	if err := appm.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
