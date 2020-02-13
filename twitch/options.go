package twitch

import (
	"net/url"
)

type Option func(client *TwitchClient)

func Server(url *url.URL) Option {
	return func(client *TwitchClient) {
		client.server = url
	}
}

func WithAuth(auth *OAuthAuthentication) Option {
	return func(client *TwitchClient) {
		client.auth = auth
	}
}

func SetChannels(channel string, additionals ...string) Option {
	return func(client *TwitchClient) {
		client.channels = append([]string{channel}, additionals...)
	}
}

func WithChannels(channels ...string) Option {
	return func(client *TwitchClient) {
		if len(channels) != 0 {
			client.channels = channels
		}
	}
}

func WithMembership() Option {
	return func(client *TwitchClient) {
		client.membership = true
	}
}

func WithTags() Option {
	return func(client *TwitchClient) {
		client.tags = true
	}
}

func WithCommands() Option {
	return func(client *TwitchClient) {
		client.commands = true
	}
}
