package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	libHttp "github.com/LF-Engineering/dev-analytics-libraries/http"
	"github.com/avast/retry-go"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	errs "github.com/pkg/errors"
)

// NewClientProvider ...
func NewClientProvider(params *Params) (*ClientProvider, error) {
	config := elasticsearch.Config{
		Addresses: []string{params.URL},
		Username:  params.Username,
		Password:  params.Password,
	}
	client, err := elasticsearch.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ClientProvider{client, params}, err
}

// CheckIfIndexExists checks if an es index exists and returns a bool depending on whether it exists or not.
func (p *ClientProvider) CheckIfIndexExists(index string) (bool, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_phrase": map[string]interface{}{
				"_index": index,
			},
		},
	}

	_, err := p.Search(index, query)
	if err != nil {
		if err.Error() == "index doesn't exist" {
			return false, nil
		}
		return false, errs.Wrap(err, "[CheckIfIndexExists] invalid request")
	}

	// index exists so return true
	return true, nil
}

// CreateIndex ...
func (p *ClientProvider) CreateIndex(index string, body []byte) ([]byte, error) {
	buf := bytes.NewReader(body)

	// Create Index request
	res, err := esapi.IndicesCreateRequest{
		Index: index,
		Body:  buf,
	}.Do(context.Background(), p.client)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	resBytes, err := toBytes(res)
	if err != nil {
		return nil, err
	}

	return resBytes, nil
}

// DeleteIndex removes existing index
func (p *ClientProvider) DeleteIndex(index string, ignoreUnavailable bool) ([]byte, error) {
	res, err := esapi.IndicesDeleteRequest{
		Index:             []string{index},
		IgnoreUnavailable: &ignoreUnavailable,
	}.Do(context.Background(), p.client)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	body, err := toBytes(res)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 200 {
		return body, nil
	}

	if res.IsError() {

		var e map[string]interface{}
		if err = jsoniter.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return nil, err
	}

	return body, nil
}

// DeleteDocumentByQuery ...
func (p *ClientProvider) DeleteDocumentByQuery(index string, query map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(query)
	if err != nil {
		return nil, err
	}

	res, err := p.client.DeleteByQuery(
		[]string{index},
		&buf)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	resBytes, err := toBytes(res)
	if err != nil {
		return nil, err
	}

	return resBytes, nil
}

// convert response to bytes
func toBytes(res *esapi.Response) ([]byte, error) {
	var resBuf bytes.Buffer
	if _, err := resBuf.ReadFrom(res.Body); err != nil {
		return nil, err
	}
	resBytes := resBuf.Bytes()
	return resBytes, nil
}

// Add ...
func (p *ClientProvider) Add(index string, documentID string, body []byte) ([]byte, error) {
	buf := bytes.NewReader(body)

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       buf,
	}

	res, err := req.Do(context.Background(), p.client)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	resBytes, err := toBytes(res)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 200 {
		return resBytes, nil
	}

	if res.IsError() {

		var e map[string]interface{}
		if err = jsoniter.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return nil, err
	}

	return resBytes, nil
}

