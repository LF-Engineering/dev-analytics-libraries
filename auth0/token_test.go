package auth0

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	// arrange

	// act
	auth0, err := NewAuth0Client("", "", "", "test", "",
		"", "", "", "")
	if err != nil {
		t.Error(err)
	}
	_, err = auth0.generateToken()
	//assert
	assert.NoError(t, err)
}
