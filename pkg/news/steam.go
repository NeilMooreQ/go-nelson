package news

import (
	"encoding/xml"
	"fmt"
	"go-nelson/pkg/structures"
	"go-nelson/pkg/utils"
	"log"
	"strings"
)

type SteamRSS struct {
	XMLName xml.Name     `xml:"rss"`
	Channel SteamChannel `xml:"channel"`
}

type SteamChannel struct {
	Title       string      `xml:"title"`
	Link        string      `xml:"link"`
	Description string      `xml:"description"`
	Items       []SteamItem `xml:"item"`
}

type SteamItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Enclosure   struct {
		URL string `xml:"url,attr"`
	} `xml:"enclosure"`
}

func ParseSteam() ([]structures.News, error) {
	log.Println("Парсинг новостей с Steam Developer")
	var news []structures.News
	fetcher := utils.NewFetcher()

	data, err := fetcher.Fetch("https://store.steampowered.com/feeds/news/group/4145017")
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении RSS ленты Steam Developer: %w", err)
	}

	reader := strings.NewReader(string(data))

	decoder := xml.NewDecoder(reader)

	var rss SteamRSS
	err = decoder.Decode(&rss)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе XML Steam Developer: %w", err)
	}

	for _, item := range rss.Channel.Items {
		if item.Title == "" || item.Link == "" {
			continue
		}

		_, err := utils.ParseRSSDate(item.PubDate)
		if err != nil {
			log.Printf("Ошибка при парсинге даты публикации Steam Developer: %v", err)
		}

		imageURL := item.Enclosure.URL
		if imageURL == "" {
			imageURL = "https://clan.fastly.steamstatic.com/images/4145017/8bfe522d8f2d91cd7dc3460771349a46ed8d6e95.jpg"
		}

		var images []string
		if imageURL != "" {
			images = append(images, imageURL)
		}

		uniqueID := getSteamID(item.GUID)
		content := utils.CleanHTML(item.Description)

		newsItem := structures.News{
			Provider:    "Steam Developer",
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

func getSteamID(guid string) string {
	if guid == "" {
		return "steam-" + utils.GenerateRandomString(8)
	}

	parts := strings.Split(guid, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		return "steam-" + lastPart
	}

	return "steam-" + strings.ReplaceAll(guid, "/", "-")
}
