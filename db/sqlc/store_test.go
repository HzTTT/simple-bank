package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := CreateAccount(t)
	account2 := CreateAccount(t)
	amount := int64(10)

	fmt.Println(">>Before:", account1.Balance, account2.Balance)

	errs := make(chan error)
	results := make(chan TransferTxResult)
	n := 10
	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			results <- result
			errs <- err
		}()
	}

	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		result := <-results
		require.NotEmpty(t, result)
		err := <-errs
		require.NoError(t, err)

		//check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		//check to entry
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		//check from entry
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)
		
		//check account balance
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		fmt.Println(">>After:", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// Check the account balance
	fromAccount, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, fromAccount)
	require.Equal(t, account1.ID, fromAccount.ID)
	require.Equal(t, account1.Owner, fromAccount.Owner)
	require.Equal(t, account1.Balance-int64(n)*amount, fromAccount.Balance)

	toAccount, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, toAccount)
	require.Equal(t, account2.ID, toAccount.ID)
	require.Equal(t, account2.Owner, toAccount.Owner)
	require.Equal(t, account2.Balance+int64(n)*amount, toAccount.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := CreateAccount(t)
	account2 := CreateAccount(t)
	amount := int64(10)

	fmt.Println(">>Before:", account1.Balance, account2.Balance)

	errs := make(chan error)

	n := 10
	for i := 0; i < n; i++ {

		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}
		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {

		err := <-errs
		require.NoError(t, err)


	}

	// Check the account balance
	fromAccount, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, fromAccount)
	require.Equal(t, account1.ID, fromAccount.ID)
	require.Equal(t, account1.Owner, fromAccount.Owner)
	require.Equal(t, account1.Balance, fromAccount.Balance)

	toAccount, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, toAccount)
	require.Equal(t, account2.ID, toAccount.ID)
	require.Equal(t, account2.Owner, toAccount.Owner)
	require.Equal(t, account2.Balance, toAccount.Balance)
}