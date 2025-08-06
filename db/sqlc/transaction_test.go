package db

import (
	"context"
	"testing"

	"github.com/chandiniv1/transfers-system/util"
	"github.com/stretchr/testify/require"
)

func createRandomTransaction(t *testing.T) Transaction {
	source := createRandomAccount(t)
	dest := createRandomAccount(t)

	arg := CreateTransactionParams{
		SourceAccountID:      source.AccountID,
		DestinationAccountID: dest.AccountID,
		Amount:               util.RandomMoney(),
	}

	tx, err := testQueries.CreateTransaction(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, tx)

	require.Equal(t, arg.SourceAccountID, tx.SourceAccountID)
	require.Equal(t, arg.DestinationAccountID, tx.DestinationAccountID)
	require.Equal(t, arg.Amount, tx.Amount)
	require.NotZero(t, tx.ID)
	require.NotZero(t, tx.CreatedAt)

	return tx
}

func TestCreateTransaction(t *testing.T) {
	createRandomTransaction(t)
}

func TestCreateTransactionZeroAmountFails(t *testing.T) {
	source := createRandomAccount(t)
	dest := createRandomAccount(t)

	arg := CreateTransactionParams{
		SourceAccountID:      source.AccountID,
		DestinationAccountID: dest.AccountID,
		Amount:               0,
	}

	tx, err := testQueries.CreateTransaction(context.Background(), arg)
	require.Error(t, err)
	require.Empty(t, tx)
	require.Contains(t, err.Error(), "check constraint")
}

func TestCreateTransactionNegativeAmountFails(t *testing.T) {
	source := createRandomAccount(t)
	dest := createRandomAccount(t)

	arg := CreateTransactionParams{
		SourceAccountID:      source.AccountID,
		DestinationAccountID: dest.AccountID,
		Amount:               -util.RandomMoney(),
	}

	tx, err := testQueries.CreateTransaction(context.Background(), arg)
	require.Error(t, err)
	require.Empty(t, tx)
	require.Contains(t, err.Error(), "check constraint")
}

func TestCreateTransactionSameAccountFails(t *testing.T) {
	account := createRandomAccount(t)

	arg := CreateTransactionParams{
		SourceAccountID:      account.AccountID,
		DestinationAccountID: account.AccountID,
		Amount:               util.RandomMoney(),
	}

	tx, err := testQueries.CreateTransaction(context.Background(), arg)
	require.Error(t, err)
	require.Empty(t, tx)
	require.Contains(t, err.Error(), "check constraint")
}
