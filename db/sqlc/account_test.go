package db

import (
	"context"
	"testing"
	"time"

	"github.com/chandiniv1/transfers-system/util"
	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func createRandomAccount(t *testing.T) Account {
	arg := CreateAccountParams{
		AccountID: util.RandomAccountID(),
		Balance:   util.RandomMoney(),
		Currency:  util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.AccountID, account.AccountID)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)
	require.NotZero(t, account.CreatedAt)

	return account
}

func TestGetAccount(t *testing.T) {
	account1 := createRandomAccount(t)
	account2, err := testQueries.GetAccount(context.Background(), account1.AccountID)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.AccountID, account2.AccountID)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt.Time, account2.CreatedAt.Time, time.Second)
}

func TestGetAccountNotFound(t *testing.T) {
	// Test with non-existent account ID
	nonExistentID := int64(99999)
	account, err := testQueries.GetAccount(context.Background(), nonExistentID)

	require.Error(t, err)
	require.Empty(t, account)
}

func TestUpdateBalance(t *testing.T) {
	account1 := createRandomAccount(t)

	amount := util.RandomMoney()
	arg := UpdateBalanceParams{
		Amount:    amount,
		AccountID: account1.AccountID,
	}

	account2, err := testQueries.UpdateBalance(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.AccountID, account2.AccountID)
	require.Equal(t, account1.Balance+amount, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt.Time, account2.CreatedAt.Time, time.Second)
}

func TestUpdateBalanceSubtract(t *testing.T) {
	account1 := createRandomAccount(t)

	maxSubtraction := account1.Balance / 2
	if maxSubtraction == 0 {
		maxSubtraction = 1
	}

	amount := -maxSubtraction
	arg := UpdateBalanceParams{
		Amount:    amount,
		AccountID: account1.AccountID,
	}

	account2, err := testQueries.UpdateBalance(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.AccountID, account2.AccountID)
	require.Equal(t, account1.Balance+amount, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.True(t, account2.Balance >= 0)
}

func TestListAccounts(t *testing.T) {
	var createdAccounts []Account
	for i := 0; i < 10; i++ {
		account := createRandomAccount(t)
		createdAccounts = append(createdAccounts, account)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 0,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.NotZero(t, account.AccountID)
		require.NotEmpty(t, account.Currency)
		require.NotZero(t, account.CreatedAt)
	}
}

func TestListAccountsWithOffset(t *testing.T) {
	for i := 0; i < 15; i++ {
		createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.LessOrEqual(t, len(accounts), 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}

func TestCreateAccountZeroBalance(t *testing.T) {
	arg := CreateAccountParams{
		AccountID: util.RandomAccountID(),
		Balance:   0,
		Currency:  util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Equal(t, int64(0), account.Balance)
}

func TestCreateAccountNegativeBalanceFails(t *testing.T) {
	arg := CreateAccountParams{
		AccountID: util.RandomAccountID(),
		Balance:   -util.RandomMoney(),
		Currency:  util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.Error(t, err)
	require.Empty(t, account)
	require.Contains(t, err.Error(), "accounts_balance_check")
}
