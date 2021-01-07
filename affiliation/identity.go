package affiliation

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/LF-Engineering/dev-analytics-libraries/elastic"
	"github.com/LF-Engineering/dev-analytics-libraries/http"
)

// Affiliations interface
type Affiliations interface {
	AddIdentity(identity *Identity) bool
}

// Affiliation struct
type Affiliation struct {
	AffBaseURL      string
	ProjectSlug     string
	ESCacheURL      string
	ESCacheUsername string
	ESCachePassword string
	Environment     string
	httpClient      *http.ClientProvider
	esClient        *elastic.ClientProvider
}

// NewAffiliationsClient consumes
//  affBaseURL, projectSlug, esCacheUrl, esCacheUsername, esCachePassword, esCacheIndex, env
func NewAffiliationsClient(affBaseURL, projectSlug, esCacheURL, esCacheUsername, esCachePassword, env string) (*Affiliation, error) {
	aff := &Affiliation{
		AffBaseURL:      affBaseURL,
		ProjectSlug:     projectSlug,
		ESCacheURL:      esCacheURL,
		ESCacheUsername: esCacheUsername,
		ESCachePassword: esCachePassword,
		Environment:     env,
	}

	httpClientProvider, esClientProvider, err := buildServices(aff)
	if err != nil {
		return nil, err
	}

	aff.esClient = esClientProvider
	aff.httpClient = httpClientProvider

	return aff, nil
}

// AddIdentity ...
func (a *Affiliation) AddIdentity(identity *Identity) bool {
	if identity == nil {
		log.Println("Repository: AddIdentity: Identity is nil")
		return false
	}
	token, err := a.ValidateToken(a.Environment)
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

	endpoint := a.AffBaseURL + "/Affiliation/" + a.ProjectSlug + "/add_identity/" + url.PathEscape(identity.Source)
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

func buildServices(a *Affiliation) (httpClientProvider *http.ClientProvider, esClientProvider *elastic.ClientProvider, err error) {
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
