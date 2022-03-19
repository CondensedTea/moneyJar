package core

import (
	"context"
	"fmt"
	"moneyjar/pkg/database"

	tg "gopkg.in/telebot.v3"
)

func (c Core) balanceCommand(tgCtx tg.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	userId := int(tgCtx.Sender().ID)

	accounts, err := c.db.GetAccounts(ctx, userId)
	if err != nil {
		c.log.Errorf("failed to get accounts: %v", err)
		msg := c.messages["failedToGetAccounts"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}

	msg := generateBalanceMessage(accounts)

	return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message(), ParseMode: tg.ModeHTML, DisableNotification: true})
}

func generateBalanceMessage(balances []database.Account) (msg string) {
	const rowTemplate = "%d) <b>@%s</b> должен_а <b>@%s</b> %.2f$\n"

	for i, account := range balances {
		balance := float64(account.Balance) / 100

		var row string
		if balance > 0 {
			row = fmt.Sprintf(rowTemplate, i+1, account.ToUserName, account.FromUserName, balance)
		} else {
			row = fmt.Sprintf(rowTemplate, i+1, account.FromUserName, account.ToUserName, balance*-1)
		}
		msg += row
	}
	return msg
}
