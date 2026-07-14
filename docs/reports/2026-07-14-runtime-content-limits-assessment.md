# Runtime Content Limits 交付复盘报告

> **日期**: 2026-07-14
> **审查人**: Codex

**关联计划**: [Secrets and Config](../spec/secrets-and-config/plans/001-bootstrap/plan.md)、[Backend Review](../spec/backend-review/plans/001-report-generation-baseline/plan.md)、[OpenAPI Bootstrap](../spec/openapi-v1-contract/plans/001-bootstrap/plan.md)、[OpenAPI Breaking Change Gate](../spec/openapi-v1-contract/plans/003-breaking-change-gate/plan.md)

**关联 Bug**: [BUG-0171](../bugs/BUG-0171.md)

## 1 复盘范围与成功证据

- HTTP、Resume、TargetJob、Practice、Report 与 provider adapter 的内容限制已收敛到统一 typed defaults、YAML override、fail-fast validation 和显式注入；RuntimeConfig 只公开五个前端 preflight 字段。
- Report 不再使用旧 48,000-byte package-local threshold；A4 当前 `report.maxFramedInputBytes=917504`。
- A3 当前要求六个 active LLM profile `max_tokens >= 16384`，report 的 output budget 为 16,384 tokens、context window 为 1,000,000 tokens。这些 token 配置与 byte guard 独立验证，不用跨单位加法宣称容量等价，TPM 也不参与单请求裁决。
- `backend/internal/review/testdata/report-boundary/input-*.json` 已删除；不再构造 default/default+1 大材料或用真实大文件证明配置逻辑。
- 前端删除本地 content-limit 默认 fallback；运行时配置未就绪时 consumer fail closed，不产生第二份默认真理源。
- 生产代码完成后，仓库根 `make test` 已通过：`ui-design` 65 项、当时的 Python/Skill 542 项（含 5,122 个 subtests）、backend `go test ./...` 全部 package、frontend 125 个 test files / 998 项 tests。后续仅修改治理文档与 Skill 合同，并单独重跑当前完整 Python/Skill suite：543 项、5,122 个 subtests 全部通过；前后端生产代码未再变化。
- `make lint-config`、`make lint-ai-profile-coverage`、49 个 plan context、20 组 BDD plan/checklist 成对审计（19 个 domain Behavior ID、3 个真实 E2E ID）、`make docs-check` 与 `git diff --check` 构成本次静态收口证据。
- 本轮没有启动场景环境，也没有运行 P0.098/P0.099/P0.101；三个目录只代表可在真实 API/UI 环境运行的 `Ready` 资产，不代表当前运行 PASS。

## 2 会话中的主要阻点/痛点

### 配置合同被跨层复制

同一默认值、override、invalid 和 limit/default+1 曾在 loader、composition、domain、frontend 与场景层重复断言。重复测试没有增加业务覆盖，反而让每次默认值调整需要同步大量数字、fixture 和 marker。

修订后只保留一组 A4 typed loader/validator owner contract。Consumer 只测试类型和注入无法自动保证的非平凡分支，例如错误映射、provider call/no-call、持久化原子性或协议读取上限，并使用小型注入值。

### 大 payload 被误当作边界证明

历史设计提交 `input-*.json`，并在内存构造默认大小、default+1 payload。业务真正需要证明的是 guard 的控制流和副作用，不是把测试数据填充到某个生产默认字节数。

修订后不再重建 62,397-byte 历史输入；该数值只保留在 Bug 根因记录中。业务控制流由一个小型 injected limit 的 admitted/overflow 测试证明，生产默认数值由 A4 typed owner contract 与 A3 profile loader/coverage lint 各自证明。只有 parser 的编码/格式行为才有理由使用真实文件 fixture；当前 report-boundary 目录仅保留 output-schema zh/en fixtures 及 manifest。

### 跨单位算术被误当成容量证明

旧 `report_budget_test.go` 直接将 UTF-8 bytes、framing reserve tokens 和 output tokens 相加后与 context-window tokens 比较。这个算式没有 tokenizer 转换合同，即使数值成立也不能证明实际请求一定被模型接受。该测试已删除；当前只独立校验 byte guard 与 profile token fields，不构造两者的虚假算术关系。

### 代码测试被包装成 E2E

旧 `test/scenarios/e2e` shell 聚合 Go test、Vitest、pytest、lint、build、fixture parity 或 provider CLI/eval，再输出场景 PASS。这些证据没有驱动已运行应用，不是端到端用户流程。

修订后的 E2E 只保留真实 HTTP API 或真实 browser UI 到 host-run backend 的流程。Report 范围只由 P0.099 验证真实 report/generating 页面、authenticated API、read-only DB 与 current-run 六图；provider reliability 归入独立 code/eval gate。

## 3 根因归类

