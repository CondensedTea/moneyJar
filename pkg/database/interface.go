package database

import "context"

type Provider interface {
	CreateUser(ctx context.Context, id int, name string) error
	UpdateAccounts(ctx context.Context, fromUser int, toUsers []int, amount int) ([]Account, error)
	UserNameToID(ctx context.Context, username string) (int, error)
	GetAccounts(ctx context.Context, userId int) ([]Account, error)
}
