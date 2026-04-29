package config

import "fmt"

// redactedMarker is the canonical placeholder rendered for any secret value.
const redactedMarker = "***"

// RedactedString wraps a configuration secret so that printing or marshaling
// the value never leaks plaintext. The struct field is intentionally
// unexported so reflection / direct access from other packages cannot
// bypass the redaction; the only legitimate plaintext escape hatch is
// Reveal(), which callers must hand directly to the SDK that consumes the
// secret without re-routing through business code (spec §4.2).
type RedactedString struct {
	v string
}

// NewRedactedString constructs a RedactedString from a plaintext value. It
// is the only public constructor; the loader uses this internally too.
func NewRedactedString(v string) RedactedString {
	return RedactedString{v: v}
}

// Reveal returns the underlying plaintext. Callers must keep the result
// out of logs, errors and JSON; redaction only applies to RedactedString
// itself, not to derived strings.
func (r RedactedString) Reveal() string {
	return r.v
}

// IsZero reports whether the underlying value is empty. Useful for
// validators to distinguish "absent" from "redacted".
func (r RedactedString) IsZero() bool {
	return r.v == ""
}

// String implements fmt.Stringer so %s / %v / Println always print ***.
func (r RedactedString) String() string {
	return redactedMarker
}

// GoString implements fmt.GoStringer so %#v never reveals plaintext.
func (r RedactedString) GoString() string {
	return redactedMarker
}

// MarshalJSON implements json.Marshaler so encoding/json output never
// reveals plaintext.
func (r RedactedString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + redactedMarker + `"`), nil
}

// MarshalText implements encoding.TextMarshaler so encoders that prefer
// text marshaling (yaml.v3, toml libs, log/slog text handler) also see ***.
func (r RedactedString) MarshalText() ([]byte, error) {
	return []byte(redactedMarker), nil
}

// Format implements fmt.Formatter so verbs not handled above (e.g. %x, %q,
// %+v) also redact. fmt's default fallback would otherwise expose the
// underlying string for some verbs; this overrides that behavior.
func (r RedactedString) Format(s fmt.State, verb rune) {
	switch verb {
	case 'q':
		fmt.Fprintf(s, "%q", redactedMarker)
	default:
		fmt.Fprint(s, redactedMarker)
	}
}
