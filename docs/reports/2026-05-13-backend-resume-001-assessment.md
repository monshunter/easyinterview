# Backend Resume 001 交付复盘报告

> **日期**: 2026-05-13
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-resume/001-asset-register-parse-and-listing`，覆盖 `registerResume` / `getResume` / `listResumes` handler、`resume_assets` store、`resume.parse` in-process drainer、B2 fixture parity、E2E.P0.034/P0.035 场景资产与 frontend workspace handoff。
- 成功证据：
  - `cd backend && go test ./...`
  - `cd backend && go test ./internal/resume/...`
  - `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResumeRegisterListHTTPScenario|TestResumeParseDrainerHTTPScenario' -count=1`
  - `test/scenarios/e2e/p0-034-resume-register-and-list/scripts/setup.sh|trigger.sh|verify.sh|cleanup.sh`
  - `test/scenarios/e2e/p0-035-resume-parse-async-job-lifecycle/scripts/setup.sh|trigger.sh|verify.sh|cleanup.sh`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`
  - `git diff --check`

## 2 会话中的主要阻点/痛点

- `getResume.not-found` fixture drift 到 Phase 5 才暴露。
  - **证据**：新增 `TestGetResumeFixtureParity` 前，handler 已有 404 行为测试，但 `openapi/fixtures/Resumes/getResume.json` 使用未声明的 `NOT_FOUND`；已记录为 [BUG-0051](../bugs/BUG-0051.md)。
  - **影响**：mock-first 对齐需要在收口阶段补测试与 fixture 修复，说明 per-operation error fixture parity 不够前置。
- BDD verify 脚本的隐私 grep 初版误判字段名。
  - **证据**：P0.034 `verify.sh` 首次失败，因为 `.test-output/.../expected-outcome.md` 中出现 schema 字段名 `guidedAnswers`，不是用户内容值。
  - **影响**：场景脚本需要区分允许出现的 contract 字段名与禁止泄漏的字段值，否则会产生 false positive。
- Cross-plan handoff 需要同时写明 generated contract 与真实 backend route 两层状态。
  - **证据**：workspace 001 先前已有 B2 generated `listResumes` handoff，本次补充 `backend-resume/001` 真实 `cmd/api` route、fixture parity 和 E2E.P0.034/P0.035 证据。
  - **影响**：下游 frontend owner 是否能从 disabled-list 切 active-list，取决于 generated client 与真实 handler 两层都就位。

## 3 根因归类

- `getResume` error fixture parity 缺口。
  - **类别**：spec-plan
  - **根因**：Phase 1-4 更强调 handler 行为和 register/list fixture parity，没有把 get/not-found error envelope parity 提前列成单独 gate。
- 隐私 grep false positive。
  - **类别**：skill
  - **根因**：场景脚本模板倾向用宽泛关键词，缺少“字段名允许、字段值禁止”的验证约束。
- Handoff 粒度不足。
  - **类别**：spec-plan
  - **根因**：workspace plan 的 handoff 先记录了 B2 generated unblock，但未区分真实 backend route availability，直到 backend-resume 完成后才补齐。

## 4 对流程资产的改进建议

- 在 backend API 实施类 plan 的 Phase 1/Phase 5 checklist 模板中，为每个 operation 明确 success 与 canonical error fixture parity。
  - **落点**：spec-plan
  - **优先级**：high
- 在 `/scenario-create` 或场景脚本示例中补一句：privacy grep 应匹配真实敏感样本值或 redaction sentinel，不应把允许的 schema field name 当成泄漏。
  - **落点**：skill
  - **优先级**：medium
- Cross-plan unblock note 应分两层描述：OpenAPI/generated/fixture ready 与 backend route/runtime ready。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：启动 `frontend-workspace-and-practice/001` 的原地 active-list 修订，把 Resume Picker 从 disabled-list 切到真实 `listResumes` 消费，并移除旧负向断言。
- 次优先级：在下一轮 backend-resume/002 前，把 `get/list/register` 之外的 Resume operation 也按 success + error fixture parity 预先列入 checklist。
- 可延后：将 scenario privacy grep 模板收敛到 skill 层，减少未来场景脚本的误报修正成本。
