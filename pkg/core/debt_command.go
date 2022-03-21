package core

import (
	"context"
	"fmt"
	"moneyjar/pkg/database"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"
	tg "gopkg.in/telebot.v3"
)

var (
	reDebtPayload   = regexp.MustCompile(`([-\d.,]+) ?([\wа-яА-Я$₽₾]+) (@[@\w, ]+); ?([\wа-яА-Я ]+)$`)
	reMentionsArray = regexp.MustCompile(`@(\w+)`)
)

func (c Core) debtCommand(tgCtx tg.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	fromUser := int(tgCtx.Sender().ID)

	if len(tgCtx.Message().Payload) == 0 {
		// Logger ignore messages without payload
		return nil
	}

	floatAmount, cur, toUsers, err := c.parsePayload(ctx, tgCtx)
	if err != nil {
		log.Errorf("failed to parse payload: %v", err)
		msg := c.messages["failedToParsePayload"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}

	// Skip useless debts
	if floatAmount == 0.0 {
		return nil
	}

	usdAmount, err := c.convertToUSD(cur, floatAmount)
	if err != nil {
		log.Errorf("failed to convert currency to USD: %v", err)
		msg := c.messages["failedToConvertCurrency"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}

	var accounts []database.Account

	if toUsers == nil {
		toUsers, err = c.db.GetAccountsWithUser(ctx, fromUser)
		if err != nil {
			log.Errorf("failed to get all accounts: %v", err)
			msg := c.messages["failedToGetAccounts"]
			return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
		}

		log.Debugf("processing @all; got accounts: %#v, toUser: %v", accounts, toUsers)
	}
	updateAccounts, err := c.db.UpdateAccounts(ctx, toUsers, usdAmount)
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

func (c Core) parsePayload(ctx context.Context, tgCtx tg.Context) (amount float64, cur currency, accounts []database.Account, err error) {
	match := reDebtPayload.FindStringSubmatch(tgCtx.Message().Payload)
	if len(match) < 3 {
		return amount, cur, accounts, fmt.Errorf("invalid payload: %d of 3 matches", len(match))
	}
	amount, err = strconv.ParseFloat(match[1], 64)
	if err != nil {
		return amount, cur, accounts, fmt.Errorf("failed to parse amount: %v", err)
	}

	cur = parseCurrency(match[2])
	if cur == "" {
		return amount, cur, accounts, fmt.Errorf("unknown currency: %s", match[1])
	}

	if match[3] == "@all" {
		accounts = nil
	} else {
		toUsernames, err := parseMentions(match[3])
		if err != nil {
			return amount, cur, accounts, fmt.Errorf("failed to parse mentions string: %v", err)
		}

		fromUser := int(tgCtx.Sender().ID)

		for _, toUsername := range toUsernames {
			account, err := c.db.UserNameToAccount(ctx, fromUser, toUsername)
			if err != nil {
				log.Errorf("failed to get userId from name %s: %v", match[3], err)
				continue
			}
			accounts = append(accounts, account)
		}
	}
	return amount, cur, accounts, nil
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
