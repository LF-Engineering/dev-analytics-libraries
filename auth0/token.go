package auth0

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/LF-Engineering/dev-analytics-libraries/elastic"
	"github.com/LF-Engineering/dev-analytics-libraries/http"
)

// HTTPClientProvider used in connecting to remote http server
type HTTPClientProvider interface {
	Request(url string, method string, header map[string]string, body []byte, params map[string]string) (statusCode int, resBody []byte, err error)
}

// ESClientProvider used in connecting to ES server
type ESClientProvider interface {
	CreateDocument(index, documentID string, body []byte) ([]byte, error)
	Search(index string, query map[string]interface{}) (bites []byte, err error)
	CreateIndex(index string, body []byte) ([]byte, error)
	Get(index string, query map[string]interface{}, result interface{}) (err error)
}

// ClientProvider ...
type ClientProvider struct {
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
	esClient         ESClientProvider
}

// NewAuth0Client ...
func NewAuth0Client(esCacheURL, esCacheUsername, esCachePassword, env, authGrantType, authClientID, authClientSecret, authAudience, authURL string) (*ClientProvider, error) {
	auth0 := &ClientProvider{
		ESCacheURL:       esCacheURL,
		ESCacheUsername:  esCacheUsername,
		ESCachePassword:  esCachePassword,
		AuthGrantType:    authGrantType,
		AuthClientID:     authClientID,
		AuthClientSecret: authClientSecret,
		AuthAudience:     authAudience,
		AuthURL:          authURL,
		Environment:      env,
	}

	httpClientProvider, esClientProvider, err := buildServices(auth0)
	if err != nil {
		return nil, err
	}

	auth0.esClient = esClientProvider
	auth0.httpClient = httpClientProvider

	return auth0, nil
}

func (a *ClientProvider) GetToken(env string) (string, error) {
	// get cached token
	var authToken string
	cachedToken, err := a.getCachedToken(env)
	if err != nil {
		log.Println(err)
	}

	if cachedToken == "" || err != nil {
		authToken, err = a.generateToken()
		if err != nil {
			return "", err
		}
		err := a.createAuthToken(env, authToken)

		return authToken, err
	}
	// check token validate
	if a.isValid(cachedToken) {
		return authToken, nil
	}

	// generate a new token if not valid
	authToken, err = a.generateToken()
	if err != nil {
		return "", err
	}

	err = a.createAuthToken(env, authToken)

	return authToken, err
}

func (a *ClientProvider) generateToken() (string, error) {
	var result Resp
	d, err := a.getLastActionDate()
	if err != nil {
		return "", err
	}
	payload := map[string]string{
		"grant_type":    a.AuthGrantType,
		"client_id":     a.AuthClientID,
		"client_secret": a.AuthClientSecret,
		"audience":      a.AuthAudience,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
	}

	// prevent new call if the last call issued since less than one hour
	if d.Add(1 * time.Hour).After(time.Now().UTC()) {
		return "", errors.New("can not request more than one token within the same hour")
	}

	// do not include ["Content-Type": "application/json"] header since its already added in the httpClient.Request implementation
	_, response, err := a.httpClient.Request(a.AuthURL, "POST", nil, body, nil)
	if err != nil {
		log.Println("GenerateToken", err)
	}

	log.Println(a.AuthURL, " ", string(body))
	err = json.Unmarshal(response, &result)
	if err != nil {
		log.Println("GenerateToken", err)
	}
	if result.AccessToken != "" {
		log.Println("GenerateToken: Token generated successfully.")
	}

	go func() {
		err := a.createLastActionDate(a.Environment)
		log.Println(err)
	}()

	return result.AccessToken, nil
}

// GetToken ...
func (a *ClientProvider) getCachedToken(env string) (string, error) {
	res, err := a.esClient.Search(strings.TrimSpace("auth0-token-cache-"+env), searchTokenQuery)
	if err != nil {
		return "", err
	}

	var e ESTokenSchema
	err = json.Unmarshal(res, &e)
	if err != nil {
		log.Println("repository: GetOauthToken: could not unmarshal the data", err)
		return "", err
	}

	if len(e.Hits.Hits) > 0 {
		data := e.Hits.Hits[0]
		return data.Source.Token, nil
	}

	return "", errors.New("GetToken: could not find the associated token")
}

// CreateAuthToken accepts
//  esCacheURL, esCacheUsername, esCachePassword, Token
func (a *ClientProvider) createAuthToken(env, Token string) error {
	log.Println("creating new auth token")
	at := AuthToken{
		Name:      "AuthToken",
		Token:     Token,
		CreatedAt: time.Now().UTC(),
	}
	doc, _ := json.Marshal(at)
	res, err := a.esClient.CreateDocument(strings.TrimSpace("auth0-token-cache-"+env), "token", doc)
	if err != nil {
		log.Println("could not write the data")
		return err
	}

	log.Println("createAuthToken: put in ES ", string(res))
	return nil
}

func buildServices(a *ClientProvider) (httpClientProvider *http.ClientProvider, esClientProvider *elastic.ClientProvider, err error) {
	esClientProvider, err = elastic.NewClientProvider(&elastic.Params{
		URL:      a.ESCacheURL,
		Username: a.ESCacheUsername,
		Password: a.ESCachePassword,
	})
	if err != nil {
		return nil, nil, err
	}

	httpClientProvider = http.NewClientProvider(time.Minute)
	return
}

var searchTokenQuery = map[string]interface{}{
	"size": 1,
	"query": map[string]interface{}{
		"match_all": map[string]interface{}{},
	},
}

func (a *ClientProvider) isValid(token string) bool {
	t, err := jwt.Parse(token, nil)
	if err != nil || !t.Valid {
		log.Println(err)
		return false
	}

	claims, _ := t.Claims.(jwt.MapClaims)
	str := fmt.Sprintf("%v", claims["exp"])
	fl, _ := strconv.ParseFloat(str, 64)
	i := int64(fl)
	s := strconv.Itoa(int(i))
	i, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Println(err)
		return false
	}

	tm := time.Unix(i, 0)
	now := time.Now()

	// return true token if valid
	if now.Before(tm) {
		return true
	}

	return false
}

func (a *ClientProvider) createLastActionDate(env string) error {
	s := struct {
		Date time.Time `json:"date"`
	}{
		Date: time.Now().UTC(),
	}
	doc, _ := json.Marshal(s)
	_, err := a.esClient.CreateDocument(strings.TrimSpace("last-auth0-token-request-"+env), "token", doc)
	if err != nil {
		log.Println("could not write the data to elastic")
		return err
	}

	return nil
}

func (a *ClientProvider) getLastActionDate() (time.Time, error) {
	now := time.Now().UTC()
	res, err := a.esClient.Search(strings.TrimSpace("last-auth0-token-request-"+a.Environment), searchTokenQuery)
	if err != nil {
		return now, err
	}
	var e LastActionSchema
	err = json.Unmarshal(res, &e)
	if err != nil {
		log.Println("repository: getLastActionDate failed", err)
		return now, err
	}

	if len(e.Hits.Hits) > 0 {
		data := e.Hits.Hits[0]
		return data.Source.Date, nil
	}

	return now, errors.New("getLastActionDate: could not find the associated date")
}
