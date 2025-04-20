package news

import (
	"go-nelson/pkg"
	"go-nelson/pkg/db"
	"go-nelson/pkg/services"
	"go-nelson/pkg/structures"
	"log"
	"time"
)

func StartNewsParser() {
	log.Println("Запуск парсера новостей")
	parseAllSources()

	ticker := time.NewTicker(60 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		parseAllSources()
	}
}

func parseAllSources() {
	log.Println("Парсинг всех источников новостей")
	var allNews []structures.News

	if pkg.Parsers.Ixbt {
		ixbtGamesNews, err := ParseIXBTGames()
		if err != nil {
			log.Printf("Ошибка при парсинге IXBT Games: %v", err)
		} else {
			allNews = append(allNews, ixbtGamesNews...)
		}
	}

	if pkg.Parsers.Stopgame {
		stopGameNews, err := ParseStopGame()
		if err != nil {
			log.Printf("Ошибка при парсинге StopGame: %v", err)
		} else {
			allNews = append(allNews, stopGameNews...)
		}
	}

	if pkg.Parsers.DTF {
		dtfNews, err := ParseDTF()
		if err != nil {
			log.Printf("Ошибка при парсинге DTF: %v", err)
		} else {
			allNews = append(allNews, dtfNews...)
		}
	}

	if pkg.Parsers.DisgustingMen {
		disgustingmenNews, err := ParseDMen()
		if err != nil {
			log.Printf("Ошибка при парсинге DisgustingMen: %v", err)
		} else {
			allNews = append(allNews, disgustingmenNews...)
		}
	}

	if pkg.Parsers.ThreeDNews {
		threedsNews, err := Parse3DNews()
		if err != nil {
			log.Printf("Ошибка при парсинге 3DNews: %v", err)
		} else {
			allNews = append(allNews, threedsNews...)
		}
	}

	if pkg.Parsers.EpicGames {
		epicGamesNews, err := ParseEpicGamesStore()
		if err != nil {
			log.Printf("Ошибка при парсинге EpicGames: %v", err)
		} else {
			allNews = append(allNews, epicGamesNews...)
		}
	}

	if pkg.Parsers.GamedevRu {
		gamedevNews, err := ParseGameDev()
		if err != nil {
			log.Printf("Ошибка при парсинге GameDev: %v", err)
		} else {
			allNews = append(allNews, gamedevNews...)
		}
	}

	if pkg.Parsers.SteamDevelopers {
		steamNews, err := ParseSteam()
		if err != nil {
			log.Printf("Ошибка при парсинге Steam Developer: %v", err)
		} else {
			allNews = append(allNews, steamNews...)
		}
	}

	filteredNews := filterExistingNews(allNews)

	if len(filteredNews) > 0 {
		processNews(filteredNews)
	}
}

func filterExistingNews(allNews []structures.News) []structures.News {
	if len(allNews) == 0 {
		return []structures.News{}
	}

	newsRepo := db.NewNewsRepository()
	var filteredNews []structures.News

	newsByProvider := make(map[string][]structures.News)
	for _, n := range allNews {
		newsByProvider[n.Provider] = append(newsByProvider[n.Provider], n)
	}

	for provider, news := range newsByProvider {
		uniqueIDs := make([]string, 0, len(news))
		for _, n := range news {
			uniqueIDs = append(uniqueIDs, n.UniqueID)
		}

		existingIDs, err := newsRepo.FindNewsByProviderAndUniqueIDs(provider, uniqueIDs)
		if err != nil {
			log.Printf("Ошибка при проверке существующих новостей для %s: %v", provider, err)
			continue
		}

		for _, n := range news {
			if !existingIDs[n.UniqueID] {
				filteredNews = append(filteredNews, n)
			}
		}
	}

	return filteredNews
}

func processNews(news []structures.News) {
	log.Printf("Обработка %d новых новостей", len(news))
	newsRepo := db.NewNewsRepository()

	for _, n := range news {
		err := newsRepo.Save(&n)
		if err != nil {
			log.Printf("Ошибка при сохранении новости: %v", err)
		}
	}

	go services.SendNews(news)
}
