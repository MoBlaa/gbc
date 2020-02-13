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

// PlatformMessage contains a message and the Platform of its origin.
type PlatformMessage struct {
	Platform   Platform
	RawMessage string
}
