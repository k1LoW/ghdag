package slk

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

type SlkClient interface {
	PostMessage(ctx context.Context, m string) error
	GetMentionLinkByName(ctx context.Context, name string) (string, error)
}

type Client struct {
	client         *slack.Client
	channelCache   map[string]slack.Channel
	userCache      map[string]slack.User
	userGroupCache map[string]slack.UserGroup
}

func NewClient() (*Client, error) {
	c := &Client{
		channelCache:   map[string]slack.Channel{},
		userCache:      map[string]slack.User{},
		userGroupCache: map[string]slack.UserGroup{},
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
		return c.postWebbookMessage(ctx, m)
	default:
		return errors.New("not found environment for Slack: SLACK_API_TOKEN or SLACK_WEBHOOK_URL")
	}
	return nil
}

func (c *Client) postMessage(ctx context.Context, m string) error {
	if os.Getenv("SLACK_CHANNEL") == "" {
		return errors.New("not found environment for Slack: SLACK_CHANNEL")
	}
	channel := os.Getenv("SLACK_CHANNEL")
	channelID, err := c.getChannelIDByName(ctx, channel)
	if err != nil {
		return err
	}
	opts := []slack.MsgOption{
		slack.MsgOptionBlocks(buildBlocks(m)...),
	}

	if username := os.Getenv("SLACK_USERNAME"); username != "" {
		opts = append(opts, slack.MsgOptionUsername(username))
	}

	if emoji := os.Getenv("SLACK_ICON_EMOJI"); emoji != "" {
		opts = append(opts, slack.MsgOptionIconEmoji(emoji))
	}

	if url := os.Getenv("SLACK_ICON_URL"); url != "" {
		opts = append(opts, slack.MsgOptionIconURL(url))
	}

	if _, _, err := c.client.PostMessageContext(ctx, channelID, opts...); err != nil {
		return err
	}
	return nil
}

func (c *Client) postWebbookMessage(ctx context.Context, m string) error {
	url := os.Getenv("SLACK_WEBHOOK_URL")
	msg := buildWebhookMessage(m)
	return slack.PostWebhookContext(ctx, url, msg)
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
			c.channelCache[channel] = cc
			if cc.Name == channel {
				cID = cc.ID
				break L
			}
		}
		if nc == "" {
			break
		}
	}
	if cID == "" {
		return "", fmt.Errorf("not found channel: %s", channel)
	}

	return cID, nil
}

func (c *Client) GetMentionLinkByName(ctx context.Context, name string) (string, error) {
	if c.client == nil {
		c.client = slack.New(os.Getenv("SLACK_API_TOKEN"))
		defer func() {
			c.client = nil
		}()
	}
	name = strings.TrimPrefix(name, "@")
	switch name {
	case "channel", "here", "everyone":
		return fmt.Sprintf("<!%s>", name), nil
	}
	if uc, ok := c.userCache[name]; ok {
		// https://api.slack.com/reference/surfaces/formatting#mentioning-users
		return fmt.Sprintf("<@%s>", uc.ID), nil
	}
	if gc, ok := c.userGroupCache[name]; ok {
		// https://api.slack.com/reference/surfaces/formatting#mentioning-groups
		return fmt.Sprintf("<!subteam^%s>", gc.ID), nil
	}

	users, err := c.client.GetUsersContext(ctx)
	if err != nil {
		return "", err
	}

	for _, u := range users {
		c.userCache[u.Name] = u
	}
	uc, ok := c.userCache[name]
	if ok {
		return fmt.Sprintf("<@%s>", uc.ID), nil
	}

	groups, err := c.client.GetUserGroupsContext(ctx)
	if err != nil {
		return "", err
	}
	for _, g := range groups {
		c.userGroupCache[g.Handle] = g
	}
	gc, ok := c.userGroupCache[name]
	if ok {
		return fmt.Sprintf("<!subteam^%s>", gc.ID), nil
	}

	return "", fmt.Errorf("not found user or usergroup: %s", name)
}

func buildWebhookMessage(m string) *slack.WebhookMessage {
	return &slack.WebhookMessage{
		Channel: os.Getenv("SLACK_CHANNEL"),
		Blocks: &slack.Blocks{
			BlockSet: buildBlocks(m),
		},
	}
}

// buildBlocks
func buildBlocks(m string) []slack.Block {
	elements := []slack.MixedElement{slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("%s | <%s|#%s> | %s", os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GHDAG_TARGET_URL"), os.Getenv("GHDAG_TARGET_NUMBER"), os.Getenv("GHDAG_TASK_ID")), false, false)}
	contextBlock := slack.NewContextBlock("footer", elements...)
	return []slack.Block{
		slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", m, false, false), nil, nil),
		contextBlock,
	}
}
