package auth0

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/LF-Engineering/dev-analytics-libraries/elastic"
	"github.com/LF-Engineering/dev-analytics-libraries/http"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// ClientProvider ...
type ClientProvider struct {
	Auth0BaseURL     string
	ESCacheURL       string
	ESCacheUsername  string
	ESCachePassword  string
	AuthGrantType    string
	AuthClientID     string
	AuthClientSecret string
	AuthAudience     string
	AuthURL          string
	Environment      string
	httpClient       *http.ClientProvider
	esClient         *elastic.ClientProvider
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

// GenerateToken ...
func (a *ClientProvider) GenerateToken() string {
	var result Resp

	headers := make(map[string]string, 0)
	headers["Content-type"] = "application/json"

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

	_, response, err := a.httpClient.Request(a.AuthURL, "POST", headers, body, nil)
	if err != nil {
		log.Println("GenerateToken", err)
	}

	log.Println(os.Getenv("AUTH_URL"), " ", string(body))
	err = json.Unmarshal(response, &result)
	if err != nil {
		log.Println("GenerateToken", err)
	}
	if result.AccessToken != "" {
		log.Println("GenerateToken: Token generated successfully.")
	}
	return result.AccessToken
}

// GetToken ...
func (a *ClientProvider) GetToken(env string) (string, error) {
	query := make(map[string]interface{}, 0)
	query["name"] = "token"
	res, err := a.esClient.Search(strings.TrimSpace("auth0-token-cache-"+env), query)
	if err != nil {
		return "", err
	}

	var e ESTokenSchema
	err = json.Unmarshal(res, &a)
	if err != nil {
		log.Println("repository: GetOauthToken: could not unmarshal the data", err)
		return "", err
	}

	if len(e.Hits.Hits) > 0 {
		data := e.Hits.Hits[0]
		return data.Source.Token, nil
	}
	log.Println("GetTokenStringESOutput: ", e.Hits.Hits[0])
	log.Println("GetToken: could not find the associated token")

	return "", errors.New("GetToken: could not find the associated token")
}

// CreateAuthToken accepts
//  esCacheURL, esCacheUsername, esCachePassword, Token
func (a *ClientProvider) CreateAuthToken(env, Token string) {
	log.Println("creating new auth token")
	at := AuthToken{
		Name:  "AuthToken",
		Token: Token,
	}
	bites, _ := json.Marshal(at)
	res, err := a.esClient.CreateDocument(strings.TrimSpace("auth0-token-cache-"+env), uuid.New().String(), bites)
	if err != nil {
		log.Println("could not write the data")
		return
	}
	log.Println("CreateAuthToken: put in ES ", string(res))
}

// UpdateAuthToken ...
func (a *ClientProvider) UpdateAuthToken(env, token string) {
	fields := fmt.Sprintf(`
		{
			"script" : {
				"source": "ctx._source.Token = '%s'",
				"lang": "painless"
			}
	}`, token)
	query := "Token: AuthToken"
	res, err := a.esClient.UpdateDocumentByQuery(strings.TrimSpace("auth0-token-cache-"+env), query, fields)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("ES Response: ", string(res))
}

// ValidateToken ...
func (a *ClientProvider) ValidateToken(env string) (string, error) {
	var authToken string
	tokenString, err := a.GetToken(env)
	if err != nil {
		log.Println(err)
	}

	if tokenString == "" {
		log.Println("ValidateToken: token is not there in elasticSearch")
		authToken = a.GenerateToken()
		a.CreateAuthToken(env, authToken)
		return authToken, nil
	}
	log.Println("ValidateToken: got the token from ES ")
	token, err := jwt.Parse(tokenString, nil)
	if token == nil {
		log.Println(err)
		return "", err
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	str := fmt.Sprintf("%v", claims["exp"])
	fl, _ := strconv.ParseFloat(str, 64)
	i := int64(fl)
	s := strconv.Itoa(int(i))
	i, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}

	tm := time.Unix(i, 0)
	now := time.Now()

	if now.Before(tm) {
		log.Println("ValidateToken: Token is valid")
		return tokenString, nil
	}
	log.Println("ValidateToken: token is expired!")
	authToken = a.GenerateToken()
	a.UpdateAuthToken(env, authToken)

	return authToken, nil
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

	httpClientProvider = http.NewClientProvider(60)
	return
}
