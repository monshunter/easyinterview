# Practice Real Interview Session 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

**关联计划**: [frontend-workspace-and-practice/002-practice-text-event-loop](../spec/frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md), [practice-voice-mvp/001-cascaded-stt-llm-tts](../spec/practice-voice-mvp/plans/001-cascaded-stt-llm-tts/plan.md)

## 1 复盘范围与成功证据

- 交付范围：按真实面试会话主路径收敛 text 与 phone 两种模式，删除右侧边栏、跳过、语音转文字 fallback、严格模式顶栏锁、独立 voice 分析面板与手动录音/提交回合控件；保留提示为用户在会话中可选的辅助动作；phone 模式默认不显示文字，只在用户点击字幕后展示与 text 共用的会话内容。
- 契约范围：删除 `turn_skipped` 事件面、更新 OpenAPI/generated client/backend practice event contract、迁移 enum source、fixture、scenario scripts、UI design contract 和 owner plan/checklist。
- 真实环境闭环证据：
  - `./test/scenarios/env-redeploy.sh all` 与后续 `./test/scenarios/env-verify.sh` PASS，backend/frontend 均在本地真实 host-run 环境中重启验证。
  - 通过真实 auth、resume API、target import API、practice plan API、practice session API 创建会话；浏览器中提交 text 回答，API 返回 `appendSessionEvent=200`，结束会话返回 `completePracticeSession=202`，报告页稳定渲染。
  - 截图证据已落在 `.test-output/real-practice-session/practice-text-real-closed-loop.png`、`.test-output/real-practice-session/practice-phone-real-default.png`、`.test-output/real-practice-session/practice-phone-real-captions.png`、`.test-output/real-practice-session/practice-generating-real-stable.png`。
- 自动化证据：
  - `node --test ui-design/ui-design-contract.test.mjs` PASS。
  - `corepack pnpm --filter @easyinterview/frontend test src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx src/app/screens/practice/PracticeScreen.test.tsx` PASS。
  - `test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/{setup,trigger,verify}.sh` PASS。
  - `test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/{setup,trigger,verify}.sh` PASS。
  - `test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/{setup,trigger,verify}.sh` PASS。
  - `test/scenarios/e2e/p0-047-practice-text-loop-privacy-and-completion/scripts/{setup,trigger,verify}.sh` PASS。
  - `make validate-fixtures`, `make lint-openapi`, `python3 scripts/lint/migrations_lint.py --repo-root .`, `make lint-backend-practice-non-current` PASS。
  - `corepack pnpm --filter @easyinterview/frontend exec tsc --noEmit` PASS。
  - `validate_context.py --context docs/spec/frontend-workspace-and-practice/plans/002-practice-text-event-loop/context.yaml --target frontend`, `sync-doc-index --check`, `make docs-check`, `git diff --check` PASS。
- Review remediation 追加证据：
  - P0.088/P0.090 route 场景已迁移到 `practice-phone-waveform`，并通过场景包装脚本 `setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh` 串行验收；trigger 日志保留在 `.test-output/e2e/p0-088-url-addressable-routing-canonical/trigger.log` 与 `.test-output/e2e/p0-090-url-routing-hash-non-current-negative/trigger.log`。
  - Report context strip 已把 current `phone` 与 legacy `voice` modality 都归一为 Phone 文案；`ReportContextStrip.test.tsx` 与 `ui-design/ui-design-contract.test.mjs` 增加回归断言并通过。
  - `backend-practice` active spec 与 002/003 plan family 已从旧 strict-hint conflict、`turn_skipped` / `skipped` 正向合同收敛到当前 optional hint、四种 event kind、四值 turn status；相关 `validate_context.py`、`sync-doc-index --check`、`make docs-check` 均通过。
  - Phone pause review remediation 已补充 `practiceVoiceTurn.test.tsx` 回归断言：暂停会丢弃当前 microphone capture、释放 media track、不提交 stale voice turn，恢复后重新自动录音；`PracticePhoneSurface` 与 `usePracticeVoiceTurn` 已同步实现。
  - OpenAPI review remediation 已把 `openapi-v1-contract` spec/history 升至 1.38，并将 `openapi/baseline/openapi-v1.0.0.yaml` 与当前 `openapi/openapi.yaml` 对齐；`openapi_inventory_test.py` 现在覆盖 owner spec + baseline 对 `skipped` / `turn_skipped` 的零残留。

