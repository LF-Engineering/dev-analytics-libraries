package affiliation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
)

var unknown string = "Unknown"
var genderAcc int64 = 0

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
	auth0ClientPrivder Auth0ClientProvider,
	slackProvider SlackProvider) (*Affiliation, error) {
	aff := &Affiliation{
		AffBaseURL:          affBaseURL,
		ProjectSlug:         projectSlug,
		httpClientProvider:  httpClientProvider,
		esClientProvider:    esClientProvider,
		auth0ClientProvider: auth0ClientPrivder,
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
	queryParams["uuid"] = identity.UUID
	queryParams["id"] = identity.ID

	endpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(a.ProjectSlug) + "/add_identity/" + url.PathEscape(identity.Source)
	_, res, err := a.httpClientProvider.Request(strings.TrimSpace(endpoint), "POST", headers, nil, queryParams)
	if err != nil {
		log.Println("AddIdentity: Could not insert the identity: ", err)
		return false
	}
	var errMsg AffiliationsResponse
	err = json.Unmarshal(res, &errMsg)
	if err != nil || errMsg.Message != "" {
		log.Println("AddIdentity: failed to add identity: ", errMsg)
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

	_, res, err := a.httpClientProvider.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
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
	endpoint := a.AffBaseURL + "/affiliation/" + "identity/" + key + "/" + value
	_, res, err := a.httpClientProvider.Request(strings.TrimSpace(endpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("GetIdentityByUser: Could not get the identity: ", err)
		return nil, err
	}

	var errMsg AffiliationsResponse
	err = json.Unmarshal(res, &errMsg)
	if err != nil || errMsg.Message != "" {
		log.Println("GetIdentityByUser: failed to get identity: ", errMsg)
		return nil, err
	}

	var ident IdentityData
	err = json.Unmarshal(res, &ident)
	if err != nil {
		return nil, err
	}

	profileEndpoint := a.AffBaseURL + "/affiliation/" + url.PathEscape(a.ProjectSlug) + "/get_profile/" + *ident.UUID
	_, profileRes, err := a.httpClientProvider.Request(strings.TrimSpace(profileEndpoint), "GET", headers, nil, nil)
	if err != nil {
		log.Println("GetIdentityByUser: Could not get the identity: ", err)
		return nil, err
	}

	err = json.Unmarshal(res, &errMsg)
	if err != nil || errMsg.Message != "" {
		log.Println("GetIdentityByUser: failed to get identity profile: ", errMsg)
		return nil, err
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

	if profile.Profile.IsBot != nil {
		identity.IsBot = profile.Profile.IsBot
	}
	if profile.Profile.Gender != nil {
		identity.Gender = profile.Profile.Gender
	}
	if profile.Profile.GenderAcc != nil {
		identity.GenderACC = profile.Profile.GenderAcc
	}

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
		return nil, errors.New("User not found")
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
