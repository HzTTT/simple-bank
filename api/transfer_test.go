package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/HzTTT/simple_bank/db/mock"
	db "github.com/HzTTT/simple_bank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestTransferAPI(t *testing.T) {
	user1, _ := randomUser(t)
	account1 := randomAccount(user1.Username)
	user2, _ := randomUser(t)
	account2 := randomAccount(user2.Username)
	account1.Currency = "USD"
	account2.Currency = "USD"
	amount := int64(10)
	toEntry := db.Entry{
		ID: 1,
		AccountID: account2.ID,
		Amount: amount,
	}
	fromEntry := db.Entry{
		ID: 2,
		AccountID: account1.ID,
		Amount: -1 * amount,
	}
	transfer := db.Transfer{
		FromAccountID: account1.ID,
		ToAccountID: account2.ID,
		Amount: amount,
	}
	transferResult := db.TransferTxResult{
		FromEntry: fromEntry,
		ToEntry: toEntry,
		Transfer: transfer,
		FromAccount: account1,
		ToAccount: account2,
	}
	newRequest := func(testCase *TestCase,server *Server) (request *http.Request, err error) {
		url := "/transfer"
		data, err := json.Marshal(testCase.request)
		require.NoError(t, err)
		request, err = http.NewRequest(http.MethodPost,url,bytes.NewReader(data)) 
		addAutgorization(t,request,server.tokenMaker,authorizationTypeBearer,user1.Username,time.Minute)
		return
	}

	testCases := []*TestCase{
		{
			name: "OK",
			request: gin.H{
				"from_account_id": account1.ID,
				"to_account_id": account2.ID,
				"amount": amount,
				"currency": "USD",
			},
			bulidStubs: func(store *mockdb.MockStore) {
				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID: account2.ID,
					Amount: amount,
				}
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(),gomock.Eq(account1.ID)).Times(1).Return(account1,nil),
					store.EXPECT().GetAccount(gomock.Any(),gomock.Eq(account2.ID)).Times(1).Return(account2,nil),
					store.EXPECT().TransferTx(gomock.Any(),gomock.Eq(arg)).Times(1).Return(transferResult,nil),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t,http.StatusOK,recorder.Code)
				require.NotEmpty(t,recorder.Body)
				requireBodyMatchTransferResult(t,transferResult,recorder.Body)
			},
			newRequest: newRequest,
		},
		{
			name: "FromAccountNotFound",
			request: gin.H{
				"from_account_id": account1.ID,
				"to_account_id": account2.ID,
				"amount": amount,
				"currency": "USD",
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(),gomock.Eq(account1.ID)).Times(1).Return(db.Account{},sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t,http.StatusNotFound,recorder.Code)
			},
			newRequest: newRequest,
		},
		{
			name: "ToAccountNotFound",
			request: gin.H{
				"from_account_id": account1.ID,
				"to_account_id": account2.ID,
				"amount": amount,
				"currency": "USD",
			},
			bulidStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(),gomock.Eq(account1.ID)).Times(1).Return(account1,nil),
					store.EXPECT().GetAccount(gomock.Any(),gomock.Eq(account2.ID)).Times(1).Return(db.Account{},sql.ErrNoRows),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t,http.StatusNotFound,recorder.Code)
			},
			newRequest: newRequest,
		},
		{
			name: "FromAccountCurrencyMismatch",
			request: gin.H{
				"from_account_id": account1.ID,
				"to_account_id": account2.ID,
				"amount": amount,
				"currency": "EUR",
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(),gomock.Eq(account1.ID)).Times(1).Return(account1,nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t,http.StatusBadRequest,recorder.Code)
			},
			newRequest: newRequest,
		},
	}

	runTestCases(t,testCases)
	
}

func requireBodyMatchTransferResult(t *testing.T, transferResult db.TransferTxResult, body *bytes.Buffer) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t,err)
	var gotResult db.TransferTxResult
	err = json.Unmarshal(data,&gotResult)
	require.NoError(t,err)
	require.Equal(t,transferResult,gotResult)
}