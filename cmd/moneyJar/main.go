package main

import (
	"flag"
	"moneyjar/pkg/config"
	"moneyjar/pkg/core"
	"moneyjar/pkg/database"
	"moneyjar/pkg/messages"
	"moneyjar/pkg/telegram"

	log "github.com/sirupsen/logrus"

	"gopkg.in/telebot.v3"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to yaml config file")
	loglevelRaw := flag.String("loglevel", "info", "Log level")
	flag.Parse()

	lvl, err := log.ParseLevel(*loglevelRaw)
	if err != nil {
		panic(err)
	}
	log.SetLevel(lvl)

	if err = config.Load(*configPath); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var (
		db   *database.Database
		tg   *telebot.Bot
		msgs map[string]string
		c    *core.Core
	)

	db, err = database.New(config.C.String("db.dsn"))
	if err != nil {
		log.Fatalf("failed to connect to dabase: %v", err)
	}
	tg, err = telegram.New(config.C.String("telegram.token"))
	if err != nil {
		log.Fatalf("failed to connect to dabase: %v", err)
	}
	msgs, err = messages.Load()
	if err != nil {
		log.Fatalf("failed to load messages file: %v", err)
	}
	c, err = core.New(db, tg, msgs)
	if err != nil {
		log.Fatalf("failed to init core module: %v", err)
	}
	log.Info("🔥 Bot is running !")
	c.Run()
}
