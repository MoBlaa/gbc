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

type OAuthAuthentication struct {
	Name  string
	Token string
}

type Message gbc.PlatformMessage

var mssgRegex = regexp.MustCompile("([@\\w=:;]+ )?([:\\w!@.]+ )?WHISPER(.*)?")

func (mess Message) IsWhisper() bool {
	return mssgRegex.MatchString(mess.RawMessage)
}

type Client struct {
	server     *url.URL
	auth       *OAuthAuthentication
	channels   []string
	membership bool
	tags       bool
	commands   bool
	mode       modes.MessageRateMode

	conn *websocket.Conn
}

func New(auth *OAuthAuthentication, opts ...Option) *Client {
	client := &Client{
		server: &url.URL{
			Scheme: "wss",
			Host:   "irc-ws.chat.twitch.tv:443",
		},
		auth:     auth,
		channels: []string{auth.Name},
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
		fmt.Sprintf("PASS oauth:%s", client.auth.Token),
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
		mssgs = append(mssgs, fmt.Sprintf("#%s", channel))
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

				log.Printf("> %s", single)
				out <- &gbc.PlatformMessage{
					Platform:   gbc.Twitch,
					RawMessage: single,
				}
			}
		}
	}()

	// Start Sender to websocket connection
	go func() {
		// This will also close the websocket, which closes the listener also
		defer client.Disconnect()
		for message := range in {
			if message.Platform == gbc.Twitch {
				err := send(client.conn, message.RawMessage)
				if err != nil {
					log.Printf("error sending message: %v", err)
					break
				}
			}
		}
	}()

	lim := Limiter{Mode: client.mode}
	return lim.Apply(out), nil
}

func (client *Client) Disconnect() {
	err := client.conn.Close()
	if err != nil {
		log.Printf("failed to close websocket connection: %v", err)
	}
}
