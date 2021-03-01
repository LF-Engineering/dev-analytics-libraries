package fixtures

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v33/github"

	jsoniter "github.com/json-iterator/go"
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
func GetGithubRepoList(endPointName string, skipREs []*regexp.Regexp, onlyREs []*regexp.Regexp) ([]string, error) {
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
					name := "/" + org + "/" + *(repo.Name)
					included := CheckIncluded(endPointName, name, skipREs, onlyREs)
					if !included {
						continue
					}
					repos = append(repos, *(repo.HTMLURL))

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

// CheckIncluded verifies whether a particular repo is in the skipped list or not.
func CheckIncluded(endPointName string, repo string, skipREs []*regexp.Regexp, onlyREs []*regexp.Regexp) bool {
	for _, skipRE := range skipREs {
		if skipRE.MatchString(repo) {
			log.Printf("%s: skipped %s (%v)\n", endPointName, repo, skipRE)
			return false
		}
	}
	if len(onlyREs) == 0 {
		return true
	}
	included := false
	for _, onlyRE := range onlyREs {
		if onlyRE.MatchString(repo) {
			log.Printf("%s: included %s (%v)\n", endPointName, repo, onlyRE)
			included = true
			break
		}
	}

	return included
}

// GetGerritRepos - return list of repos for given gerrit server (uses HTML crawler)
func GetGerritRepos(gerritURL string) ([]string, []string, error) {
	var (
		body            []byte
		err             error
		i               int
		b               byte
		projects, repos []string
		req             *http.Request
		resp            *http.Response
		result          map[string]interface{}
	)

	partials := []string{"r", "gerrit"}
	for _, partial := range partials {
		method := http.MethodGet
		if !strings.HasSuffix(gerritURL, "/") {
			gerritURL += "/"
		}
		url := gerritURL + partial + "/projects/"

		req, err = http.NewRequest(method, os.ExpandEnv(url), nil)
		if err != nil {
			fmt.Printf("new request error: %+v for %s url: %s", err, method, url)
			return projects, repos, err
		}

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("do request error: %+v for %s url: %s\n", err, method, url)
			return projects, repos, err
		}

		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("ReadAll request error: %+v for %s url: %s\n", err, method, url)
			return projects, repos, err
		}

		if resp.StatusCode == http.StatusNotFound {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("Method:%s url:%s status:%d\n%s", method, url, resp.StatusCode, body)
			return projects, repos, err
		}

		jsonStart := []byte("{")[0]
		for i, b = range body {
			if b == jsonStart {
				break
			}
		}
		body = body[i:]
		err = jsoniter.Unmarshal(body, &result)
		if err != nil {
			fmt.Printf("Bulk result unmarshal error: %+v", err)
			return projects, repos, err
		}

		for project := range result {
			ary := strings.Split(project, "/")
			org := ary[0]
			endpoint := gerritURL + partial + "/" + project
			projects = append(projects, org)
			repos = append(repos, endpoint)
		}
		break
	}
	return projects, repos, nil
}
