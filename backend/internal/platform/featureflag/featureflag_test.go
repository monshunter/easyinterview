package featureflag_test

import (
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
)

type stubClient struct{}

func (stubClient) IsEnabled(string, featureflag.FlagContext) bool   { return true }
func (stubClient) Variant(string, featureflag.FlagContext) string  { return "control" }
func (stubClient) Snapshot() map[string]featureflag.FlagDecision    { return nil }

func TestClientContract(t *testing.T) {
	var _ featureflag.FeatureFlagClient = stubClient{}
}

func TestFlagContextHasOnlyAllowedFields(t *testing.T) {
	ctx := featureflag.FlagContext{
		AnonymousDistinctID: "anon",
		AuthenticatedUserID: "uid",
		AppEnv:              "dev",
	}
	if ctx.AnonymousDistinctID != "anon" || ctx.AuthenticatedUserID != "uid" || ctx.AppEnv != "dev" {
		t.Errorf("unexpected FlagContext shape: %+v", ctx)
	}
}
