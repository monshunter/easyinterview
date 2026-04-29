package config

import (
	"time"
)

// GetString returns the value at dot-path key as a string. Empty string is
// returned for missing keys; required-key enforcement is the responsibility
// of Loader.Validate (item 1.5).
func (l *Loader) GetString(key string) string {
	if l == nil || l.k == nil {
		return ""
	}
	return l.k.String(key)
}

// GetInt returns the value at dot-path key as an int. Returns 0 when the
// key is missing or not numeric.
func (l *Loader) GetInt(key string) int {
	if l == nil || l.k == nil {
		return 0
	}
	return l.k.Int(key)
}

// GetBool returns the value at dot-path key as a bool. Returns false when
// the key is missing or not boolean.
func (l *Loader) GetBool(key string) bool {
	if l == nil || l.k == nil {
		return false
	}
	return l.k.Bool(key)
}

// GetDuration returns the value at dot-path key as a time.Duration. The
// value may be a Go duration string (e.g. "5s") or a number of seconds.
func (l *Loader) GetDuration(key string) time.Duration {
	if l == nil || l.k == nil {
		return 0
	}
	return l.k.Duration(key)
}

// GetSecret returns the value at dot-path key wrapped in RedactedString.
// Plaintext is only available via RedactedString.Reveal() so secrets cannot
// leak through fmt / JSON marshal / error wrapping.
func (l *Loader) GetSecret(key string) RedactedString {
	if l == nil {
		return RedactedString{}
	}
	if v, ok := l.secrets[key]; ok {
		return RedactedString{v: v}
	}
	if l.k != nil {
		if v := l.k.String(key); v != "" {
			return RedactedString{v: v}
		}
	}
	return RedactedString{}
}
