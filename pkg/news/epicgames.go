package news

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-nelson/pkg/structures"
	"go-nelson/pkg/utils"
	"log"
	"time"
)

type EpicGamesResponse struct {
	Data struct {
		Catalog struct {
			SearchStore struct {
				Elements []EpicGamesElement `json:"elements"`
			} `json:"searchStore"`
		} `json:"Catalog"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

type EpicGamesElement struct {
	Title         string `json:"title"`
	ID            string `json:"id"`
	Description   string `json:"description"`
	OfferType     string `json:"offerType"`
	EffectiveDate string `json:"effectiveDate"`
	ExpiryDate    string `json:"expiryDate"`
	KeyImages     []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"keyImages"`
	Price struct {
		TotalPrice struct {
			DiscountPrice   int    `json:"discountPrice"`
			OriginalPrice   int    `json:"originalPrice"`
			VoucherDiscount int    `json:"voucherDiscount"`
			Discount        int    `json:"discount"`
			CurrencyCode    string `json:"currencyCode"`
			FmtPrice        struct {
				OriginalPrice string `json:"originalPrice"`
				DiscountPrice string `json:"discountPrice"`
			} `json:"fmtPrice"`
		} `json:"totalPrice"`
	} `json:"price"`
	Promotions *struct {
		PromotionalOffers []struct {
			PromotionalOffers []struct {
				StartDate       string `json:"startDate"`
				EndDate         string `json:"endDate"`
				DiscountSetting struct {
					DiscountType       string `json:"discountType"`
					DiscountPercentage int    `json:"discountPercentage"`
				} `json:"discountSetting"`
			} `json:"promotionalOffers"`
		} `json:"promotionalOffers"`
		UpcomingPromotionalOffers []struct {
			PromotionalOffers []struct {
				StartDate       string `json:"startDate"`
				EndDate         string `json:"endDate"`
				DiscountSetting struct {
					DiscountType       string `json:"discountType"`
					DiscountPercentage int    `json:"discountPercentage"`
				} `json:"discountSetting"`
			} `json:"promotionalOffers"`
		} `json:"upcomingPromotionalOffers"`
	} `json:"promotions"`
	Categories []struct {
		Path string `json:"path"`
	} `json:"categories"`
	ProductSlug string `json:"productSlug"`
	UrlSlug     string `json:"urlSlug"`
}

func ParseEpicGamesStore() ([]structures.News, error) {
	log.Println("Парсинг новостей с Epic Games Store")
	var news []structures.News
	fetcher := utils.NewFetcher()

	data, err := fetcher.Fetch("https://store-site-backend-static.ak.epicgames.com/freeGamesPromotions")
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе API Epic Games Store: %v", err)
	}

	var response EpicGamesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("ошибка при разборе JSON-ответа Epic Games Store: %v", err)
	}

	if len(response.Errors) > 0 {
		log.Printf("API Epic Games вернул ошибки: %v", response.Errors)
	}

	currentTime := time.Now()

	for _, game := range response.Data.Catalog.SearchStore.Elements {
		if !isEligibleForFree(game) {
			continue
		}

		currentlyFree, startDate, endDate := isGameFree(game, currentTime)
		upcomingFree, upcomingStartDate, upcomingEndDate := isGameUpcomingFree(game, currentTime)

		if !currentlyFree && !upcomingFree {
			continue
		}

		imageURL := getGameImage(game)
		gameURL := getGameLink(game)

		if currentlyFree {
			newsItem := createNewsDbItem(game, imageURL, gameURL, "Сейчас бесплатно", startDate, endDate)

			if game.Price.TotalPrice.OriginalPrice > 0 && game.Price.TotalPrice.DiscountPrice == 0 {
				originalPrice := game.Price.TotalPrice.FmtPrice.OriginalPrice
				newsItem.Description += fmt.Sprintf("\n\nОбычная цена: %s", originalPrice)
			}

			news = append(news, newsItem)
		}

		if upcomingFree {
			newsItem := createNewsDbItem(game, imageURL, gameURL, "Скоро бесплатно", upcomingStartDate, upcomingEndDate)

			if game.Price.TotalPrice.OriginalPrice > 0 {
				originalPrice := game.Price.TotalPrice.FmtPrice.OriginalPrice
				newsItem.Description += fmt.Sprintf("\n\nОбычная цена: %s", originalPrice)
			}

			news = append(news, newsItem)
		}
	}

	return news, nil
}

func isEligibleForFree(game EpicGamesElement) bool {
	validType := game.OfferType == "BASE_GAME" || game.OfferType == "ADD_ON" || game.OfferType == "BUNDLE"
	if !validType {
		return false
	}

	hasFreeCategory := false
	for _, category := range game.Categories {
		if category.Path == "freegames" {
			hasFreeCategory = true
			break
		}
	}

	return hasFreeCategory
}

