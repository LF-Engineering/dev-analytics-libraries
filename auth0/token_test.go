package auth0

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LF-Engineering/dev-analytics-libraries/auth0/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetToken(t *testing.T) {
	t.Run("Test Normal Scenario", testNormalScenario)
	t.Run("Test Get Token without Cached token", testEmptyTokenCache)
	t.Run("Test Expired Token", testExpiredToken)
	t.Run("Test Generating More Than Token within one Hour", testGeneratingTwoTokensWithinHour)
	t.Run("Test with Get Last token Request Date Error", testTokenWithGetLastRequestDateError)
}

func testNormalScenario(t *testing.T) {
	// arrange
	httpClientMock := &mocks.HTTPClientProvider{}
	esClientMock := &mocks.ESClientProvider{}
	slackClientMock := &mocks.SlackProvider{}
	tokenRes := `{
"hits": {
"hits": [{"_index":"","_type":"","_id": "","_score":0,"_source":{"name": "", "token":"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlJFWTNSVGcyTTBVMk5URTRSamcwTUVRNU9VSkZRMFpGTVVGRVJUaEVRVVF5TTBZeVF6QXhOQSJ9.eyJpc3MiOiJodHRwczovL2xpbnV4Zm91bmRhdGlvbi1kZXYuYXV0aDAuY29tLyIsInN1YiI6IlhpRFNFeUxRZ0dNM0RZWEJ1SVJYVGNCVDVMaHpxMzBRQGNsaWVudHMiLCJhdWQiOiJodHRwczovL2FwaS1ndy5kZXYucGxhdGZvcm0ubGludXhmb3VuZGF0aW9uLm9yZy8iLCJpYXQiOjE2MTQxMzE5MDcsImV4cCI6MTYxNDE1MzUwNywiYXpwIjoiWGlEU0V5TFFnR00zRFlYQnVJUlhUY0JUNUxoenEzMFEiLCJzY29wZSI6ImFjY2VzczphcGkiLCJndHkiOiJjbGllbnQtY3JlZGVudGlhbHMifQ.hXv-K9wtLN59GjIsJ4cwb3Tu7hk0jbRdvuETw-1r5N3WPOzDsbNbJNb8TXnRQgsO4u0jMyTjeXyzrJ0EQuOEL1M8yMsgZdIm5CWNnwW6Gqfie_mNevSyh2bIzwFi1CDW_9D9VIypiAUWBuOtBUqZvBIkbiDKG4nAnnRANtDgPBLTxrPfZyc7T9-HvOY5JERMt6YFH-4rx1PUMpBQvknMV08eJsgCcwU0BE7u4rc-vyFjJZJkFe7XQTOOGQnDZLpsHzQYiyRbZXpDsPrFFYzl0K0ZI7NlMF3ilKVSdinZ7ufBAFEjVrpESfqBL8yHkGxCKdkl7CqzNKvqOE2kodMGjw"}}]
}}`
	lastTokenRes := `{
"hits": {
"hits": [{"_index":"","_type":"","_id": "","_score":0,"_source":{"date": "2021-02-17T19:13:43.831197652Z"}}]
}}`

	esClientMock.On("Search", "auth0-token-cache-test",
		map[string]interface{}{"query": map[string]interface{}{"term": map[string]interface{}{"_id": "token"}}, "size": 1}).Return([]byte(tokenRes), nil)
	esClientMock.On("Search", "last-auth0-token-request-test",
		map[string]interface{}{"query": map[string]interface{}{"term": map[string]interface{}{"_id": "last-token-date"}}, "size": 1}).Return([]byte(lastTokenRes), nil)
	esClientMock.On("CreateDocument", "last-auth0-token-request-test", "last-token-date", mock.Anything).Return(nil, nil)
	esClientMock.On("CreateDocument", "auth0-token-cache-test", "token", mock.Anything).Return(nil, nil)

	genPayload := `{"audience":"","client_id":"","client_secret":"","grant_type":""}`
	genRes := &Resp{AccessToken: "newToken", Scope: "", ExpiresIn: 100, TokenType: "jwt"}
	genResJSON, _ := json.Marshal(genRes)
	httpClientMock.On("Request", "localhost", "POST", mock.Anything, []byte(genPayload), mock.Anything).Return(200, genResJSON, nil)

	// act
	srv, err := NewAuth0Client("",
		"",
		"",
		"test",
		"",
		"",
		"xh02agyyaqaj07et5g0uatt15em23j7v",
		"",
		"https://linuxfoundation-dev.auth0.com",
		"",
		httpClientMock,
		esClientMock,
		slackClientMock)
	if err != nil {
		t.Error(err)
	}

	_, err = srv.GetToken()

	//assert
	assert.NoError(t, err)
}

