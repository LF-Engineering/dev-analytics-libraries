package slack

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generic helper for loading Slack API wrapper with Webhook URL
func getAPI(t *testing.T) Provider {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	require.NotEmpty(t, webhookURL, "ParameterNotFound: SLACK_WEBHOOK_URL")

	api := New(webhookURL)
	api.Payload.IconEmoji = ":robot_face:"
	api.Payload.Username = "Insights Slack - Unit Test"

	return api
}

func TestSendEmpty(t *testing.T) {
	api := getAPI(t)

	err := api.Send()
	assert.EqualError(t, err, errorEmptyText)
}

func TestSendNotEmpty(t *testing.T) {
	api := getAPI(t)
	api.SetText("TestSendNotEmpty Test Content")

	err := api.Send()
	assert.NoError(t, err, "unexpected err from api.Send()")
}

func TestSendText(t *testing.T) {
	api := getAPI(t)

	err := api.SendText("TestSendText Test Content")
	assert.NoError(t, err, "unexpected err from api.SendText()")
}
