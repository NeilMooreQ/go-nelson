package news

import (
	"encoding/xml"
	"fmt"
	"go-nelson/pkg/structures"
	"go-nelson/pkg/utils"
	"log"
	"strings"
)

type StopGameRSS struct {
	XMLName xml.Name        `xml:"rss"`
	Channel StopGameChannel `xml:"channel"`
}

type StopGameChannel struct {
	Title       string         `xml:"title"`
	Link        string         `xml:"link"`
	Description string         `xml:"description"`
	Language    string         `xml:"language"`
	PubDate     string         `xml:"pubDate"`
	Image       StopGameImage  `xml:"image"`
	Items       []StopGameItem `xml:"item"`
}

type StopGameImage struct {
	URL   string `xml:"url"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type StopGameItem struct {
	Title       string            `xml:"title"`
	Description string            `xml:"description"`
	Link        string            `xml:"link"`
	Comments    string            `xml:"comments"`
	Enclosure   StopGameEnclosure `xml:"enclosure"`
	GUID        string            `xml:"guid"`
	PubDate     string            `xml:"pubDate"`
}

type StopGameEnclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func ParseStopGame() ([]structures.News, error) {
	log.Println("Парсинг новостей с StopGame")
	var news []structures.News
	fetcher := utils.NewFetcher()

	data, err := fetcher.Fetch("https://rss.stopgame.ru/rss_all.xml")
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении RSS-ленты StopGame: %v", err)
	}

	var rss StopGameRSS
	err = xml.Unmarshal(data, &rss)
	if err != nil {
		return nil, fmt.Errorf("ошибка при декодировании XML RSS-ленты StopGame: %v", err)
	}

	for _, item := range rss.Channel.Items {
		if item.Title == "" || item.Link == "" {
			continue
		}

		description := strings.ReplaceAll(item.Description, "… […]", "")
		if description == "" {
			continue
		}

		_, err := utils.ParseRSSDate(item.PubDate)
		if err != nil {
			log.Printf("Ошибка при парсинге даты публикации новости StopGame: %v", err)
		}

		imageURL := item.Enclosure.URL
		var images []string
		if imageURL != "" {
			images = append(images, imageURL)
		}

		uniqueID := getStopGameID(item.Link)
		title := utils.CleanHTML(item.Title)
		content := utils.CleanHTML(description)

		newsItem := structures.News{
			Provider:    "StopGame",
			UniqueID:    uniqueID,
			Title:       title,
			Description: content,
			URL:         item.Link,
			Images:      images,
		}
		news = append(news, newsItem)
	}

	return news, nil
}

func getStopGameID(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) >= 5 {
		return parts[4]
	}
	return url
}