func testExpiredToken(t *testing.T) {
	// arrange
	httpClientMock := &mocks.HTTPClientProvider{}
	esClientMock := &mocks.ESClientProvider{}
	slackClientMock := &mocks.SlackProvider{}
	tokenRes := `{
"hits": {
"hits": [{"_index":"","_type":"","_id": "","_score":0,"_source":{"name": "", "token":"eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjE1ODIwMzk0OTV9.GK_gIJg4mO_8-vfJAkNGKIU4MC1oCYjsJbKidnQuw5Y"}}]
}}`
	lastTokenRes := `{
"hits": {
"hits": [{"_index":"","_type":"","_id": "","_score":0,"_source":{"date": "2021-02-17T19:13:43.831197652Z"}}]
}}`

	esClientMock.On("Search", "auth0-token-cache-test",
		map[string]interface{}{"query": map[string]interface{}{"term": map[string]interface{}{"_id": "token"}}, "size": 1}).Return([]byte(tokenRes), nil)
	esClientMock.On("Search", "last-auth0-token-request-test",
		map[string]interface{}{"query": map[string]interface{}{"term": map[string]interface{}{"_id": "last-token-date"}}, "size": 1}).Return([]byte(lastTokenRes), nil)
	esClientMock.On("CreateDocument", "last-auth0-token-request-test", "last-token-date", mock.Anything).Return(nil, nil)
	esClientMock.On("CreateDocument", "auth0-token-cache-test", "token", mock.Anything).Return(nil, nil)

	genPayload := `{"audience":"","client_id":"","client_secret":"","grant_type":""}`
	genRes := &Resp{AccessToken: "eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjc5MjUwMTYyOTV9.gPG_YA_q7An0tNtFYMEQvXJ--B-nP07UbYshQljrdMc", Scope: "", ExpiresIn: 100, TokenType: "jwt"}
	genResJSON, _ := json.Marshal(genRes)
	httpClientMock.On("Request", "localhost", "POST", mock.Anything, []byte(genPayload), mock.Anything).Return(200, genResJSON, nil)
	slackClientMock.On("SendText", "created token is not valid").Return(nil)

	// act
	srv, err := NewAuth0Client("",
		"",
		"",
		"test",
		"",
		"",
		"",
		"",
		"localhost",
		"xh02agyyaqaj07et5g0uatt15em23j7v",
		httpClientMock,
		esClientMock,
		slackClientMock)
	if err != nil {
		t.Error(err)
	}

	_, err = srv.GetToken()

	//assert
	assert.NoError(t, err)
}

func testGeneratingTwoTokensWithinHour(t *testing.T) {
	// arrange
	httpClientMock := &mocks.HTTPClientProvider{}
	esClientMock := &mocks.ESClientProvider{}
	slackClientMock := &mocks.SlackProvider{}
	tokenRes := `{
"hits": {
"hits": [{"_index":"","_type":"","_id": "","_score":0,"_source":{"name": "", "token":"eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjE1ODIwMzk0OTV9.GK_gIJg4mO_8-vfJAkNGKIU4MC1oCYjsJbKidnQuw5Y"}}]
}}`
	lastDate := time.Now().UTC().Format(time.RFC3339)
	lastTokenRes := fmt.Sprintf(`{
"hits": {
"hits": [{"_index":"","_type":"","_id": "","_score":0,"_source":{"date": "%s"}}]
}}`, lastDate)

	esClientMock.On("Search", "auth0-token-cache-test",
		map[string]interface{}{"query": map[string]interface{}{"term": map[string]interface{}{"_id": "token"}}, "size": 1}).Return([]byte(tokenRes), nil)
	esClientMock.On("Search", "last-auth0-token-request-test",
		map[string]interface{}{"query": map[string]interface{}{"term": map[string]interface{}{"_id": "last-token-date"}}, "size": 1}).Return([]byte(lastTokenRes), nil)
	esClientMock.On("CreateDocument", "last-auth0-token-request-test", "last-token-date", mock.Anything).Return(nil, nil)
	esClientMock.On("CreateDocument", "auth0-token-cache-test", "token", mock.Anything).Return(nil, nil)

	genPayload := `{"audience":"","client_id":"","client_secret":"","grant_type":""}`
	genRes := &Resp{AccessToken: "newToken", Scope: "", ExpiresIn: 100, TokenType: "jwt"}
	genResJSON, _ := json.Marshal(genRes)
	httpClientMock.On("Request", "localhost", "POST", mock.Anything, []byte(genPayload), mock.Anything).Return(200, genResJSON, nil)

	// act
	srv, err := NewAuth0Client("",
		"",
		"",
		"test",
		"",
		"",
		"",
		"",
		"localhost",
		"xh02agyyaqaj07et5g0uatt15em23j7v",
		httpClientMock,
		esClientMock,
		slackClientMock)
	if err != nil {
		t.Error(err)
	}

	_, err = srv.GetToken()

	//assert
	assert.Equal(t, errors.New("can not request more than one token within the same hour"), err)
}

