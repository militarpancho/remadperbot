package main

import (
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

	for true {
		if miscelanea.CheckOpenGreenPoints() {
			url, last_id := scraper.FindLastObject()
			if last_id != current_id {
				log.Printf("New Product found: %s", url)
				article_info := scraper.ExtractArticleInfo(url)
				botClient.PostNewArticle(article_info)
				current_id = last_id
			}
			time.Sleep(scraperInterval * time.Second)
		} else {
			secondsUntil3pm := miscelanea.SecondsICanSleep()
			if 0 < secondsUntil3pm && secondsUntil3pm < sleepTime {
				log.Printf("[%s] Remad is going to open in: %d \n", time.Now().Format(time.RFC850), secondsUntil3pm)
				time.Sleep(time.Duration(secondsUntil3pm) * time.Second)
			} else {
				time.Sleep(sleepTime * time.Second)
			}
		}
	}

}
