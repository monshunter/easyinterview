# Runtime Content Limits 交付复盘报告

> **日期**: 2026-07-14
> **审查人**: Codex

**关联计划**: [Secrets and Config](../spec/secrets-and-config/plans/001-bootstrap/plan.md)、[Backend Review](../spec/backend-review/plans/001-report-generation-baseline/plan.md)、[OpenAPI Bootstrap](../spec/openapi-v1-contract/plans/001-bootstrap/plan.md)、[OpenAPI Breaking Change Gate](../spec/openapi-v1-contract/plans/003-breaking-change-gate/plan.md)

**关联 Bug**: [BUG-0171](../bugs/BUG-0171.md)

## 1 复盘范围与成功证据

- 本次交付将 HTTP、Resume、TargetJob、Practice、Report 与四个 AI provider 的大小限制收敛到统一 typed defaults、YAML override、fail-fast validation 和显式注入；RuntimeConfig 只公开五个前端 preflight 所需字段。
- 报告生成不再使用旧 package-local threshold：内存构造的 62,397-byte 回归输入与 917,504-byte 默认边界各进入 provider 一次，917,505-byte 输入在 provider/repair/retry 前以 `REPORT_CONTEXT_TOO_LARGE` 终止且 provider calls 为 0。
- `backend/internal/review/testdata/report-boundary/` 已删除全部 `input-*.json`；测试 helper 在内存构造 exact regression/limit/limit+1 payload，并验证 canonical JSON round-trip、调用次数和持久化副作用。
- 当前场景 P0.010、P0.015、P0.034、P0.035、P0.046、P0.056、P0.058、P0.081 均通过。后端全量及相关 race、前端 126 files / 1018 tests 与 production build、OpenAPI 37/10、37 fixtures、52 个 diff-wrapper tests、11 个 owner context、docs/index/diff gate 均通过。
- 主实现提交后独立执行 `make codegen-check`，69 个 codegen-owned 输出与当前真理源字节一致；提交为 `1e5c414fb fix(runtime): centralize content size limits (BUG-0171)`。

## 2 会话中的主要阻点/痛点

- Report 的旧 package hardcode 在 provider 前拒绝了模型容量可以承载的输入。
  - **证据**：62,397-byte framed input 在旧边界下返回“材料与对话过长”，而当前 profile context window 为 1,000,000 tokens；修复后同一输入进入 provider。
  - **影响**：用户收到不可恢复的错误页面，且提示把责任错误转移给输入材料。
- 历史边界验证依赖提交的输入 JSON，而业务真正需要证明的是 byte gate 与下游副作用。
  - **证据**：删除两个输入文件后，62,397 / 917,504 / 917,505 三个 exact payload 仍可在内存稳定重建并通过 canonical round-trip；P0.056/P0.058 保持通过。
  - **影响**：真实大文件会增加仓库资产与 hash 维护成本，却不能替代 provider call/no-call 和持久化断言。
- 调高默认值后，P0.058 store 证据仍把 60,000 bytes 当作 oversized。
  - **证据**：新默认值下该样本已合法，首次 P0.058 暴露测试不再进入 oversized 分支；改为 917,505 bytes 后，provider calls=0 与 terminal persistence 重新得到证明。
  - **影响**：固定历史样本会在配置演进后悄悄测试另一条路径，形成假阳性。
- OPENAPI-005 的历史 replay 测试假定自身 audit current hash 永远等于最新 baseline。
  - **证据**：OPENAPI-006 合法 re-freeze 后该断言失败；修复为校验 OPENAPI-005 current hash 等于 OPENAPI-006 old-baseline hash，完整 52 个 wrapper tests 通过。
  - **影响**：正确追加下一条 breaking decision 反而会破坏历史审计重放，阻塞合法 baseline 演进。
- `make codegen-check` 以当前 `HEAD` 为比较基准，未提交的预期 generated diff 在主提交前必然被判为漂移。
  - **证据**：实现阶段其他 codegen/lint gate 已通过；主提交后相同 `make codegen-check` 通过且工作区无生成物变化。
  - **影响**：若不区分 pre-commit 生成正确性与 post-commit byte-stability，会提前记录无法成立的 PASS。

## 3 根因归类

- 大小限制存在多个生产真理源。
  - **类别**：spec-plan
  - 旧设计没有明确 typed config owner、consumer injection、公开/内部边界和跨字段约束；本轮已由 A4 Phase 13 与各 consumer owner 原地收敛。
- 边界测试把文件资产当作业务证明。
  - **类别**：spec-plan
  - 旧 gate 锁定 fixture hash，而没有优先锁定 exact byte construction、provider calls 和 persistence；backend-review Phase 11 已改为内存构造。
- 场景 oversized 样本没有绑定当前配置边界。
  - **类别**：spec-plan
  - 固定数字在默认值变化后失去语义；P0.058 当前 gate 已使用 default limit+1 并输出 zero-provider marker。
- OpenAPI 历史审计没有按 decision chain 重放。
  - **类别**：spec-plan
  - 旧断言比较“历史 current”与“今天 baseline”，而非下一条 accepted decision 的 old-baseline；OPENAPI-006 audit chain 已修正并加入回归测试。
- post-commit codegen 验证顺序。
  - **类别**：无需仓库改动
  - 这是 gate 的既定 HEAD 语义，本轮 checklist 与工作日志已明确先提交、再验证、再文档收口，无需改变工具。

## 4 对流程资产的改进建议

- 所有新增/修订 size limit 必须同时声明 typed default、YAML key、consumer、单位、UTF-8 byte 语义、limit/+1、跨字段 validator、公开投影与内部负向清单。
  - **落点**：secrets-and-config Phase 13 与相关 owner checklist
  - **优先级**：high
- 大小边界测试默认在内存构造 exact payload，并同时断言下游 call count 与 persistence；除解析器确需文件格式外，不提交仅用于凑 byte 数的输入文件。
  - **落点**：backend-review Phase 11、P0.056、P0.058
  - **优先级**：high
- 场景中的 oversized/limit 样本应从当前运行时默认值派生或显式写成 default+1，并在 marker 中输出实际 byte 数，避免历史字面量失效。
  - **落点**：P0.010、P0.034、P0.035、P0.046、P0.058、P0.081
  - **优先级**：high
- breaking audit replay 固定使用 decision chain：前一条 audit current hash 必须匹配后一条 accepted decision 的 old-baseline hash；不得要求历史 audit 匹配今天 baseline。
  - **落点**：OpenAPI breaking-change gate Phase 10 与 `scripts/lint/openapi_diff_test.py`
  - **优先级**：high
- 保持 `codegen-check` 的 post-commit byte-stability 语义；在 phase checklist 中明确该 gate 的执行顺序，不为适配未提交工作区而放宽比较。
  - **落点**：实施 checklist / 工作日志
  - **优先级**：medium

## 5 建议优先级与后续动作

1. 下一步优先使用 `/plan-code-review --fix secrets-and-config/001-bootstrap` 对 `1e5c414fb` 相较 `main` 做 L2 反查，重点检查所有 consumer 是否只接收注入值、RuntimeConfig 是否保持五字段 closed projection，以及 limit+1 是否都在副作用前拒绝。
2. 若后续新增新的内容类型，先扩展 A4 owner matrix 与 validator，再由具体业务 owner 消费；不要在业务 package 或前端页面新增独立默认值。
3. 不再恢复 report `input-*.json`；只有当解析器必须验证真实文件编码/格式时，才在相应解析 owner 下引入最小 fixture，而不是把它当作 byte-boundary oracle。
