package twitchclient

import (
	"fmt"
	"github.com/MoBlaa/gbc"
	"github.com/MoBlaa/gbc/twitchclient/modes"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/url"
	"regexp"
	"strings"
)

// TwitchAuthentication contains authentication information for twitch.
type TwitchAuthentication struct {
	Name  string
	Token string
}

var mssgRegex = regexp.MustCompile("([@\\w=:;]+ )?([:\\w!@.]+ )?WHISPER(.*)?")

// Message sent from/to twitch.
type Message gbc.PlatformMessage

// IsWhisper returns if the message represents a whisper message.
func (mess Message) IsWhisper() bool {
	return mssgRegex.MatchString(mess.RawMessage)
}

// Receipt extracts the first parameter of a whisper or privmsg as that represents the
// user/channel the message is extracted to.
func (mess Message) Receipt() string {
	start := strings.Index(mess.RawMessage, "WHISPER")
	if start == -1 {
		start = strings.Index(mess.RawMessage, "PRIVMSG")
	}
	if start == -1 {
		return ""
	}
	start += 8
	end := strings.Index(mess.RawMessage[start:], " ")
	return mess.RawMessage[start : start+end]
}

// Client handling communication with twitch.
type Client struct {
	server     *url.URL
	auth       *TwitchAuthentication
	channels   []string
	membership bool
	tags       bool
	commands   bool
	mode       modes.MessageRateMode

	conn *websocket.Conn
}

// New creates a new TwitchClient with default parameters applying the given options.
func New(auth *TwitchAuthentication, opts ...Option) *Client {
	client := &Client{
		server: &url.URL{
			Scheme: "wss",
			Host:   "irc-ws.chat.twitch.tv:443",
		},
		auth:     auth,
		channels: []string{auth.Name},
		mode:     modes.USER,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func send(conn *websocket.Conn, mssg string) error {
	log.Printf("< %v", mssg)
	return conn.WriteMessage(websocket.TextMessage, []byte(mssg))
}

// Connect establishes an connection to twitch. Messages sent to the `in` channel are sent to twitch
// after messaging limits are applied. Returns a channel emitting messages received from twitch.
func (client *Client) Connect(in <-chan *gbc.PlatformMessage) (<-chan *gbc.PlatformMessage, error) {
	if client.conn != nil {
		return nil, fmt.Errorf("already listening")
	}

	out := make(chan *gbc.PlatformMessage)

	// Connect to Twitch Websocket-Server
	conn, _, err := websocket.DefaultDialer.Dial(client.server.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Twitch: %w", err)
	}
	client.conn = conn

	mssgs := []string{
		fmt.Sprintf("PASS %s", client.auth.Token),
		fmt.Sprintf("NICK %s", client.auth.Name),
	}

	// Only include capabilities if enabled
	if client.membership {
		mssgs = append(mssgs, "CAP REQ :twitch.tv/membership")
	}
	if client.tags {
		mssgs = append(mssgs, "CAP REQ :twitch.tv/tags")
	}
	if client.commands {
		mssgs = append(mssgs, "CAP REQ :twitch.tv/commands")
	}

	for _, channel := range client.channels {
		mssgs = append(mssgs, fmt.Sprintf("JOIN #%s", channel))
	}
	log.Printf("Logging in as '%s'", client.auth.Name)

	// Initialize Connection with login and requesting capabilities (Tags, Memberships, etc.)
	for _, initMssg := range mssgs {
		err = send(conn, initMssg)
		if err != nil {
			return nil, err
		}
	}

	// Start listener to websocket connection
	go func() {
		defer close(out)
		for {
			_, message, err := client.conn.ReadMessage()
			if err == io.EOF {
				return
			}
			if err != nil {
				// Removed as triggered every time the connection is closed
				//elog.Error(fmt.Errorf("error reading message from twitch: %w", err))
				log.Printf("error: %v; Closing listener for twitch messages!", err)
				return
			}
			strMssg := string(message)
			mssgs := strings.Split(strings.ReplaceAll(strMssg, "\r\n", "\n"), "\n")
			for _, single := range mssgs {
				if strings.ReplaceAll(single, " ", "") == "" {
					continue
				}

				out <- &gbc.PlatformMessage{
					Platform:   gbc.Twitch,
					RawMessage: single,
				}

				if strings.HasSuffix(single, "PING :tmi.twitch.tv") {
					err = send(client.conn, "PONG :tmi.twitch.tv")
					if err != nil {
						log.Fatalf("failed to send PONG message: %v\n", err)
					}
				}
			}
		}
	}()

	// Start Sender to websocket connection
	go func() {
		// This will also close the websocket, which closes the listener also
		defer client.Disconnect()
		// Limit the output to twitch
		lim := Limiter{Mode: client.mode}
		for message := range lim.Apply(in) {
			if message.Platform == gbc.Twitch {
				err := send(client.conn, message.RawMessage)
				if err != nil {
					log.Printf("error sending message: %v", err)
					break
				}
			}
		}
	}()

	return out, nil
}

// Disconnect closes the connection to twitch.
func (client *Client) Disconnect() {
	err := client.conn.Close()
	if err != nil {
		log.Printf("failed to close websocket connection: %v", err)
	}
	client.conn = nil
}
