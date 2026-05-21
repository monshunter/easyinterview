# E2E.P0.094 expected outcome

- `GET /jd-match/profile` returns 200 + JobMatchProfile with
  `displayName` non-null, `avatarUrl / locationText / compensationText`
  null, `skills` empty array, `sources { resumes:3, jds:5, mocks:8,
  debriefs:2 }`; D-19 structural parity assertions pass; raw email
  never leaves the response.
- `GET /jd-match/agent-status` returns 200 + idle baseline.
- `GET /jd-match/recommendations?pageSize=20` returns 20 items + a
  `pageInfo.nextCursor`; the second page returns 5 items + hasMore=false.
- `GET /jd-match/recommendations/{id}` returns the detail projection.
- `POST /jd-match/recommendations/{id}/dismiss` returns 200 +
  `dismissedAt`; the row no longer appears in subsequent list calls.
- Cross-user reads (user B) on any of the routes return 404
  RESOURCE_NOT_FOUND.
- `freeNote`, `reasons`, `risks`, `source_url`, `interview_hypotheses`,
  and raw email never appear in log / audit / outbox payload.
