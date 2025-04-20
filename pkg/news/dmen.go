package news

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"go-nelson/pkg/structures"
	"go-nelson/pkg/utils"
	"log"
)

type DMenRSS struct {
	XMLName xml.Name    `xml:"rss"`
	Channel DMenChannel `xml:"channel"`
}

type DMenChannel struct {
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Description string     `xml:"description"`
	Items       []DMenItem `xml:"item"`
}

type DMenItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Creator     string `xml:"creator"`
	Enclosure   struct {
		URL  string `xml:"url,attr"`
		Type string `xml:"type,attr"`
	} `xml:"enclosure"`
}

func ParseDMen() ([]structures.News, error) {
	log.Println("Парсинг новостей с DisgustingMen")
	var news []structures.News
	fetcher := utils.NewFetcher()

	data, err := fetcher.Fetch("https://disgustingmen.com/feed/")
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе RSS-фида DisgustingMen: %v", err)
	}

	var rss DMenRSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге XML DisgustingMen: %v", err)
	}

	for _, item := range rss.Channel.Items {
		if item.Title == "" || item.Link == "" {
			continue
		}

		description := preprocessDMenDescription(item.Description)
		content := utils.CleanHTML(description)

		_, err := utils.ParseRSSDate(item.PubDate)
		if err != nil {
			log.Printf("Ошибка при парсинге даты публикации: %v", err)
		}

		hash := md5.Sum([]byte(item.Link))
		id := hex.EncodeToString(hash[:])

		imageURL := item.Enclosure.URL
		var images []string
		if imageURL != "" {
			images = append(images, imageURL)
		}

		newsItem := structures.News{
			Provider:    "DisgustingMen",
			UniqueID:    id,
			Title:       item.Title,
			Description: content,
			URL:         item.Link,
			Images:      images,
		}
		news = append(news, newsItem)
	}

	return news, nil
}

func preprocessDMenDescription(description string) string {
	description = utils.ReplaceHTMLEntity(description, "[&#8230;]", "...")

	secondPStart := utils.FindNthOccurrence(description, "<p>", 2)
	if secondPStart != -1 {
		description = description[:secondPStart]
	}

	return description
}
