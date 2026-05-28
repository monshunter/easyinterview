package auth_test

import (
	"context"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func (s *recordingChallengeStore) CompleteUserProfile(context.Context, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *verifyStore) CompleteUserProfile(context.Context, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *deleteMeStore) CompleteUserProfile(context.Context, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *sessionStore) CompleteUserProfile(context.Context, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *runtimeConfigStore) CompleteUserProfile(context.Context, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *rateLimitStore) CompleteUserProfile(context.Context, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *logoutStore) CompleteUserProfile(context.Context, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *meStore) CompleteUserProfile(_ context.Context, userID string, displayName string, _ time.Time) (auth.UserContext, error) {
	if userID != s.user.ID {
		return auth.UserContext{}, auth.ErrUserNotFound
	}
	s.user.DisplayName = displayName
	s.user.ProfileCompletionRequired = false
	return s.user, nil
}

func (s *passwordlessScenarioStore) CompleteUserProfile(_ context.Context, userID string, displayName string, _ time.Time) (auth.UserContext, error) {
	user, ok := s.usersByID[userID]
	if !ok {
		return auth.UserContext{}, auth.ErrUserNotFound
	}
	user.DisplayName = displayName
	user.ProfileCompletionRequired = false
	s.usersByID[userID] = user
	s.usersByEmail[user.Email] = user
	return user, nil
}
