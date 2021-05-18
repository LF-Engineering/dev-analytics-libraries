package users

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/LF-Engineering/dev-analytics-libraries/users/mocks"
	"github.com/stretchr/testify/assert"

	json "github.com/json-iterator/go"
)

const (
	OKStatus = 200
)

var (
	httpClientProvider    = &mocks.HTTPClientProvider{}
	auth0ClientProvider   = &mocks.Auth0ClientProvider{}
	elasticClientProvider = &mocks.ESClientProvider{}
	slackClientProvider   = &mocks.SlackProvider{}
	userStruct            = &Client{
		"PLATFORM_USER_SERVICE_ENDPOINT",
		"ELASTIC_CACHE_URL",
		"ELASTIC_CACHE_USERNAME",
		"ELASTIC_CACHE_PASSWORD",
		"AUTH0_PROD_GRANT_TYPE",
		"AUTH0_PROD_CLIENT_ID",
		"AUTH0_PROD_CLIENT_SECRET",
		"AUTH0_PROD_AUDIENCE",
		"AUTH0_TOKEN_ENDPOINT",
		"test",
		httpClientProvider,
		auth0ClientProvider,
		elasticClientProvider,
		slackClientProvider,
	}

	token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI"
	email = "lgryglicki@cncf.io"
)

func TestList(t *testing.T) {
	pageSize := "100"
	offset := "1"

	buf := &bytes.Buffer{}
	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)
	listEndpoint := userStruct.UserBaseURL + "/users?email=" + url.QueryEscape(email) + "&pageSize=" + pageSize + "&offset=" + offset

	data := map[string]interface{}{
		"Data": []map[string]interface{}{
			{
				"Name":     "Łukasz Gryglicki",
				"Username": "lgryglicki",
				"Email":    "lgryglicki@cncf.io",
				"Emails": []map[string]interface{}{
					{
						"EmailAddress": "lgryglicki@cncf.io",
						"Active":       true,
						"IsDeleted":    false,
						"IsPrimary":    true,
						"IsVerified":   true,
					},
				},
			},
		},
	}

	_ = json.NewEncoder(buf).Encode(data)
	dataBytes, _ := ioutil.ReadAll(buf)

	auth0ClientProvider.On("GetToken").Return(token, nil)
	httpClientProvider.On("Request", listEndpoint, "GET", headers, []byte(nil), map[string]string(nil)).Return(OKStatus, dataBytes, nil)

	actualResponse, _ := userStruct.List(email, pageSize, offset)
	assert.Equal(t, "Łukasz Gryglicki", actualResponse.Data[0].Name)
	assert.Equal(t, "lgryglicki", actualResponse.Data[0].Username)
	assert.Equal(t, "lgryglicki@cncf.io", actualResponse.Data[0].Emails[0].EmailAddress)
	assert.Equal(t, true, actualResponse.Data[0].Emails[0].Active)
	assert.Equal(t, false, actualResponse.Data[0].Emails[0].IsDeleted)
	assert.Equal(t, true, actualResponse.Data[0].Emails[0].IsPrimary)
	assert.Equal(t, true, actualResponse.Data[0].Emails[0].IsVerified)
	assert.Equal(t, "lgryglicki@cncf.io", actualResponse.Data[0].Email)
}