// Bulk ...
func (p *ClientProvider) Bulk(body []byte) ([]byte, error) {
	buf := bytes.NewReader(body)

	req := esapi.BulkRequest{
		Body: buf,
	}

	res, err := req.Do(context.Background(), p.client)
	if err != nil {
		log.Printf("ReqErr: %s", err.Error())
		return nil, err
	}

	if res.IsError() {
		var e map[string]interface{}
		if err = jsoniter.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	resBytes, err := toBytes(res)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 200 {
		return resBytes, nil
	}

	if res.StatusCode == 413 {
		return nil, errors.New("payload too large. decrease documents to <= 1000")
	}

	if res.IsError() {

		var e map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return nil, err
	}

	return resBytes, nil
}

// BulkInsert inserts more than one item using one request
func (p *ClientProvider) BulkInsert(data []BulkData) ([]byte, error) {
	lines := make([]interface{}, 0)

	for _, item := range data {
		indexName := item.IndexName
		index := map[string]interface{}{
			"index": map[string]string{
				"_index": indexName,
				"_id":    item.ID,
			},
		}
		lines = append(lines, index)
		lines = append(lines, "\n")
		lines = append(lines, item.Data)
		lines = append(lines, "\n")
	}

	body, err := json.Marshal(lines)
	if err != nil {
		return nil, errors.New("unable to convert body to json")
	}

	var re = regexp.MustCompile(`(}),"\\n",?`)
	body = []byte(re.ReplaceAllString(strings.TrimSuffix(strings.TrimPrefix(string(body), "["), "]"), "$1\n"))

	resData, err := p.Bulk(body)
	if err != nil {
		return nil, err
	}

	return resData, nil
}

// BulkUpdate update more than one item using one request
func (p *ClientProvider) BulkUpdate(data []BulkData) ([]byte, error) {
	lines := make([]interface{}, 0)

	for _, item := range data {
		indexName := item.IndexName
		index := map[string]interface{}{
			"update": map[string]string{
				"_index": indexName,
				"_id":    item.ID,
			},
		}
		lines = append(lines, index)
		lines = append(lines, "\n")
		lines = append(lines, item.Data)
		lines = append(lines, "\n")
	}

	body, err := json.Marshal(lines)
	if err != nil {
		return nil, errors.New("unable to convert body to json")
	}

	var re = regexp.MustCompile(`(}),"\\n",?`)
	body = []byte(re.ReplaceAllString(strings.TrimSuffix(strings.TrimPrefix(string(body), "["), "]"), "$1\n"))

	resData, err := p.Bulk(body)
	if err != nil {
		return nil, err
	}

	return resData, nil
}

// BulkDelete deletes more than one item using one request
func (p *ClientProvider) BulkDelete(data []BulkData) ([]byte, error) {
	lines := make([]interface{}, 0)

	for _, item := range data {
		indexName := item.IndexName
		deleteQuery := map[string]interface{}{
			"delete": map[string]string{
				"_index": indexName,
				"_id":    item.ID,
			},
		}
		lines = append(lines, deleteQuery)
		lines = append(lines, "\n")
	}

	body, err := json.Marshal(lines)
	if err != nil {
		return nil, errors.New("unable to convert body to json")
	}

	var re = regexp.MustCompile(`(}),"\\n",?`)
	body = []byte(re.ReplaceAllString(strings.TrimSuffix(strings.TrimPrefix(string(body), "["), "]"), "$1\n"))

	resData, err := p.Bulk(body)
	if err != nil {
		return nil, err
	}

	return resData, nil
}

// Get query result
func (p *ClientProvider) Get(index string, query map[string]interface{}, result interface{}) (err error) {
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(query)
	if err != nil {
		return err
	}

	res, err := p.client.Search(
		p.client.Search.WithIndex(index),
		p.client.Search.WithBody(&buf),
	)
	if err != nil {
		return err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	if res.StatusCode == 200 {
		// index exists so return true
		if err = json.NewDecoder(res.Body).Decode(result); err != nil {
			return err
		}

		return nil
	}

	if res.IsError() {
		if res.StatusCode == 404 {
			// index doesn't exist
			return errors.New("index doesn't exist")
		}

		var e map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
			return err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return err
	}

	return nil
}

// GetStat gets statistics ex. max min, avg
func (p *ClientProvider) GetStat(index string, field string, aggType string, mustConditions []map[string]interface{}, mustNotConditions []map[string]interface{}) (result time.Time, err error) {

	hits := &TopHitsStruct{}

	q := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{},
		},
		"aggs": map[string]interface{}{
			"stat": map[string]interface{}{
				aggType: map[string]interface{}{
					"field": field,
				},
			},
		},
	}

	if mustConditions != nil {
		q["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = mustConditions
	}

	if mustNotConditions != nil {
		q["query"].(map[string]interface{})["bool"].(map[string]interface{})["must_not"] = mustNotConditions
	}
	err = p.Get(index, q, hits)
	if err != nil {
		return time.Now().UTC(), err
	}
	date, err := time.Parse(time.RFC3339, hits.Aggregations.Stat.ValueAsString)
	if err != nil {
		return time.Now().UTC(), err
	}

	return date, nil
}

// DelayOfCreateIndex delay creating index and retry if fails
func (p *ClientProvider) DelayOfCreateIndex(ex func(str string, b []byte) ([]byte, error), uin uint, du time.Duration, index string, data []byte) error {

	retry.Attempts(uin)
	retry.Delay(du)

	err := retry.Do(func() error {
		_, err := ex(index, data)
		return err
	}, retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
		return retry.BackOffDelay(n, err, config)
	}))

	return err
}