- **spec/plan**：配置默认值、consumer 业务行为和真实 E2E 曾被写进同一验收链，导致调整一个数值需要跨层同步大量断言。
- **AGENTS.md / README**：此前没有明确区分 Behavior ID、代码级行为测试与真实 E2E，Go/Vitest/pytest/lint/build 因而被包装成场景。
- **skill**：design、implement、tdd、plan review/code review 与 scenario skills 曾倾向“有业务词就建 BDD/E2E”，缺少 BDD-N/A 和最小充分证据的退出条件。
- **no repo change needed**：62,397-byte 历史请求是根因证据，不是必须长期保存的回归 fixture；删除真实大小材料不会削弱小值注入的控制流验证。

### 3.1 修订后的测试分层

| 层级 | Owner | 允许的证据 | 不承担 |
|------|-------|------------|--------|
| 配置合同 | A4 typed loader/validator | default、override、代表性 invalid、cross-field | consumer 业务、副作用、E2E |
| Consumer code | 各业务 package/frontend module | 小值注入、错误映射、call/no-call、持久化/协议缺陷 | 重复默认数值、真实大材料 |
| Provider/profile contract | A3 loader/coverage lint + F3/evalkit | 六个 active profile 至少 16K、report 16K/1M fields、validator/judge reliability | byte/token 换算、应用 E2E |
| 全量单测 | 根 `make test` | backend `go test ./...` + frontend 全量 tests | 真实环境用户旅程 |
| E2E | `test/scenarios/e2e` | 真实 API/UI、持久化结果、用户可见状态 | Go/Vitest/pytest/lint/build 包装 |

开发时可以运行 focused test 快速反馈；阶段完成与 CI 必须从仓库根执行 `make test`，整体回归前后端单测。E2E 单独运行、单独报告，不再次编排 `make test`。

## 4 对流程资产的改进建议

- **high / AGENTS.md**：已加入配置测试适度性门禁，要求单一 owner contract、consumer 只测非平凡业务分支、配置 wiring 不单设场景环境。
- **high / AGENTS.md + scenario README**：已加入 E2E 证据边界，只承载真实 API/UI，禁止代码测试包装和 mock backend；无真实链路的旧场景直接删除。
- **high / Makefile + AGENTS.md**：已明确根 `make test` 是阶段完成和 CI 的前后端全量单测入口；focused PASS 或 E2E shell 不能替代。
- **medium / skills**：已修订 design、implement、tdd、plan review/code review 与 scenario skills，使纯配置/内部工具使用 BDD-N/A，Behavior ID 不再被强制映射为 P0 E2E；单个 `BDD-Gate` 只允许一种 evidence layer，代码 owner 的真实场景引用单独记为 `E2E-HANDOFF`，不得改变 `Ready` 状态。
- **medium / spec-plan**：Backend Review、Frontend Report 与 E2E owner 已把配置、code/eval、deterministic parity 与 P0.099 real UI 分层，不再保留已删除场景的当前 owner 引用。

## 5 建议优先级与后续动作

- **P0 / 已完成**：保持 A4 单一配置 owner、A3 active-profile floor 和小值 consumer branch 三层证据，不恢复默认尺寸 payload、跨单位公式或配置专用 BDD/E2E。
- **P1 / 按需执行**：只有需要产品验收时，才显式使用 `/scenario-run` 在真实环境分别运行 P0.098、P0.099、P0.101；结果必须与本次静态审计和根 `make test` 分开报告。
- **P2 / 持续约束**：新增 BDD 时先判定是否存在用户可观察行为；纯配置、内部工具、lint/codegen/migration 直接声明 BDD-N/A，并记录最小替代 gate。

### 5.1 当前收口结论

- A4 owner contract、A3 typed loader + active-profile coverage lint、小值 consumer tests 与 `input-*.json` zero-reference 已通过。
- 根 `make test` 已完成前后端全量单测回归；必要的 race、PostgreSQL、OpenAPI/codegen、lint、build、prompt/eval 继续作为独立 gate 报告。
- E2E 脚本不调用 Go test、Vitest/npm test、pytest、lint、build 或 provider CLI/eval；Playwright 只用于驱动真实应用 UI。
- P0.099 仅可在真实 environment evidence 完整时报告结果，不使用历史 PASS、fixture 页面或 code test marker 冒充。

### 5.2 后续原则

1. 新增内容限制时先扩展 A4 owner matrix；仅当 consumer 出现新的非平凡业务分支时才增加 focused test。
2. 默认大小调整只更新配置真理源和 owner contract；不要在多个 domain、frontend 或 E2E 复制数值。
3. 保持“根 `make test` 全量单测 / E2E 真实用户流程 / eval 模型可靠性”三条证据链独立，分别报告结果。
4. 不恢复 bytes + tokens 的跨单位容量公式；如需证明某个模型对具体请求的可接受性，必须另行定义 tokenizer/provider 契约，不得用配置单测代替。
