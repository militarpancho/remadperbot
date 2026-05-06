package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"remadperbot/db"
	"remadperbot/pkg/models"
	scraper "remadperbot/pkg/scraper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const antiquityEndpoint = scraper.RemadDetailBaseURL

var (
	token         = os.Getenv("TOKEN")
	channel_id, _ = strconv.ParseInt(os.Getenv("CHANNEL_ID"), 0, 64)
)

type botClient struct {
	Api         *tgbotapi.BotAPI
	UpdatesChan tgbotapi.UpdatesChannel
	Db          db.Database
}

type callbackData struct {
	Id     string `json:"id"`
	Action string `json:"action"`
	Url    string
}

func NewTelegramBot(db db.Database) botClient {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return botClient{
		Api:         bot,
		UpdatesChan: bot.GetUpdatesChan(u),
		Db:          db,
	}
}

func (b *botClient) HandleUpdates() {
	for update := range b.UpdatesChan {
		if update.CallbackQuery != nil {
			var cb callbackData
			err := json.Unmarshal([]byte(update.CallbackData()), &cb)
			if err != nil {
				err = fmt.Errorf("Error unmarshalling callback data: %w", err)
				fmt.Println(err.Error())
			}
			cb.Url = articleURL(cb.Id)
			if cb.Action == "update" {
				b.refreshProductStatus(update, cb)
			} else if cb.Action == "notify" {
				b.insertItemUpdate(update, cb)
			}
		}
	}
}

