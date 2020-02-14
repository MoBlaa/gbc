package twitchclient

import (
	"github.com/MoBlaa/gbc"
	"sync"
	"testing"
	"time"
)

// Used to simulate day switch
type testClock struct {
	lock        *sync.Mutex
	daySwitched bool
}

func (cl testClock) DaySwitched() bool {
	cl.lock.Lock()
	switched := cl.daySwitched
	cl.lock.Unlock()
	return switched
}

func TestDailyLimiter_Close(t *testing.T) {
	input := make(chan *gbc.PlatformMessage, 45)

	lim := dailyLimiter{Limit: 2}
	out := lim.Apply(input)

	close(input)
	select {
	case _, more := <-out:
		if more {
			t.Fatal("Not properly closing output")
		}
	case <-time.NewTimer(time.Second).C:
		t.Fatal("Timed out before close")
	}
}

// Test daily limiter limits to given amount
func TestDailyLimiter_limits(t *testing.T) {
	mssgs := []*gbc.PlatformMessage{
		{
			Platform:   gbc.Twitch,
			RawMessage: "WHISPER one :D:",
		},
		{
			Platform:   gbc.Twitch,
			RawMessage: "WHISPER two :D:",
		},
		{
			Platform:   gbc.Twitch,
			RawMessage: "WHISPER three :D:",
		},
	}

	in := make(chan *gbc.PlatformMessage)
	defer close(in)

	daily := dailyLimiter{Limit: 2, Clock: testClock{
		lock:        &sync.Mutex{},
		daySwitched: false,
	}}
	out := daily.Apply(in)

	in <- mssgs[0]
	in <- mssgs[1]
	in <- mssgs[2]

	if len(out) != 2 {
		t.Errorf("Should return only <limit> (2) number of messages")
		t.Errorf("Returned: %d", len(out))
		t.Fail()
		return
	}
}

// Test daily limiter limits to given amount
func TestDailyLimiter_resetOnDayChange(t *testing.T) {
	mssgs := []*gbc.PlatformMessage{
		{
			Platform:   gbc.Twitch,
			RawMessage: "WHISPER three :D:",
		},
		{
			Platform:   gbc.Twitch,
			RawMessage: "WHISPER three :D:",
		},
		{
			Platform:   gbc.Twitch,
			RawMessage: "WHISPER three :D:",
		},
	}

	in := make(chan *gbc.PlatformMessage)
	done := make(chan struct{})
	defer close(in)
	defer close(done)

	clock := &testClock{
		lock:        &sync.Mutex{},
		daySwitched: false,
	}
	daily := dailyLimiter{Limit: 2, Clock: clock}
	out := daily.Apply(in)

	in <- mssgs[0]
	in <- mssgs[1]

	var out1 *gbc.PlatformMessage
	var out2 *gbc.PlatformMessage
	select {
	case out1 = <-out:
	case <-time.NewTimer(20 * time.Millisecond).C:
		t.Errorf("Timed out waiting for output")
		t.Fail()
		return
	}
	select {
	case out2 = <-out:
	case <-time.NewTimer(20 * time.Millisecond).C:
		t.Errorf("Timed out waiting for output")
		t.Fail()
		return
	}

	if *out1 != *mssgs[0] || *out2 != *mssgs[1] {
		t.Errorf("Should keep order of returned messages")
		t.Errorf("Expected: %v", []*gbc.PlatformMessage{mssgs[0], mssgs[1]})
		t.Errorf("Actual:   %v", []*gbc.PlatformMessage{out1, out2})
		t.Fail()
		return
	}

	clock.lock.Lock()
	clock.daySwitched = true
	clock.lock.Unlock()

	// Write again and expect to return
	in <- mssgs[2]

	select {
	case <-out:
	case <-time.NewTimer(20 * time.Millisecond).C:
		t.Errorf("Should reset limits on day switch!")
		t.Errorf("Output-len: %d", len(out))
		t.Fail()
	}
}
