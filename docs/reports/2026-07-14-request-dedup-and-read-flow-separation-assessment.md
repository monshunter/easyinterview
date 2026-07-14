# Request Dedup and Read Flow Separation 交付复盘报告

> **日期**: 2026-07-14
> **审查人**: Codex

**关联计划**: [Frontend App Shell](../spec/frontend-shell/plans/001-app-shell-auth-settings/plan.md)、[Frontend URL Routing](../spec/frontend-shell/plans/004-url-addressable-routing/plan.md)、[Frontend Home JD Import](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md)、[Frontend Workspace](../spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md)、[Frontend Resume Listing](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)、[Backend Resume](../spec/backend-resume/plans/001-asset-register-parse-and-listing/plan.md)、[OpenAPI Breaking Change Gate](../spec/openapi-v1-contract/plans/003-breaking-change-gate/plan.md)

**关联 Bug**: [BUG-0170](../bugs/BUG-0170.md)

## 1 复盘范围与成功证据

- 生成客户端新增 per-client concurrent safe-GET single-flight，不引入 TTL 缓存；key 覆盖 method、规范化 URL/query/headers、可接受状态与 read/auth epoch。caller AbortSignal 保持独立取消语义，mutation 与鉴权变化提供失效 fence，成功/失败 settlement 都会清理 in-flight 项。
- Resume 列表改为精确九字段 closed `ResumeSummary`，backend 单次查询直接投影，fixture/generated client/frontend consumer 同批迁移；`getResume` 保持完整详情。StrictMode transport gate 证明列表初次 GET=1、打开详情前 detail GET=0、打开后 detail GET=1，失败后重试仅新增一次传输。
- JD 导入与读取拆为两个 owner：POST `/targets/import` 后进入 `/parse?targetJobId=...` 观察 queued/processing，ready 后 replace 到 `/workspace?targetJobId=...`；已解析卡片、Practice 终态与 Report 返回动作直接进入 Workspace 只读详情，不复用 Parse 动画或 import 命令。
- 同批完成 Practice 用户/助手消息安全 GFM、Settings 自定义主题仅保留 hue/saturation、Workspace 轮次卡片已进行/即将进行/未进行三态，并同步正式前端、静态原型、UI 文档与 desktop/mobile parity gate。
- 当前场景证据覆盖 P0.005、P0.006、P0.014、P0.015、P0.016、P0.018、P0.021、P0.034、P0.036、P0.037、P0.044、P0.046、P0.058、P0.059、P0.088、P0.089、P0.090、P0.098、P0.102；关键 focused 结果包括 P0.015 57/57、P0.021 36/36、P0.058 84/84、P0.102 UI contract 65/65 与 frontend 70/70。
- 提交前根级 `make test` 最终通过：UI source contract 65/65、Python 595 tests / 5122 subtests、全部 Go packages、frontend 125 files / 1010 tests。
- Chrome 实际验收保存 6 张截图与结构化 `evidence.json`：Resume list/detail、ready Workspace card 均以 Resource Timing 证明相同 GET 只有 1 次；九字段列表、三态轮次、两滑杆主题、安全 Markdown 与 JD import→Parse 进度均通过。
- OpenAPI baseline/current 同 SHA-256；`openapi-diff` 0 findings、37 operation fixtures、122 个契约测试、5 个 Prism 单测、13/13 live Prism byte-equal，以及 69 个 codegen-owned 文件零字节漂移全部通过。

## 2 会话中的主要阻点/痛点

- 浏览器自动化网络记录器把同一请求重复展示，最初看起来像修复后仍发送两次。
  - **证据**：同一页面用 Chrome `PerformanceResourceTiming` 和 Playwright `page.on('request')` 交叉统计均为 1；记录器条目数与底层传输数不一致。
  - **影响**：若把调试 UI 的累积记录当作唯一证据，会误判实现失败并诱发无效去重层。
- P0.098 的真实报告旅程 seed 只有 `session_started`，缺少可生成报告的已回答 user/assistant message。
  - **证据**：production completion 正确拒绝无回答会话；补入 opening assistant、completed user 与 linked assistant 后，setup 增加语义预检，完整场景通过且数据清理为零残留。
  - **影响**：结构合法但业务语义不足的 seed 会把生产校验误报成回归，增加跨层诊断成本。
- 广泛单测与场景已经通过时，owner 文档审计仍发现 transport/BDD handoff 没有绑定到最终 gate。
  - **证据**：新增 Home/Parse/Resume StrictMode raw transport markers、失败驱逐重试、P0.102 聚合和 Resume Prism list/detail parity 后，原 owner checklist 才具备当前可重放证据。
  - **影响**：只勾实现项而未把底层传输与 consumer handoff 写进 BDD，会让相同问题在另一个模块复发。
