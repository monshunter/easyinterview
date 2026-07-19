# Account Theme Runtime Persistence 交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付按用户选择的方案 B，将 `PATCH /me` 泛化为 `updateMe`，把主题色作为账号偏好持久化，并让应用只在登录 bootstrap / auth recovery 获取一次 `GET /me`；页面切换、主题草稿预览和保存后的 runtime 更新都不追加读取请求。
- 设置入口统一为“设置 / Settings”与明确齿轮图标；主题选择移入 Settings Appearance，Practice 同时保留全局 TopBar 和独立 Session Header。
- 关联 owner 包括 [frontend-shell/001](../spec/frontend-shell/plans/001-app-shell-auth-settings/plan.md)、[backend-auth/001](../spec/backend-auth/plans/001-email-code-session-bootstrap/plan.md)、[OpenAPI 001](../spec/openapi-v1-contract/plans/001-bootstrap/plan.md) 与 [DB migrations 001](../spec/db-migrations-baseline/plans/001-bootstrap/plan.md)。
- 根 `make test` 通过：Python 613 tests / 4615 subtests、Go 全包、frontend 127 files / 1037 tests；frontend production build 通过。
- OpenAPI 当前 baseline diff 为 0，38 operations / 10 tags、38 fixtures、Prism/dev-mock parity 和 migration lint 均通过；隔离 PostgreSQL 的 populated up/down/up、默认值与非法约束验证通过。
- 真实环境 `E2E.P0.101` run `e2e-p0-101-20260719034142-32741` 通过：Settings mount 与 route 往返均为 0 次额外 `GET /me`，保存为 1 次 `PATCH /me`，退出重登恢复 `plum`，console/page/HTTP failures 均为 0。
- Chrome current-run 验证确认 Settings 梅子主题、明确设置齿轮、Practice 双层导航与会话操作可达；精选证据使用真实 JPEG 格式保存。

## 2 会话中的主要阻点/痛点

- `codegen-check` 在生成物正确更新后仍因相对 HEAD 存在预期合同 diff 而失败。
  - **证据**：生成后的 `backend/internal/api/generated/openapi.yaml` 与源 OpenAPI 字节一致，但 target 最后执行 `git diff --exit-code`，在功能提交前必然报告新生成物 diff。
  - **影响**：幂等性与“尚未提交”被混成同一信号，增加一次误判和解释成本。
- migration lint 首次发现 `user_settings.theme` 的 enum source 未登记。
  - **证据**：真实 PostgreSQL migration/constraint test 已通过，但 `migrations_lint.py` 报 `user_settings.theme check list is not registered`。
  - **影响**：跨层实现若只看数据库行为，会漏掉迁移治理源的一致性要求。
- 共享场景数据库的 `schema_migrations` 元数据与真实 v20 net-state 不一致。
  - **证据**：首次 migrate-up 写成 dirty 5；只读审计确认 v20 表/列事实完整后，才定向修正 metadata 并由正式 migrate-up 升到 v21 clean。
  - **影响**：真实环境验证被历史元数据阻塞，且恢复必须同时避免误删共享数据和错误重放旧迁移。
- 旧四字段 `UserContext` 的 source tests 与历史 D-40 审计读取方式未随新冻结合同同步。
  - **证据**：根测试精确指出 inventory/fixture 的四字段断言，以及 D-40 直接读取可变 current source 的 digest 漂移。
  - **影响**：如果只运行局部前后端测试，会漏掉历史审计链和治理测试的当前兼容方式。

## 3 根因归类

- codegen gate 同时承担“生成幂等”与“工作树相对 HEAD 无 diff”，属于 `README / Makefile gate contract` 的语义耦合。
- enum source 漏登属于 `spec/plan` 的 migration operation matrix 未把 `migrations/enum-sources.yaml` 明列为 owner artifact。
- 场景数据库 metadata 漂移属于 `scenario-env README / tooling` 缺少“metadata 与 net-state 分歧”的预检和安全恢复指引。
- 四字段断言与历史审计链问题属于 `spec/plan` 的 consumer 清单不够完整；根级测试已正确兜底，不需要另建 Bug。
- JPEG 扩展名识别与裁剪试验属于一次性证据整理问题；已通过 `file` 校验并改正，不需要仓库规则变更。

## 4 对流程资产的改进建议

- 为 codegen 增加工作树中立的幂等 gate：在临时快照或生成前后 diff 上判断，而把“相对 HEAD 零 diff”保留为提交后 gate。
  - **落点**：Makefile、OpenAPI/codegen README
  - **优先级**：high
- OpenAPI/DB 跨层 plan 的 operation matrix 增加 `migrations/enum-sources.yaml`、source lint tests、历史 audit chain tests 三类显式 consumer。
  - **落点**：spec/plan
  - **优先级**：high
- 场景环境在 migrate-up 前比对 migration metadata 与关键 net-state marker；发现分歧时 fail closed，并提供只读诊断与定向 metadata recovery 文档，不自动清卷。
  - **落点**：scenario-env README / scripts
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最高价值事项是把 `codegen-check` 拆为“生成幂等”与“提交后 drift”两个可独立报告的 gate，避免跨层合同变更在提交前出现预期假失败。
- 随后补齐 migration net-state 预检与安全恢复说明；这项不阻塞当前主题功能，但能显著降低共享真实环境下一次升级的恢复成本。
- 当前功能没有遗留实现缺口；后续产品迭代可在同一 `updateMe` owner 下增加其他账号级显示偏好，但必须继续保持 bootstrap 单次读取和保存响应直接更新 runtime 的性能边界。
