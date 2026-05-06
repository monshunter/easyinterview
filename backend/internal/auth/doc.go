// Package auth implements the C1 passwordless session boundary.
//
// C1 keeps passwordless timing and dev sink defaults as code constants, not
// A4 runtime knobs: challenges live for 15 minutes, server-side sessions live
// for 30 days, and the third same-email or same-IP challenge request inside a
// one-minute window is deduped/rate-limited. The first-party cookie name is
// fixed by ADR-Q1 as ei_session.
package auth
