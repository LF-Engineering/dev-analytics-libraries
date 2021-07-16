package affiliation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	unknown    = "Unknown"
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Affiliations interface
type Affiliations interface {
	AddIdentity(identity *Identity) bool
}

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

// Auth0ClientProvider ...
type Auth0ClientProvider interface {
	GetToken() (string, error)
}

// Affiliation struct
type Affiliation struct {
	AffBaseURL          string
	ProjectSlug         string
	httpClientProvider  HTTPClientProvider
	esClientProvider    ESClientProvider
	auth0ClientProvider Auth0ClientProvider
	slackProvider       SlackProvider
}

// NewAffiliationsClient consumes
//  affBaseURL, projectSlug, esCacheUrl, esCacheUsername, esCachePassword, esCacheIndex, env, authGrantType, authClientID, authClientSecret, authAudience, authURL
func NewAffiliationsClient(affBaseURL string,
	projectSlug string,
	httpClientProvider HTTPClientProvider,
	esClientProvider ESClientProvider,
	auth0ClientProvider Auth0ClientProvider,
	slackProvider SlackProvider) (*Affiliation, error) {
	aff := &Affiliation{
		AffBaseURL:          affBaseURL,
		ProjectSlug:         projectSlug,
		httpClientProvider:  httpClientProvider,
		esClientProvider:    esClientProvider,
		auth0ClientProvider: auth0ClientProvider,
		slackProvider:       slackProvider,
	}

	return aff, nil
}

// AddIdentity ...
func (a *Affiliation) AddIdentity(identity *Identity) bool {
	if identity == nil {
		log.Println("AddIdentity: Identity is nil")
		return false
	}
	token, err := a.auth0ClientProvider.GetToken()
	if err != nil {
		log.Println(err)
	}
	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	queryParams := make(map[string]string, 0)
	queryParams["name"] = identity.Name
	queryParams["username"] = identity.Username
	queryParams["email"] = identity.Email
	queryParams["id"] = identity.ID
	if identity.UUID != "" {
		queryParams["uuid"] = identity.UUID
	}

	endpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(a.ProjectSlug) + "/add_identity/" + url.PathEscape(identity.Source)
	statusCode, res, err := a.httpClientProvider.Request(strings.TrimSpace(endpoint), "POST", headers, nil, queryParams)
	switch statusCode {
	case http.StatusOK, http.StatusConflict, http.StatusCreated:
		return true
	default:
		if err != nil {
			log.Println("AddIdentity: Could not insert the identity: ", err)
		}

		var errMsg AffiliationsResponse
		err = json.Unmarshal(res, &errMsg)
		if err != nil || errMsg.Message != "" {
			log.Println("AddIdentity: failed to add identity: ", errMsg)
		}

		// check if identity has been added to db
		if checkIdentity := a.GetIdentity(identity.UUID); checkIdentity != nil {
			return true
		}
	}
	time.Sleep(2 * time.Minute)
	return a.AddIdentity(identity)
}