- OpenAPI baseline 曾在 all-consumer handoff 完成前进入冻结候选状态。
  - **证据**：先恢复旧 baseline，保留 OPENAPI-005 12 条 expected findings；待 frontend/backend/fixture/Prism/scenario 全绿后再冻结，最终 baseline/current 同哈希且历史审计可重放。
  - **影响**：过早冻结会把“当前 proposed schema”误写成“已完成迁移的基线”，削弱 breaking-change gate 的证明力。
- 第一次根级 `make test` 在 594 个 Python tests 已通过后发现一条 scenario contract 仍把 Practice terminal recovery 锁定到 `/parse` 并禁止 `/workspace`。
  - **证据**：production pixel test 已精确断言 `/workspace?targetJobId=...`；修订 contract 为 Workspace 正向、Parse 负向及唯一 query key 后 focused 1/1 与第二次完整聚合均通过。
  - **影响**：只重跑改动模块会遗漏跨 owner 的旧防漂移断言，使正确的 route 修复在 closeout 才被反向判错。

## 3 根因归类

- 重复 GET 与页面级防重遗漏
  - **类别**：spec-plan
  - safe-read 去重此前没有传输层 owner，页面只能各自处理 effect 生命周期；本轮已在 frontend-shell owner 与 P0.102 固化统一策略。
- Parse/Workspace route authority 混合
  - **类别**：spec-plan
  - 命令进度和资源读取没有分别声明 route、data consumer、terminal handoff 与 negative legacy path；本轮由 Home/Workspace/Shell 原 owner 同批修订。
- Resume 列表过量 DTO
  - **类别**：spec-plan
  - OpenAPI 复用完整 `Resume` 导致所有 consumer 自然过取；OPENAPI-005、backend-resume 与 frontend-resume 现已共同定义 closed projection。
- 浏览器请求计数误判
  - **类别**：README
  - 调试记录器不保证一条记录等于一次底层传输；当前场景 README 与测试 marker 已改用 raw fetch、Resource Timing 或 Playwright request event。
- 场景 seed 只有结构约束、缺少业务可执行前置条件
  - **类别**：spec-plan
  - P0.098 setup 已新增 completed user/assistant 语义验证，避免报告路径再次使用空会话。
- closeout contract 保留旧 route 正向口径
  - **类别**：spec-plan
  - 根级 contract 现已要求 Workspace 精确 destination、targetJobId-only query 和 Parse negative；对应 owner checklist 也已恢复 completed。

## 4 对流程资产的改进建议

- 涉及“重复请求”的验收统一采用两层证据：组件/客户端 raw transport counter + 浏览器 Resource Timing 或 Playwright request event；网络面板/自动化记录器只用于定位，不作为唯一计数 oracle。
  - **落点**：相关 owner checklist 与 `test/scenarios/e2e/p0-014|015|036|037|102`
  - **优先级**：high
- route 修订固定四列 operation matrix：command operation、progress/read operation、route owner、terminal/return destination，并增加旧 route/query zero-reference。
  - **落点**：frontend-home、frontend-workspace、frontend-shell owner plan
  - **优先级**：high
- route 的 browser/pixel consumer 改动必须同步更新根级 scenario contract，并同时断言新路径存在、旧路径不存在和 query exact set；focused test 后仍运行一次根级聚合。
  - **落点**：`scripts/lint/scenario_script_contract_test.py` 与相关 owner final gate
  - **优先级**：high
- 列表 schema 默认要求 closed projection、详情字段负向清单、backend 单查询投影和 list/detail Prism 双 operation；禁止通过前端 N+1 补齐。
  - **落点**：OPENAPI-005、openapi 004、backend/frontend Resume final gate
  - **优先级**：high
- 真实旅程 seed 除 DDL/外键校验外，增加触发下一业务动作所需的语义前置检查，例如“至少一条 completed user answer 与 linked assistant reply”。
  - **落点**：P0.098 setup/README 与后续 report-producing scenario 模板
  - **优先级**：medium
- breaking schema 的冻结顺序保持为 accepted decision → old-baseline audit → all-consumer handoff → scenarios/Prism/codegen → final re-freeze；冻结后仍需从 preserved audit 重放 expected findings。
  - **落点**：openapi 003 Phase 9 final gate
  - **优先级**：high

## 5 建议优先级与后续动作

1. 本轮提交后，优先使用 `/plan-code-review --fix` 按当前 14 个 owner context 对提交相较 `main` 做一次 L2 反查，重点检查 safe-GET key/epoch、AbortSignal bypass、route negative search 与 Resume closed projection。
2. 若需要把“网络记录器不能作为唯一传输计数证据”提升为全仓通用模式，应由用户确认后再追加到 `docs/bugs/PATTERNS.md`；当前可执行场景 gate 已先落地，不阻塞本次交付。
3. 暂不引入跨请求 TTL cache、新的 Resume list endpoint 或 Parse/Workspace 兼容 route；这些都会扩大状态和接口面，违背本轮奥卡姆剃刀边界。
