package slk

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

type Client struct {
	client       *slack.Client
	channelCache map[string]*slack.Channel
	userCache    map[string]*slack.User
}

func NewClient() (*Client, error) {
	c := &Client{
		channelCache: map[string]*slack.Channel{},
		userCache:    map[string]*slack.User{},
	}
	if os.Getenv("SLACK_API_TOKEN") != "" {
		c.client = slack.New(os.Getenv("SLACK_API_TOKEN"))
	}
	return c, nil
}

func (c *Client) PostMessage(ctx context.Context, m string) error {
	switch {
	case c.client != nil:
		return c.postMessage(ctx, m)
	case os.Getenv("SLACK_API_TOKEN") != "":
		// temporary
		c.client = slack.New(os.Getenv("SLACK_API_TOKEN"))
		err := c.postMessage(ctx, m)
		c.client = nil
		return err
	case os.Getenv("SLACK_WEBHOOK_URL") != "":
		url := os.Getenv("SLACK_WEBHOOK_URL")
		msg := buildWebhookMessage(m)
		return slack.PostWebhookContext(ctx, url, msg)
	default:
		return errors.New("not found environment for Slack: SLACK_API_TOKEN or SLACK_WEBHOOK_URL")
	}
	return nil
}

func (c *Client) postMessage(ctx context.Context, m string) error {
	if os.Getenv("SLACK_CHANNEL") == "" {
		return errors.New("not found environment for Slack: SLACK_NOTIFY_CHANNEL")
	}
	channel := os.Getenv("SLACK_CHANNEL")
	channelID, err := c.getChannelIDByName(ctx, channel)
	if err != nil {
		return err
	}
	if _, _, err := c.client.PostMessageContext(ctx, channelID, slack.MsgOptionText(m, true)); err != nil {
		return err
	}
	return nil
}

func (c *Client) getChannelIDByName(ctx context.Context, channel string) (string, error) {
	channel = strings.TrimPrefix(channel, "#")
	if cc, ok := c.channelCache[channel]; ok {
		return cc.ID, nil
	}
	var (
		nc  string
		err error
		cID string
	)
L:
	for {
		var ch []slack.Channel
		p := &slack.GetConversationsParameters{
			Limit:  1000,
			Cursor: nc,
		}
		ch, nc, err = c.client.GetConversationsContext(ctx, p)
		if err != nil {
			return "", err
		}
		for _, cc := range ch {
			c.channelCache[channel] = &cc
			if cc.Name == channel {
				cID = cc.ID
				break L
			}
		}
		if nc == "" {
			break
		}
	}

	return cID, nil
}

func buildWebhookMessage(m string) *slack.WebhookMessage {
	return &slack.WebhookMessage{
		Text: m,
	}
}
