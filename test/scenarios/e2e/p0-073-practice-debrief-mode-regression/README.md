# E2E.P0.073 Practice debrief mode regression

> **场景 ID**: E2E.P0.073
> **自动化入口**: `cd backend && go test ./cmd/api -run TestE2EP0073PracticeDebriefAssistedStrictAndLegacyNegative -count=1`

验证 `goal='debrief'` 可与 `mode='assisted'` / `mode='strict'` 启动，且 legacy `mode='debrief'` 被拒绝。
