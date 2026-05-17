# Practice Voice MVP 001 Cascaded STT LLM TTS 交付复盘报告

> **日期**: 2026-05-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：`practice-voice-mvp/001-cascaded-stt-llm-tts` 全计划，从 OpenAPI / fixture / generated client，到后端 STT -> chat -> TTS orchestration、frontend voice controller、TTS playback / barge-in、`E2E.P0.007`-`E2E.P0.009` 场景与 plan lifecycle close-out。
- 成功证据：
  - `make codegen-check`
  - `make validate-fixtures`
  - `pnpm --filter @easyinterview/frontend test src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx src/api/devMockClient.test.ts src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx src/app/App.test.tsx`
  - `cd backend && go test ./internal/practice ./internal/api/practice ./internal/store/practice ./cmd/api ./internal/ai/aiclient ./internal/ai/aiclient/profile -count=1`
  - `python3 scripts/lint/ai_profile_coverage.py --repo-root .`
  - `test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/verify.sh`
  - `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/verify.sh`
  - `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/verify.sh`
  - `make docs-check`
  - `git diff --check`
- Lifecycle evidence: `docs/spec/practice-voice-mvp/plans/001-cascaded-stt-llm-tts/{plan,checklist,bdd-plan,bdd-checklist}.md` 已收口为 `completed`，`docs/spec/practice-voice-mvp/plans/INDEX.md` 已移入 Completed。

## 2 会话中的主要阻点/痛点

- `createPracticeVoiceTurn` runtime route 缺失直到 BDD 场景创建前才暴露。
  - **证据**：OpenAPI、fixtures、generated types、domain service、SQL store 均已存在，但 `backend/internal/api/practice.Handler` 和 `backend/cmd/api/main.go` 缺少 adapter / mux route；新增 `TestE2EP0007PracticeVoiceTurnHTTPRoute` 后修复前失败为 404。已记录 [BUG-0067](../bugs/BUG-0067.md)。
  - **影响**：如果只依赖 domain/frontend mock tests，完整语音主路径会在真实 HTTP runtime 下不可用。
- BDD wrapper 资产在 Phase 5 才落地，导致最后阶段同时承担“创建场景”和“发现 runtime wiring drift”。
  - **证据**：`E2E.P0.007` 创建前需要先补 handler / route，否则场景无法代表真实 backend endpoint。
  - **影响**：Phase 5 从纯验证扩展为补 runtime wiring，增加了收尾返工。
- Privacy grep 首次执行时扫入 `_test.go` fixture，产生了测试用 privacy token 命中。
  - **证据**：第一次 grep 命中 `voice_turn_service_test.go` 中刻意构造的 `raw-audio-privacy-token` / `tts-audio-privacy-token`；随后用 runtime scope 排除 `_test.go` / `*.test.*` / `__tests__` 后零命中。
  - **影响**：不是产品缺陷，但说明 privacy grep gate 必须明确 runtime scope，否则容易把 negative fixture 当成泄漏。

## 3 根因归类

- HTTP route 缺失：`spec-plan`。Operation Matrix 写明了 backend handler 责任，但 Phase 2 的完成证据集中在 service/store，没有强制 `cmd/api` route-level test。
- 场景资产后置：`spec-plan`。BDD Plan 已列出三条场景，但 checklist 允许所有 scenario wrapper 在 Phase 5 才创建，未提供早期 smoke wrapper 捕捉 runtime wiring。
- Privacy grep 误命中测试 fixture：`no repo change needed` for this session。最终 P0.009 `verify.sh` 已把 runtime grep scope 固化到场景 wrapper。

## 4 对流程资产的改进建议

- 在后续涉及新 side-effect API operation 的 plan checklist 中，把 `cmd/api` route-level test 放在 backend implementation phase，而不是只在最终 BDD phase 发现。
  - **落点**：spec-plan
  - **优先级**：high
- 对跨前后端的 BDD plan，优先在 Phase 1/2 创建最小 scenario wrapper skeleton，哪怕先只跑 focused route smoke；后续 phase 再扩展为完整用户场景。
  - **落点**：spec-plan
  - **优先级**：medium
- Privacy grep 应在 checklist 或 scenario `verify.sh` 中标明 runtime scope 与 fixture scope，特别是允许测试文件保存 redline tokens 时。
  - **落点**：spec-plan / scenario README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮实现新 API operation 时，先在对应 backend phase 增加 `cmd/api` HTTP route test，再宣称 handler/runtime wiring 完成。
- 中优先级：为未来 voice 迭代保留现有 `E2E.P0.007`-`E2E.P0.009` wrapper 作为 regression suite，避免只跑 frontend mock 或 domain tests。
- 可延后：将 runtime-scope privacy grep 抽成共享 scenario helper，减少每个场景重复写 glob 排除规则。
