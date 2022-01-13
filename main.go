package main

import (
	"fmt"
	"log"
	"remadperbot/pkg/bot"
	"remadperbot/pkg/miscelanea"
	"remadperbot/pkg/scraper"
	"time"
)

const sleepTime = 1800 // 30 minutes
const scraperInterval = 2

func main() {
	botClient := bot.NewTelegramBot()
	var current_id int
	go botClient.HandleUpdates()
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
							botClient.PostNewArticle(article_info)
						}
					}
				}
				current_id = last_id
			}
		}
		time.Sleep(scraperInterval * time.Second)
	}

}
