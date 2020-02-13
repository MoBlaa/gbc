package twitchclient

import (
	"github.com/MoBlaa/gbc"
	"github.com/MoBlaa/gbc/limiter"
	"log"
)

// DailyLimiter limits the amount of accounts the client can emit messages to.
type DailyLimiter struct {
	Limit int
	Clock limiter.Clock
}

// Apply the DailyLimiter as a pipeline step to a channel.
func (lim *DailyLimiter) Apply(in <-chan *gbc.PlatformMessage) <-chan *gbc.PlatformMessage {
	out := make(chan *gbc.PlatformMessage, lim.Limit)

	var clock limiter.Clock
	if lim.Clock == nil {
		clock = limiter.NewClock()
	} else {
		clock = lim.Clock
	}
	go func() {
		defer close(out)
		contactedAccounts := make(map[string]struct{})
		for platformMessage := range in {
			mssg := Message(*platformMessage)
			if clock.DaySwitched() {
				// Reset records of sent targets if day changes
				contactedAccounts = make(map[string]struct{})
			}

			if _, contained := contactedAccounts[mssg.Receipt()]; !contained && len(contactedAccounts) >= lim.Limit {
				// Output, that limit was reached and discard message
				log.Printf("Reached limit of unique users to send whispers to. Discarding message sent to: %v\n", mssg.Receipt())
			} else {
				// Add target and send message to output
				contactedAccounts[mssg.Receipt()] = struct{}{}
				out <- platformMessage
			}
		}
	}()

	return out
}