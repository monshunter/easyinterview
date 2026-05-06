package auth_test

import (
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func TestGeneratedAuthSurfaceIsCoveredByC1Policy(t *testing.T) {
	for _, operationID := range []string{
		"startAuthEmailChallenge",
		"verifyAuthEmailChallenge",
		"getMe",
		"deleteMe",
		"logout",
		"getRuntimeConfig",
	} {
		if !generatedHasOperation(operationID) {
			t.Fatalf("generated ServerInterface no longer exposes %s", operationID)
		}
		if _, ok := auth.SessionPolicyForOperation(operationID); !ok {
			t.Fatalf("C1/A4 session policy does not cover generated operation %s", operationID)
		}
	}
}

func TestSessionPolicyClassifiesPublicOptionalAndProtectedOperations(t *testing.T) {
	want := map[string]auth.SessionRequirement{
		"startAuthEmailChallenge":  auth.SessionPublic,
		"verifyAuthEmailChallenge": auth.SessionPublic,
		"getRuntimeConfig":         auth.SessionPublic,
		"logout":                   auth.SessionOptional,
		"getMe":                    auth.SessionRequired,
		"deleteMe":                 auth.SessionRequired,
	}
	for operationID, requirement := range want {
		got, ok := auth.SessionPolicyForOperation(operationID)
		if !ok {
			t.Fatalf("missing policy for %s", operationID)
		}
		if got != requirement {
			t.Fatalf("%s policy = %s, want %s", operationID, got, requirement)
		}
	}

	for _, route := range generated.AllRoutes {
		if _, ok := auth.SessionPolicyForOperation(route.OperationID); !ok {
			t.Fatalf("generated operation %s is not classified", route.OperationID)
		}
	}
}

func generatedHasOperation(operationID string) bool {
	for _, route := range generated.AllRoutes {
		if route.OperationID == operationID {
			return true
		}
	}
	return false
}
