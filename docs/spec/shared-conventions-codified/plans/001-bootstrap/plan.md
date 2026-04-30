# Shared Conventions Codified Bootstrap

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-04-27

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [shared-conventions-codified spec](../../spec.md) §3.1 锁定的 6 项决策落到代码：建立 `shared/conventions.yaml` 真理源、跨语言 generator、Go 共享 module（`backend/go.mod` + `internal/shared/{types,errors,idx}/`）、TS 共享 lib（`frontend/package.json` + `src/lib/{conventions,ids}/`）、UUIDv7 / Idempotency-Key 工具、错误码与枚举命名的本地可执行 lint gate，并通过本 plan 的 verification phase 证明 Go / TS 双侧测试可以编译并通过最小用例。

本 plan 只覆盖 W0 必须冻结的最小集合；后续如需扩展（本地 drift detection、prompt registry 接入、跨语言 contract test），递增 spec 与本 plan 版本，必要时 spawn `002-codegen-pipeline` 等续集 plan。远端 CI drift detection 仅在 A5 触发条件成立后再评估。

## 2 背景

[engineering-roadmap spec §5.7 / §5.8](../../../engineering-roadmap/spec.md#57-实施-wave-顺序) 把 B1 安排在 W0，与 [A1 `repo-scaffold`](../../../repo-scaffold/spec.md) 同时落地：A1 提供根目录与 Make 入口，B1 在 A1 创建的 `backend/` 与 `frontend/` 容器里写入第一份 `go.mod` / `package.json` 与共享 lib。这两件事必须先于 W1 9 份 spec 进 `/plan-review`，否则 B2 / C 全域 / D 全域会缺少共享枚举与错误码常量。

执行本 plan 前必须确认 A1 已创建根 `Makefile`、`backend/`、`frontend/`、`scripts/` 等容器目录；若 A1 尚未完成，先暂停本 plan 并实施 `repo-scaffold/001-bootstrap`。

本 plan 不接入 OpenAPI codegen（B2 持有），不实现业务 handler，不依赖 docker / db。所有产出限于仓库内文件 + Go test / TS test 跑得通。

## 3 实施步骤

本 plan v1.1 为完成态文档修订：按已验证的实际依赖关系回写 Phase 1 / Phase 2 的执行顺序与所有权边界，不改变已落地代码范围。执行顺序必须遵循依赖图：先落地真理源与 Go module，再运行归属 `backend/go.mod` 的 generator；TS toolchain 必须先于 TS helper 的原生 typecheck / test gate 完成。

### Phase 1: 跨语言真理源与 generator 前置依赖

#### 1.1 落地 `shared/conventions.yaml`

按 [00-shared-conventions.md §5](../../../../../easyinterview-tech-docs/00-shared-conventions.md#5-枚举目录) 写入 13 个上游小节覆盖的 14 个生成枚举类型（§5.13 拆为隐私请求 type/status 两个并行类型），按 §3.2 写入错误码示例（`AUTH_UNAUTHORIZED` / `TARGET_IMPORT_FAILED` / `PRACTICE_SESSION_CONFLICT` / `REPORT_NOT_READY` / `VALIDATION_FAILED` / `RATE_LIMITED` 等），按 §4.2 写入 Job 状态，按 §3 写入 `PageInfo` / `ApiError` 结构。文件结构必须可被 `gopkg.in/yaml.v3` 与 `js-yaml` 同等解析。

#### 1.2 Go module 初始化

- 在 `backend/` 下落地 `go.mod`，module path 锁定 `github.com/monshunter/easyinterview/backend`（spec D-2）；Go 版本对齐 [.tool-versions](../../../repo-scaffold/plans/001-bootstrap/plan.md#12-根-editorconfig--gitignore--tool-versions) 中的 `golang` 字段。
- 在 `backend/internal/shared/` 下创建 `types/`、`errors/`、`idx/` 三个包目录与 `doc.go` 占位；generator 在 1.3 中写入生成文件。

#### 1.3 写入 generator

在 `backend/cmd/codegen/conventions/` 下落地 generator，使 generator 归属 `backend/go.mod`：

- `main.go`（Go 实现，由 `make codegen-conventions` 调用 `go run ./backend/cmd/codegen/conventions`）：读取 `shared/conventions.yaml`，按模板渲染 `backend/internal/shared/types/enums.go`、`backend/internal/shared/types/http_dto.go`（`PageInfo` 与共享常量）、`backend/internal/shared/errors/codes.go`、`backend/internal/shared/idx/generated.go`。
- 同一个二进制额外渲染 TS 文件到 `frontend/src/lib/conventions/{enums,errors,pagination}.ts` 与 `frontend/src/lib/ids/generated.ts`；TS `errors.ts` 持有 generator 输出的 `ApiError` interface，Go `APIError` 结构归属手写 `backend/internal/shared/errors/errors.go`。
- 输出必须 idempotent：再跑一次 `git diff --exit-code` 不变。

### Phase 2: Go / TS 共享 module 骨架

#### 2.1 Go shared helpers

- `backend/internal/shared/idx/` 中手写 `RequireServerID(string) error`：拒绝 `tmp_` 前缀；`NewID()` 调用 UUIDv7 实现（依赖 `github.com/google/uuid` v1.6+）。
- `backend/internal/shared/errors/` 中手写 `APIError struct` 基类与 `Wrap(code string, msg string, retryable bool)` helper；常量由 generator 写入。

#### 2.2 Go Idempotency-Key 工具

- 在 `backend/internal/shared/idx/idempotency.go` 中手写与 TS 端一致的 `GenerateIdempotencyKey()` / `ParseIdempotencyKey()` / `IsIdempotencyKeyExpired(...)`：生成 24h TTL 的 `Idempotency-Key`（`v1.{unixSeconds}.{uuidv7}`），拒绝非十进制时间戳、非 UUIDv7 与 `tmp_` 前缀。

#### 2.3 TS workspace 与测试工具链初始化

- 在 `frontend/` 下落地最小 `package.json`（name `@easyinterview/frontend`、private true、`build` / `lint` 可保持 D1 前的占位入口，`typecheck` / `test` 必须真实调用 `tsc --noEmit` 与 vitest）。
- 安装最小 TS 验证 devDep（`typescript` / `vitest` / `@types/node`）与 `uuid >=10` 运行依赖；落地 strict `frontend/tsconfig.json`，启用 `noUncheckedIndexedAccess`。
- 在仓库根落地 `pnpm-workspace.yaml`，packages 字段含 `frontend`；若后续新增 TS 工具包，再由对应 plan 扩展 workspace packages。

#### 2.4 TS 共享 lib 与 Idempotency-Key 工具

- 在 `frontend/src/lib/conventions/` 与 `frontend/src/lib/ids/` 下创建占位 `index.ts`，由 generator 写入实际内容。
- `frontend/src/lib/ids/index.ts` 中手写 `requireServerId(s: string)` 与 `newId()`（基于 `uuid >=10` 的 UUIDv7 实现）。
- `frontend/src/lib/conventions/idempotency.ts` 与 Go 端对偶：生成 24h TTL 的 `Idempotency-Key`（UUIDv7 + 时间戳头）。

#### 2.5 Idempotency-Key 解析对偶修订

- L2 remediation: TS 端 `parseIdempotencyKey` 的时间戳解析必须与 Go 端 `strconv.ParseInt(..., 10, 64)` 的十进制语义保持一致，拒绝指数、十六进制、空白等非十进制秒数字符串。

### Phase 3: Lint 与命名约束

#### 3.1 Go lint 与错误码校验

在 `backend/.golangci.yml` 落地最小配置：启用 `revive` 的 `var-naming` 规则；同时提供本地可执行的错误码校验（可放在 generator 或 `scripts/lint/` 下），扫描 `backend/internal/shared/errors/` 与 `frontend/src/lib/conventions/errors.ts`，强制错误码常量和值为 `UPPER_SNAKE_CASE`。当前单人开发阶段不接入远端 CI；B1 必须让 `make lint` 在本地能验证该规则，A5 只记录本地质量门禁与远端 CI 延后边界。

#### 3.2 TS lint 与错误码边界

在 `frontend/.eslintrc.cjs`（或 `eslint.config.js`）中加入最小可执行规则或脚本入口：拒绝在 `frontend/src/lib/conventions/errors.ts` 之外定义错误码字面量；约束错误码必须 `UPPER_SNAKE_CASE`。D1 可在前端壳 plan 中扩展 ESLint 体系，但不得放宽 B1 的错误码边界。

L2 remediation: `scripts/lint/error_codes.py` 必须解析 `ERROR_CODES = { ... }` 内的全部条目并拒绝任何非 `UPPER_SNAKE_CASE` 的 key/value，不能只匹配已经合法的条目。

### Phase 4: Verification

#### 4.1 Generator idempotency

- 跑两次 `make codegen-conventions`；第二次后 `git status` 必须 clean。
- 删除任意一个生成文件再重跑，确认完整还原。

#### 4.2 Go test 自检

- `go test ./backend/internal/shared/...` 通过：
  - `idx_test.go` 验证 `NewID()` 返回 UUIDv7 字符串、`RequireServerID("tmp_xxx")` 返回 error。
  - `idempotency_test.go` 验证 Go 端 `Idempotency-Key` 生成、解析、24h TTL 与非法格式拒绝。
  - `errors_test.go` 验证 `Wrap(...)` 输出 JSON 满足 [00-shared-conventions.md §3.2](../../../../../easyinterview-tech-docs/00-shared-conventions.md#32-错误响应) 结构。
- `go vet ./backend/...` 通过。
- L2 remediation: 仓库根必须提供 Go workspace 或等价入口，使上述根级 `go test ./backend/...` / `go vet ./backend/...` 命令无需手动 `cd backend` 即可真实运行。

#### 4.3 TS test 自检

- `pnpm --filter @easyinterview/frontend exec tsc --noEmit` 通过。
- 加入最小 test runner（vitest 或 node:test）跑 `lib/ids` / `lib/conventions` 的最小用例：`requireServerId('tmp_x')` 抛错；枚举 union 类型在 type 层正确；TS 端 `Idempotency-Key` 与 Go 端 wire-format / TTL 语义一致。

#### 4.4 文档同步

- 确认 `docs/spec/INDEX.md` 的 `shared-conventions-codified` 行已指向真实链接且状态 / 版本 / 更新日期与 spec Header 一致；若已有内容一致，不重复改写。
- 不修改 [engineering-roadmap/001-decompose-subspecs/checklist.md](../../../engineering-roadmap/plans/001-decompose-subspecs/checklist.md) 中已经完成的 Phase 2 spawn 项；若需要重新打开父 plan，必须由 roadmap owner 明确触发。
- 把 generator 命令、Go test、TS test 的输出贴入工作日志。

## 4 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-1 到 C-5 全部成立（C-6 由 B2 plan 在引用 B1 时验证）。
- 本 plan checklist 全部勾选；Phase 4 的关键命令日志贴入工作日志。
- engineering-roadmap/001 Phase 2.2 已完成 spawn；本 plan 完结状态作为 B1 后续实施证据记录在本 checklist 与工作日志中。

## 5 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Go / TS UUIDv7 库版本漂移导致格式不一致 | Phase 1 在 `shared/conventions.yaml` 显式写出测试 UUID 样本；generator 在两侧都引用相同正则进行格式校验；当前由本地测试与 lint 门禁收口，远端 CI 仅在 A5 触发条件成立后再评估 |
| pnpm workspace 在 macOS / Linux / CI 环境差异（symlink / hoisting） | Phase 2.3 启用 `pnpm-workspace.yaml` + `package.json#packageManager` 字段锁定 pnpm 版本；A2 dev stack 在 docker 镜像中预装相同版本 |
| Generator idempotency 被 IDE 自动 format 或 import sorter 破坏 | Phase 1.3 在 generator 输出固定 import 顺序与 build tag；Phase 4.1 通过本地 `git diff --exit-code` 强制；编辑器 format 必须由 `.editorconfig` 与 generator 模板一致 |
| Go module path 与 GitHub repo 名称不一致导致 import 失败 | spec D-2 锁定为 `github.com/monshunter/easyinterview/backend`，Phase 1.2 写入 `backend/go.mod` 时直接复用此值；任何重命名必须先递增 B1 spec 版本 |
| TS 共享 lib 被前端 child（D1+）放在不同 path alias 中导致依赖混乱 | 本 plan 锁定 `frontend/src/lib/{conventions,ids}/` 路径；前端 child 只能加 alias，不能改物理路径；spec D-3 与 §4.3 对此明确禁令 |

## 6 修订记录

| 日期 | 版本 | 变更 | 证据 |
|------|------|------|------|
| 2026-04-27 | 1.2 | 对齐 A5 单人开发阶段决策：B1 的 lint/codegen drift 只要求本地质量门禁，远端 CI / PR required check / CI drift detection 不作为当前 P0 前置。 | 文档一致性修订；不新增运行时代码范围。 |
| 2026-04-27 | 1.1 | 回写 [shared-conventions-codified/001-bootstrap 交付复盘报告](../../../../reports/2026-04-27-shared-conventions-codified-001-bootstrap-assessment.md) 中确认真实的 spec-plan 漂移：重排 generator 前置依赖、明确 Go/TS `APIError` 归属、统一 13 个上游小节 / 14 个生成类型口径、补齐 Go/TS idempotency 双端 checklist 落点、把 TS toolchain 上移到 Phase 2。 | 当前代码已落地并通过 Phase 4 验证；本修订不新增运行时代码范围。 |
