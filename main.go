package main

import (
	"flag"

	"github.com/yongchengchen/goslackapp/app/api"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/sirupsen/logrus"
)

func main() {
	cx := gctx.New()

	botToken, _ := g.Cfg().Get(cx, "SLACK_BOT_TOKEN", true)
	appToken, _ := g.Cfg().Get(cx, "SLACK_SIGNING_SECRET", true)
	toAPIEndPoint, _ := g.Cfg().Get(cx, "PROXY_TO_API_ENDPOINT", true)
	toAPIToken, _ := g.Cfg().Get(cx, "PROXY_TO_API_TOKEN", true)
	appToChannel, _ := g.Cfg().Get(cx, "SLACK_CHANNEL_URL", true)

	bDebug := flag.Bool("debug", false, "Output debug message")
	flag.Parse()
	if *bDebug {
		logrus.Println("Debug ", *bDebug)
		logrus.SetLevel(logrus.DebugLevel)
	}

	api.SlackServe(cx, botToken.String(), appToken.String(), toAPIEndPoint.String(), toAPIToken.String(), appToChannel.String(), *bDebug)
}
