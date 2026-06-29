package main

import (
	"context"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func (s *practiceScenarioAuthStore) CompleteUserProfile(_ context.Context, userID string, displayName string, _ time.Time) (auth.UserContext, error) {
	user, ok := s.usersByID[userID]
	if !ok {
		return auth.UserContext{}, auth.ErrSessionInvalid
	}
	user.DisplayName = displayName
	user.ProfileCompletionRequired = false
	s.usersByID[userID] = user
	return user, nil
}
