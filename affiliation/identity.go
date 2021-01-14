package affiliation

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/LF-Engineering/dev-analytics-libraries/auth0"
	"github.com/LF-Engineering/dev-analytics-libraries/elastic"
	"github.com/LF-Engineering/dev-analytics-libraries/http"
)

// Affiliations interface
type Affiliations interface {
	AddIdentity(identity *Identity) bool
}

// Affiliation struct
type Affiliation struct {
	AffBaseURL       string
	ProjectSlug      string
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
	auth0Client      *auth0.ClientProvider
}

// NewAffiliationsClient consumes
//  affBaseURL, projectSlug, esCacheUrl, esCacheUsername, esCachePassword, esCacheIndex, env, authGrantType, authClientID, authClientSecret, authAudience, authURL
func NewAffiliationsClient(affBaseURL, projectSlug, esCacheURL, esCacheUsername, esCachePassword, env, authGrantType, authClientID, authClientSecret, authAudience, authURL string) (*Affiliation, error) {
	aff := &Affiliation{
		AffBaseURL:       affBaseURL,
		ProjectSlug:      projectSlug,
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

	httpClientProvider, esClientProvider, auth0ClientProvider, err := buildServices(aff)
	if err != nil {
		return nil, err
	}

	aff.esClient = esClientProvider
	aff.httpClient = httpClientProvider
	aff.auth0Client = auth0ClientProvider

	return aff, nil
}

// AddIdentity ...
func (a *Affiliation) AddIdentity(identity *Identity) bool {
	if identity == nil {
		log.Println("Repository: AddIdentity: Identity is nil")
		return false
	}
	token, err := a.auth0Client.ValidateToken(a.Environment)
	if err != nil {
		log.Println(err)
	}
	headers := make(map[string]string, 0)
	headers["Content-type"] = "application/json"
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	queryParams := make(map[string]string, 0)
	queryParams["name"] = identity.Name
	queryParams["username"] = identity.Username
	queryParams["email"] = identity.Email
	queryParams["uuid"] = identity.UUID

	endpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(a.ProjectSlug) + "/add_identity/" + url.PathEscape(identity.Source)
	_, res, err := a.httpClient.Request(strings.TrimSpace(endpoint), "POST", headers, nil, queryParams)
	if err != nil {
		log.Println("Repository: AddIdentity: Could not insert the identity: ", err)
		return false
	}
	var errMsg AffiliationsResponse
	err = json.Unmarshal(res, &errMsg)
	if err != nil || errMsg.Message != "" {
		log.Println("Repository: AddIdentity: failed to add identity: ", err)
		return false
	}
	return true
}

// GetIdentity ...
func (a *Affiliation) GetIdentity(uuid string) *Identity {
	if uuid == "" {
		log.Println("GetIdentity: uuid is empty")
		return nil
	}
	token, err := a.auth0Client.ValidateToken(a.Environment)
	if err != nil {
		log.Println(err)
	}
	headers := make(map[string]string, 0)
	headers["Content-type"] = "application/json"
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	endpoint := a.AffBaseURL + "/affiliation/get_identity/" + uuid

	_, res, err := a.httpClient.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("GetIdentity: Could not get the identity: ", err)
		return nil
	}
	var identity Identity
	err = json.Unmarshal(res, &identity)
	if err != nil {
		log.Println("GetIdentity: failed to unmarshal identity: ", err)
		return nil
	}
	return &identity
}

// GetOrganizations ...
func (a *Affiliation) GetOrganizations(uuid, projectSlug string) *[]Enrollment {
	if uuid == "" || projectSlug == "" {
		return nil
	}
	token, err := a.auth0Client.ValidateToken(a.Environment)
	if err != nil {
		log.Println(err)
	}
	headers := make(map[string]string, 0)
	headers["Content-type"] = "application/json"
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	endpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(projectSlug) + "/enrollments/" + uuid

	_, res, err := a.httpClient.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("GetOrganizations: Could not get the organizations: ", err)
		return nil
	}

	var response EnrollmentsResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		log.Println("GetOrganizations: failed to unmarshal enrollments response: ", err)
		return nil
	}
	return &response.Enrollments
}

