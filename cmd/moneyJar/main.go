package main

import (
	"flag"
	"moneyjar/pkg/config"
	"moneyjar/pkg/core"
	"moneyjar/pkg/database"
	"moneyjar/pkg/logger"
	"moneyjar/pkg/messages"
	"moneyjar/pkg/telegram"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to yaml config file")
	loglevelRaw := flag.String("loglevel", "info", "Log level")
	flag.Parse()

	log, err := logger.New(*loglevelRaw)
	if err != nil {
		panic(err)
	}

	if err = config.Load(*configPath); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.New(config.C.String("db.dsn"))
	if err != nil {
		log.Fatalf("failed to connect to dabase: %v", err)
	}
	tg, err := telegram.New(config.C.String("telegram.token"))
	if err != nil {
		log.Fatalf("failed to connect to dabase: %v", err)
	}
	msgs, err := messages.Load(config.C.String("messages"))
	c, err := core.New(db, tg, log, msgs)
	if err != nil {
		log.Fatalf("failed to init core module: %v", err)
	}
	log.Info("Bot is running !")
	c.Run()
}
