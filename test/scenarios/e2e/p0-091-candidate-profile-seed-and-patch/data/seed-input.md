# Seed Input — E2E.P0.081

## Users

| user | id | session cookie | user_settings |
|------|----|---------------|---------------|
| A | `01918fa3-0000-7000-8000-0000000aa101` | `ei_session=raw-session-user-a` | preferred_practice_language=`en`, ui_language=`zh-CN`, region=`CN-SH` |
| B | `01918fa3-0000-7000-8000-0000000bb201` | `ei_session=raw-session-user-b` | 同上 |
| C | `01918fa3-0000-7000-8000-0000000cc301` | `ei_session=raw-session-user-c` | 同上 |

## DB initial state

- `candidate_profiles`: 0 rows
- `experience_cards`: 0 rows
- `audit_events` for action `profile.privacy_delete`: 0 rows

测试入口 `TestProfileHTTPScenario` 在 `cmd/api` 真实路由上以这些 fixture 用户运行，trigger 调用真实 `backend/internal/profile/store` Repository、`backend/internal/profile/service` Service、`auth.SessionMiddleware`、idempotency middleware。
