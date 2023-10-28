package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/HzTTT/simple_bank/db/mock"
	db "github.com/HzTTT/simple_bank/db/sqlc"
	"github.com/HzTTT/simple_bank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	newRequest := func(testCase *TestCase,server *Server) (request *http.Request, err error) {
		url := fmt.Sprintf("/account/%d", testCase.request["accountID"])
		request, err = http.NewRequest(http.MethodGet, url, nil)
		addAutgorization(t,request,server.tokenMaker,authorizationTypeBearer,user.Username,time.Minute)
		return
	}
	testCases := []*TestCase{
		{ 
			name: "OK",
			request: gin.H{
				"accountID": account.ID,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
			newRequest: newRequest,
		},
		{ 
			name: "UnauthorizedUser",
			request: gin.H{
				"accountID": account.ID,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
			newRequest: func(testCase *TestCase, server *Server) (request *http.Request, err error) {
				url := fmt.Sprintf("/account/%d", testCase.request["accountID"])
				request, err = http.NewRequest(http.MethodGet, url, nil)
				addAutgorization(t,request,server.tokenMaker,authorizationTypeBearer,"unauthorzed_user",time.Minute)
				return
			},
		},
		{ 
			name: "NoAuthorization",
			request: gin.H{
				"accountID": account.ID,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
			newRequest: func(testCase *TestCase, server *Server) (request *http.Request, err error) {
				url := fmt.Sprintf("/account/%d", testCase.request["accountID"])
				request, err = http.NewRequest(http.MethodGet, url, nil)
				return
			},
		},
		{
			name: "NotFound",
			request: gin.H{
				"accountID": account.ID,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
			newRequest: newRequest,
		},
		{
			name: "InternalError",
			request: gin.H{
				"accountID": account.ID,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			newRequest: newRequest,
		},
		{
			name: "InvalidID",
			request: gin.H{
				"accountID": 0,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			newRequest: newRequest,
		},
	}

	runTestCases(t, testCases)
}

func TestCreateAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	newRequest := func(testCase *TestCase,server *Server) (request *http.Request, err error) {
		data, err := json.Marshal(testCase.request)
		require.NoError(t, err)
		url := "/account"
		request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
		addAutgorization(t,request,server.tokenMaker,authorizationTypeBearer,user.Username,time.Minute)
		return
	}
	testCases := []*TestCase{
		{
			name: "OK",
			request: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(
						gomock.Any(),
						gomock.Eq(db.CreateAccountParams{
							Owner:    account.Owner,
							Currency: account.Currency,
							Balance:  0,
						})).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
			newRequest: newRequest,
		},
		{
			name: "InvalidCurrency",
			request: gin.H{
				"ownwe":    account.Owner,
				"currency": "HK",
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			newRequest: newRequest,
		},
		{
			name: "InternalError",
			request: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(
						gomock.Any(),
						gomock.Eq(db.CreateAccountParams{
							Owner:    account.Owner,
							Currency: account.Currency,
							Balance:  int64(0),
						})).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			newRequest: newRequest,
		},
	}

	runTestCases(t, testCases)
}

func TestListAcoountsAPI(t *testing.T) {
	user, _ := randomUser(t)
	

	pageSize := int32(5)
	pageID := int32(5)
	accounts := make([]db.Account, pageSize)
	for i := int32(0); i < pageSize; i++ {
		accounts[i] = randomAccount(user.Username)
	}
	newRequest := func(testCase *TestCase,server *Server) (request *http.Request, err error) {
		url := fmt.Sprintf("/account?page_id=%d&page_size=%d", testCase.request["page_id"], testCase.request["page_size"])
		request, err = http.NewRequest(http.MethodGet, url, nil)
		addAutgorization(t,request,server.tokenMaker,authorizationTypeBearer,user.Username,time.Minute)

		/* request, err = http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		// Add query parameters to request URL
		q := request.URL.Query()
		q.Add("page_id", fmt.Sprintf("%d", testCase.request["page_id"]))
		q.Add("page_size", fmt.Sprintf("%d", testCase.request["page_size"]))
		request.URL.RawQuery = q.Encode() */

		return
	}

	testCases := []*TestCase{
		{
			name: "OK",
			request: gin.H{
				"page_id":   pageID,
				"page_size": pageSize,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(
						gomock.Any(),
						gomock.Eq(db.ListAccountsParams{
							Offset: (pageID - 1) * pageSize,
							Limit:  pageSize,
							Owner: user.Username,
						})).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
			newRequest: newRequest,
		},
		{
			name: "InvalidID",
			request: gin.H{
				"page_id":   int32(0),
				"page_size": pageSize,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			newRequest: newRequest,
		},
		{
			name: "InternalError",
			request: gin.H{
				"page_id":   pageID,
				"page_size": pageSize,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(
						gomock.Any(),
						gomock.Eq(db.ListAccountsParams{
							Offset: (pageID - 1) * pageSize,
							Limit:  pageSize,
							Owner: user.Username,
						})).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			newRequest: newRequest,
		},
	}

	runTestCases(t, testCases)
}

func randomAccount(owmer string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owmer,
		Balance:  util.RandMoney(),
		Currency: util.RandCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}
