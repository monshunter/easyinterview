// Package idx holds business-id helpers: UUIDv7 minting and the
// RequireServerID guard that rejects browser-only `tmp_` prefixed identifiers
// from reaching persistence layers.
//
// The package is hand-written; the generator only emits a small header file
// when shared/conventions.yaml changes the tmp prefix or sample UUIDv7.
package idx
