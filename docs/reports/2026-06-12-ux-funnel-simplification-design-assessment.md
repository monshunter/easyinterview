# UX Funnel Simplification 设计层交付复盘

> **日期**: 2026-06-12
> **审查人**: Claude

## 1 复盘范围与成功证据

本次会话完成 UX 冗余裁剪首批组合方案的设计层交付（commit `7e6ec756`，分支 `design/ux-funnel-simplification`）：

- **审查输入**：以 product-scope spec 的 P0 最短路径判据 + “如无必要勿增实体”完成全流程走查，产出 10 项冗余清单；用户批准首批 4 项（JD 导入漏斗合并·方案 A、复盘上下文选一带二、删除 `auth_reset`、删除 `ThankYouLetter` 死代码）。
- **文档修订**：`product-scope/spec.md` v1.9（新增 D-14/D-15/D-16 锁定决策与 C-14/C-15 验收场景）；`docs/ui-design/` 9 份文档原地修订并同步两处 INDEX 投影。
- **原型修订**：ParseScreen 收拢启动决策、workspace 回访枢纽化、debrief `applyContextSelection` 联动、AuthResetScreen/ThankYouLetter 删除、设置页无密码口径、canvas 移除 auth-reset 画板帧。

成功证据：

| 验证 gate | 结果 |
|-----------|------|
| `ui-design/ui-design-contract.test.mjs`（新增 4 项断言） | 20/20 PASS |
| esbuild 全量 14 个 JSX 语法编译 | 全部 OK |
| 离线浏览器冒烟（本地 React UMD + esbuild 预编译临时 harness） | parse 启动元素、立即面试登录拦截 + pendingAction 恢复直达 practice、`#auth_reset` 归一登录、debrief 选一带二出现 2 个 AUTO-FILLED 徽标，全部通过 |
| `sync-doc-index --check` | 零漂移 |
| 设计层旧口径负向搜索（忘记密码/两步验证/二次确认按钮/感谢信） | 零残留（负向规范文档除外） |

## 2 会话中的主要阻点/痛点

1. **浏览器冒烟被 CDN 依赖阻断**：原型 `index.html` 依赖 unpkg 的 react/babel-standalone，沙箱浏览器无法访问外网（`ERR_CONNECTION_CLOSED`），首轮冒烟全 false 且无显式报错定位；最终靠本地 `node_modules` 的 React UMD + 仓库内 esbuild 预编译搭建临时离线 harness 才完成运行时验证。
2. **原型 hash 路由不响应 hashchange**：`open ...#debrief` 在同页 hash 变化时不重新路由，导致两轮断言误判（先误判为渲染失败，后停留在 practice 页）；需加查询参数强制刷新。该行为与正式前端的 URL-addressable routing（frontend-shell 004）不一致，属原型已知局限但无文档提示。
3. **INDEX 投影历史漂移**：`docs/ui-design/INDEX.md` 的 auth-and-entry 行停留在 1.10/2026-05-02，而文件实际已是 1.14/2026-05-28，说明此前某次修订未同步该 INDEX；本次顺带修复。
4. **残留口径靠多轮负向搜索才找全**：设置页“密码/两步验证”行、canvas `auth-reset` 画板帧均不在首轮改动清单内，是零残留搜索补出的；契约测试原有断言未覆盖“无密码”口径。

## 3 根因归类

| # | 现象 | 根因 | 归类 |
|---|------|------|------|
| 1 | CDN 阻断浏览器验证 | 原型运行时依赖外网 CDN，仓库无离线验证路径说明 | README（ui-design） |
| 2 | hash 路由误判 | 原型按加载时 hash 路由、无 hashchange 监听，文档未提示验证方式 | README（ui-design） |
| 3 | INDEX 历史漂移 | 此前修订遗漏 `docs/ui-design/INDEX.md`（该目录不在 `sync-doc-index` 脚本覆盖范围，脚本只扫 `docs/spec/*`） | skill（sync-doc-index） |
| 4 | 残留口径分散 | 设计裁剪类变更天然跨文件，现有契约断言只覆盖已知冗余 | no repo change needed（本次已沉淀新负向断言） |

## 4 对流程资产的改进建议

1. **README（ui-design）**：在 `ui-design/` 增补“本地验证”说明——契约测试命令、原型 hash 路由按加载时解析（变更 hash 需强制刷新）、离线环境可用本地 React UMD + esbuild 预编译替代 CDN。目标资产：`ui-design/README.md`（如无则新建）或 `docs/ui-design/README.md` §1。
2. **skill（sync-doc-index）**：评估把 `docs/ui-design/INDEX.md` 纳入脚本扫描范围（其 Header/INDEX 结构与 docs/spec 相同），消除该目录投影漂移盲区。目标资产：`.agent-skills/sync-doc-index/scripts/sync-doc-index.py`。
3. **spec-plan（下游 handoff）**：正式前端落地需先原地修订 4 个 owner 的 spec/plan（frontend-home parse、frontend-workspace-and-practice、frontend-debrief、frontend-shell auth），其中 frontend-shell 已盘点 11 个 `auth_reset` 引用文件；plan 修订时必须引用 product-scope D-14/D-15/D-16 与本次更新的 UI 文档版本号。

## 5 建议优先级与后续动作

| 优先级 | 动作 | 说明 |
|--------|------|------|
| P0 | 经 `/change-intake` 或 `/implement` 进入 frontend owner plan 原地修订与实施 | 设计真理源已就位，正式前端是当前唯一未闭环层；建议从 frontend-shell auth_reset 清理（范围最小、已有引用清单）或 frontend-home/parse（用户价值最大）启动 |
| P1 | ui-design 本地验证说明（建议 1） | 一次性文档补充，消除下次原型验证的重复摸索 |
| P2 | sync-doc-index 覆盖 docs/ui-design（建议 2） | 治理增强，非阻塞 |

本批未采纳但已记录的其余 6 项冗余建议（报告 CTA 三处重复、岗位推荐联网搜索 tab 超 MVP 口径、设置占位 tab、公司情报独立页提前占位、画像页双呈现、主题过度配置）留待用户后续决策，不随本批自动进入计划。
