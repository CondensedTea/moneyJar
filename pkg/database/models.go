package database

import (
	"fmt"
	"sort"
)

// Account represents record in accounts table
type Account struct {
	ID           int
	FromUser     int    `db:"from_user"`
	FromUserName string `db:"from_user_name"`
	ToUser       int    `db:"to_user"`
	ToUserName   string `db:"to_user_name"`
	IsFlipped    bool   `db:"is_flipped"`
	Balance      int
}

func (a Account) String() string {
	ids := []int{a.FromUser, a.ToUser}
	sort.Ints(ids)
	return fmt.Sprintf("%d:%d", ids[0], ids[1])
}

// User represents record in users table
type User struct {
	ID   int
	Name string
}

// Log represents record in transactionLog table
type Log struct {
	Account       int
	BalanceChange int
	// ts            time.Time
}
