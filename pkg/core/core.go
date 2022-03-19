package core

import (
	"fmt"
	"moneyjar/pkg/config"
	"moneyjar/pkg/database"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	telegbot "gopkg.in/telebot.v3"
)

const (
	commandTimeout = 10 * time.Second
	apiTimeout     = 5 * time.Second
)

type Getter interface {
	Get(string) (*http.Response, error)
}

type Core struct {
	db database.Provider

	tg       *telegbot.Bot
	commands []telegbot.Command
	messages map[string]string

	log        *logrus.Logger
	apiKey     string
	httpClient Getter
}

func New(db database.Provider, tg *telegbot.Bot, log *logrus.Logger, msgs map[string]string) (*Core, error) {
	apiKey := config.C.String("api_key")

	c := &Core{
		db: db,

		tg:       tg,
		commands: make([]telegbot.Command, 0),
		messages: msgs,

		log:        log,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: apiTimeout},
	}

	c.addCommand("/register", "Зарегистрироваться в боте", c.registerCommand)
	c.addCommand("/debt", "Добавить долг. Формат: <сумма долга> <валюта> <@mention должника>", c.debtCommand)
	c.addCommand("/balance", "Текущие балансы", c.balanceCommand)

	if err := c.tg.SetCommands(c.commands); err != nil {
		return nil, fmt.Errorf("failed to set telegram commands: %v", err)
	}

	return c, nil
}

func (c *Core) addCommand(cmd, description string, handler telegbot.HandlerFunc) {
	c.tg.Handle(cmd, handler)
	command := telegbot.Command{
		Text:        cmd,
		Description: description,
	}
	c.commands = append(c.commands, command)
}

func (c Core) Run() {
	c.tg.Start()
}
