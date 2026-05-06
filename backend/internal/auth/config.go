package auth

import "time"

const (
	SessionCookieName = "ei_session"

	ChallengeTTL       = 15 * time.Minute
	SessionTTL         = 30 * 24 * time.Hour
	RateLimitWindow    = time.Minute
	RateLimitThreshold = 3

	DevMailSinkName = "dev-mail-sink"
)
