# Settings Simplification 交付复盘报告

> **日期**: 2026-07-15
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付按方案 A 将已登录 TopBar 收敛为单一设置齿轮，Settings 收敛为 Account / Privacy 两块，并删除 Sign-in & security、Font preset、静态产品字段及无 owner 的语言/地区/时区字段。
- OPENAPI-007 将 authenticated `UserContext` 收敛为 closed required `{id,email,displayName,profileCompletionRequired}`；Settings 显示完整账号 email，日志和 E2E 证据继续脱敏。
- 数据库 migration up/down/up、后端与前端 focused tests、37/37 fixtures、live Prism parity、OpenAPI exact oracle/codegen/diff、根 `make test`、frontend typecheck/build、owner contexts、docs/index/diff gates 均通过。
- `E2E.P0.101` 当前真实 frontend/backend/Mailpit run `e2e-p0-101-20260715114513-19516` PASS；Chrome 1440×900 与 390×844 验证无横向溢出、单一设置入口和完整账号字段。
- 后续 branch-vs-main review remediation 补齐默认 fixture `deleteMe` 后的 signed-out transition，以及 P0.101 失败 reporter 的 pre-log 脱敏边界；focused tests、scenario contract、根 `make test`、frontend typecheck/build 与文档门禁重新以当前代码通过，见 [BUG-0176](../bugs/BUG-0176.md)。

## 2 会话中的主要阻点/痛点

- 首轮真实 E2E 虽然业务断言通过，但 final URL 中保留了 percent-encoded email。
  - **证据**：增强后的 verifier 对首轮输出 RED；修复后重跑 PASS，见 [BUG-0175](../bugs/BUG-0175.md)。
  - **影响**：首轮证据被正确作废，需要清理并重新运行完整真实链路。
- Review 反查发现默认 fixture client 没有把成功 `DELETE /me` 建模为 session termination。
  - **证据**：新增 client 与 mounted App RED tests 分别观察到 post-delete `getMe` 仍 resolve、页面仍停留 Settings；修复后 10/10 focused tests 通过，见 [BUG-0176](../bugs/BUG-0176.md)。
  - **影响**：真实 backend 和 component mock 均正确，但默认开发预览组合路径会在删除成功后继续显示 authenticated 状态。
- P0.101 原有隐私检查只覆盖 trigger 成功后的 verifier，无法约束 assertion failure reporter。
  - **证据**：完整 email 被直接传给 Playwright matcher；`set -euo pipefail` 会在 matcher failure 后提前退出，后置 verifier 不执行。
  - **影响**：仅有 PASS-path verifier 会造成失败证据隐私假闭环，需要在 `tee` 前设置脱敏边界。
- 文档在最终 L1 反查时仍残留“email 仍脱敏”和“本轮未运行 P0.101”的当前口径。
  - **证据**：OpenAPI C-23、backend-auth plan 与 frontend-shell BDD checklist 在实现和真实 run 完成后仍保留旧描述。
  - **影响**：若直接按 checklist 勾选收口，会造成代码、场景事实与 owner 文档不一致。
- 浏览器 viewport 需要使用可报告实际宽高的真实 Chrome 会话校验。
  - **证据**：初始 in-app browser 未达到请求的 1440 viewport，因此没有把该截图标记为 1440；切换 Chrome 后得到精确 1440×900 与 390×844 证据。
  - **影响**：产生一次工具切换，但避免了错误的响应式验收结论。

## 3 根因归类

- URL-encoded email 与 failure reporter 漏检属于场景证据隐私规则不完整，根因类别为 `README` + scenario implementation；本轮在成功后 verifier 之外补齐 matcher-safe assertion 与 pre-log stream redaction。
- 默认 fixture 删除后仍 authenticated 属于跨 operation mock lifecycle 未被默认组合测试覆盖，根因类别为 `spec/plan` + frontend test surface；本轮已把 transition 写回 owner spec/checklist 并补 client/mounted App gates。
- 当前文档描述滞后于同轮实现和真实 run，根因类别为 `spec/plan`；本轮通过 post-pass L1 反查原地修正，不需要新建 sibling plan。
- 初始 viewport 不满足目标是一次工具能力差异，类别为 `no repo change needed`；以读取实际 viewport 并切换真实 Chrome 处理即可。

## 4 对流程资产的改进建议

- 为共享 browser E2E 证据增加可复用的敏感值变体生成/扫描 helper，统一覆盖原文、URL percent-encoded 与场景确实会产生的其他表示。
  - **落点**：`test/scenarios/` shared script + README
  - **优先级**：high
- 为 stateful fixture-backed clients 固化“命令成功后下一次 read 投影”的跨 operation contract test，避免 component 专用 mock 掩盖默认组合状态缺口。
  - **落点**：相关 owner spec/plan + default client tests
  - **优先级**：high
- 在 closeout 的 `/plan-review` 语义矩阵中固定核对“历史 Ready/未运行描述”与 current-run evidence 是否冲突。
  - **落点**：`plan-review` skill
  - **优先级**：medium
- 响应式截图验收前记录实际 `innerWidth/innerHeight/scrollWidth`，只有与目标一致时才命名和引用为对应 viewport。
  - **落点**：browser acceptance README 或场景 README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮优先把 BUG-0175/BUG-0176 的场景级修复抽成共享 evidence-redaction helper，并迁移含账号、token 或 URL 参数的 browser E2E，降低隐私 gate 重复实现风险。
- 同批可扫描其他 stateful default clients，检查 mutation 后 read 投影是否只有专用 component mock 覆盖；仅对存在跨 operation 状态的 client 增加 focused contract test。
- 随后可把 current-run 文案一致性加入 `/plan-review` 的固定 closeout 检查；viewport 工具差异维持操作级校验即可，不需要扩张产品代码或 owner plan。
