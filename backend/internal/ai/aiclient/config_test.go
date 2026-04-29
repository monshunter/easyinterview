package aiclient_test

import (
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
)

// TestNew_FailFastMatrix exercises the spec D-4 / plan 4.1 fail-fast
// matrix. Plan checklist 4.2 codifies these three core cases plus one
// non-test happy path.
func TestNew_FailFastMatrix(t *testing.T) {
	cases := []struct {
		name    string
		cfg     aiclient.Config
		opts    []aiclient.Option
		wantErr error
	}{
		{
			name:    "test-env-stub-allowed-no-gateway-OK",
			cfg:     aiclient.Config{AppEnv: aiclient.AppEnvTest},
			opts:    []aiclient.Option{aiclient.WithStubAllowed(true)},
			wantErr: nil,
		},
		{
			name:    "test-env-no-stub-no-gateway-FAIL",
			cfg:     aiclient.Config{AppEnv: aiclient.AppEnvTest},
			opts:    nil,
			wantErr: aiclient.ErrMissingGatewayConfig,
		},
		{
			name: "test-env-with-gateway-OK",
			cfg: aiclient.Config{
				AppEnv:         aiclient.AppEnvTest,
				GatewayBaseURL: "http://example",
				GatewayAPIKey:  "k",
			},
			wantErr: nil,
		},
		{
			name:    "production-no-gateway-FAIL",
			cfg:     aiclient.Config{AppEnv: "production"},
			opts:    nil,
			wantErr: aiclient.ErrMissingGatewayConfig,
		},
		{
			name:    "production-stub-allowed-no-gateway-still-FAIL",
			cfg:     aiclient.Config{AppEnv: "production"},
			opts:    []aiclient.Option{aiclient.WithStubAllowed(true)},
			wantErr: aiclient.ErrMissingGatewayConfig,
		},
		{
			name: "production-with-gateway-OK",
			cfg: aiclient.Config{
				AppEnv:         "production",
				GatewayBaseURL: "http://gateway",
				GatewayAPIKey:  "k",
			},
			wantErr: nil,
		},
		{
			name:    "staging-missing-api-key-FAIL",
			cfg:     aiclient.Config{AppEnv: "staging", GatewayBaseURL: "http://gw"},
			wantErr: aiclient.ErrMissingGatewayConfig,
		},
		{
			name:    "staging-missing-base-url-FAIL",
			cfg:     aiclient.Config{AppEnv: "staging", GatewayAPIKey: "k"},
			wantErr: aiclient.ErrMissingGatewayConfig,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := aiclient.New(tc.cfg, tc.opts...)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}
