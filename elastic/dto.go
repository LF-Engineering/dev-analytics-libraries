package elastic

import (
	"github.com/elastic/go-elasticsearch/v8"
)

// ClientProvider ...
type ClientProvider struct {
	client *elasticsearch.Client
	params *Params
}

// Params ...
type Params struct {
	URL      string
	Username string
	Password string
}

// TopHitsStruct result
type TopHitsStruct struct {
	Took         int          `json:"took"`
	Hits         Hits         `json:"hits"`
	Aggregations Aggregations `json:"aggregations"`
}

// Hits struct
type Hits struct {
	Total    Total         `json:"total"`
	MaxScore float32       `json:"max_score"`
	Hits     []*NestedHits `json:"hits"`
}

// Total struct
type Total struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

// NestedHits struct
type NestedHits struct {
	Index  string      `json:"_index"`
	Type   string      `json:"_type"`
	ID     string      `json:"_id"`
	Score  float64     `json:"_score"`
	Source interface{} `json:"_source"`
}

// Aggregations represents elastic Aggregations result
type Aggregations struct {
	Stat Stat `json:"stat"`
}

// Stat represents elastic stat result
type Stat struct {
	Value         float64 `json:"value"`
	ValueAsString string  `json:"value_as_string"`
}

// BulkData to be saved using bulkIndex
type BulkData struct {
	IndexName string `json:"index_name"`
	ID        string
	Data      interface{}
}
