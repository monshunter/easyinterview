package store

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// cursorEncoder mirrors the handler-side cursor payload so list pages emit
// values consumable by handler.decodeCursor.
type cursorEncoder struct {
	UpdatedAt time.Time `json:"u"`
	ID        string    `json:"i"`
}

func (c cursorEncoder) encode() string {
	raw, _ := json.Marshal(struct {
		UpdatedAt string `json:"u"`
		ID        string `json:"i"`
	}{
		UpdatedAt: c.UpdatedAt.UTC().Format(time.RFC3339Nano),
		ID:        c.ID,
	})
	return base64.RawURLEncoding.EncodeToString(raw)
}
