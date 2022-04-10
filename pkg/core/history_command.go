package core

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	tg "gopkg.in/telebot.v3"
)

func (c Core) historyCommand(tgCtx tg.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	var (
		userID  = int(tgCtx.Sender().ID)
		payload = tgCtx.Message().Payload
		page    = 1
		err     error
	)

	if payload != "" {
		page, err = strconv.Atoi(payload)
		if err != nil || page <= 0 {
			log.Errorf("failed to parse page: %v", err)
			msg := c.messages["failedToParsePageNumber"]
			return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
		}
	}

	logs, err := c.db.GetTransactionsForUser(ctx, userID, page)
	if err != nil {
		log.Errorf("failed to get log for user %d: %v", userID, err)
		msg := c.messages["failedToGetHistory"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}
	msg := fmt.Sprintf("История, страница %d: \n", page)

	for i, l := range logs {
		msgLine := fmt.Sprintf(
			"%d) @%s -> @%s: %.2f$; %s\n",
			i+1,
			l.FromUserName,
			l.ToUserName,
			float64(l.BalanceChange)/100.0,
			l.Comment)
		msg += msgLine
	}
	return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
}
