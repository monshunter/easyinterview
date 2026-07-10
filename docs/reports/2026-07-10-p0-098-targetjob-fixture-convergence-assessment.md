# P0.098 TargetJob Fixture Convergence 交付复盘报告

> **日期**: 2026-07-10
> **审查人**: Codex

**关联计划**: [001 Full Funnel Happy Journey](../spec/e2e-scenarios-p0/plans/001-full-funnel-happy-journey/plan.md)
**关联 Bug**: [BUG-0157](../bugs/BUG-0157.md)

## 1 复盘范围与成功证据

- 交付范围包括 P0.098 两个 harness 的 JSON request helper 去重，以及 live 场景验证中发现的 TargetJob 死 fixture 与陈旧持久化断言收敛。
- P0.098 setup/trigger/verify/cleanup 全部通过，三个目标测试实际执行；runner 日志包含 `resume_parse`、`target_import`、`report_generate` succeeded marker。
- `go test ./cmd/api -count=1`、backend `go test ./... -count=1`、`go vet ./...`、`staticcheck ./...` 均通过。
- Scoped `dupl -t 100` 为 0，死 fixture zero-reference、owner contexts、文档索引/链接/diff 与 pruning surface 通过，`real_residuals=0`。

## 2 会话中的主要阻点/痛点

- P0.098 首次 live RED 只显示 requirement 数量从 1 变成 2，表面上既可能是重复写入，也可能是 fixture 语义变化。
  - **证据**：`target_import` runner 已成功，失败发生在 `assertTargetImportPersisted`；直接读取 runtime builder 后才确认 TargetJob 使用另一套 deterministic client。
  - **影响**：若直接把期望值改成 2，会保留不可达 fixture，并继续用总数掩盖 requirement kinds。
- 场景文件保留了一个看似处理 `target.import.parse` 的分支，但运行时构造从未向 TargetJob runtime 注入该 client。
  - **证据**：`buildTargetJobRuntime` 在 test env 内部包装 `targetjob.NewDeterministicParseAIClient`，而场景 client 只注入 resume/practice/report runtime。
  - **影响**：代码阅读产生错误数据来源线索，也让 fixture 变化后断言漂移不易定位。
- 常规 backend 回归不能替代 live journey。
  - **证据**：没有 `DATABASE_URL` 时 focused/full Go test 会跳过 DB journey；显式四段场景生命周期才复现并验证修复。
  - **影响**：仅依赖 `go test ./...` 会把 Ready 场景的数据库断言漂移留到后续环境验证。

## 3 根因归类

- Fixture owner 与断言语义没有在原 full-funnel owner 中明确绑定。
  - **类别**：spec/plan
- 被 runtime builder 遮蔽的平行 fixture 与总数断言属于同一测试设计债务。
  - **类别**：spec/plan
- 初始诊断需要从场景 client 继续反查 runtime builder；这是本次具体代码形态导致的调查步骤，不构成新的通用 skill 缺陷。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- Full-funnel owner 应明确写出各 AI feature 的实际 fixture owner，并在 fixture owner 变化时要求 source-level zero-reference 与 live lifecycle 回归。
  - **落点**：`e2e-scenarios-p0/001-full-funnel-happy-journey` spec-plan
  - **优先级**：high
- 持久化场景对 enum/kind 集合应验证 expected distribution 和 unexpected zero，而不是只检查总数。
  - **落点**：相关场景 plan/checklist
  - **优先级**：high
- 可将“runtime-owned fixture 遮蔽场景级 fixture”作为候选 Bug pattern；是否加入 `docs/bugs/PATTERNS.md` 应在后续获得用户确认后处理。
  - **落点**：Bug pattern library
  - **优先级**：low

## 5 建议优先级与后续动作

- 本轮已在原 owner Phase 5 落实最高优先级项：删除平行 fixture、改为 kind 级断言并重跑 live lifecycle，无需另建实施计划。
- 下一轮技术债扫描优先查找其他 integration harness 中“runtime builder 自带 deterministic client，同时场景文件又定义同 feature fixture”的重复 ownership。
- `PATTERNS.md` 候选可延后，除非后续扫描发现第二个同类实例。
