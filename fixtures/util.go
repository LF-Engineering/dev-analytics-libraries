package fixtures

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

// CloneRepo method to clone dev-analytics-api repo as it contains the fixtures files
func CloneRepo() error {
	token := os.Getenv("GITHUB_OAUTH_TOKEN")
	stage := os.Getenv("STAGE")
	repo := "github.com/LF-Engineering/dev-analytics-api.git"
	url := fmt.Sprintf("https://%s@%s", token, repo)
	options := git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	}
	_, err := git.PlainClone("./data", false, &options)
	if err != nil {
		log.Printf("Error cloning repo: %+v", err)
	}

	// checkout branch will be test or prod based on STAGE env var
	cmd := exec.Command("git", "checkout", stage)
	cmd.Dir = "./data"
	err = cmd.Run()
	if err != nil {
		return err
	}

	// if repo cloned already pull latest
	cmd = exec.Command("git", "pull", "origin", stage)
	cmd.Dir = "./data"
	err = cmd.Run()
	if err != nil {
		return err
	}

	return err
}

// GetYamlFiles returns all the fixture yaml files
func GetYamlFiles() []string {
	var result = make([]string, 0)
	err := filepath.Walk("./data/app/services/lf/bootstrap/fixtures",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".yaml" {
				result = append(result, path)
			}
			return nil
		})
	if err != nil {
		log.Printf("Error listing yaml files: %+v", err)
	}
	return result
}

// ParseYamlFile reads the fixture yaml file and returns the model
func ParseYamlFile(path string) Fixture {
	a, err := ioutil.ReadFile(path)
	var f Fixture
	if err != nil {
		log.Printf("Error parsing yaml file: %+v", err)
		return f
	}
	err = yaml.Unmarshal(a, &f)
	if err != nil {
		log.Printf("Error unmarshalling yaml file: %+v", err)
	}
	return f
}

// FilterUniqueList returns unique elements in the data array
func FilterUniqueList(data []string) []string {
	occured := map[string]bool{}
	result := []string{}
	for element := range data {
		// check if already the mapped variable is set to true (exists) or not
		if occured[data[element]] != true {
			occured[data[element]] = true
			// Append to result slice.
			result = append(result, data[element])
		}
	}

	return result
}

// ParseOrg returns the org name by parsing the input param
func ParseOrg(endPointName string) string {
	arr := strings.Split(endPointName, "/")
	ary := []string{}
	l := len(arr) - 1
	for i, s := range arr {
		if i == l && s == "" {
			break
		}
		ary = append(ary, s)
	}
	lAry := len(ary)
	org := ary[lAry-1]
	return org
}

// GetGithubRepoList returns the list of repositories for a given endpoint
func GetGithubRepoList(endPointName string, skipREs []*regexp.Regexp) ([]string, error) {
	token := os.Getenv("GITHUB_OAUTH_TOKEN")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	url, err := url.Parse(endPointName)
	if err != nil {
		log.Printf("Error parsing endpoint url: %+v", err)
		return nil, err
	}
	repos := []string{}
	// list public repositories for org "github"
	opt := &github.RepositoryListByOrgOptions{Type: "public"}
	opt.PerPage = 100
	for {
		repoList, response, err := client.Repositories.ListByOrg(ctx, strings.TrimLeft(url.Path, "/"), opt)
		if err != nil {
			log.Printf("Error fetching repo list: %+v", err)
			return nil, err
		}
		if len(repoList) > 0 {
			org := ParseOrg(endPointName)
			for _, repo := range repoList {
				if repo.Name != nil {
					name := org + "/" + *(repo.Name)
					if CheckSkipped(endPointName, skipREs, name) {
						repos = append(repos, *(repo.HTMLURL))
					}
				}
			}
		}
		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}
	return repos, nil
}

// CheckSkipped verifies whether a particular repo is in the skipped list or not.
func CheckSkipped(endPointName string, skipREs []*regexp.Regexp, repo string) bool {
	included := true
	for _, skipRE := range skipREs {
		if skipRE.MatchString(repo) {
			included = false
			log.Printf("%s: skipped %s (%v)\n", endPointName, repo, skipRE)
			break
		}
	}
	return included
}
