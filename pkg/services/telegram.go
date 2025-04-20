package services

import (
	"context"
	"go-nelson/pkg"
	"log"
	"time"

	"github.com/go-telegram/bot"
)

var (
	telegramBot   *bot.Bot
	telegramQueue = make(chan string)
)

func StartTelegram() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var err error
	telegramBot, err = bot.New(pkg.Telegram.Token)
	if err != nil {
		log.Printf("Ошибка создания Telegram бота: %v", err)
		log.Println("Проверьте правильность токена, доступ к api.telegram.org и настройки сети")
		return
	}

	me, err := telegramBot.GetMe(ctx)
	if err != nil {
		log.Printf("Ошибка подключения к Telegram API: %v", err)
		log.Println("Проверьте правильность токена, доступ к api.telegram.org и настройки сети")
		return
	}
	log.Printf("Telegram бот успешно настроен: @%s [ID: %d]", me.Username, me.ID)

	go func() {
		log.Println("Telegram бот успешно запущен")
		telegramBot.Start(context.Background())
	}()

	go handleTelegramQueue()
}

func CloseTelegram() {
	if telegramBot != nil {
		log.Println("Telegram бот остановлен")
	}
}

func handleTelegramQueue() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case message := <-telegramQueue:
			sendToTelegram(message)
		default:
		}
	}
}

func SendTelegramMessage(message string) error {
	telegramQueue <- message
	return nil
}

func sendToTelegram(message string) {
	if telegramBot == nil || pkg.Telegram.ChannelID == "" {
		log.Println("Telegram бот не настроен или не указан ID канала")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := bot.SendMessageParams{
		ChatID:    pkg.Telegram.ChannelID,
		Text:      message,
		ParseMode: "HTML",
	}

	_, err := telegramBot.SendMessage(ctx, &params)
	if err != nil {
		log.Printf("Ошибка отправки сообщения в Telegram: %v", err)
	}
}
