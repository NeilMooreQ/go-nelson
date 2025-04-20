package services

import (
	"go-nelson/pkg/structures"
	"log"
)

func Start() {
	log.Println("Запуск всех сервисов")
	go StartDiscord()
	go StartTelegram()
	log.Println("Все сервисы успешно запущены")
}

func Close() {
	log.Println("Остановка всех сервисов")
	CloseDiscord()
	CloseTelegram()
	log.Println("Все сервисы успешно остановлены")
}

func SendNews(news []structures.News) {
	log.Printf("Отправка %d новостей во все сервисы", len(news))
	for _, n := range news {
		go SendNewsToThread(n)
	}
}
