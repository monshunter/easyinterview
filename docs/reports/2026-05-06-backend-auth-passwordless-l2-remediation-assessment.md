# Backend Auth Passwordless L2 Remediation 交付复盘报告

> **日期**: 2026-05-06
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-auth/001-passwordless-session-bootstrap` Phase 6 L2 remediation，修复 8 个 review findings：runtime auth wiring、user-scoped deleteMe idempotency、session cookie Secure policy、challenge rate-limit SQL scope、logout revoke failure response、runtime auth secret fail-fast、logout optional-session resolver error、logout revoke race touch zero-row error classification。
- 成功证据：
  - `cd backend && go test ./internal/auth -count=1`
  - `cd backend && go test ./cmd/api -count=1`
  - `cd backend && go test ./... -count=1`
  - `cd backend && go test ./cmd/api -run 'TestBuild(AuthServiceRejectsEmptyAuthSecrets|APIHandlerLogoutPropagatesSessionResolverErrors|APIHandlerLogoutKeepsKnownSessionErrorsOptional)' -count=1`
  - `cd backend && go test ./internal/auth -run TestSessionMiddlewareTreatsTouchLostRaceAsAuthState -count=1`
  - `test/scenarios/e2e/p0-003-passwordless-session-cookie/scripts/{setup,trigger,verify,cleanup}.sh`
  - `make lint`
  - `make codegen-check`
  - `make docs-check`
  - `git diff --check`
- 关联 Bug：[BUG-0013](../bugs/BUG-0013.md)。

## 2 会话中的主要阻点/痛点

- 包级 BDD 曾经给出“已闭环”的错觉。
  - **证据**：L2 review 发现 `cmd/api` 真实入口只挂 anonymous `/runtime-config`，而 BDD test 直接组合 handler / middleware。
  - **影响**：Phase 1-5 历史 PASS 没有覆盖真实 runtime route wiring，导致 Auth API 在实际入口不可达。
- Idempotency gate 只覆盖同一用户重复请求。
  - **证据**：新增两用户同 key SQL red test 后发现原 dedupe 只按 raw `Idempotency-Key` 过滤。
  - **影响**：隐私删除 handoff 存在跨用户复用 active job 的安全风险。
- 聚合 gate 暴露了局部修复中的裸 job literal。
  - **证据**：`make codegen-check` 首轮失败，提示 `backend/internal/auth/store.go` 裸 `privacy_delete` literal。
  - **影响**：修复中需要再对齐 B3 generated jobs constants，避免新增漂移。
- 初次 L2 修复仍漏掉 runtime builder 与 optional-session 错误分类。
  - **证据**：后续 review 指出 dev 默认空 auth secret 仍可构造 C1 runtime；cookie-bearing logout 在 resolver/store error 下仍返回 204。
  - **影响**：需要追加 Phase 6.6 / 6.7，并补 `cmd/api` route/builder 级负例。
- Optional-session 分类修复后仍漏掉 session touch/revoke 的 TOCTOU race。
  - **证据**：后续 review 指出两个 logout 请求重叠时，`ResolveSession` 可能已读到 active session，但 `TouchSession` 因另一请求 revoke 返回 `sql.ErrNoRows`，旧逻辑会让 optional logout 返回 500。
  - **影响**：需要追加 Phase 6.8，并补 middleware 级 race regression test。

## 3 根因归类

- Runtime wiring 证据缺失。
  - **类别**：spec-plan
  - 原 checklist 1.3 写了 generated Auth surface wiring，但历史测试未要求 `cmd/api` 入口级 route/middleware 断言。
- Cross-principal idempotency 负例不足。
  - **类别**：spec-plan
  - 原 checklist 3.5 只表达重复请求不创建重复 job，未显式要求“不同用户同 key 不共享 handoff”。
- Review 后修复需要同时跑 codegen/lint drift gate。
  - **类别**：no repo change needed
  - 现有 `make codegen-check` 已能拦截裸 job literal，本次按现有 gate 修复即可。
- Auth runtime 安全负例应覆盖 builder 与 middleware 分类，而不只覆盖 handler。
  - **类别**：spec-plan
  - 空 secret 不是 handler 行为问题，resolver/store error 也不会进入 handler；需要在 `cmd/api` builder/route 层写负例。
- Session resolver 负例应覆盖 store lookup error 与 touch/revoke race 两类阶段。
  - **类别**：spec-plan
  - 只测 `GetSessionByHash` error 不能证明 sliding renewal 与 revoke 之间的 TOCTOU classification，logout optional contract 需要覆盖 touch zero-row。

## 4 对流程资产的改进建议

- 在 backend-auth 后续计划中保留 Phase 6 的入口级 route wiring 测试作为模板。
  - **落点**：spec-plan
  - **优先级**：high
- 对所有涉及 `Idempotency-Key` / dedupe key 的计划，补“跨主体同 key”负例要求。
  - **落点**：spec-plan
  - **优先级**：high
- 对 authentication/session 类计划，补“runtime builder secret fail-fast”“optional-session 只吞已知无效认证状态”和“touch/revoke TOCTOU zero-row 归类”的固定负例。
  - **落点**：spec-plan
  - **优先级**：high
- 后续 `/plan-code-review --fix` 结束前固定运行 `make codegen-check`，尤其当修复触及 event/job/schema literal。
  - **落点**：skill
  - **优先级**：medium

## 5 建议优先级与后续动作

- 优先：在下一个 backend domain plan review 中检查是否存在“包级 handler tests 替代 runtime wiring tests”的同类盲点。
- 次优先：为 B2 / C8 后续隐私删除实现补同类 idempotency matrix，确保 `DELETE /me` 与 `POST /privacy/deletions` 最终共享语义；为后续 session-bearing API plan 增加 resolver/store error 与 touch/revoke race 的 route-level negative tests。
- 可延后：把 BUG-0013 的经验沉淀进 `PATTERNS.md`；当前已有 BUG 记录和 Phase 6 checklist 作为直接检索入口。
