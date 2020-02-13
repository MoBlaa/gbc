package twitchclient

import (
	"github.com/MoBlaa/gbc/twitchclient/modes"
	"net/url"
)

type Option func(client *Client)

func Server(url *url.URL) Option {
	return func(client *Client) {
		client.server = url
	}
}

func WithAuth(auth *OAuthAuthentication) Option {
	return func(client *Client) {
		client.auth = auth
	}
}

func SetChannels(channel string, additionals ...string) Option {
	return func(client *Client) {
		client.channels = append([]string{channel}, additionals...)
	}
}

func WithChannels(channels ...string) Option {
	return func(client *Client) {
		if len(channels) != 0 {
			client.channels = channels
		}
	}
}

func WithMembership() Option {
	return func(client *Client) {
		client.membership = true
	}
}

func WithTags() Option {
	return func(client *Client) {
		client.tags = true
	}
}

func WithCommands() Option {
	return func(client *Client) {
		client.commands = true
	}
}

func As(mode modes.MessageRateMode) Option {
	return func(client *Client) {
		client.mode = mode
	}
}
