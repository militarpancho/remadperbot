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
	var current_id int
	go botClient.HandleUpdates()
	go botClient.Notify()
	for true {
		if miscelanea.CheckOpenGreenPoints() {
			endpoint, last_id := scraper.FindLastObject()
			if last_id != current_id {
				if current_id != 0 {
					for i := current_id + 1; i <= last_id; i++ {
						url := endpoint + "/" + fmt.Sprint(i)
						article_info := scraper.ExtractArticleInfo(url, true)
						if article_info != nil {
							log.Printf("New Product found: %s", url)
							_, err := botClient.PostNewArticle(article_info)
							if err != nil {
								err = fmt.Errorf("Error posting new article: %w", err)
								fmt.Println(err.Error())
							}
						}
					}
				}
				current_id = last_id
			}
		}
		time.Sleep(scraperInterval * time.Second)
	}

}
