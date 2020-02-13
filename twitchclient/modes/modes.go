package modes

// MessageRateMode represents the modes the Twitch-Account can have which sets the amount of messages
// the bot can sent per Minute, Second and number of other accounts.
type MessageRateMode string

const (
	// USER mode represents the standard mode of a Twitch-account with smallest limits.
	USER MessageRateMode = "USER"
	// KNOWN mode represents the state of a known Twitch-Bot-Account with extended limits.
	KNOWN MessageRateMode = "KNOWN"
	// VERIFIED mode represents the highest Tier of Account-setting with maximum amount of messages allowed.
	VERIFIED MessageRateMode = "VERIFIED"
)

// ToChatPer30Seconds returns the number of messages the mode allows to be sent into a chat room per 30 seconds.
func (mode MessageRateMode) ToChatPer30Seconds() int {
	switch mode {
	case USER:
		return 20
	case KNOWN:
		return 50
	case VERIFIED:
		return 7500
	default:
		return -1
	}
}

// ToWhisperPerSecond returns the number messages the mode allows to be whispered per second.
func (mode MessageRateMode) ToWhisperPerSecond() int {
	switch mode {
	case USER:
		return 3
	case KNOWN:
		return 10
	case VERIFIED:
		return 20
	default:
		return -1
	}
}

// ToWhisperPerMinute returns the number messages the mode allows to be whispered per minute.
func (mode MessageRateMode) ToWhisperPerMinute() int {
	switch mode {
	case USER:
		return 100
	case KNOWN:
		return 200
	case VERIFIED:
		return 1200
	default:
		return -1
	}
}

// ToWhisperAccountsPerDay returns the number of accounts the mode allows sending messages to.
func (mode MessageRateMode) ToWhisperAccountsPerDay() int {
	switch mode {
	case USER:
		return 40
	case KNOWN:
		return 500
	case VERIFIED:
		return 100000
	default:
		return -1
	}
}
