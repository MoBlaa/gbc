package gbc

// The Client interface contains actions which can be performed on a Connection.
type Client interface {
	// Listen to the messages of this Client.
	Connect(in <-chan *PlatformMessage) (<-chan *PlatformMessage, error)
	// Disconnect the Client from his Platform.
	Disconnect()
}

type Platform string

const Twitch Platform = "TWITCH"

// The Location to/from which so send/receive a message to/from.
type Location struct {
	// Platform this Location is referring to.
	Platform Platform
}

// PlatformMessage contains a message and the Location of its origin.
type PlatformMessage struct {
	From       Location
	RawMessage string
}
