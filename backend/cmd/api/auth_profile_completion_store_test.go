package main

import (
	"context"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func (s *multiUserAuthStore) CompleteUserProfile(_ context.Context, userID string, displayName string, _ time.Time) (auth.UserContext, error) {
	for _, store := range s.users {
		if store.user.ID == userID {
			store.user.DisplayName = displayName
			store.user.ProfileCompletionRequired = false
			return store.user, nil
		}
	}
	return auth.UserContext{}, auth.ErrSessionInvalid
}

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
