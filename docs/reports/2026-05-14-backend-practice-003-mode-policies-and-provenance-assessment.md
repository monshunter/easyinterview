# Backend Practice 003 Mode Policies And Provenance 交付复盘报告

> **日期**: 2026-05-14
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-practice/plans/003-mode-policies-and-provenance`，覆盖 assisted/strict hint mode policy、`practice.turn.lightweight_observe` hint AI、`hint_generate` task-run provenance、graceful degrade、BDD 场景 048-051、计划生命周期收口。
- 代码证据：`cd backend && go test ./... -count=1` 通过；focused 包 `./internal/store/ai ./internal/practice ./internal/store/practice ./internal/ai/aiclient ./internal/ai/registry ./internal/migrations ./cmd/api` 通过。
- BDD 证据：`E2E.P0.048` / `049` / `050` / `051` 场景脚本按 setup → trigger → verify → cleanup 串行通过，trigger 使用 `go test -v` 并 grep 精确测试名，避免 no-op PASS。
- 工程 gate：`make codegen-check`、`make validate-fixtures`、`migrations/lint.sh`、`make lint-events`、`make codegen-events-check`、`python3 scripts/lint/conventions_drift.py --repo-root .`、`python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all`、`python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q`、`make docs-check`、`git diff --check` 均通过。
- 生命周期证据：003 `plan.md`、`checklist.md`、`test-plan.md`、`test-checklist.md`、`bdd-plan.md`、`bdd-checklist.md` 已切到 `completed`，`docs/spec/backend-practice/plans/INDEX.md` 与 `docs/spec/INDEX.md` 经 `sync-doc-index --check` 确认零 drift。

## 2 会话中的主要阻点/痛点

- 测试清单一度比真实 artifact 更乐观。
  - **证据**：`test-checklist.md` 勾选了 `TestApplyHintAIWritesAITaskRunOnF3Failure`、`TestObservabilityDecoratorWritesHintGenerateTaskRun` 等不存在的精确测试名；后续通过新增/改写测试和修订 `test-plan.md`、`test-checklist.md` 对齐到真实测试函数。
  - **影响**：如果直接按早期 checklist 收口，会形成文档 PASS 与代码事实不一致。
- 场景脚本最初只 grep package 输出。
  - **证据**：048-051 的 `verify.sh` 只检查 `github.com/monshunter/easyinterview/backend/cmd/api` 与非 `no tests to run`；后续改为 `go test -v -run '^ExactTest$'` 并 grep 精确测试名。
  - **影响**：较弱的 gate 容易漏掉正则跑偏或测试名未执行。
- 生产 runtime wiring 与 scenario harness 不一致。
  - **证据**：scenario harness 已把 Practice AI 包上 A3 observability decorator 并传入 `AITaskRuns` writer，但 `cmd/api::buildPracticeRoutes` 初始只传裸 AI client，F3/parse explicit failure 也没有生产 SQL writer。
  - **影响**：E2E harness 可能通过，但真实 API runtime 下 `hint_generate` task-run provenance 不能闭环。

## 3 根因归类

- 测试清单漂移根因：计划中的预期测试名没有在实现后做 artifact-level 反查。
  - **类别**：spec-plan / skill
- 场景脚本弱验证根因：场景模板允许 package-level PASS 作为 verify 信号，没有强制 `-v` 精确测试名证据。
  - **类别**：README / scenario skill
- 生产 wiring 缺口根因：计划同时要求 observed AI 与 explicit failed writer，但早期实现只在 scenario harness 中补齐，没有把同一能力接到 `cmd/api`。
  - **类别**：spec-plan / no repo change needed（本次已修复）

## 4 对流程资产的改进建议

- 在 `/tdd` 或 `/implement` 收口步骤中加入 named-artifact 反查：凡 checklist/test-plan 引用精确测试函数、脚本、场景 ID，必须用 `rg` 或 test listing 验证名称真实存在，并在不一致时先修正文档或补测试。
  - **落点**：skill
  - **优先级**：high
- 在 `test/scenarios/README.md` 或 `/scenario-run` 中固化 Go scenario 脚本规范：trigger 使用 `go test -v -run '^ExactTest$'`，verify grep 精确测试名并保留 `no tests to run` 负向断言。
  - **落点**：README / skill
  - **优先级**：medium
- 对涉及 observability decorator 的 backend plan，在 operation matrix 增加 "scenario harness wiring vs cmd/api production wiring" 双列，明确两者都要验证。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一步优先修订 `/tdd` 或 `/implement` 的收口规则，加入 checklist named-artifact 反查，直接降低未来虚假勾选风险。
- 然后把 Go scenario 精确测试名验证写入 `test/scenarios/README.md` 或 `/scenario-run`，避免后续场景脚本回到 package-level PASS。
- 003 本身已闭环；后续产品功能应进入 `backend-practice/004-derived-plans-debrief`，因为 003 明确不处理 retry / next_round / debrief source 派生。
