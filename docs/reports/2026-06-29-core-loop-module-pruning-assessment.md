# Core Loop Module Pruning 交付复盘报告

> **日期**: 2026-06-29
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `product-scope/001-core-loop-module-pruning` 方案 B：硬删除真实面试复盘和候选人画像模块，保留核心链路 `JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 下一轮`。
- 已完成前端、后端、OpenAPI、shared conventions/events/jobs、migrations、AI prompt/rubric/eval config、静态 UI 原型、UI 文档、场景资产和 active spec 的跨层清理。
- 主计划、checklist、BDD plan、BDD checklist 已从 `active` 收口为 `completed`；`docs/spec/product-scope/plans/INDEX.md` 已同步。
- 成功证据：
  - `make codegen` PASS。
  - `make render-openapi-fixture-examples` PASS。
  - `make docs-check` PASS。
  - `make lint-openapi` PASS：10 tags / 35 operations。
  - `make validate-fixtures` PASS：35 fixtures。
  - `make openapi-diff` PASS：baseline/current 35 operations，0 findings。
  - `make codegen-check` PASS；运行前临时暂存其覆盖的生成物路径，运行后已取消暂存，`git diff --cached --name-only` 为空。
  - Python contract tests PASS：OpenAPI/fixture/render group 57 tests + 4211 subtests；prompt/rubric/scenario/events group 36 tests；migration/conventions/events/openapi group 73 tests + 5 subtests；scenario script contract 2 tests。
  - Go gate PASS：`go test ./backend/cmd/codegen/conventions ./backend/internal/api/practice ./backend/cmd/api ./backend/internal/...`。
  - Frontend focused gate PASS：56 tests；`pnpm --filter @easyinterview/frontend build` PASS。
  - `git diff --check` PASS。

## 2 会话中的主要阻点/痛点

- 退役 spec stub 改写造成历史 plan 的 fragment 断链。
  - **证据**：`make docs-check` 首次失败，断链集中在 `backend-jobs-recommendations/plans/001.../plan.md`、`backend-profile/plans/001.../plan.md` 和 `openapi-v1-contract/spec.md`，均指向已删除的旧章节锚点。
  - **影响**：交付末段需要回补历史链接语义，否则 docs gate 不能闭环。

- 删除型计划初始 operation matrix 主要覆盖 runtime/API/DB/UI，但 lint/tooling surfaces 在后段才暴露。
  - **证据**：后段仍需清理 `prompt_hardcode_lint`、`lint_events_test`、`rubric_lint`、`authContractGate`、`conventions/render_test` 等旧复盘/画像引用。
  - **影响**：跨层删除容易遗漏防回流脚本和 codegen test，需要额外负向搜索与补丁轮次。

- `profile` 是高碰撞术语，既表示候选人画像，也出现在 Auth profile completion、AI profile、generated profile wording 中。
  - **证据**：zero-reference gate 必须人工分类保留 `completeMyProfile`、AI model profile、历史 OpenAPI rows 和 explicit negative tests。
  - **影响**：简单 grep 容易误伤有效术语，也容易让真正的 Candidate Profile runtime 残留淹没在噪声中。

- `make codegen-check` 的 clean diff 子 gate 语义容易被误读为“必须先提交才能通过”。
  - **证据**：本次实际通过方式是临时暂存 codegen-check 覆盖的生成物路径，让 target 检查生成器运行后是否产生未暂存漂移；随后取消暂存。
  - **影响**：若不理解该语义，容易把可完成 gate 错判为必须延期到 commit 后。

## 3 根因归类

- 退役 stub 断链：
  - **类别**：spec-plan / create-doc workflow
  - **根因**：deprecated stub 删除了旧章节结构，但历史 plan 和历史 OpenAPI decision rows 仍引用旧 fragment。

- Tooling surfaces 后段暴露：
  - **类别**：spec-plan
  - **根因**：删除型 plan 的 coverage matrix 没有把 `scripts/lint/*`、codegen tests、mock/runtime contract tests 作为独立 surface 列出。

- `profile` 术语误伤：
  - **类别**：spec-plan
  - **根因**：plan 明确了删除 Candidate Profile / Experience Card，但没有预先定义允许保留的同名非业务语义清单。

- codegen-check 暂存语义误读：
  - **类别**：README / plan
  - **根因**：Makefile target 注释说明了 drift gate，但没有给出“未提交大改中如何验证 intended generated diff”的操作说明。

## 4 对流程资产的改进建议

- 在删除型 plan 模板或 `plan-review` 规则中加入 `tooling/lint/codegen test surface` 检查项。
  - **落点**：spec-plan / plan-review
  - **优先级**：high

- 在 deprecated subject stub 规则里要求保留或重定向常见历史锚点，或者在退役后立即跑 `make docs-check`。
  - **落点**：create-doc / sync-doc-index / spec-plan
  - **优先级**：medium

- 为高碰撞术语的 zero-reference gate 增加 allowed-hit 分类表，例如 `profile` 分为 Auth account profile、AI profile、Candidate Profile runtime。
  - **落点**：spec-plan
  - **优先级**：medium

- 在 `Makefile` 或 development README 中补充 `codegen-check` 在未提交 generated diff 场景下的推荐验证方式：临时暂存 target 覆盖路径、运行 gate、再取消暂存。
  - **落点**：README
  - **优先级**：low

## 5 建议优先级与后续动作

- 最高价值：下一轮做删除型计划时，先在 plan/coverage matrix 中列出 `scripts/lint`、codegen tests、mock contract tests、scenario contract tests，避免实现后段才补工具残留。
- 次高价值：为退役 subject stub 制定锚点保留或重定向约定，减少历史计划断链返工。
- 可延后：补充 `codegen-check` 暂存验证说明；这属于执行便利性问题，本次已通过现有 git 语义完成验证。
