# Backend Jobs Recommendations 001 L2 Follow-up 交付复盘报告

> **日期**: 2026-05-22
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-jobs-recommendations/001-jd-match-real-backend-baseline` 第二轮 `/plan-code-review --fix` follow-up，覆盖 review 指出的 privacy runner 集成、agent scan generator context、JDMatch error envelope / retryable、以及本地 lock 文件误提交。
- 成功证据：focused privacy runner / jdmatch jobs / jdmatch handler tests PASS；`cd backend && go test ./internal/jdmatch/... -count=1` PASS；live `cmd/api` JDMatch matrix PASS；`cd backend && go test ./...` PASS；E2E.P0.097 `setup.sh -> trigger.sh -> verify.sh` PASS；`sync-doc-index --check`、`make docs-check`、`git diff --check` PASS。
- 关联记录：[BUG-0084](../bugs/BUG-0084.md) 已建档；plan / checklist / bdd-plan / bdd-checklist 更新至 v1.3 completed。

## 2 会话中的主要阻点/痛点

- service 已存在不代表 production runner 已接入。
  - **证据**：`DeleteJobMatchDataForUser` 已有 service/runtime helper，但 `privacyrunner.NewPrivacyDeleteHandler` 只拿到 upload deleter。
  - **影响**：`DELETE /api/v1/me` async privacy job 会保留 candidate_profile 与 JDMatch 5 表数据。
- AI prompt 字段名存在不代表真实 payload 有上下文。
  - **证据**：`RunRecommendationGeneratorInput` 有 `CandidateProfileJSON` / `JobsPoolJSON` 字段，但 `agent_scan` 传入空值。
  - **影响**：真实 `jd_match.recommendation` 无法基于候选画像与内部 jobs pool 推荐，也无法稳定保留 `jobMatchId` join key。
- Error contract 容易在局部 handler 中漂移。
  - **证据**：JDMatch `writeAPIError` 返回裸 `ApiError`，不同于其他 handler 的 `ApiErrorResponse` envelope。
  - **影响**：generated client 读取错误结构失败，`AI_OUTPUT_INVALID` 被错误标记为可重试。

## 3 根因归类

- Production caller 反查不够细。
  - **类别**：spec-plan
  - 现有 gate 证明了 service 与部分 runtime helper，但缺少“上游 runner/handler 是否实际注入”的反向断言。
- AI payload 只测输出，未测输入。
  - **类别**：spec-plan
  - `TestJDMatchAgentScanDrainerScenario` 原先只看 upsert/outbox 结果，没有捕获 adapter payload。
- Local tool state 缺少 ignore 规则。
  - **类别**：README / repo hygiene
  - `.claude/cache/` 已 ignore，但 scheduled lock 文件没有覆盖。

## 4 对流程资产的改进建议

- 在 completed plan 的 L2 review gate 中追加 production caller trace：service helper、runner handler、route builder、startup composition 必须串起来。
  - **落点**：plan-code-review skill / spec-plan
  - **优先级**：high
- 对 AI generator/drainer 增加 payload capture assertion，而不只检查输出结果。
  - **落点**：spec-plan
  - **优先级**：high
- 将 local agent runtime state 纳入 repo hygiene 检查。
  - **落点**：AGENTS.md 或 repo README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 高优先级：下一轮 `backend-async-runner/001-internal-job-outbox-runner` 实施前，把本次 pattern 5 的 caller trace 直接放入 runner migration gate，防止迁移时只移动 handler 注册而漏掉 domain hooks。
- 中优先级：把 AI payload capture 模式抽成 JDMatch / future AI feature 的测试 helper。
- 可延后：统一 `.claude/` / `.codex/` / `.gemini/` runtime lock 命名规则，等再次出现本地状态文件时再集中治理。
