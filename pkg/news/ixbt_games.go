package news

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go-nelson/pkg/structures"
	"go-nelson/pkg/utils"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ParseIXBTGames() ([]structures.News, error) {
	log.Println("Парсинг новостей с Ixbt Games")
	var news []structures.News
	fetcher := utils.NewFetcher()

	data, err := fetcher.Fetch("https://ixbt.games/news/")
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе страницы Ixbt Games: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге HTML: %v", err)
	}

	doc.Find("div.row.no-gutters").Each(func(i int, s *goquery.Selection) {
		titleElem := s.Find("div.card-title a")
		if titleElem.Length() == 0 {
			return
		}

		url, exists := titleElem.Attr("href")
		if !exists || strings.TrimSpace(url) == "" {
			return
		}

		url = "https://ixbt.games" + url
		title := strings.TrimSpace(titleElem.Text())
		if title == "" {
			return
		}

		descElem := s.Find("div.d-flex.d-sm-block.my-2")
		description := strings.TrimSpace(descElem.Text())

		imgElem := s.Find("img")
		imageURL, hasImage := imgElem.Attr("src")
		var images []string
		if hasImage && imageURL != "" {
			images = append(images, imageURL)
		}

		hash := md5.Sum([]byte(url))
		id := hex.EncodeToString(hash[:])

		var tags []string
		tags = append(tags, "games", "gaming")

		newsItem := structures.News{
			Provider:    "Ixbt Games",
			UniqueID:    id,
			Title:       title,
			Description: description,
			URL:         url,
			Images:      images,
			Tags:        tags,
		}

		news = append(news, newsItem)
	})

	return news, nil
}
