package slack

import (
	"encoding/json"
	"errors"

	resty "gopkg.in/resty.v1"
)

var (
	errorBadResponse = "API call gave an other-than-200 result"
	errorEmptyText   = "Refusing to send an empty message, consider calling slack.SetText() first"
)

// Provider provides an interface to interact with the Slack Webhook API
type Provider struct {
	Debug      bool
	Payload    Payload
	WebhookURL string
}

// Payload holds information about the message we want to send. Any of these values except Text may be left empty to
// use the defaults for the Webhook. IconEmoji and IconURL are mutually exclusive, do not set both.
// https://api.slack.com/incoming-webhooks#posting_with_webhooks
type Payload struct {
	Channel   string `json:"channel,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	IconURL   string `json:"icon_url,omitempty"`
	Text      string `json:"text,omitempty"`
	Username  string `json:"username,omitempty"`
}

// New creates a new invocation of this Provider wrapper
func New(webhookURL string) Provider {
	return Provider{
		WebhookURL: webhookURL,
	}
}

// SetText sets the message text, i.e. what will be displayed in the Slack window
func (a *Provider) SetText(text string) {
	a.Payload.Text = text
}

// Send sends a message to Slack
func (a *Provider) Send() error {
	if a.Payload.Text == "" {
		return errors.New(errorEmptyText)
	}
	body, err := json.Marshal(a.Payload)
	if err != nil {
		return err
	}

	restyResult, err := resty.
		SetDebug(a.Debug).
		SetHostURL(a.WebhookURL).
		R().
		SetFormData(map[string]string{"payload": string(body)}).
		Post(a.WebhookURL)

	if err != nil {
		return err
	}

	if restyResult.StatusCode() != 200 {
		return errors.New(errorBadResponse)
	}

	return nil
}

// SendText sets message text and sends the message, a convenience wrapper for a.SetText and a.Send
func (a *Provider) SendText(text string) error {
	a.SetText(text)
	return a.Send()
}
