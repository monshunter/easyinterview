# Backend Resume 002 Tailor Provenance Remediation 交付复盘报告

> **日期**: 2026-05-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-resume/002-tailor-runs-and-save-v1` L2 review follow-up，修复 `ResumeTailorRun` ready response 在 DB roundtrip 后 provenance required 字段不完整的问题。
- 修复内容：新增 `resume_tailor_runs.language / feature_flag / data_source_version` migration；更新 tailor run ready/success 写库、读取、suggestion provenance join；补充 Phase 7 remediation checklist 与 BUG-0073。
- 成功证据：
  - RED: `DATABASE_URL=*** go test ./internal/resume/store -tags=integration -run TestCompleteTailorRunSuccessWritesSuggestionsAndReadyOnlyOutbox -count=1 -v` 失败于 `tailor run provenance after DB roundtrip`。
  - GREEN: 同命令 PASS。
  - `DATABASE_URL=*** go test ./internal/resume/store -tags=integration -run 'TestCompleteTailorRunSuccessWritesSuggestionsAndReadyOnlyOutbox|TestResumeSuggestionDecisionCASIsolationAndProfileStability|TestResumeTailorRunStore' -count=1 -v` PASS。
  - `go test ./internal/resume/... -count=1` PASS。
  - `go test ./internal/resume/handler -run 'TestResumeTailorFixtureParity|TestGetResumeTailorRun|TestResumeSuggestionDecisionFixtureParity' -count=1 -v` PASS。
  - `go test ./cmd/api -run 'TestResumeTailorDrainerHTTPScenario|TestResumeTailorDrainerFailureScenario|TestResumeTailorEndpointsHTTPScenario|TestResumeSuggestionAcceptRejectHTTPScenario' -count=1 -v` PASS。
  - `python3 scripts/lint/migrations_lint.py --repo-root .` PASS；`DATABASE_URL=*** make migrate-check` PASS。
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS；`make docs-check` PASS；`git diff --check` PASS。

## 2 会话中的主要阻点/痛点

- Existing gates were green but not decisive.
  - **证据**：service fake store、handler fixture parity、cmd/api drainer fake store 均通过；只有新增 `CompleteTailorRunSuccess -> GetTailorRun` live store integration 才暴露空 `language / featureFlag / dataSourceVersion`。
  - **影响**：如果只看 fake/fixture gates，会误判 ready response 已满足 OpenAPI required provenance。
- Persisted resource 与 AI job output 被混为一个完成证据。
  - **证据**：`tailorProvenance` 已产生完整字段，但 `resume_tailor_runs` schema 只保存四个 typed columns；GET path 从 DB 读回时丢字段。
  - **影响**：job 层正确不等于 API resource 在持久化后正确。
- Completed plan 需要原地打开一个 remediation anchor。
  - **证据**：原 Phase 7 已 completed，修复时新增 7.12 remediation item、plan/checklist 版本升至 1.1，再由 sync-doc-index 恢复投影。
  - **影响**：没有原地 checklist anchor 时，修复证据容易只停留在当前会话，后续 L2 难以追踪。

## 3 根因归类

- Required persisted fields 缺少 DB roundtrip gate。
  - **类别**：spec-plan
  - **说明**：Phase 7 gate 证明 AI job 写 success payload、outbox 和 task run，但没有明确要求 ready API resource 经 DB 重新读取后仍满足 `GenerationProvenance` required 字段。
- Review workflow 对 fake store 证据的风险提示还不够显式。
  - **类别**：skill
  - **说明**：`plan-code-review` 已要求 artifact-level evidence，但可以进一步强调“schema required 字段必须有持久化 write-after-read gate，fake store 只能作为辅助证据”。

## 4 对流程资产的改进建议

- 在 `backend-resume/002` Phase 7 保留 7.12 remediation gate。
  - **落点**：spec-plan
  - **优先级**：high
  - **状态**：已完成，本次已在 plan/checklist 中固化。
- 为 L2 code review 增加 persisted-schema roundtrip 检查口径。
  - **落点**：skill
  - **优先级**：medium
  - **建议**：后续优化 `.agent-skills/plan-code-review/SKILL.md` 时，增加规则：当 OpenAPI required 字段由 DB-backed resource 返回时，必须寻找或补充 write-after-read store/integration gate；fake store/fixture parity 不单独构成 PASS。
- 将 BUG-0073 作为 schema/provenance 类复查入口。
  - **落点**：README / bug pattern
  - **优先级**：low
  - **建议**：若后续再出现类似 “job output 完整但 GET resource 丢字段” 的缺陷，可把 BUG-0073 提炼进 `docs/bugs/PATTERNS.md`。

## 5 建议优先级与后续动作

- 最高优先级：继续后续 backend-resume / frontend-resume-workshop 切真前，对所有 AI-backed ready resource 做 DB roundtrip 检查，尤其是 `GenerationProvenance` required 字段和 privacy payload allowlist。
- 次优先级：在下一次流程硬化时更新 `plan-code-review` skill，把 fake-store PASS 与 persisted-resource PASS 明确拆开。
- 可延后：等出现第二个同类 BUG 后，再把 BUG-0073 提炼为 `PATTERNS.md` 的独立模式，避免模式库过早膨胀。
