package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	scraper "remadperbot/pkg/scraper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	token         = os.Getenv("TOKEN")
	channel_id, _ = strconv.ParseInt(os.Getenv("CHANNEL_ID"), 0, 64)
)

type botClient struct {
	Api         *tgbotapi.BotAPI
	UpdatesChan tgbotapi.UpdatesChannel
}

func NewTelegramBot() botClient {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return botClient{
		Api:         bot,
		UpdatesChan: bot.GetUpdatesChan(u),
	}
}

func (b *botClient) HandleUpdates() {
	for update := range b.UpdatesChan {
		if update.CallbackQuery != nil {
			b.refreshProductStatus(update)
		}
	}
}

func (b *botClient) PostNewArticle(articleInfo *scraper.ArticleInfo) (tgbotapi.Message, error) {
	file := tgbotapi.FileBytes{
		Name:  "image.jpg",
		Bytes: articleInfo.Img,
	}
	msg := tgbotapi.NewPhoto(int64(channel_id), file)
	msg.Caption = articleInfo.Title + "\n" + strings.Join(articleInfo.Metadata[:], "\n")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = numericKeyboard(articleInfo.Url)
	return b.Api.Send(msg)
}

func (b *botClient) refreshProductStatus(update tgbotapi.Update) {
	articleInfo := scraper.ExtractArticleInfo(update.CallbackData(), false)
	var editMessage tgbotapi.EditMessageCaptionConfig
	if articleInfo != nil {
		editMessage = tgbotapi.NewEditMessageCaption(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			articleInfo.Title+"\n"+strings.Join(articleInfo.Metadata[:], "\n"),
		)
	} else {
		editMessage = tgbotapi.NewEditMessageCaption(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			fmt.Sprintf("<a href=\"%s\">%s</a>", update.CallbackData(), "Producto no disponible"),
		)
	}
	editMessage.ParseMode = "HTML"
	editMessage.ReplyMarkup = numericKeyboard(update.CallbackData())
	b.Api.Send(editMessage)
}

func numericKeyboard(url string) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Actualizar Estado", url),
		),
	)
	return &keyboard
}
