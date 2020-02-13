package limiter

import (
	"github.com/MoBlaa/gbc"
	"log"
)

// DailyLimiter limits the amount of accounts the client can emit messages to.
type DailyLimiter struct {
	Limit int
	Clock Clock
}

// FIXME: This should limit the Accounts, not the messages sent per platform
func (lim *DailyLimiter) Apply(in <-chan *gbc.PlatformMessage) <-chan *gbc.PlatformMessage {
	out := make(chan *gbc.PlatformMessage, lim.Limit)

	var clock Clock
	if lim.Clock == nil {
		clock = NewClock()
	} else {
		clock = lim.Clock
	}
	go func() {
		defer close(out)
		contactedAccounts := make(map[gbc.Platform]struct{})
		for mssg := range in {
			if clock.DaySwitched() {
				// Reset records of sent targets if day changes
				contactedAccounts = make(map[gbc.Platform]struct{})
			}

			if _, contained := contactedAccounts[mssg.Platform]; !contained && len(contactedAccounts) >= lim.Limit {
				// Output, that limit was reached and discard message
				log.Printf("Reached limit of unique users to send whispers to. Discarding message sent to: %v\n", mssg.Platform)
			} else {
				// Add target and send message to output
				contactedAccounts[mssg.Platform] = struct{}{}
				out <- mssg
			}
		}
	}()

	return out
}
