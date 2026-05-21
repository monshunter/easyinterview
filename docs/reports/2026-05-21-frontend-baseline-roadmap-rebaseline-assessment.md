# Frontend Baseline and Roadmap Rebaseline 交付复盘报告

> **日期**: 2026-05-21
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖两个范围：修复 2026-05-18 work journal 记录的两个 pre-existing frontend Vitest 失败；同步 roadmap/spec/plan 生命周期状态。
- 前端失败修复证据：
  - `pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx --run`：5 tests passed。
  - `pnpm --filter @easyinterview/frontend test src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx --run`：4 tests passed。
  - `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/scripts/setup.sh -> trigger.sh -> verify.sh -> cleanup.sh`：退出 0。
  - `pnpm --filter @easyinterview/frontend test src/api/devMockClient.test.ts --run`：9 tests passed。
  - `python3 scripts/lint/validate_fixtures.py --repo-root .`：59 fixtures OK。
  - `pnpm --filter @easyinterview/frontend test`：210 test files / 1294 tests passed。
- 文档 rebaseline 证据：
  - `validate_context.py --context docs/spec/engineering-roadmap/plans/001-decompose-subspecs/context.yaml --target docs` 通过。
  - `sync-doc-index.py --check`：0 drift。
  - plan lifecycle checkbox scan：OK。
  - `make docs-check`：Header / INDEX / link checks 通过。
  - `git diff --check`：通过。

## 2 会话中的主要阻点/痛点

- `E2E.P0.037` 仍断言 plan 001 阶段的旧 `ComingSoonTab`，但 plan 003 已把 TARGETED rewrites tab 替换为 `ResumeRewritesTab`。
  - **证据**：focused Vitest red phase 找不到 `resume-detail-tab-content-coming-soon-rewrites`，而当前组件测试已断言 `resume-rewrites-tab`。
  - **影响**：全量前端测试被历史 BDD 场景卡住，掩盖当前 UI 真理源已经前进的事实。
- `practiceVoiceTurn` 仍断言旧 `fixture-audio://` mock-only audio ref。
  - **证据**：focused Vitest red phase 收到 `data:audio/wav;base64,...`，而测试期望 `fixture-audio://practice-voice/default/chunk-001`。
  - **影响**：BUG-0072 已修正 fixture 语义后，旧测试消费者没有随契约更新。
- roadmap §5.2 与多个 plan Header 仍停留在旧生命周期状态。
  - **证据**：`docs/spec/INDEX.md` 已显示多个 P0 subject 存在，但 roadmap 仍写“未创建”；5 个 plan 的 checklist 全勾选但 plan/checklist Header 仍为 `active`。
  - **影响**：下一步规划会误判 owner 是否已创建，且 active/completed 计划边界不清晰。

## 3 根因归类

- stale scenario consumer 没有随后续 owner plan 替换 UI 内容同步更新。
  - **类别**: spec-plan。
- fixture 语义修复后缺少反向 stale consumer 搜索。
  - **类别**: spec-plan。
- roadmap 状态表依赖人工维护，未定期与 `docs/spec/INDEX.md` 和 checklist 完成度做交叉审计。
  - **类别**: spec-plan。

## 4 对流程资产的改进建议

- 在后续 plan 完成“替换旧占位/旧组件”时，把被替换 owner 的 BDD scenario test 纳入 regression gate。
  - **落点**: spec-plan。
  - **优先级**: medium。
- 对 fixture contract bugfix，增加 scoped reverse-grep：旧 scheme、旧 testid、旧 fixture sentinel 必须在 frontend tests / scenario docs 中零命中或被显式豁免。
  - **落点**: spec-plan。
  - **优先级**: medium。
- 为 roadmap rebaseline 增加 lightweight audit 命令或 checklist 项：比较 `docs/spec/INDEX.md`、workstream table、all-checked plan Header 状态。
  - **落点**: spec-plan。
  - **优先级**: low。

## 5 建议优先级与后续动作

- 最高价值：下一轮进入新后端 workstream 前，先以当前 `engineering-roadmap` v3.16 为准选择 owner，避免从旧“未创建”口径派生新计划。
- 可延后：把 lifecycle checkbox scan 固化为脚本或 docs gate；当前已用人工扫描和 `sync-doc-index` 完成本次重基线。
