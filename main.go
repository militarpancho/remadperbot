package main

import (
	"fmt"
	"remadperbot/pkg/bot"
	"remadperbot/pkg/scraper"
)

func main() {
	botClient := bot.NewTelegramBot()
	var current_id int

	//for true {
	url, last_id := scraper.FindLastObject()
	if last_id != current_id {
		fmt.Printf("New Product found: %s", url)
		article_info := scraper.ExtractArticleInfo(url)
		botClient.PostNewArticle(article_info)
	}

	//}

}
