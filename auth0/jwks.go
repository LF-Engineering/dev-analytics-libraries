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

func (a *ClientProvider) createAuthJwks(cert string) error {
	log.Println("creating new auth jwks cert string")
	at := AuthJwks{
		Name:      "AuthJwks",
		Jwks:      cert,
		CreatedAt: time.Now().UTC(),
	}
	_, err := a.esClient.UpdateDocument(fmt.Sprintf("%s%s", auth0JwksCache, a.Environment), jwksDoc, at)
	if err != nil {
		log.Println("could not write the data", err)
		return err
	}

	return nil
}

func (a *ClientProvider) getPemCert(token *jwt.Token, refreshJwks bool) (string, error) {
	cert := ""
	cert, expired, err := a.getCachedJwks()
	if err != nil {
		return cert, err
	}

	// check if the cache expired as well is not invoked via refresh token cron
	if !expired && !refreshJwks {
		return cert, nil
	}

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

	err = a.createAuthJwks(cert)
	if err != nil {
		return "", err
	}

	return cert, nil
}

func (a *ClientProvider) getCachedJwks() (string, bool, error) {
	expired := true
	res, err := a.esClient.Search(strings.TrimSpace(auth0JwksCache+a.Environment), searchJwksQuery)
	if err != nil {
		go func() {
			errMsg := fmt.Sprintf("%s-%s: error cached jwks not found\n %s", a.appName, a.Environment, err)
			err := a.slackClient.SendText(errMsg)
			fmt.Println("Err: send to slack: ", err)
		}()

		return "", expired, err
	}

	var e ESJwksSchema
	err = json.Unmarshal(res, &e)
	if err != nil {
		log.Println("repository: GetOauthJwks: could not unmarshal the data", err)
		return "", expired, err
	}

	if len(e.Hits.Hits) > 0 {
		data := e.Hits.Hits[0]
		// compare current time v/s existing cached time + 30 mins
		if data.Source.CreatedAt.Add(30*time.Minute).Unix() <= time.Now().UTC().Unix() {
			expired = false
		}

		return data.Source.Jwks, expired, nil
	}

	return "", expired, errors.New("GetJwks: could not find the associated jwks")
}

var searchJwksQuery = map[string]interface{}{
	"size": 1,
	"query": map[string]interface{}{
		"term": map[string]interface{}{
			"_id": jwksDoc,
		},
	},
}
