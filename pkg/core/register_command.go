package core

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	tg "gopkg.in/telebot.v3"
)

func (c Core) registerCommand(tgCtx tg.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	id := int(tgCtx.Sender().ID)
	username := tgCtx.Sender().Username

	if err := c.db.CreateUser(ctx, id, username); err != nil {
		log.Errorf("failed to create user: %v", err)
		msg := c.messages["failedToAddUser"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}
	msg := fmt.Sprintf(c.messages["succesifullyAddedUser"], tgCtx.Sender().Username)
	return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
}
