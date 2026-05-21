// Package profile owns the candidate profile + experience cards domain
// per backend-profile spec. It exposes the 5 Profile tag HTTP handlers,
// the candidate_profiles / experience_cards store, and a small set of
// read-only / privacy internal APIs consumed by cross-owner subsystems
// (backend-jobs-recommendations, backend internal privacy runner).
package profile
