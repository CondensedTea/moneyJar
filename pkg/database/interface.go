package database

import "context"

// Provider is database interface
type Provider interface {
	CreateUser(ctx context.Context, id int, name string) error
	UpdateAccounts(ctx context.Context, toAccounts []int, amount int) ([]Account, error)
	UserNamesToAccount(ctx context.Context, fromUserID int, toUserName string) (int, error)
	GetAccounts(ctx context.Context, userID int) ([]Account, error)
}
