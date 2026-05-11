# Seed Input

- App entry: `<App client={createDevMockClient()} />`
- Initial route: default Home route.
- Runtime config fixture: `openapi/fixtures/Auth/getRuntimeConfig.json`
- Auth fixtures:
  - `openapi/fixtures/Auth/getMe.json`
  - `openapi/fixtures/Auth/startAuthEmailChallenge.json`
  - `openapi/fixtures/Auth/verifyAuthEmailChallenge.json`
  - `openapi/fixtures/Auth/logout.json`
- User input:
  - email: `alice@example.com`
  - verification token: `654321`

No Kind resources, database rows, browser storage, or external network calls are
seeded for this scenario.
