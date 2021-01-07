package affiliation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// GenerateToken ...
func (a *Affiliation) GenerateToken() string {
	var result Resp

	headers := make(map[string]string, 0)
	headers["Content-type"] = "application/json"

	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     os.Getenv("CLIENT_ID"),
		"client_secret": os.Getenv("CLIENT_SECRET"),
		"audience":      os.Getenv("CLIENT_AUDIENCE"),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
	}
	authURL := fmt.Sprintf("https://%s/oauth/token", os.Getenv("AUTH_URL"))

	_, response, err := a.httpClient.Request(authURL, "POST", headers, body, nil)
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
func (a *Affiliation) GetToken(env string) (string, error) {
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
func (a *Affiliation) CreateAuthToken(env, Token string) {
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
func (a *Affiliation) UpdateAuthToken(env, token string) {
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
func (a *Affiliation) ValidateToken(env string) (string, error) {
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
