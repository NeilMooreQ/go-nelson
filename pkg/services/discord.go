package services

import (
	"bytes"
	"fmt"
	"go-nelson/pkg"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"image/jpeg"

	"go-nelson/pkg/structures"
	"go-nelson/pkg/utils"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/image/webp"
)

var discordSession *discordgo.Session
var discordNewsChannel = make(chan structures.News, 500)
var forumTagsCache map[string]string

var requiredTags = []string{
	"EGS",
	"DTF",
	"Ixbt",
	"Steam",
	"Disgusting",
	"3Dnews",
	"GameDev",
	"StopGame",
}

func StartDiscord() {
	log.Println("Запуск Discord сервиса")
	var err error

	discordSession, err = discordgo.New("Bot " + pkg.Discord.Token)
	if err != nil {
		log.Printf("Ошибка при создании Discord сессии: %v", err)
		return
	}

	err = discordSession.Open()
	if err != nil {
		log.Printf("Ошибка при подключении к Discord: %v", err)
		return
	}

	initForumTags()

	go handleDiscordQueue()
	log.Println("Discord сервис успешно запущен")
}

func initForumTags() {
	log.Println("Инициализация тегов форума Discord")
	forumTagsCache = make(map[string]string)

	forumChannel, err := discordSession.Channel(pkg.Discord.NewsForumId)
	if err != nil {
		log.Printf("Ошибка при получении информации о форуме: %v", err)
		return
	}

	if forumChannel.Type != discordgo.ChannelTypeGuildForum {
		log.Printf("Канал %s не является форумом", pkg.Discord.NewsForumId)
		return
	}

	for _, tag := range forumChannel.AvailableTags {
		forumTagsCache[strings.ToLower(tag.Name)] = tag.ID
	}

	for _, tagName := range requiredTags {
		if _, exists := forumTagsCache[strings.ToLower(tagName)]; !exists {
			createForumTag(tagName)
		}
	}
}

func createForumTag(tagName string) {
	log.Printf("Создание нового тега '%s' для форума Discord", tagName)

	forumChannel, err := discordSession.Channel(pkg.Discord.NewsForumId)
	if err != nil {
		log.Printf("Ошибка при получении информации о форуме для создания тега: %v", err)
		return
	}

	newTag := discordgo.ForumTag{
		Name: tagName,
	}

	updatedTags := append(forumChannel.AvailableTags, newTag)

	channelEdit := &discordgo.ChannelEdit{
		AvailableTags: &updatedTags,
	}

	updatedForum, err := discordSession.ChannelEdit(pkg.Discord.NewsForumId, channelEdit)
	if err != nil {
		log.Printf("Ошибка при создании тега '%s': %v", tagName, err)
		return
	}

	for _, tag := range updatedForum.AvailableTags {
		if strings.EqualFold(tag.Name, tagName) {
			forumTagsCache[strings.ToLower(tagName)] = tag.ID
			break
		}
	}
}

func CloseDiscord() {
	log.Println("Закрытие соединения с Discord")
	if discordSession != nil {
		discordSession.Close()
	}
}

func handleDiscordQueue() {
	log.Println("Запуск обработчика очереди сообщений Discord")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case news := <-discordNewsChannel:
			err := sendToDiscordWithRateLimiting(news)
			if err != nil {
				log.Printf("Ошибка при отправке новости в Discord: %v", err)
				if strings.Contains(err.Error(), "rate limit") {
					time.Sleep(10 * time.Second)
					discordNewsChannel <- news
				}
			}

			time.Sleep(1 * time.Second)
		default:
		}
	}
}

func SendNewsToThread(news structures.News) error {

	select {
	case discordNewsChannel <- news:
		return nil
	default:
		log.Printf("ОШИБКА: Канал Discord переполнен, новость отброшена: %s", news.Title)
		return fmt.Errorf("канал discord переполнен")
	}
}

func getTagForProvider(provider string) string {
	var tagName string

	switch provider {
	case "DTF":
		tagName = "DTF"
	case "IxbtGames":
		tagName = "Ixbt"
	case "Steam":
		tagName = "Steam"
	case "Disgustingmen":
		tagName = "Disgusting"
	case "3DNews":
		tagName = "3Dnews"
	case "Epic Games Store":
		tagName = "EGS"
	case "GameDev.ru":
		tagName = "GameDev"
	case "StopGame":
		tagName = "StopGame"
	default:
		return ""
	}

	return forumTagsCache[strings.ToLower(tagName)]
}

