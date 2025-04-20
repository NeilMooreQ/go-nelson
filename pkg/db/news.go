package db

import (
	"context"
	"go-nelson/pkg/structures"
	"log"
	"time"

	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/operator"
	opts "github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NewsRepository struct {
	collection *qmgo.Collection
}

func NewNewsRepository() *NewsRepository {
	coll := GetCollection("news")

	ctx := context.Background()
	indexOpt := options.Index().SetUnique(true)

	err := coll.CreateOneIndex(ctx, opts.IndexModel{
		Key:          []string{"provider", "unique_id"},
		IndexOptions: indexOpt,
	})
	if err != nil {
		log.Printf("Error creating index on news collection: %v", err)
	}

	return &NewsRepository{
		collection: coll,
	}
}

func (r *NewsRepository) Save(news *structures.News) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"provider":  news.Provider,
		"unique_id": news.UniqueID,
	}

	existingNews := &structures.News{}
	err := r.collection.Find(ctx, filter).One(existingNews)

	if err == nil {
		news.Id = existingNews.Id
		news.UpdateAt = time.Now()

		err = r.collection.UpdateOne(ctx, filter, bson.M{
			operator.Set: news,
		})

		return err
	} else {
		if news.CreateAt.IsZero() {
			news.CreateAt = time.Now()
			news.UpdateAt = time.Now()
		}

		_, err = r.collection.InsertOne(ctx, news)
		return err
	}
}

func (r *NewsRepository) FindByID(id string) (*structures.News, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	news := &structures.News{}
	err = r.collection.Find(ctx, bson.M{"_id": objID}).One(news)

	return news, err
}

func (r *NewsRepository) FindByProviderAndUniqueID(provider, uniqueID string) (*structures.News, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	news := &structures.News{}
	err := r.collection.Find(ctx, bson.M{
		"provider":  provider,
		"unique_id": uniqueID,
	}).One(news)

	return news, err
}

func (r *NewsRepository) UpdateDiscordInfo(newsID, threadID, messageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(newsID)
	if err != nil {
		return err
	}

	update := bson.M{
		operator.Set: bson.M{
			"discord_thread_id":  threadID,
			"discord_message_id": messageID,
			"updateat":           time.Now(),
		},
	}

	return r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
}

func (r *NewsRepository) UpdateTelegramInfo(newsID, messageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(newsID)
	if err != nil {
		return err
	}

	update := bson.M{
		operator.Set: bson.M{
			"telegram_message_id": messageID,
			"updateat":            time.Now(),
		},
	}

	return r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
}

func (r *NewsRepository) FindRecent(page, limit int64) ([]*structures.News, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := make([]*structures.News, 0)

	err := r.collection.Find(ctx, bson.M{}).
		Sort("-createat").
		Skip(page * limit).
		Limit(limit).
		All(&result)

	return result, err
}

func (r *NewsRepository) FindNewsByProviderAndUniqueIDs(provider string, uniqueIDs []string) (map[string]bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := make(map[string]bool)

	cursor := r.collection.Find(ctx, bson.M{
		"provider": provider,
		"unique_id": bson.M{
			operator.In: uniqueIDs,
		},
	})

	var foundNews []*structures.News
	err := cursor.All(&foundNews)

	for _, news := range foundNews {
		result[news.UniqueID] = true
	}

	return result, err
}
