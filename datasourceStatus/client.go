package datasourceStatus

import (
	"errors"
	"fmt"
	"github.com/LF-Engineering/dev-analytics-libraries/uuid"
	jsoniter "github.com/json-iterator/go"
	"time"
)

const cacheIndex = "sds-datasource-status"

// ESClientProvider used in connecting to ES server
type ESClientProvider interface {
	CreateDocument(index, documentID string, body []byte) ([]byte, error)
	Get(index string, query map[string]interface{}, result interface{}) error
	UpdateDocument(index string, id string, body interface{}) ([]byte, error)
}

// ClientProvider ...
type ClientProvider struct {
	esClient    ESClientProvider
	environment string
}

// NewClientProvider ...
func NewClientProvider(esClient ESClientProvider, environment string) (*ClientProvider, error) {
	client := &ClientProvider{
		esClient:    esClient,
		environment: environment,
	}

	return client, nil
}

// StoreDatasourceStatus ...
func (s *ClientProvider) StoreDatasourceStatus(status Status) error {
	if status.Datasource == "" || status.ProjectSlug == "" || status.Endpoint == "" {
		return errors.New("err : status datasource, project slug and endpoint are all required")
	}

	docID, err := uuid.Generate(status.ProjectSlug, status.Datasource, status.Endpoint)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	status.CreatedAt = now
	status.UpdatedAt = now
	b, err := jsoniter.Marshal(status)
	if err != nil {
		return err
	}

	index := fmt.Sprintf("%s-%s", cacheIndex, s.environment)
	_, err = s.esClient.CreateDocument(index, docID, b)
	if err != nil {
		if err.Error() == "create document failed" {
			return s.updateDocument(status, index, docID)
		}
		return err
	}

	return nil
}

// PullDatasourceStatus ...
func (s *ClientProvider) PullDatasourceStatus(projectSlug string, datasource string, endpoint string) (*Status, error) {
	docID, err := uuid.Generate(projectSlug, datasource, endpoint)
	if err != nil {
		return &Status{}, err
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"_id": map[string]string{
					"value": docID},
			},
		},
	}

	var res TopHits
	err = s.esClient.Get(fmt.Sprintf("%s-%s", cacheIndex, s.environment), query, res)
	if err != nil {
		return &Status{}, err
	}

	if len(res.Hits.Hits) == 1 {
		return &res.Hits.Hits[0].Source, nil
	}

	return &Status{}, fmt.Errorf("error getting Datasource Status, %v", res)
}

func (s *ClientProvider) updateDocument(status Status, index string, docID string) error {
	doc := map[string]interface{}{
		"project_slug":          status.ProjectSlug,
		"datasource":            status.Datasource,
		"endpoint":              status.Endpoint,
		"updated_at":            status.UpdatedAt,
		"error_message":         status.ErrorMessage,
		"status":                status.Status,
		"last_successful_event": status.LastSuccessfulEvent,
	}

	_, err := s.esClient.UpdateDocument(index, docID, doc)
	if err != nil {
		return err
	}
	return nil
}
