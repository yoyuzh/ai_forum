// Package main starts the outbox-publisher process.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"ai-forum/backend/internal/bootstrap"
	"ai-forum/backend/internal/config"
)

func main() {
	cfg, err := config.Load(configPath())
	if err == nil {
		err = config.Validate(cfg)
	}
	if err != nil {
		log.Fatal(err)
	}
	app, err := bootstrap.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := bootstrap.RunProcess(context.Background(), app, app.NewOutboxPublisher(), 10*time.Second); err != nil {
		log.Fatal(err)
	}
}

func configPath() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return "config/config.dev.yaml"
}
