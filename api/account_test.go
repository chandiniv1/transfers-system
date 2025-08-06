package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/chandiniv1/transfers-system/db/mock"
	db "github.com/chandiniv1/transfers-system/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCreateAccountAPI(t *testing.T) {
	account := db.Account{
		AccountID: 1,
		Currency:  "USD",
		Balance:   1000,
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
				"account_id": account.AccountID,
				"currency":   account.Currency,
				"balance":    account.Balance,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "Invalid Input",
			body: gin.H{
				"currency": "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "Internal Error",
			body: gin.H{
				"account_id": 1,
				"currency":   "USD",
				"balance":    1000,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, errors.New("db error"))
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name: "Unique Violation Error",
			body: gin.H{
				"account_id": 1,
				"currency":   "USD",
				"balance":    1000,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, &pq.Error{Code: pq.ErrorCode("23505")})
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, rec.Code)
			},
		},
		{
			name: "Foreign Key Violation Error",
			body: gin.H{
				"account_id": 1,
				"currency":   "USD",
				"balance":    1000,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, &pq.Error{Code: pq.ErrorCode("23503")})
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, rec.Code)
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

			url := "/accounts"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(rec, req)
			tc.checkResponse(rec)
		})
	}
}

func TestGetAccountAPI(t *testing.T) {
	account := db.Account{
		AccountID: 1,
		Currency:  "USD",
		Balance:   1000,
	}

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rec *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.AccountID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.AccountID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name:      "Not Found",
			accountID: account.AccountID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.AccountID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
		{
			name:      "Internal Error",
			accountID: account.AccountID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.AccountID)).
					Times(1).
					Return(db.Account{}, errors.New("db error"))
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name:      "Invalid ID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
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

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(rec, req)
			tc.checkResponse(rec)
		})
	}
}

func TestListAccountsAPI(t *testing.T) {
	accounts := []db.Account{
		{
			AccountID: 1,
			Currency:  "USD",
			Balance:   1000,
		},
		{
			AccountID: 2,
			Currency:  "EUR",
			Balance:   2000,
		},
		{
			AccountID: 3,
			Currency:  "CAD",
			Balance:   1500,
		},
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rec *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "?page_id=1&page_size=5",
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Limit:  5,
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name:  "Internal Error",
			query: "?page_id=1&page_size=5",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Account{}, errors.New("db error"))
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name:  "Invalid Page ID",
			query: "?page_id=0&page_size=5",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name:  "Invalid Page Size",
			query: "?page_id=1&page_size=15",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name:  "Missing Query Parameters",
			query: "?page_id=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
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

			url := "/accounts" + tc.query
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(rec, req)
			tc.checkResponse(rec)
		})
	}
}
