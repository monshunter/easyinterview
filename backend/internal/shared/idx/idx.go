package idx

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// ErrTmpPrefix is returned when an id with the browser-only `tmp_` prefix
// reaches a server-side ingress that expects a persisted UUIDv7.
var ErrTmpPrefix = errors.New("idx: tmp_ prefixed identifier is not a server id")

// ErrInvalidUUIDv7 is returned when the input is not a syntactically valid
// UUIDv7 string. Empty strings count as invalid.
var ErrInvalidUUIDv7 = errors.New("idx: not a valid UUIDv7")

var (
	uuidRegexOnce sync.Once
	uuidRegex     *regexp.Regexp
)

func uuidv7Pattern() *regexp.Regexp {
	uuidRegexOnce.Do(func() {
		uuidRegex = regexp.MustCompile(UUIDv7RegexExpr)
	})
	return uuidRegex
}

// NewID returns a freshly minted UUIDv7 string. UUIDv7 is time-ordered, so
// generated ids sort lexically by creation moment, which is the property that
// makes them safe as primary keys for cursor pagination.
func NewID() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Errorf("idx: uuid.NewV7 failed: %w", err))
	}
	return id.String()
}

// RequireServerID is the ingress guard that rejects browser-only `tmp_` ids and
// any string that is not a syntactically valid UUIDv7. Persistence layers and
// any handler accepting client-supplied identifiers should call this before
// trusting the value.
func RequireServerID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: empty string", ErrInvalidUUIDv7)
	}
	if strings.HasPrefix(id, TmpIDPrefix) {
		return fmt.Errorf("%w: %q", ErrTmpPrefix, id)
	}
	if !uuidv7Pattern().MatchString(id) {
		return fmt.Errorf("%w: %q", ErrInvalidUUIDv7, id)
	}
	return nil
}
