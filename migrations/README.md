# migrations

PostgreSQL schema 迁移脚本、enum source manifest 与 backfill manifest 落点。

Owner subspec: [db-migrations-baseline](../docs/spec/db-migrations-baseline/spec.md)（当前 Contract active spec）。

当前迁移工具已由 B4 定型为 `golang-migrate`，唯一可执行入口是
[`backend/cmd/migrate`](../backend/cmd/migrate/main.go)。不要在本目录另起 Go module，
也不要直接绕过 wrapper 调裸 `golang-migrate`。

常用入口：

- `make migrate-up`
- `make migrate-down`
- `make migrate-status`
- `make migrate-create NAME=add_example`
- `make migrate-check`

新增迁移必须使用 `make migrate-create NAME=...` 生成严格递增的
`NNNNNN_<name>.up.sql` / `NNNNNN_<name>.down.sql` 文件。修改 schema、enum source、
backfill manifest 或 migration wrapper 后，至少运行 `make migrate-check`，并按
[backend README](../backend/README.md) 与 B4 spec 补充对应 Go 测试。

本仓库尚未上线；当前产品范围以 active spec 为准。Pre-launch 迁移链可以保留中间态
DDL 和最终态 cleanup，但当前 baseline inventory 必须以 B4 spec、现行 migration lint
和 privacy matrix 为准。
