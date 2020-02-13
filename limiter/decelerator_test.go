package limiter

import (
	"github.com/MoBlaa/gbc"
	"testing"
	"time"
)

func TestDecelerator_Closing(t *testing.T) {
	input := make(chan *gbc.PlatformMessage)

	dec := Decelerator(time.Second * 5)
	out := dec.Apply(input)

	close(input)
	select {
	case _, more := <-out:
		if more {
			t.Log("received event before closed")
		}
	case <-time.NewTimer(time.Second).C:
		t.Fatal("Timed out before closing")
	}
}

func TestDecelerator_Buffer(t *testing.T) {
	input := make(chan *gbc.PlatformMessage, 10)
	done := make(chan struct{})
	defer close(input)
	defer close(done)

	events := []*gbc.PlatformMessage{
		{
			Platform:   gbc.Twitch,
			RawMessage: ":sender!sender@sender.tmi.twitch.tv PRIVMSG #target :Message1",
		}, {
			Platform:   gbc.Twitch,
			RawMessage: ":sender!sender@sender.tmi.twitch.tv PRIVMSG #target :Message2",
		}, {
			Platform:   gbc.Twitch,
			RawMessage: ":sender!sender@sender.tmi.twitch.tv PRIVMSG #target :Message3",
		},
	}

	dec := Decelerator(500 * time.Millisecond)
	output := dec.Apply(input)

	for _, event := range events {
		input <- event
	}

	oEvents := <-output
	if len(oEvents) != 2 {
		t.Errorf("No bundled output. Length is %d", len(oEvents))
		t.Fail()
	}
}

func TestDecelerator_Time(t *testing.T) {
	input := make(chan *gbc.PlatformMessage, 10)
	defer close(input)

	events := []*gbc.PlatformMessage{
		{
			Platform:   gbc.Twitch,
			RawMessage: ":sender!sender@sender.tmi.twitch.tv PRIVMSG #target :Message1",
		}, {
			Platform:   gbc.Twitch,
			RawMessage: ":sender!sender@sender.tmi.twitch.tv PRIVMSG #target :Message2",
		}, {
			Platform:   gbc.Twitch,
			RawMessage: ":sender!sender@sender.tmi.twitch.tv PRIVMSG #target :Message3",
		},
	}

	dec := Decelerator(50 * time.Millisecond)
	output := dec.Apply(input)

	start := time.Now()
	for _, event := range events {
		input <- event
	}
	<-output
	<-output
	elapsed := time.Since(start)

	if elapsed.Round(time.Millisecond) < 45*time.Millisecond {
		t.Errorf("Took less time then expected. Expected: %v, Actual: %v", 50*time.Millisecond, elapsed)
		t.Fail()
	}
	if elapsed.Round(time.Millisecond) > 55*time.Millisecond {
		t.Errorf("Took less time then expected. Expected: %v, Actual: %v", 50*time.Millisecond, elapsed)
		t.Fail()
	}
}
