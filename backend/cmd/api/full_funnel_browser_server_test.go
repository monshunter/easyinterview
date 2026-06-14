package main

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

const (
	fullFunnelBrowserServerEnv         = "EI_E2E_P0_099_SERVER"
	fullFunnelBrowserBackendPortEnv    = "EI_E2E_BACKEND_PORT"
	fullFunnelBrowserFrontendOriginEnv = "EI_E2E_FRONTEND_ORIGIN"
	fullFunnelBrowserStatePathEnv      = "EI_E2E_STATE_PATH"
)

type fullFunnelBrowserState struct {
	APIBaseURL         string `json:"apiBaseUrl"`
	FrontendOrigin     string `json:"frontendOrigin"`
	UserID             string `json:"userId"`
	UserEmail          string `json:"userEmail"`
	ResumeID           string `json:"resumeId"`
	SessionCookieName  string `json:"sessionCookieName"`
	SessionCookieValue string `json:"sessionCookieValue"`
}

func TestE2EP0099ScenarioBackendServer(t *testing.T) {
	if os.Getenv(fullFunnelBrowserServerEnv) != "1" {
		t.Skip(fullFunnelBrowserServerEnv + " is not set; skipping browser scenario backend server")
	}

	h := newFullFunnelJourneyHarnessWithTimeout(t, 30*time.Minute)
	seed := h.seedReadyResume(t)

	port := fullFunnelEnvDefault(fullFunnelBrowserBackendPortEnv, "18099")
	frontendOrigin := fullFunnelEnvDefault(fullFunnelBrowserFrontendOriginEnv, "http://127.0.0.1:4174")
	statePath := fullFunnelEnvDefault(
		fullFunnelBrowserStatePathEnv,
		filepath.Join(scenarioRepoRoot(t), ".test-output", "e2e", "p0-099-full-funnel-fullstack-ui-journey", "state.json"),
	)

	listener, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		t.Fatalf("listen P0.099 backend server: %v", err)
	}

	serverCtx, stopSignals := signal.NotifyContext(h.ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	mux := http.NewServeMux()
	mux.HandleFunc("/__e2e/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"resumeId": seed.ResumeID,
		})
	})
	mux.Handle("/", h.handler)

	srv := &http.Server{
		Handler: fullFunnelBrowserCORS(frontendOrigin, mux),
	}
	serverErr := make(chan error, 1)
	go func() {
		err := srv.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	h.kernel.Start(serverCtx)

	state := fullFunnelBrowserState{
		APIBaseURL:         "http://" + listener.Addr().String() + "/api/v1",
		FrontendOrigin:     frontendOrigin,
		UserID:             h.userID,
		UserEmail:          fullFunnelJourneyEmail,
		ResumeID:           seed.ResumeID,
		SessionCookieName:  h.cookie.Name,
		SessionCookieValue: h.cookie.Value,
	}
	if err := writeFullFunnelBrowserState(statePath, state); err != nil {
		t.Fatalf("write P0.099 state: %v", err)
	}
	t.Logf("P0.099 backend server listening at %s", state.APIBaseURL)

	select {
	case <-serverCtx.Done():
	case err := <-serverErr:
		if err != nil {
			t.Fatalf("serve P0.099 backend server: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown P0.099 backend server: %v", err)
	}
	if err := h.kernel.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown P0.099 backend kernel: %v", err)
	}
}

func fullFunnelBrowserCORS(frontendOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == frontendOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept,Accept-Language,Content-Type,Idempotency-Key,Traceparent,X-Client-Version,X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeFullFunnelBrowserState(path string, state fullFunnelBrowserState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o600)
}

func fullFunnelEnvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