func testEmptyTokenCache(t *testing.T) {
	// arrange
	httpClientMock := &mocks.HTTPClientProvider{}
	esClientMock := &mocks.ESClientProvider{}
	slackClientMock := &mocks.SlackProvider{}

	esClientMock.On("Search", "auth0-token-cache-test",
		map[string]interface{}{"query": map[string]interface{}{"term": map[string]interface{}{"_id": "token"}}, "size": 1}).Return(nil, errors.New("not found"))
	esClientMock.On("Search", "last-auth0-token-request-test",
		map[string]interface{}{"query": map[string]interface{}{"term": map[string]interface{}{"_id": "last-token-date"}}, "size": 1}).Return(nil, errors.New("index doesn't exist"))
	esClientMock.On("CreateDocument", "last-auth0-token-request-test", "last-token-date", mock.Anything).Return(nil, nil)
	esClientMock.On("CreateDocument", "auth0-token-cache-test", "token", mock.Anything).Return(nil, nil)

	genPayload := `{"audience":"","client_id":"","client_secret":"","grant_type":""}`
	genRes := &Resp{AccessToken: "newToken", Scope: "", ExpiresIn: 100, TokenType: "jwt"}
	genResJSON, _ := json.Marshal(genRes)
	httpClientMock.On("Request", "localhost", "POST", mock.Anything, []byte(genPayload), mock.Anything).Return(200, genResJSON, nil)
	slackClientMock.On("SendText", "created token is not valid").Return(nil)
	slackClientMock.On("SendText", "not found").Return(nil)

	// act
	srv, err := NewAuth0Client("",
		"",
		"",
		"test",
		"",
		"",
		"",
		"",
		"localhost",
		"xh02agyyaqaj07et5g0uatt15em23j7v",
		httpClientMock,
		esClientMock,
		slackClientMock)
	if err != nil {
		t.Error(err)
	}

	_, err = srv.GetToken()

	//assert
	assert.Equal(t, errors.New("created token is not valid"), err)
}

func testTokenWithGetLastRequestDateError(t *testing.T) {
	// arrange
	httpClientMock := &mocks.HTTPClientProvider{}
	esClientMock := &mocks.ESClientProvider{}
	slackClientMock := &mocks.SlackProvider{}

	esClientMock.On("Search", "auth0-token-cache-test",
		map[string]interface{}{"query": map[string]interface{}{"term": map[string]interface{}{"_id": "token"}}, "size": 1}).Return(nil, errors.New("not found"))
	esClientMock.On("Search", "last-auth0-token-request-test",
		map[string]interface{}{"query": map[string]interface{}{"term": map[string]interface{}{"_id": "last-token-date"}}, "size": 1}).Return(nil, errors.New("es is down"))
	esClientMock.On("CreateDocument", "last-auth0-token-request-test", "last-token-date", mock.Anything).Return(nil, nil)
	esClientMock.On("CreateDocument", "auth0-token-cache-test", "token", mock.Anything).Return(nil, nil)

	genPayload := `{"audience":"","client_id":"","client_secret":"","grant_type":""}`
	genRes := &Resp{AccessToken: "newToken", Scope: "", ExpiresIn: 100, TokenType: "jwt"}
	genResJSON, _ := json.Marshal(genRes)
	httpClientMock.On("Request", "localhost", "POST", mock.Anything, []byte(genPayload), mock.Anything).Return(200, genResJSON, nil)
	slackClientMock.On("SendText", "not found").Return(nil)

	// act
	srv, err := NewAuth0Client("",
		"",
		"",
		"test",
		"",
		"",
		"",
		"",
		"localhost",
		"xh02agyyaqaj07et5g0uatt15em23j7v",
		httpClientMock,
		esClientMock,
		slackClientMock)
	if err != nil {
		t.Error(err)
	}

	_, err = srv.GetToken()

	//assert
	assert.Equal(t, errors.New("es is down"), err)
}
