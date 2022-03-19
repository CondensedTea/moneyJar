package core

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	tg "gopkg.in/telebot.v3"
)

var (
	reDebtPayload   = regexp.MustCompile(`([-\d.,]+) ([\wа-яА-Я]+) (@[@\w, ]+)$`)
	reMentionsArray = regexp.MustCompile(`@(\w+)`)
)

func (c Core) debtCommand(tgCtx tg.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	fromUser := int(tgCtx.Sender().ID)

	floatAmount, cur, toUsers, err := c.parsePayload(tgCtx)
	if err != nil {
		c.log.Errorf("failed to parse payload: %v", err)
		msg := c.messages["failedToParsePayload"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}

	if floatAmount == 0.0 {
		return nil
	}

	usdAmount, err := c.convertToUSD(cur, floatAmount)
	if err != nil {
		c.log.Errorf("failed to convert currency to USD: %v", err)
		msg := c.messages["failedToConvertCurrency"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}

	if len(toUsers) == 1 && toUsers[0] == 0 {
		return tgCtx.Send("unimplemented", &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}
	balances, err := c.db.UpdateAccounts(ctx, fromUser, toUsers, usdAmount)
	if err != nil {
		c.log.Errorf("failed to update accounts: %v", err)
		msg := c.messages["failedToUpdateBalance"]
		return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message()})
	}

	var msg = "Баланс обновлен успешно\n"
	msg += generateBalanceMessage(balances)
	return tgCtx.Send(msg, &tg.SendOptions{ReplyTo: tgCtx.Message(), ParseMode: tg.ModeHTML})
}

func (c Core) parsePayload(tgCtx tg.Context) (amount float64, cur currency, userIDs []int, err error) {
	match := reDebtPayload.FindStringSubmatch(tgCtx.Message().Payload)
	if len(match) < 3 {
		return amount, cur, userIDs, fmt.Errorf("invalid payload: %d of 3 matches", len(match))
	}
	amount, err = strconv.ParseFloat(match[1], 64)
	if err != nil {
		return amount, cur, userIDs, fmt.Errorf("failed to parse amount: %v", err)
	}

	cur = parseCurrency(match[2])
	if cur == "" {
		return amount, cur, userIDs, fmt.Errorf("unknown currency: %s", match[1])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if match[3] == "all" {
		userIDs = []int{0}
	} else {
		usernames, err := parseMentions(match[3])
		if err != nil {
			return amount, cur, userIDs, fmt.Errorf("failed to parse mentions string: %v", err)
		}

		for _, username := range usernames {
			user, err := c.db.UserNameToID(ctx, username)
			if err != nil {
				c.log.Warnf("failed to get userId from name %s: %v", match[3], err)
				continue
			}
			userIDs = append(userIDs, user)
		}
	}
	return amount, cur, userIDs, nil
}

func parseMentions(mentions string) (usernames []string, err error) {
	match := reMentionsArray.FindAllStringSubmatch(mentions, -1)
	if len(match) < 1 {
		return nil, fmt.Errorf("bad string: %s", mentions)
	}
	for _, mention := range match {
		if len(mention) < 2 {
			logrus.Warnf("mention matched, but didnt have match group: %s", mention)
			continue
		}
		usernames = append(usernames, mention[1])
	}
	return usernames, nil
}
