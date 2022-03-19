package telegram

import (
	"time"

	tg "gopkg.in/telebot.v3"
)

func New(token string) (*tg.Bot, error) {
	opts := tg.Settings{
		Token:  token,
		Poller: &tg.LongPoller{Timeout: 10 * time.Second},
	}
	return tg.NewBot(opts)
}
