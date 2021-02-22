package slk

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/slack-go/slack"
)

type Client struct {
	client *slack.Client
}

func NewClient() (*Client, error) {
	c := &Client{}
	if os.Getenv("SLACK_API_TOKEN") != "" {
		c.client = slack.New(os.Getenv("SLACK_API_TOKEN"))
	}
	return c, nil
}

func (c *Client) PostMessage(ctx context.Context, m string) error {
	switch {
	case os.Getenv("SLACK_API_TOKEN") != "":
		return fmt.Errorf("not implemented: %s", "slack api message")
	case os.Getenv("SLACK_WEBHOOK_URL") != "":
		url := os.Getenv("SLACK_WEBHOOK_URL")
		msg := buildWebhookMessage(m)
		return slack.PostWebhookContext(ctx, url, msg)
	default:
		return errors.New("not found environment for Slack [SLACK_WEBHOOK_URL SLACK_API_TOKEN]")
	}
	return nil
}

func buildWebhookMessage(m string) *slack.WebhookMessage {
	return &slack.WebhookMessage{
		Text: m,
	}
}
