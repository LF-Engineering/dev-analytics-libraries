package fixtures

import (
	"testing"

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
