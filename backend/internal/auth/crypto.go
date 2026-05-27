package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
)

type TokenGenerator interface {
	GenerateToken() (string, error)
}

type SecureTokenGenerator struct{}

func (SecureTokenGenerator) GenerateToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

type SixDigitCodeGenerator struct{}

func (SixDigitCodeGenerator) GenerateToken() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate email code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashWithPepper(pepper, value string) string {
	sum := sha256.Sum256([]byte(pepper + "\x00" + value))
	return hex.EncodeToString(sum[:])
}
