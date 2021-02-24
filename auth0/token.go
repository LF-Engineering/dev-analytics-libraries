package auth0

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// HTTPClientProvider used in connecting to remote http server
type HTTPClientProvider interface {
	Request(url string, method string, header map[string]string, body []byte, params map[string]string) (statusCode int, resBody []byte, err error)
}

// ESClientProvider used in connecting to ES server
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
	AuthSecret       string
	Environment      string
	httpClient       HTTPClientProvider
	esClient         ESClientProvider
	slackClient      SlackProvider
}

// NewAuth0Client ...
func NewAuth0Client(esCacheURL,
	esCacheUsername,
	esCachePassword,
	env,
	authGrantType,
	authClientID,
	authClientSecret,
	authAudience,
	authURL string,
	authSecret string,
	httpClient HTTPClientProvider,
	esClient ESClientProvider,
	slackClient SlackProvider) (*ClientProvider, error) {
	auth0 := &ClientProvider{
		ESCacheURL:       esCacheURL,
		ESCacheUsername:  esCacheUsername,
		ESCachePassword:  esCachePassword,
		AuthGrantType:    authGrantType,
		AuthClientID:     authClientID,
		AuthClientSecret: authClientSecret,
		AuthAudience:     authAudience,
		AuthURL:          authURL,
		AuthSecret:       authSecret,
		Environment:      env,
		httpClient:       httpClient,
		esClient:         esClient,
		slackClient:      slackClient,
	}

	return auth0, nil
}

// GetToken ...
func (a *ClientProvider) GetToken() (string, error) {
	// get cached token
	var authToken string
	cachedToken, err := a.getCachedToken()
	if err != nil {
		log.Println(err)
	}

	if cachedToken == "" || err != nil {
		authToken, err = a.generateToken()
		if err != nil {
			return "", err
		}
		err := a.createAuthToken(authToken)

		return authToken, err
	}
	// check token validity
	if a.isValid(cachedToken) {
		return authToken, nil
	}

	// generate a new token if not valid
	authToken, err = a.generateToken()
	if err != nil {
		return "", err
	}

	go func() {
		err = a.createAuthToken(authToken)
		log.Println(err)
	}()

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
		log.Println("generateToken : unmarshal payload error :", err)
	}

	// prevent new call if the last call issued since less than one hour
	if d.Add(1 * time.Hour).After(time.Now().UTC()) {
		return "", errors.New("can not request more than one token within the same hour")
	}

	// do not include ["Content-Type": "application/json"] header since its already added in the httpClient.Request implementation
	_, response, err := a.httpClient.Request(a.AuthURL, "POST", nil, body, nil)
	if err != nil {
		go func() {
			err := a.slackClient.SendText(err.Error())
			fmt.Println("Err: send to slack: ", err)
		}()
		log.Println("Err: GenerateToken ", err)
	}
	go func() {
		err = a.createLastActionDate()
		log.Println(err)
	}()

	log.Println(a.AuthURL, " ", string(body))
	err = json.Unmarshal(response, &result)
	if err != nil {
		log.Println("GenerateToken", err)
	}
	if result.AccessToken != "" {
		log.Println("GenerateToken: Token generated successfully.")
	}
	if !a.isValid(result.AccessToken) {
		go func() {
			err := a.slackClient.SendText("created token is not valid")
			fmt.Println("Err: send to slack: ", err)
		}()
		return "", errors.New("created token is not valid")
	}

	return result.AccessToken, nil
}

func (a *ClientProvider) getCachedToken() (string, error) {
	res, err := a.esClient.Search(strings.TrimSpace(auth0TokenCache+a.Environment), searchTokenQuery)
	if err != nil {
		go func() {
			err := a.slackClient.SendText(err.Error())
			fmt.Println("Err: send to slack: ", err)
		}()
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

func (a *ClientProvider) createAuthToken(token string) error {
	log.Println("creating new auth token")
	at := AuthToken{
		Name:      "AuthToken",
		Token:     token,
		CreatedAt: time.Now().UTC(),
	}
	doc, _ := json.Marshal(at)
	res, err := a.esClient.CreateDocument(strings.TrimSpace(auth0TokenCache+a.Environment), tokenDoc, doc)
	if err != nil {
		log.Println("could not write the data")
		return err
	}

	log.Println("createAuthToken: put in ES ", string(res))
	return nil
}

var searchTokenQuery = map[string]interface{}{
	"size": 1,
	"query": map[string]interface{}{
		"term": map[string]interface{}{
			"_id": tokenDoc,
		},
	},
}

var searchCacheQuery = map[string]interface{}{
	"size": 1,
	"query": map[string]interface{}{
		"term": map[string]interface{}{
			"_id": lastTokenDate,
		},
	},
}

func (a *ClientProvider) isValid(token string) bool {
	p, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}

		cert, err := a.getPemCert(t)
		if err != nil {
			return nil, err
		}

		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		if err != nil {
			return nil, err
		}

		return key, nil
	})
	if err != nil || !p.Valid {
		log.Println(err)
		return false
	}

	return p.Valid
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func (a *ClientProvider) getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(a.AuthURL + "/.well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k, _ := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key.")
		return cert, err
	}

	return cert, nil
}

func (a *ClientProvider) createLastActionDate() error {
	s := struct {
		Date time.Time `json:"date"`
	}{
		Date: time.Now().UTC(),
	}
	doc, _ := json.Marshal(s)
	_, err := a.esClient.CreateDocument(strings.TrimSpace(lastAuth0TokenRequest+a.Environment), lastTokenDate, doc)
	if err != nil {
		log.Println("could not write the data to elastic")
		return err
	}

	return nil
}

func (a *ClientProvider) getLastActionDate() (time.Time, error) {
	now := time.Now().UTC()
	res, err := a.esClient.Search(strings.TrimSpace(lastAuth0TokenRequest+a.Environment), searchCacheQuery)
	if err != nil && err.Error() == "index doesn't exist" {
		return now.Add(-2 * time.Hour), nil
	}
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
