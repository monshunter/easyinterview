# Account Theme L2 Review Remediation 交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 [BUG-0191](../bugs/BUG-0191.md)：Settings 主题保存的迟到响应 ownership、账号主题 projection fail-closed、dev mock combined-invalid 原子拒绝，以及产品/OpenAPI/BDD owner 漂移。
- 关联 owner 为 [frontend-shell/001](../spec/frontend-shell/plans/001-app-shell-auth-settings/plan.md)、[product-scope/001](../spec/product-scope/plans/001-core-loop-module-pruning/plan.md) 与 [OpenAPI 001](../spec/openapi-v1-contract/plans/001-bootstrap/plan.md)。product-scope 与 OpenAPI 本轮恢复 completed；frontend-shell 因既有 Phase 15.3 locale/Chrome 三项未完成而继续保持 active。
- RED 证据为 frontend focused `4 failed, 41 passed` 与 owner-doc focused `2 failed`；GREEN 为 frontend 4 files / 45 tests、owner-doc focused 2 tests、完整 owner lint 43 tests / 28 subtests。
- 根 `make test` 通过：Python 615 tests / 4615 subtests、Go 全包、frontend 127 files / 1042 tests；typecheck、production build、OpenAPI 38/38、fixture、codegen、diff、migration、docs/index/context/diff 与 residual gates 均通过。
- 真实环境 `E2E.P0.101` run `e2e-p0-101-20260719082610-75505` PASS：theme save 1 PATCH、Settings/路由切换 0 次额外 `/me`、重登恢复 plum、console/page/非预期 HTTP failures 为 0，cleanup PASS。

## 2 会话中的主要阻点/痛点

- 原完成态 owner 把 failure/retry/leave 等行为标为 PASS，但缺少 race、invalid projection 与 combined-invalid 反例，且 active product/OpenAPI 文本仍声明旧方案。
  - **证据**：新增测试先得到 4 个前端失败和 2 个文档失败；product D-21 与 OpenAPI §4.2/handoff 分别保留 TopBar theme / `CompleteProfileRequest` 旧口径。
  - **影响**：历史 checklist 状态掩盖了真实竞态、mock parity 和 truth-source 冲突。
- 离开 Settings 的新测试最初把 authenticated fixture 的确认主题误认为 ocean。
  - **证据**：fixture 的 `authenticated` scenario 实际为 plum；测试再次点击 plum 后没有产生草稿，却期待离开恢复 ocean。
  - **影响**：造成一次无效 GREEN 迭代；读取真实 fixture 后改为 plum → ocean 草稿 → 恢复 plum，并撤回为错误前提增加的额外路由逻辑。
- `make migrate-check` 首次在未加载 dev-stack env 的 shell 中缺少 `DATABASE_URL`。
  - **证据**：migration lint 通过后命令明确退出 `DATABASE_URL is required`；按 scenario-env 合同加载本地 ignored env 后同一 gate 通过。
  - **影响**：没有实现返工，但必须区分环境前置缺失与迁移断言失败。

## 3 根因归类

- 历史 BDD/checklist 与 active owner 文本失真属于 `spec/plan`；本轮已原地重开、增加 executable semantic gates 并按真实剩余项恢复 lifecycle。
- fixture 主题误读属于一次性测试编写错误，类别为 `no repo change needed`；现有“实际读取当前文件”规则足够，修正测试与撤回多余实现即可。
- migration env 缺失属于 `no repo change needed`；`scenario-env` README/skill 已明确单一 env 来源，本轮按现有合同恢复，无需新增环境兼容层。
- 迟到异步响应 ownership 是可复用风险，但是否进入通用模式库需要单独确认，目标资产为 `docs/bugs/PATTERNS.md`。

## 4 对流程资产的改进建议

- 将“异步响应写回 auth/runtime 前验证 mounted + request generation + current identity”提炼为前端通用检查模式，并附 race test 模板。
  - **落点**：`docs/bugs/PATTERNS.md`
  - **优先级**：high
- 后续账号偏好字段继续复用单一 closed runtime normalizer 与 absent/present-invalid mock parity gate，不在 screen、runtime、mock 各自维护宽松解析。
  - **落点**：frontend-shell spec/plan
  - **优先级**：medium
- 保持 frontend-shell/001 为 active，完成既有 `BDD.SHELL.AUTH.LOCALE.001` 的中文 gate、英文切换和 current-run Chrome 证据后再恢复 completed。
  - **落点**：frontend-shell/001 Phase 15.3
  - **优先级**：high

## 5 建议优先级与后续动作

- 下一步最高优先级是继续 `/implement` frontend-shell/001 Phase 15.3，补齐 auth loading/error locale 的真实 Chrome 验收；这是该 owner 当前唯一明确未完成的主题，不应被本轮 account-theme PASS 掩盖。
- 若用户确认沉淀通用模式，再单独更新 `docs/bugs/PATTERNS.md`；本轮不擅自写入可复用治理规则。
- 账号主题修复本身没有遗留实现或验收缺口；后续提交时可使用 `fix(settings): harden account theme review regressions (BUG-0191)`，并在 work journal 关联 BUG-0191。
