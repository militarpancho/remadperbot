package bot

import (
	"log"
	"os"
	"strconv"

	scraper "remadperbot/pkg/scraper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	token         = os.Getenv("TOKEN")
	channel_id, _ = strconv.Atoi(os.Getenv("CHANNEL_ID_TEST"))
)

type botClient struct {
	api *tgbotapi.BotAPI
}

func NewTelegramBot() botClient {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	return botClient{
		api: bot,
	}
}

func (b *botClient) PostNewArticle(articleInfo scraper.ArticleInfo) (tgbotapi.Message, error) {
	file := tgbotapi.FileBytes{
		Name:  "image.jpg",
		Bytes: articleInfo.Img,
	}
	msg := tgbotapi.NewPhoto(int64(channel_id), file)

	return b.api.Send(msg)
}

// log.Printf("Authorized on account %s", bot.Self.UserName)

// u := tgbotapi.NewUpdate(0)
// u.Timeout = 60

// updates := bot.GetUpdatesChan(u)

// for update := range updates {
// 	if update.Message != nil { // If we got a message
// 		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
// 		msg.ReplyToMessageID = update.Message.MessageID

// 		bot.Send(msg)
// 	}
// }
