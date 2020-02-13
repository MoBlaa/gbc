package limiter

import (
	"fmt"
	"github.com/MoBlaa/gbc"
	"testing"
	"time"
)

func TestLimiter_Close(t *testing.T) {
	input := make(chan *gbc.PlatformMessage, 45)

	lim := Limiter{
		Duration: 500 * time.Millisecond,
		Limit:    2,
	}
	out := lim.Apply(input)

	close(input)
	select {
	case _, more := <-out:
		if more {
			t.Fatal("Not properly closing output")
		}
	case <-time.NewTimer(10 * time.Second).C:
		t.Fatal("Timed out before close")
	}
}

func TestLimiter_limits(t *testing.T) {
	in := make(chan *gbc.PlatformMessage, 5)
	defer close(in)

	lim := Limiter{
		Duration: 500 * time.Millisecond,
		Limit:    2,
	}
	out := lim.Apply(in)

	for i := 1; i <= 5; i++ {
		start := time.Now()
		in <- &gbc.PlatformMessage{
			Platform:   gbc.Twitch,
			RawMessage: fmt.Sprintf("PRIVMSG #blaaabot%d", i),
		}
		in <- &gbc.PlatformMessage{
			Platform:   gbc.Twitch,
			RawMessage: fmt.Sprintf("PRIVMSG #blaaabot%d%d", i, i),
		}

		select {
		case <-out:
			select {
			case <-out:
			case <-time.NewTimer(500 * time.Millisecond).C:
				t.Errorf("Timedout waiting for second output")
				t.Fail()
				return
			}
			took := time.Since(start).Round(time.Millisecond)
			if took >= 505*time.Millisecond || took <= 495*time.Millisecond {
				t.Errorf("Timeout between input and output is not as expected:")
				t.Errorf("Took: %v", took)
				t.Errorf("Actual: %v", 500*time.Millisecond)
				t.Fail()
				return
			}
		case <-time.NewTimer(600 * time.Millisecond).C:
			t.Errorf("Missed message nr. %d", i)
			t.Fail()
		}
	}
}