// GetIdentity ...
func (a *Affiliation) GetIdentity(uuid string) *Identity {
	if uuid == "" {
		log.Println("GetIdentity: uuid is empty")
		return nil
	}
	token, err := a.auth0ClientProvider.GetToken()
	if err != nil {
		log.Println(err)
	}
	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	endpoint := a.AffBaseURL + "/affiliation/get_identity/" + uuid

	_, res, err := a.httpClientProvider.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
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
	token, err := a.auth0ClientProvider.GetToken()
	if err != nil {
		log.Println(err)
	}
	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	endpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(projectSlug) + "/enrollments/" + uuid

	statusCode, res, err := a.httpClientProvider.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	switch statusCode {
	case http.StatusBadRequest, http.StatusNotFound:
		log.Println("GetOrganizations: Could not get the enrollment: ", err)
		return nil
	case http.StatusOK:
	default:
		if err != nil {
			log.Println("GetOrganizations: Could not get the organizations: ", err)
			return nil
		}

		var errMsg AffiliationsResponse
		err = json.Unmarshal(res, &errMsg)
		if err != nil || errMsg.Message != "" {
			log.Println("GetOrganizations: failed to get organizations: ", errMsg)
		}

		time.Sleep(2 * time.Minute)
		return a.GetOrganizations(uuid, projectSlug)
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
	token, err := a.auth0ClientProvider.GetToken()
	if err != nil {
		log.Println(err)
	}
	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)

	endpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(projectSlug) + "/get_profile/" + uuid

	_, res, err := a.httpClientProvider.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
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
		nilKeyOrValueErr := "GetIdentityByUser: key or value is null"
		return nil, fmt.Errorf(nilKeyOrValueErr)
	}
	token, err := a.auth0ClientProvider.GetToken()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)
	endpoint := a.AffBaseURL + "/affiliation/" + "identity/" + key + "/" + value
	statusCode, res, err := a.httpClientProvider.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	switch statusCode {
	case http.StatusBadRequest, http.StatusNotFound:
		log.Println("GetIdentityByUser: Could not get the identity: l", err)
		return nil, errors.New("identity not found")
	case http.StatusOK:
	default:
		if err != nil {
			log.Println("GetIdentityByUser: Could not get the identity: ", err)
			return nil, err
		}

		var errMsg AffiliationsResponse
		err = json.Unmarshal(res, &errMsg)
		if err != nil || errMsg.Message != "" {
			log.Println("GetIdentityByUser: failed to get identity: ", errMsg)
		}

		time.Sleep(2 * time.Minute)
		return a.GetIdentityByUser(key, value)
	}

	var ident IdentityData
	err = json.Unmarshal(res, &ident)
	if err != nil {
		return nil, err
	}

	profileEndpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(a.ProjectSlug) + "/get_profile/" + *ident.UUID
	statusCode, profileRes, err := a.httpClientProvider.Request(strings.TrimSpace(profileEndpoint), "GET", headers, nil, nil)
	switch statusCode {
	case http.StatusBadRequest, http.StatusNotFound:
		log.Println("GetIdentityByUser: Could not get the identity: ", err)
		return nil, errors.New("identity not found")
	case http.StatusOK:
	default:
		if err != nil {
			log.Println("GetIdentityByUser: Could not get the identity: ", err)
		}

		var errMsg AffiliationsResponse
		err = json.Unmarshal(res, &errMsg)
		if err != nil || errMsg.Message != "" {
			log.Println("GetIdentityByUser: failed to get identity profile: ", errMsg)
		}

		time.Sleep(2 * time.Minute)
		return a.GetIdentityByUser(key, value)
	}

	var profile UniqueIdentityFullProfile
	err = json.Unmarshal(profileRes, &profile)
	if err != nil {
		return nil, err
	}
	var identity AffIdentity
	identity.UUID = ident.UUID
	if ident.Name != nil {
		identity.Name = *ident.Name
	}

	if ident.Username != nil {
		identity.Username = *ident.Username
	}

	if ident.Email != nil {
		identity.Email = *ident.Email
	}

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

	if profile.Profile.Name != nil {
		identity.Name = *profile.Profile.Name
	}

	return &identity, nil

}

// GetProfileByUsername ...
func (a *Affiliation) GetProfileByUsername(username string, projectSlug string) (*AffIdentity, error) {
	if username == "" && projectSlug == "" {
		nilKeyOrValueErr := "GetProfileByUsername: username or projectSlug is null"
		log.Println(nilKeyOrValueErr)
		return nil, fmt.Errorf(nilKeyOrValueErr)
	}

	token, err := a.auth0ClientProvider.GetToken()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	headers := make(map[string]string, 0)
	headers["Authorization"] = fmt.Sprintf("%s %s", "Bearer", token)
	endpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(projectSlug) + "/get_profile_by_username/" + url.PathEscape(username)
	statusCode, res, err := a.httpClientProvider.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("GetProfileByUsername: Could not get the profile: ", err)
		return nil, err
	}

	if statusCode != 200 {
		return nil, errors.New("user not found")
	}

	var profile UniqueIdentityFullProfile
	err = json.Unmarshal(res, &profile)
	if err != nil {
		return nil, err
	}
	var identity AffIdentity

	for _, value := range profile.Identities {
		if value.Source == "github" {
			profileIdentity := value
			identity.UUID = profileIdentity.UUID
			if profileIdentity.Name != nil {
				identity.Name = *profileIdentity.Name
			} else {
				identity.Name = unknown
			}

			if profileIdentity.Email != nil {
				identity.Email = *profileIdentity.Email
			}

			identity.ID = &profileIdentity.ID
		}
	}

	if len(profile.Identities) == 0 {
		identity.Name = unknown
		identity.ID = &unknown
		identity.UUID = &unknown
	}

	identity.Username = username

	if profile.Profile != nil {
		identity.IsBot = profile.Profile.IsBot
	}

	if profile.Enrollments == nil {
		identity.OrgName = &unknown
		identity.MultiOrgNames = make([]string, 0)
	}

	if len(profile.Enrollments) > 1 {
		identity.OrgName = a.getUserOrg(profile.Enrollments)
		for _, org := range profile.Enrollments {
			identity.MultiOrgNames = append(identity.MultiOrgNames, org.Organization.Name)
		}
	} else if len(profile.Enrollments) == 1 {
		identity.OrgName = &profile.Enrollments[0].Organization.Name
		identity.MultiOrgNames = append(identity.MultiOrgNames, profile.Enrollments[0].Organization.Name)
	}

	return &identity, nil
}

// Get Most Recent Org Name where user has multiple enrollments
func (a *Affiliation) getUserOrg(enrollments []*Enrollments) *string {
	var result string
	var lowest, startTime int64
	now := time.Now()

	for _, enrollment := range enrollments {
		startTime = now.Unix() - enrollment.Start.Unix()

		if lowest == 0 {
			lowest = startTime
			result = enrollment.Organization.Name
		} else if startTime < lowest {
			lowest = startTime
			result = enrollment.Organization.Name
		}
	}

	return &result
}
