package fixtures

import "regexp"

// Fixture struct represents a fixture model
type Fixture struct {
	Native struct {
		Slug string `yaml:"slug"`
	} `yaml:"native"`
	DataSources []struct {
		Slug     string `yaml:"slug"`
		Projects []struct {
			Name      string `yaml:"name"`
			Endpoints []struct {
				Name  string `yaml:"name"`
				Flags struct {
					Type string `yaml:"type"`
				} `yaml:"flags"`
				Skip    []string         `yaml:"skip"`
				SkipREs []*regexp.Regexp `yaml:"-"`
			} `yaml:"endpoints"`
			Flags interface{} `yaml:"flags"`
		} `yaml:"projects,omitempty"`
		Config []struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		} `yaml:"config,omitempty"`
		MaxFrequency string `yaml:"max_frequency,omitempty"`
	} `yaml:"data_sources"`
}

// Repo struct represents a single repo model
type Repo struct {
	Htmlurl string `json:"html_url"`
}

// RepoList represents list of Repos
type RepoList []Repo
