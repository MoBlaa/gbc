// Package gbc is a connector library for chat providers mainly supporting development of bots.
//
// It's aimed to support the basic operations like sending and receiving messages, as well
// as adjusting message output to the platform to their messaging rates.
//
// Currently only Twitch IRC Chat is supported with messaging rates and automatically replying
// to `PING` messages.
//
// ###### Usage example with the Twitch Client
//
// ```go
// 	// Authentication is required
// 	auth := &twitchclient.TwitchAuthentication{
//		Name:  "mo_blaa",
//		Token: "oauth:8vt1t2fd2ye84sf9x0wny1o8qdaw7s",
//	}
// 	// Creating a client with some options. By default Membership, Tags and Commands are disabled.
//	client := twitchclient.New(auth, twitchclient.WithMembership(), twitchclient.WithTags(), twitchclient.WithCommands())
//
//	// Create an input channel, which will be used to emit messages to Twitch
//	in := make(chan *gbc.PlatformMessage)
//	defer close(in)
//
// 	// Connect and return a channel which emits messages received from Twitch.
//	out, err := client.Connect(in)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create a channel to wait for
//	waitc := make(chan struct{})
//
//	// Start a goroutine to listen for messages in parallel
//	go func() {
//		defer close(waitc)
//		// If connection to Twitch is lost, the output channel gets closed and the loop finishes
//		for mssg := range out {
//			log.Printf("> %s", mssg.RawMessage){
//		}
//	}()
//
//	// Maybe do some fancy graceful shutdown or other things
//
//	<-waitc // Wait for the client to quit
// ```
package gbc