## 2 会话中的主要阻点/痛点

- Phone 默认层一开始仍残留题干文字。
  - **证据**：真实截图准备阶段发现 phone layer 仍有 `practice-phone-question` / `QuestionHeader`；随后补充 UI contract 与 frontend test，断言 phone 默认层不显示题干，字幕开启后才显示 transcript。
  - **影响**：如果只跑组件测试而不做真实截图验收，会把“电话模式无文字”的核心体验误判为已完成。

- 旧右侧边栏 test id 在 P0 场景中残留。
  - **证据**：P0.044 / P0.047 首轮场景失败，因为测试仍寻找 `practice-rightpanel-cta-finish`；实际产品已迁移为中心会话层底部的 `practice-finish-cta`。
  - **影响**：产品删面后，场景资产如果不一起迁移，会把正确实现误报为失败，也会掩盖真正的端到端回归。

- 真实环境会话数据准备仍缺少一条稳定的 repo-tracked helper。
  - **证据**：`manual_form` target import 返回 `TARGET_IMPORT_FAILED`，`manual_text` import 可 queued 但本地 runner 没有 target.import consumer；本次用真实 API 创建 practice plan/session 仍闭环成功，但准备链路有额外人工组合成本。
  - **影响**：后续做截图验收时容易把数据准备问题误认为 practice 会话问题，且每轮都需要重新拼接 auth、resume、target、plan、session 请求。

- Review 阶段暴露出旧 consumer 与 active owner docs 不一致。
  - **证据**：P0.088/P0.090 仍断言 `practice-voice-waveform`，report handoff 的 `modality=phone` 仍显示 Text，`backend-practice` owner docs 仍描述 strict hint conflict、`turn_skipped` / `skipped` 旧合同。
  - **影响**：如果只接受 feature branch 的主路径截图与局部测试，旧 route consumer、report handoff 与 active backend spec 会在 review 后继续漂移，后续实现者会从错误 owner truth source 继续开发。

## 3 根因归类

- Phone 模式文字显示边界没有在最初 gate 中显式落成 negative assertion。
  - **类别**：spec-plan

- 删除 UI surface 时，runtime 搜索覆盖了组件与文案，但早期没有把 scenario test id 全量迁移作为独立 checklist gate。
  - **类别**：spec-plan

- 本地真实验收依赖 auth + resume + target + practice 多 API 组合，但场景工具层还没有“最小 practice session seed + screenshot artifact”入口。
  - **类别**：README / scenario tooling

- Review 修复前没有把 legacy route consumer、report handoff modality、active backend owner docs 纳入同一个 post-pass reconcile gate。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 在 practice owner plan 中保留“phone 默认层不显示题干/转写/聊天文字，字幕开启后才显示共用 transcript”的 contract、unit、real-env screenshot gate。
  - **落点**：spec-plan
  - **优先级**：high

- 对任何删 UI surface 的 plan，增加 changed-testid sweep：同时覆盖 `frontend/src`、`ui-design/src`、`test/scenarios`、scenario data、docs 中的旧 test id / component name / user-facing label。
  - **落点**：spec-plan
  - **优先级**：high

- 新增一个 repo-tracked real practice smoke helper：用真实 auth/API 创建 resume、ready target 或可用 fallback target、practice plan、session，并统一输出 browser screenshot artifact 路径。
  - **落点**：test/scenarios/README.md / scenario tooling
  - **优先级**：medium

- 在真实面试会话 owner gate 中追加 post-pass reconcile：legacy route hash/canonical consumer、report handoff modality、backend active spec/plan family 必须与当前 phone/text 合同一起复查。
  - **落点**：spec-plan
  - **优先级**：high

## 5 建议优先级与后续动作

- high：本轮变更进入 `/work-journal` 前，保留截图 artifact 与真实 API 结果 JSON 作为提交说明的验收证据，避免仅以组件测试代替真实闭环。
- high：提交前把本轮 review remediation 的 P0.088/P0.090 场景包装脚本、report modality 单测、backend-practice context/doc gate 一起作为验收证据写入工作日志。
- medium：单独开一个小 bug/plan 处理本地 real practice seed helper 与 `manual_form` / `manual_text` import runner 间隙；它不阻塞本轮 practice 会话 UI 与真实闭环，但会降低后续验收成本。
