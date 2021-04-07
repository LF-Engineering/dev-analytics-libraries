package users

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/LF-Engineering/dev-analytics-libraries/elastic"

	"github.com/LF-Engineering/dev-analytics-libraries/auth0"
	"github.com/LF-Engineering/dev-analytics-libraries/http"

	json "github.com/json-iterator/go"
)

// Auth0ClientProvider ...
type Auth0ClientProvider interface {
	GetToken() (string, error)
}

// HTTPClientProvider ...
type HTTPClientProvider interface {
	Request(string, string, map[string]string, []byte, map[string]string) (int, []byte, error)
}

// ESClientProvider ...
type ESClientProvider interface {
	CreateDocument(index, documentID string, body []byte) ([]byte, error)
	Search(index string, query map[string]interface{}) ([]byte, error)
	CreateIndex(index string, body []byte) ([]byte, error)
	Get(index string, query map[string]interface{}, result interface{}) error
}

// SlackProvider ...
type SlackProvider interface {
	SendText(text string) error
}

// Usr struct
type Usr struct {
	UserBaseURL      string
	ESCacheURL       string
	ESCacheUsername  string
	ESCachePassword  string
	AuthGrantType    string
	AuthClientID     string
	AuthClientSecret string
	AuthAudience     string
	AuthURL          string
	Environment      string
	httpClient       HTTPClientProvider
	auth0Client      Auth0ClientProvider
	esClient         ESClientProvider
	slackProvider    SlackProvider
}

// ListUsers ...
func (u *Usr) ListUsers(email string, pageSize string, offset string) (*ListUsersResponse, error) {
	token, err := u.auth0Client.GetToken()
	if err != nil {
		log.Println("ListUsers", err)
		return nil, err
	}
	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	if offset == "" {
		offset = "0"
	}

	if pageSize == "" {
		pageSize = "100"
	}
	var endpoint string
	if email != "" {
		endpoint = u.UserBaseURL + "/users?email=" + url.QueryEscape(email) + "&pageSize=" + url.PathEscape(pageSize) + "&offset=" + url.PathEscape(offset)
	} else {
		endpoint = u.UserBaseURL + "/users?pageSize=" + url.PathEscape(pageSize) + "&offset=" + url.PathEscape(offset)
	}
	_, res, err := u.httpClient.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("ListUsers: Could not get the users list: ", err)
		return nil, err
	}
	var response ListUsersResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		log.Println("ListUsers: failed to unmarshal ListUsersResponse: ", err)
		log.Println("ListUsers: response: ", res)
		return nil, err
	}
	return &response, nil
}

// NewClient consumes
// userBaseURL, esCacheUrl, esCacheUsername, esCachePassword, esCacheIndex, env, authGrantType, authClientID, authClientSecret, authAudience, authURL
func NewClient(userBaseURL, esCacheURL, esCacheUsername,
	esCachePassword, env, authGrantType, authClientID, authClientSecret,
	authAudience, authURL string, slackProvider SlackProvider) (*Usr, error) {
	user := &Usr{
		UserBaseURL:      userBaseURL,
		ESCacheURL:       esCacheURL,
		ESCacheUsername:  esCacheUsername,
		ESCachePassword:  esCachePassword,
		AuthGrantType:    authGrantType,
		AuthClientID:     authClientID,
		AuthClientSecret: authClientSecret,
		AuthAudience:     authAudience,
		AuthURL:          authURL,
		Environment:      env,
		slackProvider:    slackProvider,
	}

	httpClientProvider, auth0ClientProvider, esClientProvider, err := buildServices(user)
	if err != nil {
		return nil, err
	}

	user.httpClient = httpClientProvider
	user.auth0Client = auth0ClientProvider
	user.esClient = esClientProvider

	return user, nil
}

func buildServices(u *Usr) (*http.ClientProvider, *auth0.ClientProvider, *elastic.ClientProvider, error) {
	esClientProvider, err := elastic.NewClientProvider(&elastic.Params{
		URL:      u.ESCacheURL,
		Username: u.ESCacheUsername,
		Password: u.ESCachePassword,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	httpClientProvider := http.NewClientProvider(time.Minute)

	auth0ClientProvider, err := auth0.NewAuth0Client(u.Environment,
		u.AuthGrantType,
		u.AuthClientID,
		u.AuthClientSecret,
		u.AuthAudience,
		u.AuthURL,
		httpClientProvider,
		esClientProvider,
		u.slackProvider,
		"Library.User")
	if err != nil {
		return nil, nil, nil, err
	}

	return httpClientProvider, auth0ClientProvider, esClientProvider, nil
}
