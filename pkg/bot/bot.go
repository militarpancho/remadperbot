package bot

import (
	"log"
	"os"
	"strconv"
	"strings"

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
	msg.Caption = articleInfo.Title + "\n" + strings.Join(articleInfo.Metadata[:], "\n")
	msg.ParseMode = "HTML"
	return b.api.Send(msg)
}
