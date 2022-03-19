package database

import "time"

type Account struct {
	ID           int
	FromUser     int    `db:"from_user"`
	FromUserName string `db:"from_user_name"`
	ToUser       int    `db:"to_user"`
	ToUserName   string `db:"to_user_name"`
	Balance      int
}

type User struct {
	ID   int
	Name string
}

type Log struct {
	Account       int
	BalanceChange int
	ts            time.Time
}
