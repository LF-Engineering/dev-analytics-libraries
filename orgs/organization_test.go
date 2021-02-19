package orgs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/LF-Engineering/dev-analytics-libraries/orgs/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	OKStatus = 200
)

var (
	httpClientProvider    = &mocks.HTTPClientProvider{}
	auth0ClientProvider   = &mocks.Auth0ClientProvider{}
	elasticClientProvider = &mocks.ESClientProvider{}
	slackClientProvider   = &mocks.SlackProvider{}
	orgStruct             = &Org{
		os.Getenv("ORG_SERVICE_ENDPOINT"),
		os.Getenv("ELASTIC_CACHE_URL"),
		os.Getenv("ELASTIC_CACHE_USERNAME"),
		os.Getenv("ELASTIC_CACHE_PASSWORD"),
		os.Getenv("AUTH0_PROD_GRANT_TYPE"),
		os.Getenv("AUTH0_PROD_CLIENT_ID"),
		os.Getenv("AUTH0_PROD_CLIENT_SECRET"),
		os.Getenv("AUTH0_PROD_AUDIENCE"),
		os.Getenv("AUTH0_TOKEN_ENDPOINT"),
		"test",
		httpClientProvider,
		auth0ClientProvider,
		elasticClientProvider,
		"secret",
		slackClientProvider,
	}

	token   = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI"
	orgName = "linux"
)

func TestLookupOrganization(t *testing.T) {
	buf := &bytes.Buffer{}
	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)
	lookupEndpoint := orgStruct.OrgBaseURL + "/lookup?name=" + url.PathEscape(orgName)

	data := (map[string]string{
		"ID":      "v03fs-3",
		"Name":    "Linux Foundation, US",
		"Link":    "linuxfoundation.com",
		"LogoURL": "linuxfoundationlogo.com/logo.png",
	})

	_ = json.NewEncoder(buf).Encode(data)
	dataBytes, _ := ioutil.ReadAll(buf)

	auth0ClientProvider.On("GetToken").Return(token, nil)
	httpClientProvider.On("Request", lookupEndpoint, "GET", headers, []byte(nil), map[string]string(nil)).Return(OKStatus, dataBytes, nil)

	actualResponse, _ := orgStruct.LookupOrganization(orgName)
	assert.Equal(t, actualResponse.Name, "Linux Foundation, US")
	assert.Equal(t, actualResponse.ID, "v03fs-3")
}

func TestSearchOrganization(t *testing.T) {
	pageSize := "100"
	offset := "1"

	buf := &bytes.Buffer{}
	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)
	searchEndpoint := orgStruct.OrgBaseURL + "/orgs/search?name=" + url.PathEscape(orgName) + "&pageSize=" + pageSize + "&offset=" + offset

	data := map[string]interface{}{
		"Data": []map[string]string{
			{
				"ID":      "v03fs-3",
				"Name":    "Linux Foundation, US",
				"Link":    "linuxfoundation.com",
				"LogoURL": "linuxfoundationlogo.com/logo.png",
			},
			{
				"ID":      "v03fs-5",
				"Name":    "Linux Foundation, APAC",
				"Link":    "linuxfoundation.com",
				"LogoURL": "linuxfoundationl.com/logo.png",
			},
		},
	}

	_ = json.NewEncoder(buf).Encode(data)
	dataBytes, _ := ioutil.ReadAll(buf)

	auth0ClientProvider.On("GetToken").Return(token, nil)
	httpClientProvider.On("Request", searchEndpoint, "GET", headers, []byte(nil), map[string]string(nil)).Return(OKStatus, dataBytes, nil)

	actualResponse, _ := orgStruct.SearchOrganization(orgName, pageSize, offset)
	assert.Equal(t, actualResponse.Data[0].Name, "Linux Foundation, US")
	assert.Equal(t, actualResponse.Data[0].ID, "v03fs-3")
	assert.Equal(t, actualResponse.Data[1].Name, "Linux Foundation, APAC")
	assert.Equal(t, actualResponse.Data[1].ID, "v03fs-5")
}

func TestSearchOrganizationSpecialChars(t *testing.T) {
	pageSize := "100"
	offset := "1"
	orgName = "ą ę jest ż"
	buf := &bytes.Buffer{}
	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)
	searchEndpoint := orgStruct.OrgBaseURL + "/orgs/search?name=" + url.PathEscape(orgName) + "&pageSize=" + pageSize + "&offset=" + offset

	data := map[string]interface{}{
		"Data": []map[string]string{
			{
				"ID":      "v03fs-7",
				"Name":    "ą ę jest ż",
				"Link":    "jest.com",
				"LogoURL": "jestlogo.com",
			},
		},
	}

	_ = json.NewEncoder(buf).Encode(data)
	dataBytes, _ := ioutil.ReadAll(buf)

	auth0ClientProvider.On("GetToken").Return(token, nil)
	httpClientProvider.On("Request", searchEndpoint, "GET", headers, []byte(nil), map[string]string(nil)).Return(OKStatus, dataBytes, nil)

	actualResponse, _ := orgStruct.SearchOrganization(orgName, pageSize, offset)
	assert.Equal(t, actualResponse.Data[0].Name, "ą ę jest ż")
	assert.Equal(t, actualResponse.Data[0].ID, "v03fs-7")
}
