package telegram

import (
	"time"

	"gopkg.in/telebot.v3"
)

// New returns configured telebot.Bot
func New(token string) (*telebot.Bot, error) {
	opts := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	return telebot.NewBot(opts)
}
