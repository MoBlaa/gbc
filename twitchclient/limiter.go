package twitchclient

import (
	"github.com/MoBlaa/gbc"
	"github.com/MoBlaa/gbc/internal"
	"github.com/MoBlaa/gbc/twitchclient/modes"
	"sync"
	"time"
)

// limiter adjusts message output to message limits of Twitch.
// These are categorized as:
// - whispers to accounts per day
// - whisper per second
// - whisper per minute
// - chat message per 30 seconds
type limiter struct {
	Mode modes.MessageRateMode
}

// Apply the limiter as a pipeline step to the given channel.
func (lim *limiter) Apply(in <-chan *gbc.PlatformMessage) <-chan *gbc.PlatformMessage {
	whispers := make(chan *gbc.PlatformMessage)
	chats := make(chan *gbc.PlatformMessage)

	// Split the input into whisper and channel messages
	go func() {
		defer close(whispers)
		defer close(chats)
		for mssg := range in {
			if Message(*mssg).IsWhisper() {
				whispers <- mssg
			} else {
				chats <- mssg
			}

		}
	}()

	//// Start Limiters
	// Create channel limiting the chat output
	limit := &internal.Limiter{
		Duration: 30 * time.Second,
		Limit:    lim.Mode.ToChatPer30Seconds(),
	}
	chatOut := limit.Apply(chats)
	//// Chain Whisper-limits
	// Limit daily contacted accounts
	daily := dailyLimiter{Limit: lim.Mode.ToWhisperAccountsPerDay()}
	whisperAccOut := daily.Apply(whispers)
	// Limit Messages whispered per minute
	minLimiter := &internal.Limiter{
		Duration: time.Minute,
		Limit:    lim.Mode.ToWhisperPerMinute(),
	}
	whisperMinuteOut := minLimiter.Apply(whisperAccOut)
	// Limit Messages whispered per second
	secLimiter := &internal.Limiter{
		Duration: time.Second,
		Limit:    lim.Mode.ToWhisperPerSecond(),
	}
	whisperOut := secLimiter.Apply(whisperMinuteOut)

	// Merge output of chatOut and whisperOut
	return fanIn(chatOut, whisperOut)
}

func fanIn(in ...<-chan *gbc.PlatformMessage) <-chan *gbc.PlatformMessage {
	var wg sync.WaitGroup
	out := make(chan *gbc.PlatformMessage)

	// Reads output of one input channel
	output := func(ch <-chan *gbc.PlatformMessage) {
		defer wg.Done()
		for m := range ch {
			out <- m
		}
	}
	wg.Add(len(in))

	// Start Goroutine per input channel
	for _, ch := range in {
		go output(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