// Search ...
func (p *ClientProvider) Search(index string, query map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(query)
	if err != nil {
		return nil, err
	}

	res, err := p.client.Search(
		p.client.Search.WithIndex(index),
		p.client.Search.WithBody(&buf),
	)
	if err != nil {
		if strings.Contains(err.Error(), "server is not Elasticsearch") {
			fmt.Println("esssssss", p.params)
		}
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	if res.StatusCode == 200 {
		var in interface{}
		// index exists so return true
		if err = json.NewDecoder(res.Body).Decode(&in); err != nil {
			return nil, err
		}

		bites, err := jsoniter.Marshal(in)
		if err != nil {
			return nil, err
		}

		return bites, nil
	}

	if res.IsError() {
		if res.StatusCode == 404 {
			// index doesn't exist
			return nil, errors.New("index doesn't exist")
		}

		var e map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return nil, err
	}

	return nil, errors.New("search failed")
}

// SearchWithNoIndex for querying across multiple indices for example: GET _search?allow_no_indices=true
func (p *ClientProvider) SearchWithNoIndex(query map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(query)
	if err != nil {
		return nil, err
	}
	res, err := p.client.Search(
		p.client.Search.WithAllowNoIndices(true),
		p.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	if res.StatusCode == 200 {
		var in interface{}
		if err = json.NewDecoder(res.Body).Decode(&in); err != nil {
			return nil, err
		}

		bites, err := jsoniter.Marshal(in)
		if err != nil {
			return nil, err
		}
		return bites, nil
	}

	if res.IsError() {
		if res.StatusCode == 404 {
			// index doesn't exist
			return nil, errors.New("index doesn't exist")
		}

		var e map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return nil, err
	}

	return nil, errors.New("search failed")
}

// CreateDocument ...
func (p *ClientProvider) CreateDocument(index, documentID string, body []byte) ([]byte, error) {
	buf := bytes.NewReader(body)

	// Create es document request
	res, err := esapi.CreateRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       buf,
	}.Do(context.Background(), p.client)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		var in interface{}
		if err = json.NewDecoder(res.Body).Decode(&in); err != nil {
			return nil, err
		}

		bites, err := jsoniter.Marshal(in)
		if err != nil {
			return nil, err
		}
		return bites, nil
	}

	if res.IsError() {
		if res.StatusCode == 404 {
			// index doesn't exist
			return nil, errors.New("index doesn't exist")
		}

		var e map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return nil, err
	}

	return nil, errors.New("create document failed")
}

// UpdateDocumentByQuery ...
func (p *ClientProvider) UpdateDocumentByQuery(index, query, fields string) ([]byte, error) {
	// update es document request
	res, err := p.client.UpdateByQuery(
		[]string{index},
		p.client.UpdateByQuery.WithQuery(query),
		p.client.UpdateByQuery.WithBody(strings.NewReader(fields)))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	resBytes, err := toBytes(res)
	if err != nil {
		return nil, err
	}

	return resBytes, nil
}

// ReadWithScroll scrolls through the pages of size given in the query and adds up the scrollID in the result
// Which is expected in the subsequent function call to get the next page, empty result indicates the end of the page
func (p *ClientProvider) ReadWithScroll(index string, query map[string]interface{}, result interface{}, scrollID string) (err error) {
	var res *esapi.Response
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	if scrollID == "" {
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(query)
		if err != nil {
			return err
		}

		res, err = p.client.Search(
			p.client.Search.WithIndex(index),
			p.client.Search.WithBody(&buf),
			p.client.Search.WithScroll(time.Minute),
		)
	} else {
		res, err = p.client.Scroll(p.client.Scroll.WithScrollID(scrollID), p.client.Scroll.WithScroll(time.Minute))
	}
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK {
		if err = json.NewDecoder(res.Body).Decode(result); err != nil {
			return err
		}

		return nil
	}
	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			// index doesn't exist
			return errors.New("index doesn't exist")
		}

		var e map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
			return err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return err
	}
	return nil
}

// UpdateDocument update elastic single document
func (p *ClientProvider) UpdateDocument(index string, id string, body interface{}) ([]byte, error) {

	m := make(map[string]interface{})
	m["doc"] = body
	b, err := jsoniter.Marshal(m)
	if err != nil {
		return nil, err
	}
	buf := strings.NewReader(string(b))

	// Create Index request
	res, err := esapi.UpdateRequest{
		Index:      index,
		DocumentID: id,
		Body:       buf,
		Refresh:    "true",
	}.Do(context.Background(), p.client)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	if res.StatusCode == http.StatusOK {
		var in interface{}
		if err = json.NewDecoder(res.Body).Decode(&in); err != nil {
			return nil, err
		}

		bites, err := jsoniter.Marshal(in)
		if err != nil {
			return nil, err
		}
		return bites, nil
	}

	if res.IsError() {
		if res.StatusCode == 404 {
			// index doesn't exist
			return nil, errors.New("index doesn't exist")
		}

		var e map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return nil, err
	}

	return nil, errors.New("update document failed")
}

