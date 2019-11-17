package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rizkyduut/released_bot"
	"github.com/rizkyduut/released_bot/dbadapter"
	"github.com/subosito/gotenv"
	"log"
	"os"
)

const (
	cmdLatest  = "latest"
	cmdDeploy  = "deploy"
	cmdGroup   = "group"
	cmdService = "service"
	cmdHelp    = "help"
)

func main() {
	gotenv.Load()
	redisServer := os.Getenv("REDIS_SERVER")
	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	redisConfig := &dbadapter.Config{
		Host:     redisServer,
		Password: "",
	}
	redisAdapter := dbadapter.NewRedisAdapter(redisConfig, "released_bot")
	rb := releasedbot.New(redisAdapter)

	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		var botHandler releasedbot.Handler
		botData := &releasedbot.BotData{
			Sender:           update.Message.From.UserName,
			RawMessage:       update.Message.Text,
			Command:          update.Message.Command(),
			CommandArguments: update.Message.CommandArguments(),
		}
		switch botData.Command {
		case cmdHelp:
			botHandler = rb.HelpHandler
		case cmdLatest:
			botHandler = rb.LatestHandler
		case cmdDeploy:
			botHandler = rb.DeployHandler
		case cmdService:
			botHandler = rb.ServiceHandler
		case cmdGroup:
			botHandler = rb.GroupHandler
		default:
			botHandler = rb.DefaultHandler
		}

		response, err := botHandler(botData)
		if err != nil {
			response = fmt.Sprintf("Error: %s", err.Error())
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyToMessageID = update.Message.MessageID
		_, err = bot.Send(msg)
		if err != nil {

		}
	}
}
