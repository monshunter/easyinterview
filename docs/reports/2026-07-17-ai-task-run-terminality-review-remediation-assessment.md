# AI Task Run Terminality Review Remediation 交付复盘报告

> **日期**: 2026-07-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 修复 branch-vs-main review 发现的 `resume_tailor` task-run 与 async job 终态矛盾，并清理全量 staticcheck 的 S1016。
- 原地重开 A3 `001-aiclient-and-profile-bootstrap` Phase 17，新增 D-19/C-23、TDD checklist 与可执行替代 gate。
- RED 证明空 suggestions 与空白 bullet 会在业务失败后留下 success task run；GREEN 将约束前移到 output schema，并保持 decorator 为唯一 writer。
- `make test`、`make build`、`make lint`、全量 `go vet` / `staticcheck`、context、docs/index 与 prompt lint 全部通过。

## 2 会话中的主要阻点/痛点

- 单写者重构的局部目标正确，但历史 Phase 16 只验证“没有重复写入”，没有验证 writer 的完成时点与所有业务 decoder 的接受集合一致。
  - **证据**：`{"suggestions":[]}` 同时得到 task-run `success/ok` 与 async job `AI_OUTPUT_INVALID`。
  - **影响**：观测数据会误导故障定位，并让成功率指标与实际业务终态分叉。
- Go 与 Python 对 JSON Schema `pattern` 的调用方式不同：Go validator 使用 substring match，prompt lint 使用 full match。
  - **证据**：初版 `\\S` 在 Go focused test 通过，却被 `make lint-prompts` 拒绝；改为 `.*\\S.*` 后双端一致。
  - **影响**：若只跑单语言 gate，tracked schema 可能出现实现间语义漂移。

## 3 根因归类

- **spec-plan**：缺少“通用 schema validation 必须覆盖业务 decoder 拒绝条件”的终态约束。
- **test**：缺少结构合法但业务不可用的 decorator integration negative case。
- **无需额外架构改动**：现有 output-schema subset 已支持 `minItems`、`minLength`、`pattern`，无需新增 post-decode writer 或公共 API。

## 4 对流程资产的改进建议

- 已落地 D-19/C-23 和 Phase 17，要求 task run、业务 job 与副作用终态一致。
  - **落点**：A3 spec/plan/checklist
  - **优先级**：high
- 已将“持久化终态早于业务输出终态”沉淀为 Bug Pattern 12，后续 AI domain review 应逐一对比 schema 与 decoder 接受集合。
  - **落点**：`docs/bugs/PATTERNS.md`
  - **优先级**：high
- 后续若新增 schema `pattern`，同时执行 Go owner test 与 Python prompt lint，不能只依赖一端。
  - **落点**：现有 Phase 17 substitute gates
  - **优先级**：medium

## 5 建议优先级与后续动作

- 当前修复已闭环，无需扩展公共 API 或新增 E2E。
- 下一步建议在提交前通过 `/work-journal` 记录 Phase 17、BUG-0185、完整门禁证据，并使用 `fix(ai): close review findings (BUG-0185)` 原子提交。
