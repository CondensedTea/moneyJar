package database

import (
	"context"
	"fmt"
	"math"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // nolint:revive
	log "github.com/sirupsen/logrus"
)

// Database wraps DB-related logic
type Database struct {
	conn *sqlx.DB
}

// New returns new Database
func New(dsn string) (*Database, error) {
	conn, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &Database{
		conn: conn,
	}, nil
}

// CreateUser creates new User in database
func (db Database) CreateUser(ctx context.Context, id int, name string) error {
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer func() {
		commitErr := tx.Commit()
		if commitErr != nil {
			log.Errorf("failed to commit: %v", err)
		}
	}()

	const createUserQuery = "insert into users (id, name) values ($1, $2)"
	_, err = tx.ExecContext(ctx, createUserQuery, id, name)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback: %v", err)
		}
		return fmt.Errorf("failed to create user: %v", err)
	}

	const createAccountsQuery = `
		insert into
		    accounts (from_user, to_user) 
		    select id from_user, $1 to_user from users where id != $1`
	_, err = tx.ExecContext(ctx, createAccountsQuery, id)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback: %v", err)
		}
		return fmt.Errorf("failed to create accounts: %v", err)
	}
	return nil
}

// UpdateAccounts updates accounts from one user to multiple users
func (db Database) UpdateAccounts(ctx context.Context, toAccounts []int, amount int) ([]Account, error) {
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer func() {
		commitErr := tx.Commit()
		if commitErr != nil {
			log.Errorf("db.UpdateAccounts: failed to commit: %v", err)
		}
	}()
	var accounts []Account

	dividedAmount := int(math.Round(float64(amount) / float64(len(toAccounts))))

	const balanceQuery = `
		update 
			accounts
		set 
			balance = balance + $1 
		where 
		      id = $2
		returning id, from_user, to_user, balance`

	const logQuery = `insert into transactionlog (account, balance_change) values ($1, $2)`

	for _, toAccount := range toAccounts {
		var account Account

		if err = tx.GetContext(ctx, &account, balanceQuery, dividedAmount, toAccount); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, fmt.Errorf("failed to rollback: %v", err)
			}
			return nil, fmt.Errorf("failed to update balance: %v", err)
		}

		if _, err = tx.ExecContext(ctx, logQuery, account.ID, amount); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, fmt.Errorf("failed to rollback: %v", err)
			}
			return nil, fmt.Errorf("failed to insert log record: %v", err)
		}

		(&account).FromUserName, err = db.userIDtoName(ctx, account.FromUser)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, fmt.Errorf("failed to rollback: %v", err)
			}
			return nil, fmt.Errorf("failed to resolve FromUser name by id: %v", err)
		}
		(&account).ToUserName, err = db.userIDtoName(ctx, account.ToUser)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, fmt.Errorf("failed to rollback: %v", err)
			}
			return nil, fmt.Errorf("failed to resolve ToUser name by id: %v", err)
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// UserNamesToAccount converts username to user ID
func (db Database) UserNamesToAccount(ctx context.Context, fromUserID int, toUsername string) (int, error) {
	const query = `
		select 
		       id 
		from 
		     accounts
		where 
		      (from_user = $1 and to_user = (select id from users where name = $2))
		   or 
		      (to_user = $1 and from_user = (select id from users where name = $2))`

	var id int
	if err := db.conn.QueryRowxContext(ctx, query, fromUserID, toUsername).Scan(&id); err != nil {
		return 0, fmt.Errorf("failed to get user by name: %v", err)
	}
	return id, nil
}

func (db Database) userIDtoName(ctx context.Context, userID int) (string, error) {
	const query = `select name from users where id = $1`

	var name string
	if err := db.conn.QueryRowxContext(ctx, query, userID).Scan(&name); err != nil {
		return "", fmt.Errorf("failed to get user by id: %v", err)
	}
	return name, nil
}

// GetAccounts returns accounts connected to user
func (db Database) GetAccounts(ctx context.Context, userID int) ([]Account, error) {
	const query = `
		select
		       a.id,
		       from_user,
		       u1.name from_user_name,
		       to_user,
		       u2.name to_user_name,
		       balance
		from
		     accounts a
		         join users u1 on u1.id = a.from_user
		         join users u2 on u2.id = a.to_user
		where
		      from_user = $1
		   or
		      to_user = $1`

	var accounts []Account
	if err := db.conn.SelectContext(ctx, &accounts, query, userID); err != nil {
		return nil, fmt.Errorf("failed to get list of accounts for user %d: %v", userID, err)
	}
	return accounts, nil
}
