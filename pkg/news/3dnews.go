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

type ThreeDNewsRSS struct {
	XMLName xml.Name          `xml:"rss"`
	Channel ThreeDNewsChannel `xml:"channel"`
}

type ThreeDNewsChannel struct {
	Title       string           `xml:"title"`
	Link        string           `xml:"link"`
	Description string           `xml:"description"`
	PubDate     string           `xml:"pubDate"`
	Items       []ThreeDNewsItem `xml:"item"`
}

type ThreeDNewsItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Category    string `xml:"category"`
	Enclosure   struct {
		URL    string `xml:"url,attr"`
		Length string `xml:"length,attr"`
		Type   string `xml:"type,attr"`
	} `xml:"enclosure"`
}

func Parse3DNews() ([]structures.News, error) {
	log.Println("Парсинг новостей с 3DNews")
	var news []structures.News
	fetcher := utils.NewFetcher()

	data, err := fetcher.Fetch("https://3dnews.ru/news/rss/")
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе RSS-фида 3DNews: %v", err)
	}

	var rss ThreeDNewsRSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге XML 3DNews: %v", err)
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

		hash := md5.Sum([]byte(item.Link))
		id := hex.EncodeToString(hash[:])

		imageURL := item.Enclosure.URL
		var images []string
		if imageURL != "" {
			images = append(images, imageURL)
		}

		category := determine3DNewsCategory(item.Category)
		var tags []string
		if category != "" {
			tags = append(tags, category)
		}

		newsItem := structures.News{
			Provider:    "3DNews",
			UniqueID:    id,
			Title:       item.Title,
			Description: content,
			URL:         item.Link,
			Images:      images,
			Tags:        tags,
		}
		news = append(news, newsItem)
	}

	return news, nil
}

func determine3DNewsCategory(categoryText string) string {
	categoryText = strings.ToLower(categoryText)

	if strings.Contains(categoryText, "игры") ||
		strings.Contains(categoryText, "gamesblender") ||
		strings.Contains(categoryText, "игровые консоли") {
		return "Игры"
	}

	if strings.Contains(categoryText, "видеокарты") ||
		strings.Contains(categoryText, "жесткие диски") ||
		strings.Contains(categoryText, "мониторы") ||
		strings.Contains(categoryText, "ноутбуки") ||
		strings.Contains(categoryText, "пк") ||
		strings.Contains(categoryText, "серверы") ||
		strings.Contains(categoryText, "корпуса") ||
		strings.Contains(categoryText, "аудио") ||
		strings.Contains(categoryText, "видео периферия") ||
		strings.Contains(categoryText, "мобильные телефоны") ||
		strings.Contains(categoryText, "смартфоны") ||
		strings.Contains(categoryText, "планшетные компьютеры") ||
		strings.Contains(categoryText, "разгон и замеры производительности") {
		return "Железо"
	}

	if strings.Contains(categoryText, "по ") ||
		strings.Contains(categoryText, "драйверы") ||
		strings.Contains(categoryText, "искусственный интеллект") ||
		strings.Contains(categoryText, "новости сети") ||
		strings.Contains(categoryText, "im-клиенты") ||
		strings.Contains(categoryText, "нанотехнологии") ||
		strings.Contains(categoryText, "на острие науки") ||
		strings.Contains(categoryText, "космос") ||
		strings.Contains(categoryText, "мир роботехники") ||
		strings.Contains(categoryText, "финансовые новости") {
		return "Софт"
	}

	return ""
}
