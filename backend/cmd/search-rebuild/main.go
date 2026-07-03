// Package main rebuilds the Elasticsearch search read model from MySQL.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/search"
)

func main() {
	cfg, err := config.Load(configPath())
	if err == nil {
		err = validateSearchRebuildConfig(cfg)
	}
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewMySQL(cfg.MySQL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	esClient, err := search.NewES(cfg.Elasticsearch)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := search.NewSyncHandler(db, search.NewESIndexStore(esClient)).RebuildAll(ctx); err != nil {
		log.Fatal(err)
	}
	log.Print("search rebuild complete")
}

func validateSearchRebuildConfig(cfg *config.Config) error {
	if cfg.MySQL.Host == "" || cfg.MySQL.Port == 0 || cfg.MySQL.Username == "" || cfg.MySQL.Database == "" {
		return fmt.Errorf("missing mysql config")
	}
	if len(cfg.Elasticsearch.Addresses) == 0 || cfg.Elasticsearch.Addresses[0] == "" {
		return fmt.Errorf("missing elasticsearch addresses")
	}
	return nil
}

func configPath() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return "config/config.dev.yaml"
}
