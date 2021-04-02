package auth0

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	UpdateDocument(index string, id string, body interface{}) ([]byte, error)
}

// SlackProvider ...
type SlackProvider interface {
	SendText(text string) error
}

// ClientProvider ...
type ClientProvider struct {
	AuthGrantType    string
	AuthClientID     string
	AuthClientSecret string
	AuthAudience     string
	AuthURL          string
	Environment      string
	httpClient       HTTPClientProvider
	esClient         ESClientProvider
	slackClient      SlackProvider
	appName          string
}

// NewAuth0Client ...
func NewAuth0Client(env,
	authGrantType,
	authClientID,
	authClientSecret,
	authAudience,
	authURL string,
	httpClient HTTPClientProvider,
	esClient ESClientProvider,
	slackClient SlackProvider,
	appName string) (*ClientProvider, error) {
	auth0 := &ClientProvider{
		AuthGrantType:    authGrantType,
		AuthClientID:     authClientID,
		AuthClientSecret: authClientSecret,
		AuthAudience:     authAudience,
		AuthURL:          authURL,
		Environment:      env,
		httpClient:       httpClient,
		esClient:         esClient,
		slackClient:      slackClient,
		appName:          appName,
	}

	return auth0, nil
}

// GetToken ...
func (a *ClientProvider) GetToken() (string, error) {
	// get cached token
	authToken, err := a.getCachedToken()
	if err != nil {
		log.Println(err)
	}

	if authToken == "" || err != nil {
		authToken, err = a.refreshToken()
		if err != nil {
			return "", err
		}

		return authToken, err
	}

	// check token validity
	ok, claims, err := a.isValid(authToken)
	if ok {
		go func() {
			// refresh token before expiry by X minutes
			if claims.VerifyExpiresAt(time.Now().Add(5*time.Minute).Unix(), false) == false {
				if _, err := a.refreshToken(); err != nil {
					log.Printf("Error refresh auth0 token %s\n", err.Error())
				}
			}
		}()

		return authToken, nil
	}

	// generate a new token if not valid
	return a.refreshToken()
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
	_, response, err := a.httpClient.Request(fmt.Sprintf("%s/oauth/token", a.AuthURL), "POST", nil, body, nil)
	if err != nil {
		go func() {
			errMsg := fmt.Sprintf("%s-%s: error generating a new token\n %s", a.appName, a.Environment, err)
			if err := a.slackClient.SendText(errMsg); err != nil {
				log.Println(" Err: GenerateToken ", a.Environment, err)
			}

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
	ok, _, err := a.isValid(result.AccessToken)
	if !ok || err != nil {
		go func() {
			errMsg := fmt.Sprintf("%s-%s: error validating the newly created token\n %s", a.appName, a.Environment, err)
			err := a.slackClient.SendText(errMsg)
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
			errMsg := fmt.Sprintf("%s-%s: error cached token not found\n %s", a.appName, a.Environment, err)
			err := a.slackClient.SendText(errMsg)
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
	_, err := a.esClient.UpdateDocument(fmt.Sprintf("%s%s", auth0TokenCache, a.Environment), tokenDoc, at)
	if err != nil {
		log.Println("could not write the data")
		return err
	}

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

func (a *ClientProvider) isValid(token string) (bool, jwt.MapClaims, error) {
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
	if err != nil {
		return false, nil, err
	}

	claims, ok := p.Claims.(jwt.MapClaims)
	if !ok {
		return false, nil, err
	}

	return p.Valid, claims, err
}

// Jwks result from auth0 well know keys
type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

// JSONWebKeys auth0 token key
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
	_, resp, err := a.httpClient.Request(fmt.Sprintf("%s/oauth/.well-known/jwks.json", a.AuthURL), "GET", nil, nil, nil)
	if err != nil {
		return cert, err
	}

	var jwks = Jwks{}
	if err := json.Unmarshal(resp, &jwks); err != nil {
		return cert, err
	}

	for _, k := range jwks.Keys {
		if token.Header["kid"] == k.Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + k.X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("unable to find appropriate key")
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

func (a *ClientProvider) refreshToken() (string, error) {
	authToken, err := a.generateToken()
	if err != nil {
		return "", err
	}

	return authToken, a.createAuthToken(authToken)
}
