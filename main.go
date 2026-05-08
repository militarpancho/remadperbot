package main

import (
	"fmt"
	"log"
	"os"
	"remadperbot/db"
	"remadperbot/pkg/bot"
	"remadperbot/pkg/miscelanea"
	"remadperbot/pkg/scraper"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const sleepTime = 1800 // 30 minutes
const scraperInterval = 2

var (
	db_user     = os.Getenv("POSTGRES_USER")
	db_password = os.Getenv("POSTGRES_PASSWORD")
	db_name     = os.Getenv("POSTGRES_DB")
)

func main() {
	db, err := db.Initialize(db_user, db_password, db_name)
	if err != nil {
		err = fmt.Errorf("Cannot connect with db: %w", err)
		fmt.Println(err.Error())
		os.Exit(1)
	}
	botClient := bot.NewTelegramBot(db)
	scraperClient := scraper.NewClient()
	seenProducts := map[string]bool{}
	catalogSeeded := false
	go botClient.HandleUpdates()
	go botClient.Notify()
	postStartupDebugArticle(&botClient, scraperClient)
	for true {
		if miscelanea.CheckOpenGreenPoints() {
			if !catalogSeeded {
				if err := seedSeenProducts(seenProducts, scraperClient); err != nil {
					err = fmt.Errorf("Error seeding current catalog: %w", err)
					fmt.Println(err.Error())
				} else {
					catalogSeeded = true
				}
			} else {
				articleInfos, err := scraperClient.ArticleInfosUntilKnown(seenProducts, true)
				if err != nil {
					err = fmt.Errorf("Error finding new articles: %w", err)
					fmt.Println(err.Error())
				}
				for _, articleInfo := range articleInfos {
					log.Printf("New Product found: %s", articleInfo.Url)
					_, err := botClient.PostNewArticle(articleInfo)
					if err != nil {
						err = fmt.Errorf("Error posting new article: %w", err)
						fmt.Println(err.Error())
						break
					}
					seenProducts[articleInfo.ID] = true
				}
			}
		}
		time.Sleep(scraperInterval * time.Second)
	}
}

type articlePoster interface {
	PostNewArticle(*scraper.ArticleInfo) (tgbotapi.Message, error)
}

func postStartupDebugArticle(botClient articlePoster, scraperClient scraper.Client) {
	articleInfo, err := scraperClient.LatestArticleInfo(true)
	if err != nil {
		err = fmt.Errorf("Error finding startup debug article: %w", err)
		fmt.Println(err.Error())
		return
	}
	log.Printf("Startup debug product found: %s", articleInfo.Url)
	_, err = botClient.PostNewArticle(articleInfo)
	if err != nil {
		err = fmt.Errorf("Error posting startup debug article: %w", err)
		fmt.Println(err.Error())
	}
}

func seedSeenProducts(seenProducts map[string]bool, scraperClient scraper.Client) error {
	antiques, err := scraperClient.CatalogPage(0)
	if err != nil {
		return err
	}
	for _, antiquity := range antiques {
		seenProducts[antiquity.Hash] = true
	}
	return nil
}
