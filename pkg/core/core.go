package core

import (
	"fmt"
	"moneyjar/pkg/config"
	"moneyjar/pkg/database"
	"net/http"
	"time"

	"gopkg.in/telebot.v3"
)

const (
	commandTimeout = 10 * time.Second
	apiTimeout     = 15 * time.Second
)

// Getter represents http.Get interface
type Getter interface {
	Get(string) (*http.Response, error)
}

// Core contains business logic of bot
type Core struct {
	db database.Provider

	tg       *telebot.Bot
	commands []telebot.Command
	messages map[string]string

	apiKey     string
	httpClient Getter
}

// New returns new Core
func New(db database.Provider, tg *telebot.Bot, msgs map[string]string) (*Core, error) {
	apiKey := config.C.String("api_key")

	c := &Core{
		db: db,

		tg:       tg,
		commands: make([]telebot.Command, 0),
		messages: msgs,

		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: apiTimeout},
	}

	c.addCommand("/register", "Зарегистрироваться в боте", c.registerCommand)
	c.addCommand("/debt", "Добавить долг для @пользователя", c.debtCommand)
	c.addCommand("/balance", "Текущие счета", c.balanceCommand)

	if err := c.tg.SetCommands(c.commands); err != nil {
		return nil, fmt.Errorf("failed to set telegram commands: %v", err)
	}

	return c, nil
}

func (c *Core) addCommand(cmd, description string, handler telebot.HandlerFunc) {
	c.tg.Handle(cmd, handler)
	command := telebot.Command{
		Text:        cmd,
		Description: description,
	}
	c.commands = append(c.commands, command)
}

// Run starts telegram bot
func (c Core) Run() {
	c.tg.Start()
}
