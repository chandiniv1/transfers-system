package db

import (
	"context"
	"fmt"
)

// TransferTxParams contains the input parameters for transferring money
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult contains the result of a successful transfer transaction
type TransferTxResult struct {
	Transaction Transaction `json:"transaction"`
	FromAccount Account     `json:"from_account"`
	ToAccount   Account     `json:"to_account"`
}

// TransferTx creates a transaction, updates balances, and returns updated accounts.
func (s *SQLStore) TransferTx(ctx context.Context, args TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transaction, err = q.CreateTransaction(ctx, CreateTransactionParams{
			SourceAccountID:      args.FromAccountID,
			DestinationAccountID: args.ToAccountID,
			Amount:               args.Amount,
		})
		if err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		if args.FromAccountID < args.ToAccountID {
			err = updateBalances(ctx, q, args.FromAccountID, -args.Amount, args.ToAccountID, args.Amount)
		} else {
			err = updateBalances(ctx, q, args.ToAccountID, args.Amount, args.FromAccountID, -args.Amount)
		}
		if err != nil {
			return fmt.Errorf("failed to update balances: %w", err)
		}

		result.FromAccount, err = q.GetAccount(ctx, args.FromAccountID)
		if err != nil {
			return fmt.Errorf("failed to fetch from_account: %w", err)
		}

		result.ToAccount, err = q.GetAccount(ctx, args.ToAccountID)
		if err != nil {
			return fmt.Errorf("failed to fetch to_account: %w", err)
		}

		return nil
	})

	return result, err
}

// updateBalances updates balances of two accounts atomically
func updateBalances(ctx context.Context, q *Queries, acc1ID, amt1, acc2ID, amt2 int64) error {
	_, err := q.UpdateBalance(ctx, UpdateBalanceParams{
		AccountID: acc1ID,
		Amount:    amt1,
	})
	if err != nil {
		return err
	}

	_, err = q.UpdateBalance(ctx, UpdateBalanceParams{
		AccountID: acc2ID,
		Amount:    amt2,
	})

	return err
}
