# Backend TargetJob Contract Remediation 交付复盘报告

> **日期**: 2026-05-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `/plan-code-review backend-targetjob/001-targetjob-import-and-parse-bootstrap backend --base-rev 49a55c0c0c45b94f931ff013856232b3b36f0463` 后续确认的 TargetJob L2 contract drift，包括 error envelope、pagination envelope、parse runtime、AI output validation 与 BDD HTTP evidence 口径。
- 代码证据：`backend/internal/targetjob/handler.go` 改为 generated `ApiErrorResponse`；`service.go` 回填 `pageInfo.pageSize`；`parse_executor.go` 从 `PromptResolution` 组装 A3 payload 并严格校验 requirements；`cmd/api/main.go` 在 `APP_ENV=test` 注入 deterministic JSON parse fixture。
- 文档证据：plan/checklist/spec 升到 1.4 并勾选 7.5-7.10；`docs/spec/backend-targetjob/spec.md`、plans INDEX、BDD plan/checklist、scenario INDEX 与 p0-010..013 README / scripts 均对齐 `cmd-api-http` evidence 状态。
- 验证通过：context validator、`go test ./internal/targetjob/... ./cmd/api -count=1`、`go test ./internal/ai/aiclient/... -count=1`、p0-010..013 `TEST_RUN_ID=targetjob-http-20260508` setup / trigger / verify / cleanup、`make validate-fixtures`、`make lint-events`、`make lint-config`、`migrations_lint`、`make docs-check`、`make codegen-check`、`git diff --check`。
- Bug 记录：[BUG-0028](../bugs/BUG-0028.md) 已记录本次 multi-component contract drift。

## 2 会话中的主要阻点/痛点

- Generated contract 与 handler tests 脱节。
  - **证据**：历史 handler tests 只看 status / substring，未反序列化 generated `ApiErrorResponse`，导致 legacy `{"errors":[...]}` 留存。
  - **影响**：OpenAPI consumer 可能按 `error` envelope 解析失败。
- Package proxy 与真实 BDD evidence 曾经容易混淆。
  - **证据**：scenario scripts 曾需要 `ALLOW_TARGETJOB_PACKAGE_PROXY=1` 才跑包级 go test，但 verify 曾写 `status=passed` / `method=go-test`；7.10 后 p0-010..013 已改为 `cmd/api` HTTP harness。
  - **影响**：若未来脚本回退，review 可能再次把 TDD proxy 当成 `auth -> HTTP API -> cmd/api drainer` 场景 PASS。
- Test runtime 与 package fake 不是同一条 AI 链路。
  - **证据**：包级 E2E test 注入 JSON fake，`cmd/api` runtime 使用 A3 generic stub，stub 返回非 JSON content。
  - **影响**：真实 runtime 成功路径不能由当前 test env 闭合。

## 3 根因归类

- Generated envelope 漂移。
  - **类别**：spec-plan
  - **说明**：checklist 缺少对 generated response envelope 的语义断言，历史 status-only tests 覆盖不足。
- BDD evidence 误读风险。
  - **类别**：spec-plan / README
  - **说明**：scenario README、verify output 与 plan gate 没有用机器可读字段区分 proxy evidence 和 BDD evidence。
- Runtime fixture 缺口。
  - **类别**：spec-plan
  - **说明**：计划要求 `APP_ENV=test` 可闭合 parse 成功路径，但未明确 deterministic fixture 必须挂在 `cmd/api` runtime，而不是只在包级 fake 中存在。

## 4 对流程资产的改进建议

- 在 backend plan-code-review 检查模板中加入 generated envelope semantic assertions。
  - **落点**：skill / spec-plan
  - **优先级**：high
- 对所有 scenario verify output 固定要求机器可读的 `method` 与 `validBddEvidence`；真实场景必须输出 `method=cmd-api-http` / `validBddEvidence=true`。
  - **落点**：test scenario README / spec-plan
  - **优先级**：high
- 对涉及 AI runtime 的计划补充 “package fake 与 cmd/api runtime fixture 必须分别验证” 的 gate。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- **最高优先级**：确认是否按 `/tdd` 生命周期规则把 `backend-targetjob/001-targetjob-import-and-parse-bootstrap` plan/checklist/BDD docs 切换为 `completed`，随后用 `/work-journal` 做 phase commit。
- **次优先级**：把 `method` / `validBddEvidence` 作为场景框架层 lint，防止其他包级 proxy 脚本再次冒充 BDD PASS。
- **可延后**：将 F3 runtime package 接入替换当前 `StaticPromptRegistry` contract bridge；当前 dev / staging / prod fail-closed 语义已经保留。
