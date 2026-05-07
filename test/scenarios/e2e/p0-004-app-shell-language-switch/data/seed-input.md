# E2E.P0.004 Seed Input

- Runtime config fixture: `openapi/fixtures/Auth/getRuntimeConfig.json`
  `default` scenario (`defaultUiLanguage: zh-CN`).
- User context fixture: `openapi/fixtures/Auth/getMe.json` `unauthenticated`
  scenario.
- Initial route: `home`.
- User action sequence:
  1. Observe default Chinese shell.
  2. Switch TopBar language control to `en`.
  3. Navigate to auth / profile / settings / placeholder shells.
  4. Submit D1 auth operations through fixture-backed generated client.