// GetIndices get all indices based on a specific pattern , or you can use _all to get all indices
func (p *ClientProvider) GetIndices(pattern string) ([]string, error) {
	res, err := esapi.IndicesGetRequest{
		Index:  []string{pattern},
		Pretty: true,
	}.Do(context.Background(), p.client)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	resBytes, err := toBytes(res)
	if err != nil {
		return nil, err
	}

	var ind map[string]interface{}
	err = json.Unmarshal(resBytes, &ind)
	if err != nil {
		return nil, err
	}

	var indices []string
	for key := range ind {
		indices = append(indices, key)
	}
	return indices, nil
}

// Count get documents count based on query
func (p *ClientProvider) Count(index string, query map[string]interface{}) (int, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(query)
	if err != nil {
		return 0, err
	}

	res, err := p.client.Count(
		p.client.Count.WithIndex(index),
		p.client.Count.WithBody(&buf),
	)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()

	if res.StatusCode == 200 {
		result := make(map[string]interface{})
		// index exists so return true
		if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
			return 0, err
		}

		floatCount := result["count"].(float64)
		count := int(floatCount)
		return count, nil
	}

	if res.IsError() {
		if res.StatusCode == 404 {
			// index doesn't exist
			return 0, errors.New("index doesn't exist")
		}

		var e map[string]interface{}
		if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
			return 0, err
		}

		err = fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
		return 0, err
	}

	return 0, nil
}

// CreateUUID calls checkIfUUIDExists
// if a uuid exists in the index then it generates and returns a new one
func (p *ClientProvider) CreateUUID(index string) (string, error) {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	ok, err := p.CheckIfUUIDExists(index, newUUID.String())
	if err != nil {
		return "", errs.Wrap(err, "CreateUUID")
	}

	if ok {
		p.CreateUUID(index)
	}

	return newUUID.String(), nil
}

// CheckIfUUIDExists checks whether a uuid exists as a document id in an index.
// Returns true or false depending on whether the uuid exists or not
func (p *ClientProvider) CheckIfUUIDExists(index, uuidString string) (bool, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_phrase": map[string]interface{}{
				"_id": fmt.Sprintf("%s", uuidString),
			},
		},
	}

	res, err := p.Search(index, query)
	if err != nil {
		return false, errs.Wrap(err, "checkIfUUIDExists invalid request")
	}

	var data TopHitsStruct
	if err = json.Unmarshal(res, &data); err != nil {
		return false, errs.Wrap(err, "checkIfUUIDExists failed to parse json")
	}

	for _, v := range data.Hits.Hits {
		if uuidString == v.ID {
			return true, nil
		}
	}
	return false, nil
}

// UpdateFieldByQuery updates a single field in an es document
//  params es credentials
//  index es index
//  matchFieldName es field name to match in query eg author_uuid
//  matchValue es field value to match in query eg 97fa918c612a2fda17ba5aa1e1fc933a00e020d7
//  updateFieldName es field to update eg author_name
//  updateValue es field value to update to eg Rob Underwood
func (p *ClientProvider) UpdateFieldByQuery(params Params, index, matchFieldName, matchValue, updateFieldName, updateValue string) (bool, error) {
	httpClientProvider := libHttp.NewClientProvider(time.Minute)
	url := fmt.Sprintf("https://%s:%s@%s/%s/_update_by_query", params.Username, params.Password, params.URL, index)

	updateQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"match_phrase": map[string]interface{}{
							fmt.Sprintf("%s", matchFieldName): fmt.Sprintf("%s", matchValue),
						},
					},
				},
			},
		},
		"script": fmt.Sprintf("ctx._source.'%s' = '%s';", updateFieldName, updateValue),
	}

	updateBytes, err := json.Marshal(updateQuery)
	if err != nil {
		return false, err
	}

	statusCode, _, err := httpClientProvider.Request(url, http.MethodPost, nil, updateBytes, nil)
	if statusCode == http.StatusOK {
		return true, nil
	}
	return false, err
}
