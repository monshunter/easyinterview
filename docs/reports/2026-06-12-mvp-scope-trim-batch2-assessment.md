# MVP Scope Trim Batch 2 设计层交付复盘

> **日期**: 2026-06-12
> **审查人**: Claude

## 1 复盘范围与成功证据

本次会话完成第二批 UX 冗余裁剪的设计层交付（commit `79ee280c`，分支 `design/ux-funnel-simplification`，32 文件 +1056/−3323）：

- **决策输入**：首批复盘报告记录的 6 项未采纳建议全部裁剪（报告 CTA 三处重复、岗位推荐联网搜索 tab、设置占位 tab、公司情报独立页、画像页双呈现、主题过度配置），叠加用户指令：删除岗位推荐模块、简历模块扁平化（移除轻量问答 / 版本树 / 主版本继承，改写建议仅“采纳”，采纳后确认前预览选覆盖原简历或保存为新简历）。
- **文档修订**：`product-scope/spec.md` v2.0（D-17~D-21 锁定决策、删除主流程 B、重写 M5、新增 C-16/C-17/C-18）；`docs/ui-design/` 12 份原地修订（含 `resume-module`、`jd-resume-management` 2.0 重写）+ README/TEMPLATES 口径清理 + 两处 INDEX 投影同步。
- **原型修订**：删除 `screen-jd-match.jsx`、`CompanyIntelScreen`、legacy 树形简历工坊 dead code（约 570 行）与报告死组件；`jd_match`/`company_intel` route 归一；导航收敛 4 项；简历工坊重写为平铺 + 采纳收口；报告 CTA 单点；设置双 tab；主题四预设；`resumeVersionId` 全量更名 `resumeId`。

成功证据：

| 验证 gate | 结果 |
|-----------|------|
| `ui-design/ui-design-contract.test.mjs`（重写并扩展） | 28/28 PASS |
| esbuild 全量 14 个 JSX 语法编译 | 全部 OK |
| 离线浏览器冒烟（本地 React UMD + esbuild 预编译临时 harness） | 导航 4 项、`#jd_match`→home、`#company_intel`→workspace、简历平铺列表、采纳→预览并保存→另存为新简历全链路、报告唯一 Header CTA + 复练计划无重复按钮 + 题目回顾标记不开练、设置仅双 tab、主题菜单四预设无滑杆、parse 启动决策完整，全部通过 |
| `sync-doc-index --check` | 零漂移 |
| 设计层旧口径负向搜索（jd_match / company_intel / 主版本 / 轻量问答 / customAccent / resumeVersionId 等） | 零残留（负向规范文档除外） |

## 2 会话中的主要阻点/痛点

1. **esbuild 验证假阳性后被退出码复检纠正**：首轮语法检查把 stderr 重定向丢弃，`npx esbuild` 实际报 `command not found` 却显示全部 OK；复检改用退出码才暴露 ui-design 目录没有可解析的 esbuild，需要用仓库根 `node_modules/.pnpm/node_modules/.bin/esbuild`。
   - **证据**：会话中两次 esbuild 运行输出对比。
   - **影响**：若未复检，14 个 JSX 的“语法验证”将是虚假证据；离线冒烟 harness 也因此多一轮探路（React UMD 路径同样靠 find 摸索）。
2. **首批漏网口径靠本批语义负向搜索补获**：`ui-architecture.md` §6.3 用户菜单页面列表仍残留 `AuthReset`（首批 D-16 漏网）；`removed-modules` §5 “替代归属”块、profile 页“来自保存搜索 / 关注岗位”来源标签等岗位推荐语义关联词，均不在模块名关键词表内，需要多轮扩词扫描才找全。
   - **证据**：本批 grep 命中清单与修复 diff。
   - **影响**：每轮裁剪都要重新构造负向词表，残留发现依赖人工扩词。
3. **会话级中断三次（Bash 安全分类器临时不可用 ×1、输出 token 上限 ×2）**：均顺利恢复（改用 Read+Edit 替代 Bash、续写中断点），未造成返工，但拉长了交付节奏。
   - **证据**：会话事实。
   - **影响**：轻微；无仓库改动需求。

## 3 根因归类

| # | 现象 | 根因 | 归类 |
|---|------|------|------|
| 1 | esbuild 假阳性 / 工具路径摸索 | 首批复盘已建议的“ui-design 本地验证 README”仍未落地：契约测试命令、esbuild 真实路径、离线 smoke harness 搭法、hash 路由需整页刷新等知识只存在于复盘报告中，每批都重新探路 | README（ui-design） |
| 2 | 首批漏网口径与语义关联词残留 | 负向搜索词表会话内临时构造、不沉淀；本批已把第二批口径固化为 28 项契约断言（含 jd_match / customAccent / resumeVersionId / 拒绝按钮等负向断言），后续回归由测试承接 | 无需仓库改动（本批已沉淀） |
| 3 | 会话中断 | 平台环境因素 | 无需仓库改动 |

## 4 对流程资产的改进建议

1. **README（ui-design）**：新建 `ui-design/README.md`（或扩充 `docs/ui-design/README.md` §1），固化本地验证手册：`node --test ui-design-contract.test.mjs`、esbuild 路径 `node_modules/.pnpm/node_modules/.bin/esbuild`（ui-design 目录内 `npx esbuild` 不可用）、离线冒烟 harness 搭法（本地 React UMD + 预编译 JS + 按 index.html 顺序加载）、原型 hash 路由按加载时解析（验证不同 route 必须整页刷新并带查询参数防缓存）。该建议在首批复盘中为 P1，本批再次付出重复摸索成本，应升级优先级。
2. **spec-plan（下游 handoff）**：本批裁剪命中大量已完成实现，下游 owner 必须原地修订而非新建 sibling：
   - `frontend-home-job-picks-and-parse`（jd_match plan 002 范围删除）与正式前端 jd_match 代码；
   - `backend-jobs-recommendations` owner spec 与 backend jd-match 实现、`openapi` jobmatch tag（12 个 operation）、相关 events/migrations——属于接口与数据结构删改，按 AGENTS.md §4.1 须先与用户对齐删除节奏（立即删除 vs 标记 deprecated）；
   - `frontend-resume-workshop` / `backend-resume`：版本树与 tailor 契约改为平铺资产，`resumeVersionId`→`resumeId` 为 B2 破坏性变更，需走 openapi ADR/diff gate；
   - `frontend-report-dashboard`（CTA 单点）、`frontend-shell`（设置占位 tab、主题 accent、导航四项）；
   - `engineering-roadmap` D-12 行的 Job Picks 口径。
3. **skill（sync-doc-index）**：首批建议的 `docs/ui-design/INDEX.md` 纳入脚本扫描范围仍未落地，本批 13 行投影继续手工维护；维持 P2 建议不变。

## 5 建议优先级与后续动作

| 优先级 | 动作 | 说明 |
|--------|------|------|
| P0 | 与用户对齐岗位推荐 / 简历版本契约的下游删改节奏，然后逐 owner 经 `/change-intake` 原地修订 spec/plan 并 `/implement` | 设计真理源已就位；本批与首批不同，命中多个已完成的前后端实现与 OpenAPI/migrations 契约，删改属于架构级决策，不应由 Agent 单方面排序 |
| P1 | ui-design 本地验证 README（建议 1，首批 P1 复提升级） | 两批会话均为同一摸索付费，一次性文档补充即可消除 |
| P2 | sync-doc-index 覆盖 docs/ui-design（建议 3，首批 P2 维持） | 治理增强，非阻塞 |
