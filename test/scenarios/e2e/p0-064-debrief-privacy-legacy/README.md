# E2E.P0.064 Debrief Privacy + Legacy Negative

> **场景 ID**: E2E.P0.064
> **关联需求**: backend-debrief C-14, C-15
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

## Given

请求中包含 marker `__SECRET_RAW_TEXT__`，旧 debrief/mistakes/growth 口径必须不在 runtime surface 出现。

## When

执行 debrief outbox、audit、task run 和 legacy negative lint gates。

## Then

outbox / audit / task-run metadata 不泄漏 raw text；legacy terms 在 runtime / contract / fixtures / scenario runtime 中 0 命中。
