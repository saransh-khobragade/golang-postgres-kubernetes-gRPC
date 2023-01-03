package db

import (
	"context"
	"database/sql"
	"fmt"
)

// *Queries will be used for standalone queries
// *sql.DB instance will be needed to do sql operation
// Compositing-Combining both to multiple set of operation batch way

// Store provides all functions to executes db queries and transactions
type Store struct {
	*Queries         //ref to instance provided by sqlc
	db       *sql.DB //ref to database
}

// Newstore creates a new store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,      //ref to sql
		Queries: New(db), //New method provided by sqlc
	}
}

// execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil { //if there is a error need to check if here is rollback error or not
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err %v ,rb err %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id`
	ToAccountID   int64 `json:"to_account_id`
	Amount        int64 `json:"amount`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

//TransferTx performs a money transfer from one account to another
//It create transfer records,add account entries and update account balance within a single database transaction

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		//For creating transfer details,closure variable result
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		//For creating from FromEntry details
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		//For creating from ToEntry details
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
