# E2E.P0.072 Practice derived source isolation privacy

> **场景 ID**: E2E.P0.072
> **自动化入口**: `cd backend && go test ./cmd/api -run TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy -count=1`

验证 missing、cross-user、wrong-target、draft 和 empty source 均返回统一 validation envelope，且不泄露 source id 或 debrief raw text。