// GetProfile ...
func (a *Affiliation) GetProfile(uuid, projectSlug string) *ProfileResponse {
	if uuid == "" || projectSlug == "" {
		return nil
	}
	token, err := a.auth0Client.ValidateToken(a.Environment)
	if err != nil {
		log.Println(err)
	}
	headers := make(map[string]string, 0)
	headers["Content-type"] = "application/json"
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	endpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(projectSlug) + "/get_profile/" + uuid

	_, res, err := a.httpClient.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("GetProfile: Could not get the profile: ", err)
		return nil
	}

	var response ProfileResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		log.Println("GetProfile: failed to unmarshal profile response: ", err)
		return nil
	}
	return &response
}

// GetIdentityByUser ...
func (a *Affiliation) GetIdentityByUser(key string, value string) (*AffIdentity, error) {
	if key == "" || value == "" {
		nilKeyOrValueErr := "repository: GetIdentityByUser: key or value is null"
		log.Println(nilKeyOrValueErr)
		return nil, fmt.Errorf(nilKeyOrValueErr)
	}
	token, err := a.auth0Client.ValidateToken(a.Environment)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	headers := make(map[string]string, 0)
	headers["Content-type"] = "application/json"
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	endpoint := a.AffBaseURL + "/affiliation/" + "identity/" + key + "/" + value
	_, res, err := a.httpClient.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("Repository: GetIdentityByUser: Could not get the identity: ", err)
		return nil, err
	}

	// Todo : need review
	//var errMsg AffiliationsResponse
	//err = json.Unmarshal(res, &errMsg)
	//if err != nil || errMsg.Message != "" {
	//	log.Println("Repository: GetIdentityByUser: failed to get identity: ", err)
	//	return nil, err
	//}

	var ident IdentityData
	err = json.Unmarshal(res, &ident)
	if err != nil {
		return nil, err
	}

	profileEndpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(a.ProjectSlug) + "/get_profile/" + *ident.UUID
	_, profileRes, err := a.httpClient.Request(strings.TrimSpace(profileEndpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("Repository: GetIdentityByUser: Could not get the identity: ", err)
		return nil, err
	}

	var profile UniqueIdentityFullProfile
	err = json.Unmarshal(profileRes, &profile)
	if err != nil {
		return nil, err
	}
	var identity AffIdentity
	identity.UUID = ident.UUID
	identity.Name = *ident.Name
	identity.Username = *ident.Username
	identity.Email = *ident.Email
	identity.ID = &ident.ID

	identity.IsBot = profile.Profile.IsBot
	identity.Gender = profile.Profile.Gender
	identity.GenderACC = profile.Profile.GenderAcc

	if len(profile.Enrollments) > 1 {
		identity.OrgName = &profile.Enrollments[0].Organization.Name
		for _, org := range profile.Enrollments {
			identity.MultiOrgNames = append(identity.MultiOrgNames, org.Organization.Name)
		}
	} else if len(profile.Enrollments) == 1 {
		identity.OrgName = &profile.Enrollments[0].Organization.Name
		identity.MultiOrgNames = append(identity.MultiOrgNames, profile.Enrollments[0].Organization.Name)
	}

	return &identity, nil

}

func buildServices(a *Affiliation) (httpClientProvider *http.ClientProvider, esClientProvider *elastic.ClientProvider, auth0ClientProvider *auth0.ClientProvider, err error) {
	esClientProvider, err = elastic.NewClientProvider(&elastic.Params{
		URL:      a.ESCacheURL,
		Username: a.ESCacheUsername,
		Password: a.ESCachePassword,
	})
	if err != nil {
		return
	}

	httpClientProvider = http.NewClientProvider(time.Minute)

	auth0ClientProvider, err = auth0.NewAuth0Client(a.ESCacheURL, a.ESCacheUsername, a.ESCachePassword, a.Environment, a.AuthGrantType, a.AuthClientID, a.AuthClientSecret, a.AuthAudience, a.AuthURL)
	if err != nil {
		return
	}

	return
}
