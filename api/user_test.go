package api

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"

	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/HzTTT/simple_bank/db/mock"
	db "github.com/HzTTT/simple_bank/db/sqlc"
	"github.com/HzTTT/simple_bank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"

	"github.com/stretchr/testify/require"
)

type eqMatcherUserArg struct {
	arg      db.CreateUserParams
	password string
}

func (e eqMatcherUserArg) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}
	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqMatcherUserArg) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqUserArg(x db.CreateUserParams, pasaword string) gomock.Matcher {
	return eqMatcherUserArg{arg: x, password: pasaword}
}

func TestCreateUserAPI(t *testing.T) {

	user, password := randomUser(t)

	newRequest := func(testCase *TestCase,server *Server) (request *http.Request, err error) {
		body, err := json.Marshal(testCase.request)
		if err != nil {
			return nil, err
		}
		request = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		return request, nil
	}

	testCases := []*TestCase{
		{
			name: "OK",
			request: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqUserArg(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, user, recorder.Body)
			},
			newRequest: newRequest,
		},
		{
			name: "InternalError",
			request: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
			newRequest: newRequest,
		},
		{
			name: "DuplicateUsername",
			request: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
			newRequest: newRequest,
		},
		{
			name: "InvalidUsername",
			request: gin.H{
				"username":  "invalid-user#1",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			newRequest: newRequest,
		},
		{
			name: "InvalidEmail",
			request: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     "invalid-email",
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
			newRequest: newRequest,
		},
		{
			name: "TooShortPassword",
			request: gin.H{
				"username":  user.Username,
				"password":  "123",
				"full_name": user.FullName,
				"email":     user.Email,
			},
			bulidStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
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

func TestLoginUserAPI(t *testing.T) {
	user, password := randomUser(t)

	newRequest := func(testCase *TestCase,server *Server) (request *http.Request, err error) {
		body, err := json.Marshal(testCase.request)
		if err != nil {
			return nil, err
		}
		request = httptest.NewRequest(http.MethodPost,"/user/login",bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		return request, nil
	}
	testCases := []*TestCase{
		{
			name: "OK",
			request: gin.H{
				"username":  user.Username,
				"password":  password,
			},
			bulidStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetUser(gomock.Any(),gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchLoginResponse(t, user, recorder.Body)
			},
			newRequest: newRequest,
		},
	}

	runTestCases(t,testCases)
}

func requireBodyMatchUser(t *testing.T, user db.User, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser userResponse
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	user = db.User{
		Username:       util.RandOwner(),
		FullName:       util.RandOwner(),
		Email:          util.RandomEmail(),
		HashedPassword: hashedPassword,
	}
	return user, password
}
func requireBodyMatchLoginResponse(t *testing.T, user db.User, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var loginnRespon loginUserResponse
	err = json.Unmarshal(data, &loginnRespon)
	require.NoError(t, err)

	require.Equal(t, user.Username, loginnRespon.User.Username)
	require.Equal(t, user.FullName, loginnRespon.User.FullName)
	require.Equal(t, user.Email, loginnRespon.User.Email)
	require.NotZero(t,loginnRespon.AccessToken)
	fmt.Println(loginnRespon.AccessToken)
}