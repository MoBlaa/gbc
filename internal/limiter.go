package internal

import (
	"github.com/MoBlaa/gbc"
	"time"
)

// Limiter adjusts message input to limits in a specific timespan.
// The problem of limiting is solved by periodically sending
// one event incoming to the output where the time of a period depends on the
// maximum number of messages (limit) in a given duration.
type Limiter struct {
	Duration time.Duration
	Limit    int
}

// Apply the Limiter as a pipeline step to a channel.
func (lim *Limiter) Apply(in <-chan *gbc.PlatformMessage) <-chan *gbc.PlatformMessage {
	// No buffers as the sender
	out := make(chan *gbc.PlatformMessage)

	go func() {
		defer close(out)
		for range time.NewTicker(lim.Duration / time.Duration(lim.Limit)).C {
			mssg, ok := <-in
			if !ok {
				break
			}
			out <- mssg
		}
	}()

	return out
}
