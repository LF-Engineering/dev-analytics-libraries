package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	http2 "github.com/LF-Engineering/dev-analytics-libraries/http"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"gopkg.in/yaml.v2"
)

func main() {
	fileList := make([]string, 0)
	err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		if ok := strings.HasSuffix(path, ".yaml"); ok {
			fileList = append(fileList, path)
		}
		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	esIndices := make(map[string]map[string]map[string]interface{})
	if len(os.Getenv("ES_URL")) == 0 {
		log.Fatal("ES_URL environment variable not set")
	}
	requestURL := fmt.Sprintf("%s%s", os.Getenv("ES_URL"), "/_aliases?pretty=true")
	resp, err := http.Get(requestURL)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err.Error())
		}
	}()

	err = json.NewDecoder(resp.Body).Decode(&esIndices)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("len(esIndices)", len(esIndices))

	finalObject := make(map[string]map[string]bool)

	projects, err := getProjectsFromProjectsService()
	if err != nil {
		log.Fatal(err)
	}

	sfProjects := make(map[string]project)
	for _, proj := range projects {
		sfProjects[proj.Slug] = proj
	}

	for k, aliases := range esIndices {
		if strings.HasSuffix(strings.TrimSpace(k), "github-pull_request") {
			continue
		}

		if strings.HasSuffix(strings.TrimSpace(k), "github-repository") {
			continue
		}

		if strings.HasSuffix(strings.TrimSpace(k), "-raw") {
			continue
		}

		aliases := aliases["aliases"]
		if aliases != nil {
			for alias := range aliases {
				finalObject[alias] = make(map[string]bool, 0)
			}
		}

		finalObject[k] = make(map[string]bool, 0)
	}

	config := RDSConfig{
		URL: os.Getenv("SH_DB"),
	}

	dbConn, err := ConnectDatabase(&config)
	if err != nil {
		log.Fatal(err)
	}
	dbService := New(dbConn)

	mappings := make(map[string]SlugMapping)
	slugMapping, err := dbService.GetSlugMapping()
	if err != nil {
		log.Fatal(err)
	}

	for _, mapping := range slugMapping {
		mappings[mapping.DAName] = mapping
	}

	finalOutputInInsightsAndSlugMappingNotInSfdc := make([]OutputStruct, 0)
	finalOutputNotInSlugMappingAndSfdc := make([]OutputStruct, 0)
	var DASlugs []string
	for _, file := range fileList {
		yamlFile, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("yamlFile.Get err #%v ", err)
		}

		var c Fixture
		err = yaml.Unmarshal(yamlFile, &c)
		if err != nil {
			log.Println(file)
			log.Fatalf("Unmarshal: %v", err)
		}
		DASlugs = append(DASlugs, c.Native.Slug)

		for _, datasource := range c.DataSources {
			projSlugFormatted := strings.ReplaceAll(c.Native.Slug, "/", "-")
			datasourceSlugFormatted := strings.ReplaceAll(datasource.Slug, "/", "-")
			index := fmt.Sprintf("sds-%s-%s", projSlugFormatted, datasourceSlugFormatted)

			if _, ok := finalObject[index]; !ok {
				if strings.HasSuffix(strings.TrimSpace(index), "github-pull_request") {
					continue
				}

				if strings.HasSuffix(strings.TrimSpace(index), "github-repository") {
					continue
				}

				if strings.HasSuffix(strings.TrimSpace(index), "-raw") {
					continue
				}
				// log indices we don't have in elastic search
				fmt.Println("index not in es: ", index)
				continue
			}

			if _, ok := mappings[c.Native.Slug]; ok {
				if _, exists := sfProjects[mappings[c.Native.Slug].SFName]; !exists {
					finalOutputInInsightsAndSlugMappingNotInSfdc = append(finalOutputInInsightsAndSlugMappingNotInSfdc, OutputStruct{
						IndexName:    index,
						InsightsSlug: mappings[c.Native.Slug].DAName,
						Disabled:     mappings[c.Native.Slug].IsDisabled,
					})
				}
			} else {
				finalOutputNotInSlugMappingAndSfdc = append(finalOutputNotInSlugMappingAndSfdc, OutputStruct{
					IndexName:    index,
					InsightsSlug: c.Native.Slug,
					Disabled:     c.Disabled,
				})
			}
		}

	}

	csvFile, err := os.Create("./data_in_da_and_slug_mapping_not_in_sfdc.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)

	for _, line := range finalOutputInInsightsAndSlugMappingNotInSfdc {
		var row []string
		row = append(row, line.IndexName)
		row = append(row, line.InsightsSlug)
		row = append(row, strconv.FormatBool(line.Disabled))
		if err := writer.Write(row); err != nil {
			log.Fatal(err)
		}
	}
	writer.Flush()

	csvFile, err = os.Create("./data_not_in_slug_mapping_and_sfdc.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	writer = csv.NewWriter(csvFile)

	for _, line := range finalOutputNotInSlugMappingAndSfdc {
		var row []string
		row = append(row, line.IndexName)
		row = append(row, line.InsightsSlug)
		row = append(row, strconv.FormatBool(line.Disabled))
		if err := writer.Write(row); err != nil {
			log.Fatal(err)
		}
	}
	writer.Flush()
}

// Fixture struct
type Fixture struct {
	Native struct {
		Slug string `yaml:"slug"`
	} `yaml:"native"`
	DataSources []struct {
		Slug     string `yaml:"slug,omitempty"`
		Projects []struct {
			Name      string `yaml:"name"`
			Endpoints []struct {
				Name string `yaml:"name"`
			} `yaml:"endpoints"`
			Flags interface{} `yaml:"flags"`
		} `yaml:"projects"`
		//Name     []string `yaml:"name"`
	} `yaml:"data_sources"`
	Disabled bool `yaml:"disabled"`
}

// RDSConfig struct
type RDSConfig struct {
	URL string
}

// ConnectDatabase initializes database connection
func ConnectDatabase(config *RDSConfig) (*sqlx.DB, error) {
	var dbConnection *sqlx.DB

	dbConnection, err := sqlx.Connect("mysql", config.URL)
	if err != nil {
		return nil, err
	}

	dbConnection.SetMaxOpenConns(1)    // The default is 0 (unlimited)
	dbConnection.SetMaxIdleConns(0)    // defaultMaxIdleConnections = 2, (0 is unlimited)
	dbConnection.SetConnMaxLifetime(0) // 0, connections are reused forever.

	fmt.Println("DB Service Initialized!")
	return dbConnection, nil
}

// service ...
type service struct {
	db *sqlx.DB
}

// SlugMapping struct
type SlugMapping struct {
	DAName     string      `db:"da_name" json:"da_name"`
	SfID       string      `db:"sf_id" json:"sf_id"`
	SFName     string      `db:"sf_name" json:"sf_name"`
	IsDisabled bool        `db:"is_disabled" json:"is_disabled"`
	CreatedAt  interface{} `db:"created_at" json:"created_at"`
}

// New creates new db service instance with given db
func New(db *sqlx.DB) Service {
	return &service{
		db: db,
	}
}

// Service ...
type Service interface {
	GetSlugMapping() ([]SlugMapping, error)
}

// GetSlugMapping returns a list of data from the slug_mapping table
func (s *service) GetSlugMapping() (mapping []SlugMapping, err error) {
	query := `
		SELECT *
		FROM
		slug_mapping;`
	err = s.db.Select(&mapping, query)
	return
}

// getProjectsFromProjectsService returns a list of all projects from the projects service
func getProjectsFromProjectsService() ([]project, error) {
	httpClient := http2.NewClientProvider(60 * time.Second)
	if len(os.Getenv("TOKEN")) == 0 {
		return nil, errors.New("TOKEN environment variable not set")
	}
	if len(os.Getenv("PROJECTS_SERVICE_BASE_URL")) == 0 {
		return nil, errors.New("PROJECTS_SERVICE_BASE_URL environment variable not set")
	}
	headers := make(map[string]string)
	headers["Accept"] = "application/json"
	headers["Authorization"] = os.Getenv("TOKEN")

	url := fmt.Sprintf("%s%s", os.Getenv("PROJECTS_SERVICE_BASE_URL"), "projects?pageSize=100000")
	_, resp, err := httpClient.Request(url, http.MethodGet, headers, nil, nil)
	if err != nil {
		return nil, err
	}

	var projects Data
	err = json.Unmarshal(resp, &projects)
	if err != nil {
		return nil, err
	}
	return projects.Data, nil
}

// project struct
type project struct {
	AutoJoinEnabled bool   `json:"AutoJoinEnabled" yaml:"AutoJoinEnabled"`
	Description     string `json:"Description" yaml:"Description"`
	Parent          string `json:"Parent" yaml:"Parent"`
	Name            string `json:"Name" yaml:"Name"`
	ProjectLogo     string `json:"ProjectLogo" yaml:"ProjectLogo"`
	RepositoryURL   string `json:"RepositoryURL" yaml:"RepositoryURL"`
	Slug            string `json:"Slug" yaml:"Slug"`
	StartDate       string `json:"StartDate" yaml:"StartDate"`
	Status          string `json:"Status" yaml:"Status"`
	Website         string `json:"Website" yaml:"Website"`
	Category        string `json:"Category" yaml:"Category"`
	CreatedDate     string `json:"CreatedDate" yaml:"CreatedDate"`
	EndDate         string `json:"EndDate" yaml:"EndDate"`
	Foundation      struct {
		ID      string `json:"ID" yaml:"ID"`
		LogoURL string `json:"LogoURL" yaml:"LogoURL"`
		Name    string `json:"Name" yaml:"Name"`
	} `json:"Foundation" yaml:"Foundation"`
	ID               string `json:"ID" yaml:"ID"`
	ModifiedDate     string `json:"ModifiedDate" yaml:"ModifiedDate"`
	OpportunityOwner struct {
		Email     string `json:"Email" yaml:"Email"`
		FirstName string `json:"FirstName" yaml:"FirstName"`
		ID        string `json:"ID" yaml:"ID"`
		LastName  string `json:"LastName" yaml:"LastName"`
	} `json:"OpportunityOwner" yaml:"OpportunityOwner"`
	Owner struct {
		Email     string `json:"Email" yaml:"Email"`
		FirstName string `json:"FirstName" yaml:"FirstName"`
		ID        string `json:"ID" yaml:"ID"`
		LastName  string `json:"LastName" yaml:"LastName"`
	} `json:"Owner" yaml:"Owner"`
	ProjectType    string    `json:"ProjectType" yaml:"ProjectType"`
	SystemModStamp string    `json:"SystemModStamp" yaml:"SystemModStamp"`
	Type           string    `json:"Type" yaml:"Type"`
	Projects       []project `json:"Projects" yaml:"Projects"`
}

// Data struct
type Data struct {
	Data []project `json:"Data"`
}

// OutputStruct struct
type OutputStruct struct {
	IndexName    string
	InsightsSlug string
	Disabled     bool
}
