package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	mockdb "github.com/HzTTT/simple_bank/db/mock"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name          string
	request       gin.H
	bulidStubs    func(store *mockdb.MockStore)
	checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	newRequest    func(testCase *TestCase) (request *http.Request, err error)
}

func TestMain(m *testing.M) {
    gin.SetMode(gin.TestMode)
    os.Exit(m.Run())
}

func runTestCases(t *testing.T, testCases []*TestCase) {
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			//mock store
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)

			//bulid stubs
			tc.bulidStubs(store)

			//new request
			request, err := tc.newRequest(tc)
			require.NoError(t, err)
			//new recorder as response
			recorder := httptest.NewRecorder()
			//start test server
			server := NewServer((store))
			//send request
			server.router.ServeHTTP(recorder, request)

			//check response
			tc.checkResponse(t, recorder)

		})
	}
}