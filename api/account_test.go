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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		bulidStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
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
			accountID: account.ID,
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(),gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{},sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound,recorder.Code)
			},
		},
		{
			name: "InternalError",
			accountID: account.ID,
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(),gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{},sql.ErrConnDone)	
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t,http.StatusInternalServerError,recorder.Code)
			},
		},
		{
			name: "InvalidID",
			accountID: -1,
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(),gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t,http.StatusBadRequest,recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name,func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)

			//bulid stubs
			tc.bulidStubs(store)

			//start test server and send request
			server := NewServer((store))
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/account/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)

			//check response
			tc.checkResponse(t,recorder)

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
