# E2E.P0.061 Debrief Get Draft/Completed + Cross-User Isolation

> **场景 ID**: E2E.P0.061
> **关联需求**: backend-debrief C-6, C-7, C-8
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

## Given

用户 A 拥有 draft 与 completed 两种 debrief，用户 B 不应读取用户 A 的 debrief。

## When

调用 `GET /debriefs/{debriefId}`。

## Then

draft response 不含 completed-only 字段，completed response 带完整 questions / riskItems / provenance，cross-user 与 missing id 均返回 `DEBRIEF_NOT_FOUND`。
