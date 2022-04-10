package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"moneyjar/pkg/database"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"
	tg "gopkg.in/telebot.v3"
)

var (
	reDebtPayload   = regexp.MustCompile(`([-\d.,]+) ?([\wа-яА-Я$₽₾]+) ([@\w, ]+);? ?([\wа-яА-Я ]+)?`)
	reMentionsArray = regexp.MustCompile(`@(\w+)`)

	errFailedToGetAllAccounts = errors.New(`failed to get all accounts`)
)

type debtPayload struct {
	amount   float64
	currency currency
	accounts []database.Account
	comment  string
}

func (c Core) debtCommand(tgCtx tg.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	if len(tgCtx.Message().Payload) == 0 {
		log.Debug("ignored message without payload")
		return nil
	}

	debt, err := c.parsePayload(ctx, tgCtx)
	if err != nil {
		log.Errorf("failed to parse payload: %v", err)

		var msg string
		switch {
		case errors.Is(sql.ErrNoRows, err):
			msg = c.messages["unknownUsersInPayload"]
		case errors.Is(errFailedToGetAllAccounts, err):
			msg = c.messages["failedToGetAccounts"]
		default:
			msg = c.messages["failedToParsePayload"]
		}
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}

	// Skip useless debts
	if debt.amount == 0.0 {
		return nil
	}

	usdAmount, err := c.convertToUSD(debt.currency, debt.amount)
	if err != nil {
		log.Errorf("failed to convert currency to USD: %v", err)
		msg := c.messages["failedToConvertCurrency"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}

	updateAccounts, err := c.db.UpdateAccounts(ctx, debt.accounts, usdAmount, debt.comment)
	if err != nil {
		log.Errorf("failed to update accounts: %v", err)
		msg := c.messages["failedToUpdateBalance"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}

	if len(updateAccounts) < 1 {
		// If no accounts were updated we should return warning
		msg := c.messages["zeroBalancesWereUpdated"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}

	var msg = "Баланс обновлен успешно: \n"
	msg += generateBalanceMessage(updateAccounts)
	return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message(), ParseMode: tg.ModeHTML})
}

func (c Core) parsePayload(ctx context.Context, tgCtx tg.Context) (*debtPayload, error) {
	match := reDebtPayload.FindStringSubmatch(tgCtx.Message().Payload)
	if len(match) < 5 {
		return nil, fmt.Errorf("invalid payload: %d of 5 matches", len(match))
	}
	amount, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %v", err)
	}

	cur := parseCurrency(match[2])
	if cur == "" {
		return nil, fmt.Errorf("unknown currency: %s", match[1])
	}

	var accounts []database.Account

	if match[3] == "@all" {
		fromUser := int(tgCtx.Sender().ID)
		accounts, err = c.db.GetAccountsWithUser(ctx, fromUser)
		if err != nil {
			log.Errorf("failed to get all accounts: %v", err)
			return nil, fmt.Errorf("%w: %v", errFailedToGetAllAccounts, err)
		}
	} else {
		toUsernames, err := parseMentions(match[3])
		if err != nil {
			return nil, fmt.Errorf("failed to parse mentions string: %v", err)
		}

		fromUser := int(tgCtx.Sender().ID)

		for _, toUsername := range toUsernames {
			account, err := c.db.UserNameToAccount(ctx, fromUser, toUsername)
			if err != nil {
				return nil, fmt.Errorf("failed to get userId from name %s: %v", match[3], err)
			}
			accounts = append(accounts, account)
		}
	}

	return &debtPayload{
		amount:   amount,
		currency: cur,
		accounts: accounts,
		comment:  match[4],
	}, nil
}

func parseMentions(mentions string) (usernames []string, err error) {
	match := reMentionsArray.FindAllStringSubmatch(mentions, -1)
	if len(match) < 1 {
		return nil, fmt.Errorf("bad string: %s", mentions)
	}
	for _, mention := range match {
		if len(mention) < 2 {
			log.Warnf("mention matched, but didnt have match group: %s", mention)
			continue
		}
		usernames = append(usernames, mention[1])
	}
	return usernames, nil
}
