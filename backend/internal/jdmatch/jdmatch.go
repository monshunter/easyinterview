// Package jdmatch is the backend-jobs-recommendations domain root: HTTP
// handlers, store repositories, service orchestrators, and the in-process
// agent_scan background job all live under this package tree.
//
// Spec: docs/spec/backend-jobs-recommendations/spec.md
// Owner plan: docs/spec/backend-jobs-recommendations/plans/001-jd-match-real-backend-baseline/plan.md
package jdmatch

import (
	stderrs "errors"
	"time"
)

// AgentScanStatus mirrors the spec D-X status enum (idle / scanning /
// error) registered in migrations/enum-sources.yaml.
type AgentScanStatus string

const (
	AgentScanStatusIdle     AgentScanStatus = "idle"
	AgentScanStatusScanning AgentScanStatus = "scanning"
	AgentScanStatusError    AgentScanStatus = "error"
)

// AllAgentScanStatuses lists every defined value in declaration order;
// runtime callers should range over this constant when they need a stable
// canonical set (status filter, lint, debug dump).
var AllAgentScanStatuses = []AgentScanStatus{
	AgentScanStatusIdle,
	AgentScanStatusScanning,
	AgentScanStatusError,
}

// AgentScanRecord is the on-disk projection of an `agent_scans` row.
// Handlers and services consume it via the store Repository; the
// generated `AgentScanStatus` DTO from `backend/internal/api/generated`
// is constructed by handler.
type AgentScanRecord struct {
	ID                  string
	UserID              string
	Status              AgentScanStatus
	StartedAt           *time.Time
	FinishedAt          *time.Time
	LastScanAt          *time.Time
	NextScanAt          *time.Time
	ErrorMessage        *string
	RecommendationCount int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Domain-level error sentinels. Handler / service / store layers map
// their implementation-specific errors to one of these so the HTTP
// layer can convert to the canonical RESOURCE_NOT_FOUND / cross-user
// 404 envelopes without leaking storage-specific details.
var (
	// ErrNotFound covers both "row not present" and "row present but
	// owned by another user" (spec D-6 cross-user isolation).
	ErrNotFound = stderrs.New("jdmatch: resource not found")
	// ErrInvalidStatus is returned when the caller hands a status
	// value outside AllAgentScanStatuses.
	ErrInvalidStatus = stderrs.New("jdmatch: invalid agent_scan status")
	// ErrUserIDRequired is returned when a repository / service is
	// called with an empty userID.
	ErrUserIDRequired = stderrs.New("jdmatch: userID is required")
	// ErrValidationFailed maps to a 400 VALIDATION_FAILED response and
	// is reused by createSavedSearch / searchJobs validators.
	ErrValidationFailed = stderrs.New("jdmatch: validation failed")
)
