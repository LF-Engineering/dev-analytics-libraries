package orgs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/LF-Engineering/dev-analytics-libraries/elastic"

	"github.com/LF-Engineering/dev-analytics-libraries/auth0"
	"github.com/LF-Engineering/dev-analytics-libraries/http"
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

// Org struct
type Org struct {
	OrgBaseURL       string
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
	AuthSecret       string
	slackProvider    SlackProvider
}

// SearchOrganization ...
func (o *Org) SearchOrganization(name string, pageSize string, offset string) (*SearchOrganizationResponse, error) {
	if name == "" {
		log.Println("SearchOrganization: name param is empty")
		return nil, errors.New("SearchOrganization: name param is empty")
	}
	token, err := o.auth0Client.GetToken()
	if err != nil {
		log.Println(err)
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
	endpoint := o.OrgBaseURL + "/orgs/search?name=" + url.PathEscape(name) + "&pageSize=" + url.PathEscape(pageSize) + "&offset=" + url.PathEscape(offset)
	_, res, err := o.httpClient.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("SearchOrganization: Could not get the organization list: ", err)
		return nil, err
	}
	var response SearchOrganizationResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		log.Println("SearchOrganization: failed to unmarshal SearchOrganizationResponse: ", err)
		return nil, err
	}
	return &response, nil
}

// LookupOrganization ...
func (o *Org) LookupOrganization(name string) (*Organization, error) {
	if name == "" {
		log.Println("LookupOrganization: name param is empty")
		return nil, errors.New("LookupOrganization: name param is empty")
	}
	token, err := o.auth0Client.GetToken()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	endpoint := o.OrgBaseURL + "/lookup?name=" + url.PathEscape(name)
	_, res, err := o.httpClient.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("LookupOrganization: Could not get the organization: ", err)
		return nil, err
	}

	var response Organization
	err = json.Unmarshal(res, &response)
	if err != nil {
		log.Println("LookupOrganization: failed to unmarshal LookupOrganizationResponse: ", err)
		return nil, err
	}
	return &response, nil
}

// NewClient consumes
// orgBaseURL, esCacheUrl, esCacheUsername, esCachePassword, esCacheIndex, env, authGrantType, authClientID, authClientSecret, authAudience, authURL
func NewClient(orgBaseURL, esCacheURL, esCacheUsername,
	esCachePassword, env, authGrantType, authClientID, authClientSecret,
	authAudience, authURL, authSecret string, slackProvider SlackProvider) (*Org, error) {
	org := &Org{
		OrgBaseURL:       orgBaseURL,
		ESCacheURL:       esCacheURL,
		ESCacheUsername:  esCacheUsername,
		ESCachePassword:  esCachePassword,
		AuthGrantType:    authGrantType,
		AuthClientID:     authClientID,
		AuthClientSecret: authClientSecret,
		AuthAudience:     authAudience,
		AuthURL:          authURL,
		Environment:      env,
		AuthSecret:       authSecret,
		slackProvider:    slackProvider,
	}

	httpClientProvider, auth0ClientProvider, esClientProvider, err := buildServices(org)
	if err != nil {
		return nil, err
	}

	org.httpClient = httpClientProvider
	org.auth0Client = auth0ClientProvider
	org.esClient = esClientProvider

	return org, nil
}

func buildServices(o *Org) (*http.ClientProvider, *auth0.ClientProvider, *elastic.ClientProvider, error) {
	esClientProvider, err := elastic.NewClientProvider(&elastic.Params{
		URL:      o.ESCacheURL,
		Username: o.ESCacheUsername,
		Password: o.ESCachePassword,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	httpClientProvider := http.NewClientProvider(time.Minute)
	auth0ClientProvider, err := auth0.NewAuth0Client(o.ESCacheURL, o.ESCacheUsername, o.ESCachePassword, o.Environment, o.AuthGrantType, o.AuthClientID, o.AuthClientSecret, o.AuthAudience, o.AuthURL, o.AuthSecret, o.httpClient, o.esClient, o.slackProvider)
	if err != nil {
		return nil, nil, nil, err
	}

	return httpClientProvider, auth0ClientProvider, esClientProvider, nil
}
