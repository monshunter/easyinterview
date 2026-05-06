// Package auth implements the C1 passwordless session boundary.
//
// C1 keeps passwordless timing and dev sink defaults as code constants, not
// A4 runtime knobs: challenges live for 15 minutes, server-side sessions live
// for 30 days, and the third same-email or same-IP challenge request inside a
// one-minute window is deduped/rate-limited. The first-party cookie name is
// fixed by ADR-Q1 as ei_session.
//
// Frontend-shell handoff:
//   - D1 consumes only the B2 generated auth operations:
//     startAuthEmailChallenge, verifyAuthEmailChallenge, getMe, logout, and
//     the public getRuntimeConfig endpoint with C1's session-aware resolver.
//   - Browser authentication state is the HttpOnly first-party ei_session
//     cookie minted by verifyAuthEmailChallenge. Frontend code must not store a
//     bearer token, raw session token, password credential, OAuth credential,
//     or custom session marker in local/session storage.
//   - startAuthEmailChallenge always returns an accepted response for accepted
//     or rate-limited requests and does not reveal account existence. In local
//     and test flows, the dev mail sink retrieval boundary is the only place
//     that exposes the transient verification link; logs, dispatch payloads,
//     audit events, and metrics must stay redacted.
//   - Missing, invalid, revoked, or expired sessions surface as the B1 error
//     envelope with AUTH_UNAUTHORIZED. Logout is optional-session and always
//     clears the cookie, so frontend pendingAction recovery should treat logout
//     as idempotent.
//   - pendingAction is owned by frontend-shell: C1 only mints the session and
//     serves /me/runtime-config context after verification; the route, params,
//     and action label to restore after login remain frontend state.
package auth
