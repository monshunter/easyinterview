package auth

import "errors"

// ErrUserNotFound indicates the userID did not match a non-deleted users
// row. Callers should fall back to a non-PII anonymous display name rather
// than block the whole endpoint. (The cross-owner UserIdentity projection
// that originally lived beside this error was removed with the JD-Match
// module per product-scope v2.1 D-17; the error stays because the auth
// service and store classify missing users with it.)
var ErrUserNotFound = errors.New("auth: user identity not found")
