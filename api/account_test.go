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

	mockdb "github.com/HzTTT/simple_bank/db/mock"
	db "github.com/HzTTT/simple_bank/db/sqlc"
	"github.com/HzTTT/simple_bank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name          string
	request       gin.H
	bulidStubs    func(store *mockdb.MockStore)
	checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
}

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []testCase{
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
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc, func() (request *http.Request, err error) {
				url := fmt.Sprintf("/account/%d", tc.request["accountID"])
				request, err = http.NewRequest(http.MethodGet, url, nil)
				return
			})

		})
	}
}

func TestCreateAccountAPI(t *testing.T) {
	account := randomAccount()
	testCases := []testCase{
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
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc, func() (request *http.Request, err error) {
				data, err := json.Marshal(tc.request)
				require.NoError(t, err)
				url := "/account"
				request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
				return
			})
		})
	}
}

func TestListAcoountsAPI(t *testing.T) {
	pageSize := int32(5)
	pageID := int32(5)
	accounts := make([]db.Account, pageSize)
	for i := int32(0); i < pageSize; i++ {
		accounts[i] = randomAccount()
	}

	testCases := []testCase{
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
							Limit:  int32(pageSize),
						})).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc, func() (request *http.Request, err error) {
				url := fmt.Sprintf("/account?page_id=%d&page_size=%d", tc.request["page_id"], tc.request["page_size"])
				request, err = http.NewRequest(http.MethodGet, url, nil)
				/* 
				request, err = http.NewRequest(http.MethodGet, url, nil)
				require.NoError(t, err)

				// Add query parameters to request URL
				q := request.URL.Query()
				q.Add("page_id", fmt.Sprintf("%d", tc.request["page_id"]))
				q.Add("page_size", fmt.Sprintf("%d", tc.request["page_size"]))
				request.URL.RawQuery = q.Encode()
				 */
				return
			})

		})
	}

}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandOwner(),
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

func runTestCase(t *testing.T, testCase testCase, newRequest func() (request *http.Request, err error)) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	//bulid stubs
	testCase.bulidStubs(store)

	//start test server and send request
	server := NewServer((store))
	recorder := httptest.NewRecorder()

	request, err := newRequest()
	require.NoError(t, err)
	server.router.ServeHTTP(recorder, request)

	//check response
	testCase.checkResponse(t, recorder)
}
