// Package stub is the deterministic AIClient provider used by unit tests
// and offline contract tests. It computes its output as a hash of the
// profile name and a canonicalized payload, so the same input always
// produces the same response (spec §4.4).
//
// Spec §3.1 D-4 forbids enabling this provider outside APP_ENV=test unless
// the caller passes WithAllowed(true) explicitly. New() enforces that gate.
package stub
