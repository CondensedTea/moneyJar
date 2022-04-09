package database

import "context"

// Provider is database interface
type Provider interface {
	CreateUser(ctx context.Context, id int, name string) error
	UpdateAccounts(ctx context.Context, toAccounts []Account, amount int, comment string) ([]Account, error)
	UserNameToAccount(ctx context.Context, fromUserID int, toUsername string) (Account, error)
	GetAccountsWithUser(ctx context.Context, userID int) ([]Account, error)
	GetTransactionsForUser(ctx context.Context, userID, page int) ([]Log, error)
}
