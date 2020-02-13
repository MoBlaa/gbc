package twitchclient

import (
	"github.com/MoBlaa/gbc/twitchclient/modes"
	"net/url"
)

// Option to be applied to a client. Options are used for a good maintainable and fluid construction of twitch clients.
type Option func(client *Client)

// Server to use to connect to twitch.
func Server(url *url.URL) Option {
	return func(client *Client) {
		client.server = url
	}
}

// SetAuth sets the Authentication used to connect to twitch.
func SetAuth(auth *TwitchAuthentication) Option {
	return func(client *Client) {
		client.auth = auth
	}
}

// SetChannels sets the channels to which the client should receive messages from.
func SetChannels(channel string, additionals ...string) Option {
	return func(client *Client) {
		client.channels = append([]string{channel}, additionals...)
	}
}

// WithChannels adds channels to the one noted in other options and the default ones.
func WithChannels(channels ...string) Option {
	return func(client *Client) {
		if len(channels) != 0 {
			client.channels = channels
		}
	}
}

// WithMembership enables the client to request that the messages from twitch should contain membership information.
func WithMembership() Option {
	return func(client *Client) {
		client.membership = true
	}
}

// WithTags enables the client to request that the messages from twitch should contain information in tags.
func WithTags() Option {
	return func(client *Client) {
		client.tags = true
	}
}

// WithCommands enables the client to request that the messages from twitch should contain commands.
func WithCommands() Option {
	return func(client *Client) {
		client.commands = true
	}
}

// As sets the messaging rate mode.
func As(mode modes.MessageRateMode) Option {
	return func(client *Client) {
		client.mode = mode
	}
}
