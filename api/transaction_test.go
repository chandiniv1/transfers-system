package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/chandiniv1/transfers-system/db/mock"
	db "github.com/chandiniv1/transfers-system/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestCreateTransferAPI(t *testing.T) {
	amount := int64(10)
	createdAt := pgtype.Timestamptz{}
	_ = createdAt.Scan(time.Now())

	fromAccount := db.Account{
		AccountID: 1,
		Currency:  "USD",
		Balance:   1000,
		CreatedAt: createdAt,
	}

	toAccount := db.Account{
		AccountID: 2,
		Currency:  "USD",
		Balance:   500,
		CreatedAt: createdAt,
	}

	transferResult := db.TransferTxResult{
		Transaction: db.Transaction{
			ID:                   1,
			SourceAccountID:      fromAccount.AccountID,
			DestinationAccountID: toAccount.AccountID,
			Amount:               amount,
			CreatedAt:            createdAt,
		},
		FromAccount: db.Account{
			AccountID: fromAccount.AccountID,
			Currency:  fromAccount.Currency,
			Balance:   fromAccount.Balance - amount,
			CreatedAt: fromAccount.CreatedAt,
		},
		ToAccount: db.Account{
			AccountID: toAccount.AccountID,
			Currency:  toAccount.Currency,
			Balance:   toAccount.Balance + amount,
			CreatedAt: toAccount.CreatedAt,
		},
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rec *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": fromAccount.AccountID,
				"to_account_id":   toAccount.AccountID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.AccountID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.AccountID)).Times(1).Return(toAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(transferResult, nil)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "Invalid JSON",
			body: gin.H{
				"from_account_id": "invalid",
				"to_account_id":   toAccount.AccountID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "Missing Required Fields",
			body: gin.H{
				"from_account_id": fromAccount.AccountID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "From Account Not Found",
			body: gin.H{
				"from_account_id": fromAccount.AccountID,
				"to_account_id":   toAccount.AccountID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.AccountID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.AccountID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
		{
			name: "From Account Internal Error",
			body: gin.H{
				"from_account_id": fromAccount.AccountID,
				"to_account_id":   toAccount.AccountID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.AccountID)).Times(1).Return(db.Account{}, errors.New("db error"))
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.AccountID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name: "From Account Currency Mismatch",
			body: gin.H{
				"from_account_id": fromAccount.AccountID,
				"to_account_id":   toAccount.AccountID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				wrongCurrencyAccount := fromAccount
				wrongCurrencyAccount.Currency = "EUR"
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.AccountID)).Times(1).Return(wrongCurrencyAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.AccountID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "To Account Not Found",
			body: gin.H{
				"from_account_id": fromAccount.AccountID,
				"to_account_id":   toAccount.AccountID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.AccountID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.AccountID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
		{
			name: "To Account Internal Error",
			body: gin.H{
				"from_account_id": fromAccount.AccountID,
				"to_account_id":   toAccount.AccountID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.AccountID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.AccountID)).Times(1).Return(db.Account{}, errors.New("db error"))
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name: "To Account Currency Mismatch",
			body: gin.H{
				"from_account_id": fromAccount.AccountID,
				"to_account_id":   toAccount.AccountID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				wrongCurrencyAccount := toAccount
				wrongCurrencyAccount.Currency = "EUR"
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.AccountID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.AccountID)).Times(1).Return(wrongCurrencyAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "TransferTx Internal Error",
			body: gin.H{
				"from_account_id": fromAccount.AccountID,
				"to_account_id":   toAccount.AccountID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.AccountID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.AccountID)).Times(1).Return(toAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(db.TransferTxResult{}, errors.New("transfer failed"))
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name: "Invalid Amount - Zero",
			body: gin.H{
				"from_account_id": fromAccount.AccountID,
				"to_account_id":   toAccount.AccountID,
				"amount":          0, // invalid amount
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "Invalid Account ID - Zero",
			body: gin.H{
				"from_account_id": 0,
				"to_account_id":   toAccount.AccountID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			rec := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/transactions"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(rec, req)
			tc.checkResponse(rec)
		})
	}
}

// func TestValidAccount(t *testing.T) {
// 	account := db.Account{
// 		AccountID: 1,
// 		Currency:  "USD",
// 		Balance:   1000,
// 		CreatedAt: time.Now(),
// 	}

// 	testCases := []struct {
// 		name          string
// 		accountID     int64
// 		currency      string
// 		buildStubs    func(store *mockdb.MockStore)
// 		checkResponse func(rec *httptest.ResponseRecorder)
// 		expectValid   bool
// 	}{
// 		{
// 			name:      "OK",
// 			accountID: account.AccountID,
// 			currency:  account.Currency,
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.AccountID)).Times(1).Return(account, nil)
// 			},
// 			checkResponse: func(rec *httptest.ResponseRecorder) {
// 				// This test won't write to recorder since it's called within transfer
// 				// We test this indirectly through transfer tests
// 			},
// 			expectValid: true,
// 		},
// 		{
// 			name:      "Account Not Found",
// 			accountID: account.AccountID,
// 			currency:  account.Currency,
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.AccountID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
// 			},
// 			checkResponse: func(rec *httptest.ResponseRecorder) {
// 				// This is tested through the transfer tests
// 			},
// 			expectValid: false,
// 		},
// 		{
// 			name:      "Internal Error",
// 			accountID: account.AccountID,
// 			currency:  account.Currency,
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.AccountID)).Times(1).Return(db.Account{}, errors.New("db error"))
// 			},
// 			checkResponse: func(rec *httptest.ResponseRecorder) {
// 				// This is tested through the transfer tests
// 			},
// 			expectValid: false,
// 		},
// 		{
// 			name:      "Currency Mismatch",
// 			accountID: account.AccountID,
// 			currency:  "EUR", // different currency
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.AccountID)).Times(1).Return(account, nil)
// 			},
// 			checkResponse: func(rec *httptest.ResponseRecorder) {
// 				// This is tested through the transfer tests
// 			},
// 			expectValid: false,
// 		},
// 	}

// 	// Note: validAccount is a helper function that's primarily tested through
// 	// the createTransfer tests above. These test cases document the behavior
// 	// but the actual testing happens in the transfer integration tests.
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			store := mockdb.NewMockStore(ctrl)
// 			tc.buildStubs(store)

// 			server := newTestServer(t, store)

// 			// Create a gin context for testing
// 			rec := httptest.NewRecorder()
// 			ctx, _ := gin.CreateTestContext(rec)

// 			// Test the validAccount function directly
// 			returnedAccount, valid := server.validAccount(ctx, tc.accountID, tc.currency)

// 			// Verify the validity matches expectation
// 			require.Equal(t, tc.expectValid, valid)

// 			if tc.expectValid {
// 				require.Equal(t, account.AccountID, returnedAccount.AccountID)
// 				require.Equal(t, account.Currency, returnedAccount.Currency)
// 			}
// 		})
// 	}
// }
