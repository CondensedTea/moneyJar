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
		    accounts (from_user, to_user, is_flipped) 
		    select $1 to_user, id from_user, false from users where id != $1
		    union
		    select id to_user, $1 from_user, true from users where id != $1`
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
func (db Database) UpdateAccounts(ctx context.Context, toAccounts []Account, amount int, comment string) ([]Account, error) {
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

	if len(toAccounts) > 1 {
		partsPlusAuthor := len(toAccounts) + 1
		amount = int(math.Round(float64(amount) / float64(partsPlusAuthor)))
	}

	const balanceQuery = `
		update
			accounts
		set
		    balance = case when from_user = $2 then balance + $1 else balance end
		where
		    (from_user = $3 and to_user = $2) or (from_user = $2 and to_user = $3)
		returning from_user, to_user, balance, is_flipped`

	const logQuery = `insert into transactionlog (from_user, to_user, balance_change, comment) values ($1, $2, $3, $4)`

	for _, toAccount := range toAccounts {
		var updatedAccounts []Account

		err = tx.SelectContext(ctx, &updatedAccounts, balanceQuery, amount, toAccount.FromUser, toAccount.ToUser)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, fmt.Errorf("failed to rollback: %v", err)
			}
			return nil, fmt.Errorf("failed to update balance: %v", err)
		}

		if _, err = tx.ExecContext(ctx, logQuery, toAccount.FromUser, toAccount.ToUser, amount, comment); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, fmt.Errorf("failed to rollback: %v", err)
			}
			return nil, fmt.Errorf("failed to insert log record: %v", err)
		}

		updatedAccounts = mergeDuplicateAccounts(updatedAccounts)
		if len(updatedAccounts) > 1 {
			return nil, fmt.Errorf("updated accounts after merging still not 1: %d", len(updatedAccounts))
		}
		account := updatedAccounts[0]
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

// UserNameToAccount gets user ID from username
func (db Database) UserNameToAccount(ctx context.Context, fromUserID int, toUsername string) (Account, error) {
	const query = `
		select 
		       is_flipped, to_user
		from 
		     accounts
		where 
		      from_user = $1
		  and 
		      to_user = (select id from users where name = $2)`

	var (
		isFlipped bool
		userID    int
	)
	if err := db.conn.QueryRowxContext(ctx, query, fromUserID, toUsername).Scan(&isFlipped, &userID); err != nil {
		return Account{}, fmt.Errorf("failed to get user by name: %v", err)
	}
	return Account{FromUser: fromUserID, ToUser: userID, IsFlipped: isFlipped}, nil
}

func (db Database) userIDtoName(ctx context.Context, userID int) (string, error) {
	const query = `select name from users where id = $1`

	var name string
	if err := db.conn.QueryRowxContext(ctx, query, userID).Scan(&name); err != nil {
		return "", fmt.Errorf("failed to get user by id: %v", err)
	}
	return name, nil
}

// GetAccountsWithUser returns accounts connected to user
func (db Database) GetAccountsWithUser(ctx context.Context, userID int) ([]Account, error) {
	const query = `
		select
		       a.id,
		       from_user,
		       u1.name from_user_name,
		       to_user,
		       u2.name to_user_name,
		       balance,
		       is_flipped
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

	accounts = mergeDuplicateAccounts(accounts)

	return accounts, nil
}

// GetTransactionsForUser returns log records with given users
func (db Database) GetTransactionsForUser(ctx context.Context, userID, page int) ([]Log, error) {
	var logs []Log

	if page < 1 {
		return nil, fmt.Errorf("page number can not be less 1")
	}

	const offsetStep = 10
	const query = `
		select 
		       (select name from users where id = from_user) as from_user_name,
		       (select name from users where id = to_user) as to_user_name,
		       balance_change,
		       comment,
		       ts
		from
		     transactionlog
		where
		      from_user = $1 or to_user = $1
		order by ts desc
		limit 10
		offset $2
		`

	if err := db.conn.SelectContext(ctx, &logs, query, userID, (page-1)*offsetStep); err != nil {
		return nil, err
	}
	return logs, nil
}

// mergeDuplicateAccounts is needed because we store two records for single user-to-user relation.
// It puts accounts in hash map with key as sorted user IDs and sums balances in same pairs.
func mergeDuplicateAccounts(accounts []Account) (resultAccounts []Account) {
	hashMap := make(map[string]*Account)

	for i, account := range accounts {
		if account.IsFlipped == false {
			key := account.String()
			hashMap[key] = &accounts[i]
		}
	}

	for _, account := range accounts {
		if account.IsFlipped == true {
			key := account.String()
			hashMap[key].Balance += account.Balance * -1
		}
	}
	for _, account := range hashMap {
		resultAccounts = append(resultAccounts, *account)
	}
	return resultAccounts
}