func (b *botClient) Notify() {
	for true {
		item_updates, err := b.Db.GetAllItemUpdates()
		if err != nil {
			err = fmt.Errorf("error getting db record: %w", err)
			fmt.Println(err.Error())
		}
		for _, itemUpdate := range item_updates.ItemUpdates {
			var status string
			article_info := scraper.ExtractArticleInfo(articleURL(itemUpdate.ID), false)
			if article_info != nil {
				status = article_info.Status
			} else {
				status = "No disponible"
			}
			if status != itemUpdate.Status {
				article_info = scraper.ExtractArticleInfo(articleURL(itemUpdate.ID), true)
				users, err := b.Db.GetAllUsersByItemUpdate(itemUpdate.ID)
				if err != nil {
					err = fmt.Errorf("error getting db record: %w", err)
					fmt.Println(err.Error())
				}
				err = b.PostItemUpdate(article_info, &users)
				if err != nil {
					err = fmt.Errorf("error posting item update: %w", err)
					fmt.Println(err.Error())
				}
				itemUpdate.Status = status
				_, err = b.Db.UpdateItemUpdate(itemUpdate.ID, itemUpdate)
				if err != nil {
					err = fmt.Errorf("error updating item status: %w", err)
					fmt.Println(err.Error())
				}
				if itemUpdate.Status == "No disponible" {
					err := b.Db.DeleteUsersItemUpdate(itemUpdate.ID)
					if err != nil {
						err = fmt.Errorf("error removing db record: %w", err)
						fmt.Println(err.Error())
					}
					err = b.Db.DeleteItemUpdate(itemUpdate.ID)
					if err != nil {
						err = fmt.Errorf("error removing db record: %w", err)
						fmt.Println(err.Error())
					}
				}
			}
		}
		time.Sleep(3 * time.Second)
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
	msg.ReplyMarkup = numericKeyboard(articleID(articleInfo))
	message, err := b.Api.Send(msg)
	return message, err
}

func (b *botClient) PostItemUpdate(articleInfo *scraper.ArticleInfo, users *models.UserList) error {
	var file tgbotapi.FileBytes
	var caption string
	for _, user := range users.Users {
		user_id, _ := strconv.Atoi(user.ID)
		if articleInfo != nil {
			file = tgbotapi.FileBytes{
				Name:  "image.jpg",
				Bytes: articleInfo.Img,
			}
			caption = articleInfo.Title + "\nCambio en el estado del artículo: \n" + articleInfo.Metadata[3]
			msg := tgbotapi.NewPhoto(int64(user_id), file)
			msg.Caption = caption
			msg.ParseMode = "HTML"
			_, err := b.Api.Send(msg)
			if err != nil {
				return err
			}
		} else {
			caption = "\nCambio en el estado de un artículo al que seguias a: \n Estado: <b>No disponible</b>"
			msg := tgbotapi.NewMessage(int64(user_id), caption)
			msg.ParseMode = "HTML"
			_, err := b.Api.Send(msg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *botClient) refreshProductStatus(update tgbotapi.Update, cb callbackData) {
	articleInfo := scraper.ExtractArticleInfo(cb.Url, false)
	var editMessage tgbotapi.EditMessageCaptionConfig
	if articleInfo != nil {
		currentTime := time.Now()
		editMessage = tgbotapi.NewEditMessageCaption(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			articleInfo.Title+"\n"+strings.Join(articleInfo.Metadata[:], "\n")+"\n"+fmt.Sprintf("Actualizado a las %s", currentTime.Format("15:04:05 2006-01-02")),
		)
		editMessage.ReplyMarkup = numericKeyboard(articleID(articleInfo))
	} else {
		editMessage = tgbotapi.NewEditMessageCaption(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			fmt.Sprintf("<a href=\"%s\">%s</a>", cb.Url, "Producto no disponible"),
		)
	}
	editMessage.ParseMode = "HTML"
	_, err := b.Api.Send(editMessage)
	if err != nil {
		err = fmt.Errorf("error refreshing status: %w", err)
		fmt.Println(err.Error())
	}
}

func (b *botClient) insertItemUpdate(update tgbotapi.Update, cb callbackData) {
	articleInfo := scraper.ExtractArticleInfo(cb.Url, false)
	if articleInfo == nil {
		return
	}
	status := articleInfo.Status
	err := b.Db.AddUsersItemUpdate(cb.Id, status, fmt.Sprint(update.SentFrom().ID))
	if err != nil {
		err = fmt.Errorf("error inserting db record: %w", err)
		fmt.Println(err.Error())
		callback := tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "Ya estás suscrito a las alertas de este producto")
		if _, err = b.Api.Request(callback); err != nil {
			err = fmt.Errorf("error sending callback: %w", err)
			fmt.Println(err.Error())
		}
	} else {
		msg := tgbotapi.NewMessage(update.SentFrom().ID, fmt.Sprintf("🔔🔔🔔 Te has suscrito a las alertas del artículo %s. Recibiriras una alerta cuando el artículo cambie de estado.", articleInfo.Title))
		msg.ParseMode = "HTML"
		_, err = b.Api.Send(msg)
		var alert_message string
		if err != nil {
			alert_message = "Para poder suscribirte, empieza una conversacion con el bot"
		} else {
			alert_message = "Te has suscrito a las alertas de este producto"
		}
		callback := tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, alert_message)
		if _, err = b.Api.Request(callback); err != nil {
			err = fmt.Errorf("error sending callback: %w", err)
			fmt.Println(err.Error())
		}
	}
}

func numericKeyboard(id string) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Actualizar Estado", fmt.Sprintf("{\"id\": \"%s\", \"action\":\"update\"}", id)),
			tgbotapi.NewInlineKeyboardButtonData("👀 Informarme de Cambios", fmt.Sprintf("{\"id\": \"%s\", \"action\":\"notify\"}", id)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🤖 Abrir Alertas Remad Bot", "https://t.me/remadperbot"),
		),
	)
	return &keyboard
}

func articleURL(id string) string {
	return antiquityEndpoint + id
}

func articleID(articleInfo *scraper.ArticleInfo) string {
	if articleInfo.ID != "" {
		return articleInfo.ID
	}
	splitURL := strings.Split(articleInfo.Url, "/")
	return splitURL[len(splitURL)-1]
}
