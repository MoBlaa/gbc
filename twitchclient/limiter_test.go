package twitchclient

import (
	"fmt"
	"github.com/MoBlaa/gbc"
	"github.com/MoBlaa/gbc/twitchclient/modes"
	"testing"
	"time"
)

func TestTwitchLimiter_Close(t *testing.T) {
	input := make(chan *gbc.PlatformMessage, 45)

	lim := Limiter{Mode: modes.USER}
	out := lim.Apply(input)

	close(input)
	select {
	case _, more := <-out:
		if more {
			t.Fatal("Not properly closing output")
		}
	case <-time.NewTimer(2 * time.Second).C:
		t.Fatal("Timed out before close")
	}
}

func TestFanin_Close(t *testing.T) {
	inputs := []chan *gbc.PlatformMessage{
		make(chan *gbc.PlatformMessage, 45),
		make(chan *gbc.PlatformMessage, 45),
	}
	out := fanIn(inputs[0], inputs[1])

	close(inputs[0])
	close(inputs[1])
	select {
	case _, more := <-out:
		if more {
			t.Fatal("Not properly closing output")
		}
	case <-time.NewTimer(2 * time.Second).C:
		t.Fatal("Timed out before close")
	}
}

// Tests for whisper account limit being reached
func TestTwitchLimiter_Accounts(t *testing.T) {
	// generate 41 as 40 is the maximum number of accounts to send messages to for USER-Mode
	var mssgs []*gbc.PlatformMessage
	for i := 0; i < 41; i++ {
		mssgs = append(mssgs, &gbc.PlatformMessage{
			Platform:   gbc.Twitch,
			RawMessage: fmt.Sprintf("WHISPER testuser%d :D:", i+1),
		})
	}

	in := make(chan *gbc.PlatformMessage, 45)
	defer close(in)

	lim := Limiter{Mode: modes.USER}
	out := lim.Apply(in)

	for i, mssg := range mssgs {
		in <- mssg

		// Wait for response and check if 41th message times out
		select {
		case <-out:
		case <-time.NewTimer(2 * time.Second).C:
			if i != 40 {
				// Last one timed out -> success
				t.Fatalf("Response to a message timed out: %d", i)
			}
			return
		}
	}
}

func TestTwitchLimiter_WhisperPerSecond(t *testing.T) {
	expTimeout := time.Second / time.Duration(3)
	mssgs := []*gbc.PlatformMessage{
		{
			Platform:   gbc.Twitch,
			RawMessage: "WHISPER testTarget :D:",
		},
		{
			Platform:   gbc.Twitch,
			RawMessage: "WHISPER testTarget :D:",
		}, {
			Platform:   gbc.Twitch,
			RawMessage: "WHISPER testTarget :D:",
		},
	}

	in := make(chan *gbc.PlatformMessage, 5)
	defer close(in)

	lim := Limiter{Mode: modes.USER}
	out := lim.Apply(in)

	for i, mssg := range mssgs {
		start := time.Now()
		in <- mssg

		select {
		case <-out:
			took := time.Since(start).Round(time.Millisecond)
			if took < expTimeout {
				t.Fatalf("timeout between messages is to big :: expected: %v, actual: %v", expTimeout, took)
				return
			}
		case <-time.NewTimer(2 * time.Second).C:
			t.Fatalf("Response timed out for message nr: %d", i+1)
			return
		}
	}
}

func TestTwitchLimiter_WhisperPerMinute(t *testing.T) {
	var mssgs []*gbc.PlatformMessage
	for i := 0; i < 101; i++ {
		mssgs = append(mssgs, &gbc.PlatformMessage{
			Platform:   gbc.Twitch,
			RawMessage: "WHISPER test :D:",
		})
	}

	in := make(chan *gbc.PlatformMessage, 102)
	defer close(in)

	lim := Limiter{Mode: modes.USER}
	out := lim.Apply(in)

	for _, mssg := range mssgs {
		in <- mssg
	}

	start := time.Now()
	for i := 0; i < 101; i++ {
		<-out
	}
	took := time.Since(start).Round(time.Minute)

	t.Logf("Took %v", took)
	if took < time.Minute {
		t.Fatalf("Should take more then a minute for all messages :: actual: %v", took)
	}
}

func TestTwitchLimiter_ChatPer30Seconds(t *testing.T) {
	expTimeout := (30 * time.Second) / time.Duration(20)
	mssgs := []*gbc.PlatformMessage{
		{
			Platform:   gbc.Twitch,
			RawMessage: "PRIVMSG #test :D:",
		},
		{
			Platform:   gbc.Twitch,
			RawMessage: "PRIVMSG #test :D:",
		},
	}

	in := make(chan *gbc.PlatformMessage)
	defer close(in)

	lim := &Limiter{Mode: modes.USER}
	out := lim.Apply(in)

	for i, mssg := range mssgs {
		start := time.Now()
		in <- mssg

		select {
		case <-out:
			took := time.Since(start).Round(100 * time.Millisecond)
			if took < expTimeout {
				t.Log(i)
				t.Fatalf("Timeout between messages to a chat should have a minimum timeout of %v :: actual: %v", expTimeout, took)
				return
			}
		case <-time.NewTimer(2 * time.Second).C:
			t.Fatalf("Message nr. %d timed out", i)
			return
		}
	}
}