func isGameFree(game EpicGamesElement, currentTime time.Time) (bool, time.Time, time.Time) {
	if game.Promotions == nil || len(game.Promotions.PromotionalOffers) == 0 {
		if game.Price.TotalPrice.OriginalPrice == 0 ||
			(game.Price.TotalPrice.OriginalPrice > 0 && game.Price.TotalPrice.DiscountPrice == 0) {
			return true, time.Now(), time.Now().AddDate(0, 0, 7)
		}
		return false, time.Time{}, time.Time{}
	}

	for _, promo := range game.Promotions.PromotionalOffers {
		for _, offer := range promo.PromotionalOffers {
			startDate, err := time.Parse(time.RFC3339, offer.StartDate)
			if err != nil {
				log.Printf("Ошибка при парсинге даты начала акции: %v", err)
				continue
			}

			endDate, err := time.Parse(time.RFC3339, offer.EndDate)
			if err != nil {
				log.Printf("Ошибка при парсинге даты окончания акции: %v", err)
				continue
			}

			if currentTime.After(startDate) && currentTime.Before(endDate) {
				if offer.DiscountSetting.DiscountType == "PERCENTAGE" &&
					(offer.DiscountSetting.DiscountPercentage == 0 ||
						offer.DiscountSetting.DiscountPercentage == 100 ||
						game.Price.TotalPrice.DiscountPrice == 0) {
					return true, startDate, endDate
				}
			}
		}
	}

	if game.Price.TotalPrice.OriginalPrice > 0 && game.Price.TotalPrice.DiscountPrice == 0 {
		return true, time.Now(), time.Now().AddDate(0, 0, 7)
	}

	return false, time.Time{}, time.Time{}
}

func isGameUpcomingFree(game EpicGamesElement, currentTime time.Time) (bool, time.Time, time.Time) {
	if game.OfferType == "ADD_ON" && game.ExpiryDate != "" {
		expiryDate, err := time.Parse(time.RFC3339, game.ExpiryDate)
		if err == nil && expiryDate.After(currentTime) {
			effectiveDate, err := time.Parse(time.RFC3339, game.EffectiveDate)
			if err == nil && effectiveDate.After(currentTime) {
				return true, effectiveDate, expiryDate
			}
		}
	}

	if game.Promotions == nil || len(game.Promotions.UpcomingPromotionalOffers) == 0 {
		return false, time.Time{}, time.Time{}
	}

	for _, promo := range game.Promotions.UpcomingPromotionalOffers {
		for _, offer := range promo.PromotionalOffers {
			startDate, err := time.Parse(time.RFC3339, offer.StartDate)
			if err != nil {
				log.Printf("Ошибка при парсинге даты начала предстоящей акции: %v", err)
				continue
			}

			endDate, err := time.Parse(time.RFC3339, offer.EndDate)
			if err != nil {
				log.Printf("Ошибка при парсинге даты окончания предстоящей акции: %v", err)
				continue
			}

			if startDate.After(currentTime) &&
				(offer.DiscountSetting.DiscountType == "PERCENTAGE" &&
					(offer.DiscountSetting.DiscountPercentage == 0 ||
						offer.DiscountSetting.DiscountPercentage == 100)) {
				return true, startDate, endDate
			}
		}
	}

	return false, time.Time{}, time.Time{}
}

func getGameImage(game EpicGamesElement) string {
	for _, image := range game.KeyImages {
		if image.Type == "OfferImageTall" {
			return image.URL
		}
	}

	for _, image := range game.KeyImages {
		if image.Type == "OfferImageWide" {
			return image.URL
		}
	}

	for _, image := range game.KeyImages {
		if image.Type == "Thumbnail" {
			return image.URL
		}
	}

	return ""
}

func getGameLink(game EpicGamesElement) string {
	baseURL := "https://store.epicgames.com/ru/p/"

	if game.ProductSlug != "" && game.ProductSlug != "null" {
		return baseURL + game.ProductSlug
	}

	if game.UrlSlug != "" && game.UrlSlug != "null" {
		return baseURL + game.UrlSlug
	}

	return "https://store.epicgames.com/ru/free-games"
}

func createNewsDbItem(game EpicGamesElement, imageURL, gameURL, status string, startDate, endDate time.Time) structures.News {
	hash := md5.Sum([]byte(game.ID + status))
	id := hex.EncodeToString(hash[:])

	var dateRange string
	if !startDate.IsZero() && !endDate.IsZero() {
		dateRange = fmt.Sprintf("%s - %s",
			startDate.Format("02.01.2006"),
			endDate.Format("02.01.2006"))
	}

	var title, content string
	gameType := getGameType(game.OfferType)

	var availableNow, availableSoon string
	switch gameType {
	case "Игра":
		availableNow = "доступна бесплатно"
		availableSoon = "скоро будет бесплатной"
	case "Дополнение":
		availableNow = "доступно бесплатно"
		availableSoon = "скоро будет бесплатным"
	default:
		availableNow = "доступен бесплатно"
		availableSoon = "скоро будет бесплатным"
	}

	if status == "Сейчас бесплатно" {
		title = fmt.Sprintf("%s %s %s в Epic Games Store", gameType, game.Title, availableNow)
		content = fmt.Sprintf("%s\n\n%s %s в период: %s",
			game.Description, gameType, availableNow, dateRange)
	} else {
		title = fmt.Sprintf("%s %s %s в Epic Games Store", gameType, game.Title, availableSoon)
		content = fmt.Sprintf("%s\n\n%s %s в период: %s",
			game.Description, gameType, availableSoon, dateRange)
	}

	var images []string
	if imageURL != "" {
		images = append(images, imageURL)
	}

	return structures.News{
		Provider:    "Epic Games Store",
		UniqueID:    id,
		Title:       title,
		Description: content,
		URL:         gameURL,
		Images:      images,
	}
}

func getGameType(offerType string) string {
	switch offerType {
	case "BASE_GAME":
		return "Игра"
	case "ADD_ON":
		return "Дополнение"
	case "BUNDLE":
		return "Набор"
	default:
		return "Контент"
	}
}
