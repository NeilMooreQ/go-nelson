package news

import (
	"encoding/xml"
	"fmt"
	"go-nelson/pkg/structures"
	"go-nelson/pkg/utils"
	"log"
	"strings"

	"golang.org/x/net/html/charset"
)

type GameDevRSS struct {
	XMLName xml.Name       `xml:"rss"`
	Channel GameDevChannel `xml:"channel"`
}

type GameDevChannel struct {
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	Description string        `xml:"description"`
	Items       []GameDevItem `xml:"item"`
}

type GameDevItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Category    string `xml:"category"`
	GUID        string `xml:"guid"`
}

func ParseGameDev() ([]structures.News, error) {
	log.Println("Парсинг новостей с GameDev.ru")
	var news []structures.News
	fetcher := utils.NewFetcher()

	data, err := fetcher.Fetch("https://gamedev.ru/rss")
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении RSS ленты GameDev: %w", err)
	}

	reader := strings.NewReader(string(data))

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	var rss GameDevRSS
	err = decoder.Decode(&rss)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе XML GameDev: %w", err)
	}

	for _, item := range rss.Channel.Items {
		if item.Title == "" || item.Link == "" {
			continue
		}

		_, err := utils.ParseRSSDate(item.PubDate)
		if err != nil {
			log.Printf("Ошибка при парсинге даты публикации GameDev: %v", err)
		}

		imageURL := utils.ExtractImageURL(item.Description)
		var images []string
		if imageURL != "" {
			images = append(images, imageURL)
		}

		uniqueID := getGameDevID(item.Link)
		content := utils.CleanHTML(item.Description)

		newsItem := structures.News{
			Provider:    "GameDev.ru",
			UniqueID:    uniqueID,
			Title:       item.Title,
			Description: content,
			URL:         item.Link,
			Images:      images,
		}
		news = append(news, newsItem)
	}

	return news, nil
}

func getGameDevID(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		idParts := strings.Split(lastPart, ".")
		if len(idParts) > 0 {
			return "gamedev-" + idParts[0]
		}
	}
	return "gamedev-" + strings.ReplaceAll(url, "/", "-")
}
