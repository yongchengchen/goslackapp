package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/yongchengchen/goslackapp/app/service"
)

var sydneyTimeZone, _ = time.LoadLocation("Australia/Sydney")

var proxyToAPIEndPoint = ""
var proxyToAPIToken = ""
var slackChannelUrl = ""
var slackbotToken = ""

var gfCtx context.Context
var logger = g.Log()

func handleSlashCommand(cmd slack.SlashCommand) {
	// Handle the command here
	action := ""
	if cmd.Command == "/clock-in" || cmd.Command == "/clock-out" {
		if cmd.Text == "test" {
			logger.Infof(gfCtx, "Ignored test command %s %s\n", cmd.Command, cmd.Text)
			// return
		}
		action = strings.Replace(cmd.Command, "/", "", 1)
		action = strings.Replace(action, "-", "", 1)
	} else {
		logger.Infof(gfCtx, "Ignored command %s\n", cmd.Command)
		return
	}
	currentTime := time.Now().In(sydneyTimeZone).Format("2006-01-02 15:04:05")

	logger.Printf(gfCtx, "Tony debug receive command %s:%s,%s\n", currentTime, cmd.UserName, cmd.UserID)
	payload := fmt.Sprintf("{\"event\": \"%s\", \"slack_uid\": \"%s\"}", action, cmd.UserID)
	if !service.RequestHttpApi(proxyToAPIEndPoint, proxyToAPIToken, []byte(payload), true) {
		payload := fmt.Sprintf("{\"text\": \"%s:%s clock in `failed`, please try again\"}", currentTime, cmd.UserName)
		service.RequestHttpApi(slackChannelUrl, slackbotToken, []byte(payload), true)
		logger.Infof(gfCtx, "%s:%s,%s   --> failed \n", currentTime, cmd.UserName, cmd.UserID)
	}
}

func SlackServe(ctx context.Context, botToken string, appToken string, toApiEndPoint string, toApiToken string, channelUrl string, bDebug bool) {
	gfCtx = ctx
	slackbotToken = "Bearer " + botToken
	proxyToAPIEndPoint = toApiEndPoint
	proxyToAPIToken = toApiToken
	slackChannelUrl = channelUrl
	api := slack.New(botToken, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	client := socketmode.New(
		api,
		socketmode.OptionDebug(bDebug),
		socketmode.OptionLog(log.New(logger, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	defer watchRecover()

	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeInteractive:
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					logger.Infof(gfCtx, "Ignored %+v\n", evt)
					continue
				}

				if callback.Type == slack.InteractionTypeBlockActions {
					logger.Info(gfCtx, "Received block action")
					// Handle Block Actions
				}

			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					logger.Infof(gfCtx, "Ignored %+v\n", evt)
					continue
				}

				logger.Infof(gfCtx, "Received Slash Command: %+v\n", cmd)
				client.Ack(*evt.Request, map[string]interface{}{
					"text": fmt.Sprintf("Processing %s %s...", cmd.UserName, cmd.Command),
				})

				// Handle the slash command
				go handleSlashCommand(cmd)

			case socketmode.EventTypeConnecting:
				logger.Info(gfCtx, "Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnected:
				logger.Info(gfCtx, "Connected to Slack with Socket Mode.")
			case socketmode.EventTypeDisconnect:
				logger.Info(gfCtx, "Disconneted to Slack with Socket Mode. Restart app")
				panic("Slack Disconneted")
			default:
				logger.Infof(gfCtx, "Ignored %+v\n", evt)
			}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.RunContext(ctx)
}

func watchRecover() {
	if err := recover(); err != nil {
		logger.Infof(gfCtx, "App crashed with error: %v. Restarting...\n", err)
		restartApp()
	}
}

func restartApp() {
	executable, err := os.Executable()
	if err != nil {
		logger.Fatalf(gfCtx, "Failed to get executable: %v\n", err)
	}

	cmd := exec.Command(executable)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		logger.Fatalf(gfCtx, "Failed to restart app: %v\n", err)
	}

	logger.Infof(gfCtx, "Restarted app with PID %d\n", cmd.Process.Pid)
	os.Exit(0) // Terminate the current instance of the app
}
