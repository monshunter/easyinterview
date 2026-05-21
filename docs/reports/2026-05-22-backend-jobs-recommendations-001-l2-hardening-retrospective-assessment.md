# Backend Jobs Recommendations 001 L2 Hardening 交付复盘报告

> **日期**: 2026-05-22
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-jobs-recommendations/001-jd-match-real-backend-baseline` 的 post-reopen `/plan-code-review --fix` 收口，覆盖真实 A3/F3 runtime adapter、F3 prompt `jobMatchId` contract、search completed outbox、AI_OUTPUT_INVALID、privacy delete transaction rollback、BDD wrapper false-green hardening、Bug/报告/计划证据同步。
- 成功证据：`cd backend && go test ./...` PASS；`cd backend && go test ./internal/ai/registry -count=1` PASS；`cd backend && go test ./internal/jdmatch/...` PASS；focused live `cmd/api` JD-Match matrix PASS；E2E.P0.094-P0.097 `setup.sh -> trigger.sh -> verify.sh` 全部 PASS；`sync-doc-index --check`、`make docs-check`、`git diff --check` PASS。
- 关联记录：[BUG-0082](../bugs/BUG-0082.md) 与 [BUG-0083](../bugs/BUG-0083.md) 已建档，L2 assessment report 已从 No-Go 更新为 Fixed / Go。

## 2 会话中的主要阻点/痛点

- 已有 completed 状态会掩盖 production composition 缺口。
  - **证据**：`main.go` 仍以 `stubJDMatchAI{}` / `generatorAIAdapter{}` 构建 JD-Match runtime，直到 L2 反查 production startup path 才发现。
  - **影响**：如果只按 checklist PASS 收口，真实部署会绕过 A3 AIClient / F3 registry。
- Prompt truth source 与 runtime parser 没有同测。
  - **证据**：F3 prompt 未要求 `jobMatchId`，但 search handler 只能用 returned IDs join `jd_match_recommendations`。
  - **影响**：真实 provider 即使返回合法推荐字段，也可能无法被 search endpoint 投影。
- Cross-owner event / privacy rollback 只在局部层有证据。
  - **证据**：search path 缺 `jd_match.search.completed` outbox；privacy delete service 有顺序测试，但 cmd/api runtime 未包事务。
  - **影响**：B3/B4 contract 在 runtime 层不闭环。

## 3 根因归类

- Production composition 反查不足。
  - **类别**：spec-plan
  - Plan gate 曾强调 cmd/api wiring，但没有明确要求“main startup path 不得传 stub adapter”。
- Prompt-output 与 runtime join key 缺少同一 preflight。
  - **类别**：spec-plan
  - F3 gate 主要验证 registry load / profile coverage，缺少业务关键输出键与 parser 的配套断言。
- Scenario wrapper 从 false-green 修复后仍需要持续防回退。
  - **类别**：README
  - 场景框架已有 README，但 wrapper authoring checklist 可以更明确要求 package-level success marker 与 fail marker rejection。

## 4 对流程资产的改进建议

- 在 backend-jobs-recommendations 后续 plan/checklist 模板中追加 production composition gate。
  - **落点**：spec-plan
  - **优先级**：high
- 为 F3 cross-owner additive 增加“prompt required output marker + runtime parser” preflight 模式。
  - **落点**：spec-plan
  - **优先级**：high
- 在 scenario README 或 wrapper template 中固化 verify 最小断言：目标测试名、package-level `ok`、拒绝 `--- FAIL` / package `FAIL` / skip / no-test。
  - **落点**：README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 高优先级：下一次 JD-Match 或 AI feature plan 启动前，把 production composition gate 与 prompt/parser preflight 写入对应 plan gate。
- 中优先级：抽取本次 P0.094-P0.097 wrapper hardening 规则到 scenario framework README，减少每个场景重复补救。
- 可延后：将本次 A3/F3 adapter 测试模式整理为 shared helper，等第二个业务域需要类似 adapter 时再做，避免过早抽象。
