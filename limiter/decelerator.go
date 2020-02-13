package limiter

import (
	"github.com/MoBlaa/gbc"
	"time"
)

type Decelerator time.Duration

// Decelerator periodically reads all buffered events of an input-channel
// and returns them as array.
// This is a pre-step for bundling events.
func (dec *Decelerator) Apply(in <-chan *gbc.PlatformMessage) <-chan []*gbc.PlatformMessage {
	out := make(chan []*gbc.PlatformMessage, 10000)

	go func() {
		defer close(out)
		for next := range in {
			// Collect all buffered input events
			var input []*gbc.PlatformMessage
			input = append(input, next)
			// Read buffered values
			for i := 0; i < len(in); i++ {
				next, more := <-in
				if !more {
					break
				}
				input = append(input, next)
			}
			if input != nil {
				out <- input
			}
			<-time.NewTicker(time.Duration(*dec)).C
		}
	}()

	return out
}
