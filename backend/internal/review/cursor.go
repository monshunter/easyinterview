package review

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidCursor = errors.New("review: invalid cursor")

type cursorPayload struct {
	CreatedAt string `json:"createdAt"`
	ID        string `json:"id"`
}

func EncodeCursor(createdAt time.Time, id string) string {
	payload := cursorPayload{
		CreatedAt: createdAt.UTC().Format(time.RFC3339Nano),
		ID:        strings.TrimSpace(id),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func DecodeCursor(cursor string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(cursor))
	if err != nil || len(raw) == 0 {
		return time.Time{}, "", ErrInvalidCursor
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var payload cursorPayload
	if err := dec.Decode(&payload); err != nil {
		return time.Time{}, "", ErrInvalidCursor
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return time.Time{}, "", ErrInvalidCursor
	}
	createdAt, err := time.Parse(time.RFC3339Nano, payload.CreatedAt)
	if err != nil {
		return time.Time{}, "", ErrInvalidCursor
	}
	if payload.CreatedAt != createdAt.UTC().Format(time.RFC3339Nano) {
		return time.Time{}, "", ErrInvalidCursor
	}
	id := strings.TrimSpace(payload.ID)
	if _, err := uuid.Parse(id); err != nil {
		return time.Time{}, "", fmt.Errorf("%w: invalid id", ErrInvalidCursor)
	}
	return createdAt.UTC(), id, nil
}
