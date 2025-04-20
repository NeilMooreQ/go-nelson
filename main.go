package main

import (
	"log"

	"go-nelson/pkg"
	"go-nelson/pkg/db"
	"go-nelson/pkg/news"
	"go-nelson/pkg/services"
)

func main() {
	err := pkg.LoadConfig("configs.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = db.Initialize(pkg.MongoDB.URI, pkg.MongoDB.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	services.Start()

	news.StartNewsParser()
}
