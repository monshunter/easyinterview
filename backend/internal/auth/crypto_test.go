package auth_test

import (
	"regexp"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func TestSixDigitCodeGeneratorReturnsNumericCode(t *testing.T) {
	code, err := (auth.SixDigitCodeGenerator{}).GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if !regexp.MustCompile(`^[0-9]{6}$`).MatchString(code) {
		t.Fatalf("code = %q, want six digits", code)
	}
}
