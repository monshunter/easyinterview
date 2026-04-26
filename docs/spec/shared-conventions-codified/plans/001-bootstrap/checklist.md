# Shared Conventions Codified Bootstrap Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-26

**关联计划**: [plan](./plan.md)

## Phase 1: 跨语言真理源与 generator

- [x] 1.1 写入 `shared/conventions.yaml`：13 类枚举 + 6 个已记录错误码示例 + Job 状态 + `PageInfo` / `ApiError` 结构（与 [00-shared-conventions.md](../../../../easyinterview-tech-docs/00-shared-conventions.md) §3 / §4 / §5 严格对齐）
- [x] 1.2 落地 `backend/cmd/codegen/conventions/main.go`：从 YAML 渲染 Go 与 TS 文件，输出 idempotent；接入根 `Makefile` 的 `codegen-conventions` target

## Phase 2: Go / TS 共享 module 骨架

- [x] 2.1 落地 `backend/go.mod`（module path `github.com/monshunter/easyinterview/backend`）+ `backend/internal/shared/{types,errors,idx}/` 目录与 `doc.go` 占位
- [x] 2.2 在 `backend/internal/shared/idx/` 实现 `NewID()` (UUIDv7) 与 `RequireServerID()`（拒绝 `tmp_` 前缀）；在 `backend/internal/shared/errors/` 实现 `APIError struct` 与 `Wrap()` helper
- [x] 2.3 落地 `frontend/package.json`（name `@easyinterview/frontend`、私有、`build`/`lint`/`test` script 占位，依赖 `uuid >=10`）+ 仓库根 `pnpm-workspace.yaml`
- [x] 2.4 在 `frontend/src/lib/{conventions,ids}/` 创建占位 `index.ts`，并实现 `requireServerId()`、`newId()`（UUIDv7）、`Idempotency-Key` 24h TTL 工具

## Phase 3: Lint 与命名约束

- [x] 3.1 落地 `backend/.golangci.yml`（启用 `revive var-naming`）+ 本地可执行错误码校验，确保 `make lint` 能拦截非 `UPPER_SNAKE_CASE` 错误码
- [x] 3.2 落地 `frontend/.eslintrc.cjs`（或 `eslint.config.js`）+ 本地可执行边界校验，拒绝在 `lib/conventions/errors.ts` 之外定义错误码字面量

## Phase 4: Verification

- [x] 4.1 跑两次 `make codegen-conventions`，第二次 `git diff --exit-code` 通过；删除任一生成文件后再跑可还原
- [x] 4.2 `go test ./backend/internal/shared/...` 与 `go vet ./backend/...` 通过；`idx_test.go` / `errors_test.go` 覆盖 spec C-3 / 错误响应结构
- [x] 4.3 `pnpm --filter @easyinterview/frontend exec tsc --noEmit` 通过；最小 vitest / node:test 用例覆盖 `requireServerId('tmp_x')` 抛错与枚举 union 类型
- [x] 4.4 确认 `docs/spec/INDEX.md` 中 `shared-conventions-codified` 行为真实链接 + 真实状态；不改写已经完成的 `engineering-roadmap/001-decompose-subspecs` Phase 2 spawn 项；将 generator / Go test / TS test 输出贴入工作日志
