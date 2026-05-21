package main

import (
	"database/sql"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	profilehandler "github.com/monshunter/easyinterview/backend/internal/profile/handler"
	profileservice "github.com/monshunter/easyinterview/backend/internal/profile/service"
	profilestore "github.com/monshunter/easyinterview/backend/internal/profile/store"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

// profileRoutes carries the Profile tag handler + idempotency middleware that
// cmd/api wires onto /api/v1/profiles/* routes.
type profileRoutes struct {
	Handler     *profilehandler.Handler
	Service     *profileservice.Service
	Store       profile.Store
	Idempotency *idempotency.Middleware
}

// buildProfileRoutes composes the backend-profile runtime against the shared
// *sql.DB and idempotency configuration used elsewhere in cmd/api.
func buildProfileRoutes(loader *config.Loader, db *sql.DB) profileRoutes {
	repo := profilestore.NewRepository(db)
	settings := profilestore.NewSettingsReader(db)
	audit := profilestore.NewAuditTombstoneWriter(db)
	svc := profileservice.New(profileservice.Options{
		Store: repo,
		Audit: audit,
	})
	handler := profilehandler.New(profilehandler.Options{
		Store:    repo,
		Settings: settings,
		Session:  currentUserFromContext,
		NewID:    idx.NewID,
	})
	ttl := time.Duration(sharedtypes.IdempotencyKeyTTLSeconds) * time.Second
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return profileRoutes{
		Handler: handler,
		Service: svc,
		Store:   repo,
		Idempotency: idempotency.New(idempotency.MiddlewareOptions{
			Store:     idempotency.NewSQLStore(db),
			KeyPepper: loader.GetSecret("auth.challengeTokenPepper").Reveal(),
			TTL:       ttl,
		}),
	}
}
