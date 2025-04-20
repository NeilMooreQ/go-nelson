package news

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"go-nelson/pkg/structures"
	"go-nelson/pkg/utils"
	"log"
	"strings"
)

type DTFRSS struct {
	XMLName xml.Name   `xml:"rss"`
	Channel DTFChannel `xml:"channel"`
}

type DTFChannel struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Link        string    `xml:"link"`
	Image       DTFImage  `xml:"image"`
	Items       []DTFItem `xml:"item"`
	PubDate     string    `xml:"pubDate"`
}

type DTFImage struct {
	URL   string `xml:"url"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type DTFItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Creator     string `xml:"creator"`
	PubDate     string `xml:"pubDate"`
	Comments    string `xml:"comments"`
	Author      string `xml:"author"`
	Enclosure   struct {
		URL    string `xml:"url,attr"`
		Length string `xml:"length,attr"`
		Type   string `xml:"type,attr"`
	} `xml:"enclosure"`
}

func ParseDTF() ([]structures.News, error) {
	log.Println("Парсинг новостей с DTF")
	var news []structures.News
	fetcher := utils.NewFetcher()

	data, err := fetcher.Fetch("https://dtf.ru/rss")
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе RSS-фида DTF: %v", err)
	}

	var rss DTFRSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге XML DTF: %v", err)
	}

	for _, item := range rss.Channel.Items {
		if item.Title == "" || item.Link == "" {
			continue
		}

		content := utils.CleanHTML(item.Description)

		_, err := utils.ParseRSSDate(item.PubDate)
		if err != nil {
			log.Printf("Ошибка при парсинге даты публикации: %v", err)
		}

		idStr := item.Link
		if item.GUID != "" {
			idStr = item.GUID
		}
		hash := md5.Sum([]byte(idStr))
		id := hex.EncodeToString(hash[:])

		var images []string
		if item.Enclosure.URL != "" && strings.Contains(item.Enclosure.Type, "image") {
			images = append(images, item.Enclosure.URL)
		}

		newsItem := structures.News{
			Provider:    "DTF",
			UniqueID:    id,
			Title:       utils.CleanCDATA(item.Title),
			Description: content,
			URL:         item.Link,
			Images:      images,
		}
		news = append(news, newsItem)
	}

	return news, nil
}
