package db

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/qiniu/qmgo"
)

var (
	client       *qmgo.Client
	database     *qmgo.Database
	collections  map[string]*qmgo.Collection
	initializeMu sync.Mutex
	initialized  bool
)

func Initialize(uri, dbName string) error {
	initializeMu.Lock()
	defer initializeMu.Unlock()

	if initialized {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = qmgo.NewClient(ctx, &qmgo.Config{Uri: uri})
	if err != nil {
		return err
	}

	database = client.Database(dbName)
	collections = make(map[string]*qmgo.Collection)

	initialized = true
	return nil
}

func GetCollection(name string) *qmgo.Collection {
	if !initialized {
		log.Printf("Warning: Database not initialized when trying to get collection %s", name)
		return nil
	}

	if col, ok := collections[name]; ok {
		return col
	}

	col := database.Collection(name)
	collections[name] = col
	return col
}

func Close() error {
	if !initialized || client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Close(ctx)
	if err != nil {
		log.Printf("Error closing MongoDB connection: %v", err)
		return err
	}

	initialized = false
	return nil
}