func sendToDiscordWithRateLimiting(news structures.News) error {
	if discordSession == nil {
		return fmt.Errorf("discord бот не настроен")
	}

	if pkg.Discord.NewsForumId == "" {
		return fmt.Errorf("не указан ID канала форума для новостей")
	}

	title := formatTitle(news.Title)

	threadParams := &discordgo.ThreadStart{
		Name:                title,
		AutoArchiveDuration: 10080,
	}

	tagID := getTagForProvider(news.Provider)
	if tagID != "" {
		threadParams.AppliedTags = []string{tagID}
	}

	description := makeDescription(news.URL, news.Description, news.Tags, news.Title, news.Provider, len(news.Images) > 0)

	messageSend := &discordgo.MessageSend{
		Content: description,
	}

	if len(news.Images) > 0 {
		processAndAttachImage(news.Images[0], messageSend)
	}

	_, err := discordSession.ForumThreadStartComplex(pkg.Discord.NewsForumId, threadParams, messageSend)
	if err != nil {
		log.Printf("Ошибка создания треда для новости '%s': %v", news.Title, err)
		return err
	}

	return nil
}

func downloadImage(url string) ([]byte, string, string, error) {
	fetcher := utils.NewFetcher()

	imageData, err := fetcher.Fetch(url)
	if err != nil {
		return nil, "", "", err
	}

	contentType := http.DetectContentType(imageData)

	fileName := filepath.Base(url)
	if fileName == "" || fileName == "." || strings.Contains(fileName, "?") {
		ext := getExtensionFromContentType(contentType)
		fileName = fmt.Sprintf("image%s", ext)
	}

	return imageData, contentType, fileName, nil
}

func convertWebpToJpg(webpData []byte) ([]byte, string, string, error) {
	webpImg, err := webp.Decode(bytes.NewReader(webpData))
	if err != nil {
		return nil, "", "", err
	}

	jpgBuf := new(bytes.Buffer)

	err = jpeg.Encode(jpgBuf, webpImg, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, "", "", err
	}

	jpgData := jpgBuf.Bytes()
	fileName := "image.jpg"

	return jpgData, "image/jpeg", fileName, nil
}

func processAndAttachImage(imageURL string, messageData *discordgo.MessageSend) {
	imageData, contentType, fileName, err := downloadImage(imageURL)
	if err != nil {
		log.Printf("Ошибка при скачивании изображения: %v", err)
		return
	}

	if len(imageData) > 25*1024*1024 {
		return
	}

	if strings.Contains(contentType, "webp") {
		convertedData, newContentType, newFileName, err := convertWebpToJpg(imageData)
		if err != nil {
			log.Printf("Ошибка при конвертации webp в jpg: %v", err)
		} else {
			imageData = convertedData
			contentType = newContentType
			fileName = newFileName
		}
	}

	if strings.Contains(fileName, "?") {
		ext := getExtensionFromContentType(contentType)
		fileName = fmt.Sprintf("image%s", ext)
	}

	if len(imageData) < 25*1024*1024 {
		messageData.Files = []*discordgo.File{
			{
				Name:   fileName,
				Reader: bytes.NewReader(imageData),
			},
		}
	}
}

func getExtensionFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg"):
		return ".jpg"
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	default:
		return ".jpg"
	}
}

func formatTitle(title string) string {
	if len(title) > 95 {
		return title[:95] + "..."
	}
	return title
}

func makeDescription(url, description string, tags []string, title, provider string, withImage bool) string {
	var result strings.Builder

	if len(tags) > 0 {
		result.WriteString(fmt.Sprintf("**Теги**: %s\n\n", strings.Join(tags, ", ")))
	}

	if len(title) > 95 {
		result.WriteString(fmt.Sprintf("# %s\n\n", title))
	}

	if len(description) > 3800 {
		description = description[:3800] + "..."
	}
	result.WriteString(description)

	if withImage {
		result.WriteString(fmt.Sprintf("\n\n[Подробнее](<%s>) `🔒 %s`", url, provider))
	} else {
		result.WriteString(fmt.Sprintf("\n\n[Подробнее](%s) `🔒 %s`", url, provider))
	}

	return result.String()
}
