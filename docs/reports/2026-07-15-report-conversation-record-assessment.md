# Report Conversation Record 交付复盘报告

> **日期**: 2026-07-15
> **审查人**: Codex

**关联计划**: [Backend report generation](../spec/backend-review/plans/001-report-generation-baseline/plan.md) / [Frontend report dashboard](../spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/plan.md) / [Real API/UI journeys](../spec/e2e-scenarios-p0/plans/001-real-api-ui-journeys/plan.md)

## 1 复盘范围与成功证据

- 本次交付把面试会话记录落成 report-owned、`reportId`-only 的只读页面，覆盖 UI 设计文档、正式前端、OpenAPI/generated client、backend handler/service/store、fixture 和真实场景证据；主线合并保持已删除 Demo runtime 不复活。
- 根 `make test` 完成 backend/frontend 全量代码回归；focused tests 覆盖非法 report locator hidden-404、strict message order、Markdown security、route/back、OpenAPI fixture 和场景证据校验。
- `E2E.P0.099` current run `e2e-p0-099-20260715T021319Z-57232` 在真实 host-run frontend/backend/PostgreSQL 环境通过：六张 Chrome full-page desktop/mobile 图片、ready/generating no-OCR 审计、Report → Conversation → Back、authenticated API 与只读 PostgreSQL binding 均为 PASS。
- [BUG-0173](../bugs/BUG-0173.md) 与 [BUG-0174](../bugs/BUG-0174.md) 已修复并有回归测试；前者关闭非法 UUID locator 穿透，后者收窄证据隐私门禁到项目用户数据与 secret。
- 本次主线集成在 `main` 上合并 `feat/report-conversation-record-0715`，冲突处理保留当前 report `3/2/2/2/1` 布局计划和已删除 UI Demo 边界，将原分支 Bug 编号冲突重排为 BUG-0173/0174。当前 `make test` 通过 Python 551 tests / 4493 subtests、Go 全包和前端 124 files / 989 tests；`make build`、`make docs-check`、`make codegen-check` 与 UI Demo pruning gate 均通过。本次合并不把既有 P0.099 run 冒充为新的 E2E 执行。

## 2 会话中的主要阻点/痛点

- Hybrid 场景的自动阶段只能到达 `MANUAL_REQUIRED`，必须在同一 current-run 输出目录中补齐真实 Chrome 六图和人工 no-OCR 审计后才能转为 PASS。
  - **证据**：首次 verify 缺少 current browser artifacts；补齐 exact-six、conversation navigation 和 `manual-visual-audit.json` 后，`verify.sh` 输出 `P0_099_MANUAL_VISUAL_AUDIT_BOUND_PASS` 与 PASS。
  - **影响**：若只读取早期 marker，容易把资产 Ready、自动 preflight 和最终 current-run PASS 混为一谈。
- 证据校验器曾把任何 PNG metadata chunk 当作隐私泄露，普通 `iCCP` 色彩配置阻断真实浏览器证据。
  - **证据**：新增对照测试先复现 benign `iCCP` 被拒绝，再证明内容级扫描允许技术 profile、拒绝含 `ei_session=` 的 metadata。
  - **影响**：扩大隐私门禁到开发过程文件结构，既没有提高用户数据保护，也增加了无效验收成本。
- 各 owner checklist 保留了“本轮未运行/Ready”历史语句，真实 PASS 后需要同步清理，避免当前事实分叉。
  - **证据**：backend-review、frontend-report-dashboard 与 e2e-scenarios-p0 的 BDD checklist 均存在旧 current-run 描述，本轮已按同一 run ID 回填。
  - **影响**：只更新场景输出而不回填 owner 文档，会让后续 review 误判交付状态。
- 功能虽在独立分支完成验收，但该 commit 一度不在当前 `main` 的可达历史中，导致用户在主线截图中看不到入口。
  - **证据**：遗漏分支与当前 main 的公共祖先早于 UI Demo pruning 和 report layout 修订；合并时同时出现 Bug ID add/add 冲突、owner doc 冲突和旧 prototype/pixel-parity 资产回流。
  - **影响**：“feature branch 验收通过”被误当成“当前主线已交付”；简单选择 theirs 还会复活已删除的双实现合同。

## 3 根因归类

- Hybrid preflight 与浏览器验收分段属于真实运行合同，不是流程缺陷；根因类别为 **无需仓库改动**，执行时必须保留同一 run identity。
- PNG metadata 误判源于证据隐私范围未在 owner spec 中明确；根因类别为 **spec-plan**，并由场景 validator 测试缺口放大。
- PASS 后 owner 文档滞后属于 **spec-plan** 生命周期同步问题；current-run 结果需要在场景 checklist 和跨 owner BDD checklist 使用同一证据坐标。
- 已验收功能与当前目标分支不可达属于 **spec-plan / closeout** 交付坐标缺口；“验收证据”没有同时记录 target branch reachability，而后续主线删除工作又使盲目合并具有回流风险。

## 4 对流程资产的改进建议

- 保持 e2e-scenarios-p0 D-7 与 `PRIVACY-SCOPE-GATE` 为当前规则：只拒绝项目用户数据、认证材料和运行 secret，文件完整性、digest 与 metadata 内容风险独立验证。
  - **落点**：spec-plan / scenario README
  - **优先级**：high
- Hybrid 场景收口时以最终 `verify.sh` result artifact 为唯一 current-run 结果，并要求浏览器审计继续复用 setup 生成的 run ID 和输出目录。
  - **落点**：scenario README
  - **优先级**：medium
- 真实 E2E PASS 后执行 scoped stale-current-run 搜索，至少覆盖 owner checklist 中的 `MANUAL_REQUIRED`、`本轮未运行` 和 `Ready`，但不修改其他未运行场景。
  - **落点**：spec-plan checklist
  - **优先级**：medium
- 功能验收收口时同时记录 `git branch --contains <commit>` 或等价的 target-branch reachability；如未进入用户当前主线，结论必须明确为“分支已验收、尚未集成”。后续合并先运行已删除资产负向 gate，不以冲突为零推断语义无回流。
  - **落点**：spec-plan closeout checklist / work-journal closeout
  - **优先级**：high

## 5 建议优先级与后续动作

- 最高优先级规则已在本轮完成：隐私范围写入 spec/plan/checklist/README，并以 benign/sensitive metadata 对照测试锁定。
- 最高价值的下一步是在后续 `/work-journal` / plan closeout 中固化 target-branch reachability，避免再次出现“已验收但主线无功能”。
- 功能后续建议回到 `/implement frontend-report-dashboard/001-report-screen-and-generating-handoff frontend`，完成尚未实施的 Phase 12 `3/2/2/2/1` 布局与当前 P0.099 真实验收；`E2E.P0.098` / `E2E.P0.101` 不属于本次会话记录集成的阻塞项。
- 可以延后的优化：若后续多个 hybrid 场景重复出现 run identity 回填成本，再评估在场景框架 README 中增加统一的 current-run closeout 模板；单次交付不需要新增抽象。
