# Backend Profile 001 L2 Remediation 交付复盘报告

> **日期**: 2026-05-21
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：历史 `backend-profile/001-candidate-profile-and-experience-cards` 的 L2 code review `--fix` 闭环（实体已随 product-scope/001 删除；当前删除证据由 [Core Loop Module Pruning](../spec/product-scope/plans/001-core-loop-module-pruning/plan.md) 承接），覆盖场景编号证据漂移、legacy negative gate、隐私删除失败恢复、event lint 和 plan v1.2 收口。
- 成功证据：
  - P0.091 / P0.092 / P0.093 场景全部完成 setup -> trigger -> verify -> cleanup 并通过。
  - 隐私删除新增 failure rollback integration gate：`DATABASE_URL=... go test -tags=integration ./internal/profile/store/... -count=1` PASS。
  - 完整后端 gate：`go test ./...` PASS；`DATABASE_URL=... go test ./cmd/api -run TestProfileHTTPScenario -count=1` PASS。
  - 前端与 lint gate：`pnpm --filter @easyinterview/frontend test src/api/devMockClient.test.ts` PASS；`pnpm --filter @easyinterview/frontend typecheck` PASS；`PYTHONPATH=scripts/lint python3 -m unittest scripts.lint.lint_events_test` PASS。
  - 契约 / 文档 gate：`make lint-openapi`、`make validate-fixtures`、`make openapi-diff`、`make codegen-check`、`make docs-check`、`validate_context.py`、`sync-doc-index --check`、`git diff --check` 全 PASS。
  - 负向搜索：旧 P0.081-083 / `p0-08*` 在当前 profile 场景与 HTTP scenario 测试中 0 命中；`mistake|growth|drill|experiences|star` 在 `backend/internal/profile/` 0 命中。
  - Bug 记录：[BUG-0081](../bugs/BUG-0081.md)。

## 2 会话中的主要阻点/痛点

- **场景 ID 与 output path 漂移在 completed 状态后仍存在**
  - **证据**：P0.091 / P0.092 / P0.093 目录已经重命名，但 README、data、setup、trigger、verify、cleanup 和 `profile_http_scenario_test.go` 注释仍残留 P0.081 / P0.082 / P0.083 或旧 `.test-output` 路径。
  - **影响**：BDD 资产可执行但证据指向旧编号，容易让后续 owner 误读场景归属，也会削弱 checklist 的完成可信度。

- **隐私删除 failure path 没有执行级证据**
  - **证据**：原 store 删除路径没有显式事务 wrapper，audit tombstone writer 只支持 success metadata；P0.093 verify 只检查成功删除与 PII redaction。
  - **影响**：profile 删除失败时可能出现 cards 已删、profile 未删且无 failure audit 的不可追踪状态，不符合 privacy lifecycle 的恢复路径要求。

- **复合 codegen gate 曾被遗留解释掩盖**
  - **证据**：`make codegen-check` 同时暴露 profile `resume_parse` source type 与 frontend `debrief_generate` naked literal；前者需要窄 allowlist，后者需要真实代码修复。
  - **影响**：如果继续把 target failure 解释成 pre-existing，完成态 plan 会缺少可重复重跑的完整 gate。

## 3 根因归类

- 场景 ID / output path 漂移
  - **类别**：spec-plan + test README。计划重命名场景编号时，没有把旧编号和旧产物路径零残留搜索写入 Phase 6 gate。

- 隐私删除 failure path 漏测
  - **类别**：spec-plan。P0.093 原 gate 对 success path 足够，但没有把 rollback 与 failure audit 纳入必须执行的 integration assertion。

- `make codegen-check` 复合失败处理不足
  - **类别**：skill / plan。L2 review 必须把复合 gate 分解到具体错误并修掉本轮可修项；仅标注 pre-existing 会让 target 失去重跑价值。

## 4 对流程资产的改进建议

- 在 backend-profile plan 的隐私删除类条目中保留 Phase 6.3 作为模板参考：privacy lifecycle gate 必须同时列出 success audit、failure audit、rollback 和 PII redaction。
  - **落点**：spec-plan
  - **优先级**：high

- 后续 `/plan-code-review --fix` 对 completed BDD plan 做 close-out 时，固定执行“旧 scenario ID / 旧 output path / retired product term”负向搜索，并把命中视作 blocking finding。
  - **落点**：plan-code-review skill
  - **优先级**：high

- 对 `make codegen-check` 这类复合 target，报告中应写清每个失败点的处置：true positive 直接修复、domain false positive 加窄 allowlist 与测试、确认为外部遗留才进入 follow-up。
  - **落点**：plan-code-review skill + docs/development gate guidance
  - **优先级**：medium

## 5 建议优先级与后续动作

- 推荐下一步：继续对 `backend-jobs-recommendations/001-jd-match-real-backend-baseline` 做 `/plan-review --fix` 或 `/plan-code-review` 前置校对，重点复用本次经验，先锁定 scenario ID reserve、operation matrix、privacy/failure path 和复合 gate 的可重跑性。
- 备选路径：若优先收敛 backend-profile，可先对 [BUG-0081](../bugs/BUG-0081.md) 提到的 privacy lifecycle gate 规则抽象成 `/plan-code-review` skill 改进项，再进入下一个 backend subject。
