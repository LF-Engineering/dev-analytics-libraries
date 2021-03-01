package fixtures

import (
	"testing"

	"net/url"

	"github.com/stretchr/testify/assert"
)

// Check the format date util method returns expected computed value
func TestFilterUniqueList(t *testing.T) {
	data := []string{"A1", "A2", "A3", "A4", "A2", "A5", "A3"}
	expected := []string{"A1", "A2", "A3", "A4", "A5"}
	result := FilterUniqueList(data)
	assert.Equal(t, result, expected)
}

func TestParseOrg(t *testing.T) {
	endPoint := "https://github.com/stretchr"
	expected := "stretchr"
	result := ParseOrg(endPoint)
	assert.Equal(t, result, expected)
}

func TestGetGerritRepo(t *testing.T) {
	// Invalid endpoint test
	endPoint := "http://dummygerritendpoint"
	var expected []string
	projects, repos, err := GetGerritRepos(endPoint)
	_, ok := err.(*url.Error)
	assert.Equal(t, expected, projects)
	assert.Equal(t, expected, repos)
	assert.Equal(t, true, ok)
}
