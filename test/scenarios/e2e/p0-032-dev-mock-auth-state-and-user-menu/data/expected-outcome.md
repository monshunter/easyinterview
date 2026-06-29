# Expected Outcome

The scenario is successful when all of the following evidence is produced:

- Initial Home render is unauthenticated and shows `topbar-login` plus
  `topbar-register`.
- Passwordless mock login changes the mounted App to authenticated state without
  replacing the client.
- Authenticated TopBar exposes `topbar-user-chip` and `topbar-user-avatar`.
- `topbar-user-settings` and `topbar-user-logout` are
  absent until `topbar-user-chip` opens the dropdown.
- Dropdown header includes `Alice Example` and `ali***@example.com`.
- Settings action routes to `route-settings`; no `topbar-user-profile` action is exposed.
- Logout confirmation routes through `route-auth_logout`, calls the generated
  logout operation, and refreshes `/me` to unauthenticated.
- The final TopBar state shows `登录` / `注册` again.
- Evidence log contains `dev mock unauthenticated login avatar dropdown settings logout`.
- Evidence log does not contain retired inline or prototype-source markers.
