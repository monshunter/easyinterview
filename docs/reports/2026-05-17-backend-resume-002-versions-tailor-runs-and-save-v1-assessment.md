# Backend Resume 002 Versions Tailor Runs And Save v1 交付复盘报告

> **日期**: 2026-05-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-resume/002-versions-tailor-runs-and-save-v1`，覆盖 `confirmResumeStructuredMaster`、version read/update/branch、`requestResumeTailor`、`getResumeTailorRun`、resume_tailor drainer、suggestion accept/reject、B2 fixture parity、E2E.P0.074-P0.080 场景，以及 frontend-resume-workshop handoff。
- 成功证据：
  - `cd backend && go test ./...`
  - `cd backend && go test ./internal/resume/... -count=1`
  - `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResume.*HTTPScenario|TestResumeTailorDrainer.*' -count=1`
  - `make validate-fixtures`
  - `test/scenarios/e2e/p0-074...` through `p0-080...` `setup -> trigger -> verify -> cleanup`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`
  - `git diff --check`

## 2 会话中的主要阻点/痛点

- Phase 9 全量 gate 暴露 cross-owner baseline 漂移。
  - **证据**：`cd backend && go test ./...` 初次失败在 `backend/cmd/codegen/openapi` 的 `58-row table` 期望，以及 `shared/fixtures/conventions-parity.json` 缺少 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`。
  - **影响**：业务实现已完成后仍需要回补 codegen / shared conventions parity，说明新增 operation/error code 的 cross-owner baseline gate 可以更前置。
- P0.080 privacy gate 初始是 no-op。
  - **证据**：`go test ./internal/resume/jobs -run 'TestOutboxPrivacy|TestAuditPrivacy|TestAiTaskRunsPrivacy' -count=1 -v` 返回 `[no tests to run]`，随后补充了 `TestOutboxPrivacyForTailorCompletedEvent`、`TestAiTaskRunsPrivacyForTailorDrainer`、`TestAuditPrivacyForTailorDrainer`。
  - **影响**：如果只依赖脚本名称和历史 PASS，privacy regression 场景会误判为已覆盖。
- P0.080 verify 的隐私 grep 首版匹配了说明文档。
  - **证据**：`verify.sh` 首次在 `.test-output/.../seed-input.md` 中命中 `prompt body` / `suggested bullet` 描述文字，而非 runtime evidence 泄漏。
  - **影响**：场景脚本需要区分 runner log / runtime evidence 与 seed/expected prose，否则隐私 gate 会产生 false positive。
- frontend-resume-workshop handoff 文档仍停留在等待 backend 的口径。
  - **证据**：frontend spec 仍写 “backend-resume 真实落地后” 和 002/003 未创建，Phase 9 需要补 1.1 history/spec，明确 backend-resume/002 已 ready。
  - **影响**：下游 owner 启动 002/003 时可能继续按 mock-only 或 pending backend 判断。

## 3 根因归类

- 新 operation/error code 的 codegen + conventions baseline 没有在 Phase 1 收口时形成单独 final gate。
  - **类别**：spec-plan
- 场景脚本只写了测试名意图，没有先证明该 regex 匹配到真实测试。
  - **类别**：spec-plan
- privacy grep 作用域过宽，把说明文档当作 runtime 证据扫描。
  - **类别**：skill
- Cross-plan handoff 需要由完成方主动把 “generated/fixture ready” 与 “real backend ready” 分层写入下游 owner spec。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 后续 backend API plan 中，新增 operation 或 error code 的阶段应固定包含 `cd backend && go test ./cmd/codegen/openapi ./internal/shared/types -count=1` 或等价 baseline parity gate。
  - **落点**：spec-plan
  - **优先级**：high
- 场景创建模板应要求 `verify.sh` 检查具体测试名，并显式拒绝 `[no tests to run]`；新增 privacy 场景前先跑一次 regex no-op 检查。
  - **落点**：skill
  - **优先级**：high
- privacy grep 应默认扫描 runner log / runtime output，而不是 seed / expected prose；若必须扫描整个 `.test-output`，应使用 sentinel value 而非通用字段描述词。
  - **落点**：skill
  - **优先级**：medium
- Cross-plan handoff note 应写入 consumer owner spec/history，而不只写在当前 owner checklist。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：创建并审查 `frontend-resume-workshop/002-create-flow-and-onboarding`，以 backend-resume/002 real handler ready 为前提，同时先核对 backend-upload real handoff。
- 次优先级：把 P0.080 中沉淀的 privacy/no-op 脚本模式回补到 `/scenario-create` 示例或 README。
- 可延后：将 backend API cross-owner baseline gate 抽成共享 checklist snippet，先在 backend-resume/003 或下一轮 backend owner plan 中复用验证。
