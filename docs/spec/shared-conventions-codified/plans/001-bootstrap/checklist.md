# Shared Conventions Codified Bootstrap Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-04-27

**关联计划**: [plan](./plan.md)

## Phase 1: 跨语言真理源与 generator 前置依赖

- [x] 1.1 写入 `shared/conventions.yaml`：13 个上游 §5 小节 / 14 个生成枚举类型 + 6 个已记录错误码示例 + Job 状态 + `PageInfo` / `ApiError` 结构（与 [00-shared-conventions.md](../../../../easyinterview-tech-docs/00-shared-conventions.md) §3 / §4 / §5 严格对齐）
- [x] 1.2 落地 `backend/go.mod`（module path `github.com/monshunter/easyinterview/backend`）+ `backend/internal/shared/{types,errors,idx}/` 目录与 `doc.go` 占位，作为 generator 运行前置依赖
- [x] 1.3 落地 `backend/cmd/codegen/conventions/main.go`：从 YAML 渲染 Go 与 TS 文件，输出 idempotent；接入根 `Makefile` 的 `codegen-conventions` target

## Phase 2: Go / TS 共享 module 骨架

- [x] 2.1 在 `backend/internal/shared/idx/` 实现 `NewID()` (UUIDv7) 与 `RequireServerID()`（拒绝 `tmp_` 前缀）；在 `backend/internal/shared/errors/` 实现 `APIError struct` 与 `Wrap()` helper
- [x] 2.2 在 `backend/internal/shared/idx/` 实现与 TS 端 wire-format 一致的 `Idempotency-Key` 生成 / 解析 / 24h TTL 工具，覆盖 spec C-4 的 Go 侧
- [x] 2.3 落地 `frontend/package.json`（name `@easyinterview/frontend`、私有、`build`/`lint` 占位，`typecheck`/`test` 真实可执行，依赖 `uuid >=10`，devDep 含 `typescript` / `vitest` / `@types/node`）+ 仓库根 `pnpm-workspace.yaml` + strict `frontend/tsconfig.json`
- [x] 2.4 在 `frontend/src/lib/{conventions,ids}/` 创建占位 `index.ts`，并实现 `requireServerId()`、`newId()`（UUIDv7）、`Idempotency-Key` 24h TTL 工具，覆盖 spec C-4 的 TS 侧
- [x] 2.5 L2 remediation: TS `parseIdempotencyKey` 拒绝指数 / 十六进制 / 空白等非十进制时间戳，保持与 Go `ParseIdempotencyKey` 的 wire-format 校验一致

## Phase 3: Lint 与命名约束

- [x] 3.1 落地 `backend/.golangci.yml`（启用 `revive var-naming`）+ 本地可执行错误码校验，确保 `make lint` 能拦截非 `UPPER_SNAKE_CASE` 错误码
- [x] 3.2 落地 `frontend/.eslintrc.cjs`（或 `eslint.config.js`）+ 本地可执行边界校验，拒绝在 `lib/conventions/errors.ts` 之外定义错误码字面量
- [x] 3.3 L2 remediation: `scripts/lint/error_codes.py` 必须拒绝 `ERROR_CODES` 对象内任何小写 / 非法 key 或 value，不能因正则只匹配合法条目而漏报

## Phase 4: Verification

- [x] 4.1 跑两次 `make codegen-conventions`，第二次 `git diff --exit-code` 通过；删除任一生成文件后再跑可还原
- [x] 4.2 `go test ./backend/internal/shared/...` 与 `go vet ./backend/...` 通过；`idx_test.go` / `idempotency_test.go` / `errors_test.go` 覆盖 spec C-3 / C-4 / 错误响应结构
- [x] 4.3 `pnpm --filter @easyinterview/frontend exec tsc --noEmit` 通过；最小 vitest / node:test 用例覆盖 `requireServerId('tmp_x')` 抛错、枚举 union 类型、TS `Idempotency-Key` wire-format / TTL 语义
- [x] 4.4 确认 `docs/spec/INDEX.md` 中 `shared-conventions-codified` 行为真实链接 + 真实状态；不改写已经完成的 `engineering-roadmap/001-decompose-subspecs` Phase 2 spawn 项；将 generator / Go test / TS test 输出贴入工作日志
- [x] 4.5 L2 remediation: 仓库根级 `go test ./backend/internal/shared/...` 与 `go vet ./backend/...` 必须真实通过，不依赖手动切换到 `backend/`
- [x] 4.6 文档一致性修订：对齐 A5 单人开发阶段决策，B1 只要求本地 lint/codegen 质量门禁，远端 CI / PR required check / CI drift detection 不作为当前 P0 前置
