package idx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// IdempotencyKeyVersion is the wire-format version prefix for keys produced by
// GenerateIdempotencyKey. Bumping it indicates a backward-incompatible change
// to either the format or the semantics of expiry; the parser only accepts
// keys that match the current version.
const IdempotencyKeyVersion = "v1"

// ErrIdempotencyKeyMalformed is returned when the input is not a syntactically
// valid Idempotency-Key string for the current version.
var ErrIdempotencyKeyMalformed = errors.New("idx: idempotency key malformed")

// IdempotencyKey is the parsed representation of a wire-format key. The
// embedded UUID is the request-scoped UUIDv7 originally minted on the client
// or server; IssuedAt is the explicit timestamp prefix, used for TTL checks
// without parsing the UUIDv7 timestamp bytes.
type IdempotencyKey struct {
	Version  string
	IssuedAt time.Time
	UUID     string
}

// GenerateIdempotencyKey returns a freshly minted key valid until
// IssuedAt + IdempotencyKeyTTLSeconds. The wire format is
// `v1.{unixSeconds}.{uuidv7}` so the TS side can parse it without any
// UUIDv7 byte unpacking.
func GenerateIdempotencyKey() string {
	return formatIdempotencyKey(time.Now(), NewID())
}

func formatIdempotencyKey(issuedAt time.Time, uuid string) string {
	return fmt.Sprintf("%s.%d.%s", IdempotencyKeyVersion, issuedAt.Unix(), uuid)
}

// ParseIdempotencyKey validates and decodes a key without checking expiry.
// It rejects empty input, wrong version, non-numeric timestamps, non-UUIDv7
// bodies, and any value carrying the browser-only `tmp_` prefix.
func ParseIdempotencyKey(key string) (IdempotencyKey, error) {
	if key == "" {
		return IdempotencyKey{}, fmt.Errorf("%w: empty", ErrIdempotencyKeyMalformed)
	}
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return IdempotencyKey{}, fmt.Errorf("%w: expected 3 dot-separated parts, got %d", ErrIdempotencyKeyMalformed, len(parts))
	}
	if parts[0] != IdempotencyKeyVersion {
		return IdempotencyKey{}, fmt.Errorf("%w: version %q (only %q is accepted)", ErrIdempotencyKeyMalformed, parts[0], IdempotencyKeyVersion)
	}
	unixSec, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return IdempotencyKey{}, fmt.Errorf("%w: timestamp not numeric: %s", ErrIdempotencyKeyMalformed, parts[1])
	}
	if err := RequireServerID(parts[2]); err != nil {
		return IdempotencyKey{}, fmt.Errorf("%w: %v", ErrIdempotencyKeyMalformed, err)
	}
	return IdempotencyKey{
		Version:  parts[0],
		IssuedAt: time.Unix(unixSec, 0),
		UUID:     parts[2],
	}, nil
}

// IsIdempotencyKeyExpired returns true when the key was issued more than
// IdempotencyKeyTTLSeconds ago relative to `now`. Pass time.Now() in
// production; tests inject a fixed clock.
func IsIdempotencyKeyExpired(key string, now time.Time) (bool, error) {
	parsed, err := ParseIdempotencyKey(key)
	if err != nil {
		return false, err
	}
	age := now.Sub(parsed.IssuedAt)
	return age > time.Duration(IdempotencyKeyTTLSeconds)*time.Second, nil
}
