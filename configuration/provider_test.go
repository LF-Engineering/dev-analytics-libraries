package configuration

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGet(t *testing.T) {
	// arrange
	s := NewLocalConfigStorage()

	// act
	srv := NewProvider(s)
	URL := "http://localhost:9200"
	err := srv.Set(EsURL, URL)

	// assert
	assert.NoError(t, err)
	esURL, err := srv.Get(EsURL)
	assert.NoError(t, err)
	assert.Equal(t, URL, esURL)

}
