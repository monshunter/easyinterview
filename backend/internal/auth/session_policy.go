package auth

import "github.com/monshunter/easyinterview/backend/internal/api/generated"

type SessionRequirement string

const (
	SessionPublic   SessionRequirement = "public"
	SessionOptional SessionRequirement = "optional"
	SessionRequired SessionRequirement = "required"
)

var publicOperations = map[string]SessionRequirement{
	"startAuthEmailChallenge":  SessionPublic,
	"verifyAuthEmailChallenge": SessionPublic,
	"getRuntimeConfig":         SessionPublic,
	"logout":                   SessionOptional,
}

// SessionPolicyForOperation classifies every B2 generated operation for C1
// middleware. Public and optional operations are explicit exceptions; all
// other current generated operations require a first-party session.
func SessionPolicyForOperation(operationID string) (SessionRequirement, bool) {
	if requirement, ok := publicOperations[operationID]; ok {
		return requirement, true
	}
	for _, route := range generated.AllRoutes {
		if route.OperationID == operationID {
			return SessionRequired, true
		}
	}
	return "", false
}
