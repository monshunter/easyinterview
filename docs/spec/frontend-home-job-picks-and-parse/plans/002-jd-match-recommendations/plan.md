# 002 JD Match Recommendations (Recommended / Search / Watchlist)

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-09

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

在 plan `001-home-jd-import-and-parse` 已交付的 jd_match P1 placeholder shell 基础上，把 `ui-design/src/screen-jd-match.jsx` 完整三 tab（Recommended / Search / Watchlist）+ Profile snapshot chip 数据驱动 + AGENT scan status badge + Save / Mark not relevant / Confirm interview / Open source 闭环 + 自然语言 Search + Saved searches + Watchlist + Market signals 一次性源级复刻到正式 frontend；通过 OpenAPI `JobMatch` tag + 12 operationId + fixture-backed transport 完成 P0 UI/BDD 闭环；为未来 `backend-jobs-recommendations` subspec 提供清晰的契约 handoff。

完成本计划后，用户在 frontend dev server 上能够：

1. 进入 `jd_match` 路由看到 Hero + Profile snapshot chip（驱动来源 `getJobMatchProfile`）+ AGENT ACTIVE badge（驱动来源 `getAgentScanStatus`）+ 三 tab 标签（带数量 badge）
2. Recommended tab：JobMatchCard 列表（驱动来源 `listJobRecommendations`）+ 右侧 sticky JDDetail（reasons / risks / highlights / INTEL 条件渲染 / Action bar 4 button）+ Save / Unsave + Mark not relevant + Confirm interview → parse + Open source new tab
3. Search tab：自然语言搜索框 + Run web search → `searchJobs`；in-flight 期间渲染单一动态加载文案直到 results / failure；Saved searches 网格（list + create）；Results 2 列网格 + 4 个 chip filter 纯 client-side 切换
4. Watchlist tab：列表（驱动来源 `listWatchlist`）+ Market signals 4 卡（驱动来源 `getMarketSignals`）+ chevron 切回 Recommended tab 并 selected = `WatchlistItem.linkedJobMatchId`
5. 全部用户行为通过 generated `JobMatch` client + fixture-backed mock transport 闭环；所有 side-effect 操作（add / remove / dismiss / search / createSavedSearch）携带 `Idempotency-Key`；所有隐私字段（query / saved-search label / watchlist label / sourceJobUrl / freeNote / linkedJobMatchId）不进 console / URL / localStorage / telemetry；i18n zh/en 完整切换；warm/light + dark + customAccent 三态可见变化；desktop + mobile pixel parity 通过

## 2 背景

`frontend-home-job-picks-and-parse` spec v1.1 已通过 D-1 决策把 jd_match 完整三 tab 业务锁给本 plan。plan 001（completed 2026-05-08）落地了 home + parse + jd_match P1 placeholder shell（E2E.P0.014/015/016/017）。`ui-design/src/screen-jd-match.jsx`（652 行）早已落地完整三 tab + Save/Mark not relevant/Saved searches/Agent active status/Market signals/Network intel 全部交互。

仓库当前没有任何 backend recommendations 实现 / OpenAPI 契约 / fixture / prompt baseline / migration。spec v1.2 D-1 把 jd_match 业务范围改写为「契约先行 + frontend fixture 消费」模式，与 plan 001 D-2 generated client + fixture-backed transport 模式一致：本 plan 在真实 backend handler 落地之前先扩展 OpenAPI `JobMatch` tag + 12 operationId + fixture，frontend 通过 generated client 一次性源级复刻 ui-design 三 tab 完整业务；真实 backend handler / service / store / agent scan pipeline / 真实联网搜索 / 候选池抓取由独立未来 subspec `backend-jobs-recommendations` 承接，本 plan 不依赖也不实现真实 backend。

ui-design 真理源由用户独立维护：用户已（或将）独立完成 `ui-design/src/screen-jd-match.jsx` SearchTab `searching=true` 区域简化（从 5 步 AGENT panel 改为单一动态加载文案），plan 002 不修改 ui-design 静态文件，所有 frontend 复刻以届时 ui-design 真理源为准。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior（用户可感知 UI + API 行为 + 业务流程 + 端到端功能）
- **TDD 策略**: Red-Green-Refactor 入口为 `pnpm --filter @easyinterview/frontend test`（Vitest）；每个 Phase 在新增组件前先写失败测试，覆盖 DOM 锚点、控件类型、props/state、generated client 调用断言（含 request body schema、`Idempotency-Key` header、URL/state 隐私反查）；`pnpm --filter @easyinterview/frontend test:pixel-parity` 在 Phase 6 升级 `frontend/tests/pixel-parity/jd_match.spec.ts` 覆盖三 tab × desktop + mobile × warm/dark/customAccent；新增组件文件位于 `frontend/src/app/screens/jd_match/` 子目录；测试文件与组件 colocate（`*.test.tsx`）。
- **BDD 策略**: Feature plan requires BDD；本 plan 在 `bdd-plan.md` 定义 5 个场景 `E2E.P0.022 / E2E.P0.023 / E2E.P0.024 / E2E.P0.025 / E2E.P0.026`，`bdd-checklist.md` 跟踪每个场景资产创建与执行；主 `checklist.md` 在每个 Phase 末尾保留 `BDD-Gate:` 项引用对应场景 ID；`p0-017-jd-match-placeholder` scenario 同步升级 verify 脚本（不再 grep placeholder，断言三 tab 数据驱动渲染）。
- **替代验证 gate**: 不适用（feature plan，已有完整 BDD + TDD 双层覆盖）

## 3.5 Coverage Matrix

> 行号引用 `ui-design/src/screen-jd-match.jsx` 截止 2026-05-09 的版本；用户独立修订 SearchTab `searching=true` 区域后行号会偏移，以届时真理源为准。

| 类别 | 覆盖描述 | UI Source Anchor | Phase | 验证入口 |
|------|----------|------------------|-------|---------|
| Primary path | 进入 jd_match → Hero/Profile/AGENT 渲染 → Recommended 列表 → JDDetail 详情 → Confirm interview → `nav("parse", { source: "jd_match", sourceJobMatchId })` | `screen-jd-match.jsx::JDMatchScreen` line 4 + `JobMatchCard` line 335 + `JDDetail` line 377 Action bar | 2+3 | E2E.P0.022 + Vitest `jd_match/RecommendedTab.test.tsx` + `jd_match/JDDetail.test.tsx` |
| Alternate · Save/Unsave | 点击 Save → `addToWatchlist` + Idempotency-Key + window.eiToast；再点 → `removeFromWatchlist` + toast | `JDMatchScreen::toggleSaved` line 181 | 3 | Vitest + E2E.P0.022 sub-case |
| Alternate · Mark not relevant | `markJobNotRelevant` + Idempotency-Key + 隐藏卡片 + 自动选下一张 + toast | `JDMatchScreen::markNotRelevant` line 196 | 3 | Vitest + E2E.P0.022 sub-case |
| Alternate · Open source | `window.open(url, "_blank", "noopener,noreferrer")` | `JDMatchScreen::openSource` line 213 | 3 | Vitest spy + E2E.P0.022 sub-case |
| Alternate · Search query + filter | 输入 query → Run → `searchJobs` + Idempotency-Key；4 chip filter 纯 client-side 切换 | `JDMatchScreen::runSearch` line 173 + `filterPredicate` line 226 | 4 | Vitest + E2E.P0.023 |
| Alternate · Saved search create | `saveCurrentAsWatch` → `createSavedSearch` + Idempotency-Key + toast | `JDMatchScreen::saveCurrentAsWatch` line 219 | 4 | Vitest + E2E.P0.023 |
| Alternate · Watchlist chevron handoff | `openJob(id)` → tab 切 Recommended + selected = `WatchlistItem.linkedJobMatchId` | `JDMatchScreen::openJob` line 235 + `WatchlistTab::findJobIdFor` line 594 | 5 | Vitest + E2E.P0.024 |
| Alternate · Auth pending action | 未登录 side-effect 操作触发 `requestAuth({ type: "jd_match_action", route: "jd_match", params: { tab, selectedJobMatchId, action } })`；登录后回到 jd_match 自动重新触发 | D1 `requestAuth` + plan 001 C-7 模式 | 3+4+5 | Vitest `jd_match/JDMatchAuthGate.test.tsx` + E2E.P0.025 |
| Failure / recovery · listJobRecommendations 4xx/5xx | failed variant → 错误占位 + retry button | n/a (error state) | 3 | Vitest + fixture variant |
| Failure / recovery · addToWatchlist 4xx revert | save 调用失败 → button 状态 revert + error toast；卡片不进入 watchlist | n/a | 3 | Vitest + E2E.P0.022 sub-case |
| Failure / recovery · markJobNotRelevant 4xx revert | dismiss 失败 → 卡片重新显示 + error toast | n/a | 3 | Vitest + E2E.P0.022 sub-case |
| Failure / recovery · searchJobs failed / slow-response | failed → inline error + 保留输入；slow-response → 单一加载文案持续显示直到 settle，不超时 | n/a | 4 | Vitest + E2E.P0.023 |
| Failure / recovery · createSavedSearch 4xx | 失败 → toast + 不写入 saved searches | n/a | 4 | Vitest + E2E.P0.023 |
| Failure / recovery · getMarketSignals partial-data | 部分 signal 缺失 → 渲染部分卡 + 缺失值 fallback | n/a | 5 | Vitest + E2E.P0.024 |
| Boundary · empty recommendations | `listJobRecommendations` 返回 [] → empty state DOM + CTA「调整搜索 / 完善画像」 | n/a | 2+3 | Vitest fixture variant |
| Boundary · empty saved searches | `listSavedSearches` 返回 [] → empty grid + 仅 Save current 按钮可见 | n/a | 4 | Vitest |
| Boundary · empty watchlist | `listWatchlist` 返回 [] → empty state + 引导「去 Recommended 添加」 | n/a | 5 | Vitest + E2E.P0.024 |
| Boundary · 12+ items 列表 cap | Recommended 列表展示全部不 cap（按 backend cursor），Search results 2 列网格 cap 6 | `screen-jd-match.jsx::SearchTab` results | 3+4 | Vitest |
| Boundary · unknown linkedJobMatchId | chevron handoff 时 `linkedJobMatchId` 不在当前 Recommended 列表 → tab 切但 selected fallback 第一张 + warning toast | `WatchlistTab::findJobIdFor` line 594 | 5 | Vitest + E2E.P0.024 |
| Preflight · UI truth source | `ui-design/src/screen-jd-match.jsx::SearchTab` 当前 `searching=true` 区域必须已是单一动态加载文案；若仍是 5 步 AGENT panel，停止 frontend 实施并先修订 ui-design/docs-ui truth source | `ui-design/src/screen-jd-match.jsx` + `docs/ui-design/` | 0 | source grep + `/design` handoff when needed |
| Preflight · B2 owner inventory | 新增 `JobMatch` tag 前，B2 owner truth source 与 validators 必须从 12 tag / 34 endpoint additive 升级到 13 tag / 46 endpoint；不得让旧 34-operation gate 继续作为 `make validate-fixtures` 完成标准 | `openapi-v1-contract` + inventory/fixture validators | 0+1 | contract lint + validator unit test |
| Cross-layer contract · 12 operation schema | 全部 12 operationId 在 `openapi/openapi.yaml` 声明 schema 完整、tag=`JobMatch`；codegen drift 0 | OpenAPI `JobMatch` tag | 1 | contract test + typecheck + `make validate-fixtures` |
| Cross-layer contract · Idempotency-Key 5 操作 | `addToWatchlist` / `removeFromWatchlist` / `markJobNotRelevant` / `searchJobs` / `createSavedSearch` request 必带 `Idempotency-Key` header | OpenAPI shared `Idempotency-Key` parameter | 3+4 | Vitest request header 反查 |
| Cross-layer contract · GenerationProvenance 必填 | AI 生成字段（score / reasons / risks / highlights / hypotheses / interviewHypotheses）必带 provenance；缺失时进入降级展示，不允许前端推断 | OpenAPI `GenerationProvenance` | 3 | Vitest |
| Cross-layer contract · fixture variant 完整 | 12 operation 各自 fixture 至少 default + 一个 edge variant；通过 `mock-contract-suite` per-operation per-scenario 切换 | `openapi/fixtures/JobMatch/*.json` | 1 | `make validate-fixtures` + mock-contract-suite parity test |
| Cross-layer contract · WatchlistItem.linkedJobMatchId | backend 必填字段；frontend chevron handoff 直接读取，不 string match | OpenAPI `WatchlistItem` schema | 5 | Vitest |
| Privacy / security · query zero-leak | 自然语言搜索 query 不进 console / URL / localStorage / telemetry；mockTransport 仅记 call status | n/a | 4 | redact lint + Vitest spy + E2E.P0.023 |
| Privacy / security · saved-search label / watchlist label / sourceJobUrl / freeNote / linkedJobMatchId zero-leak | 全部隐私字段不进 logger / URL / localStorage / telemetry payload；fixture redact lint 必须覆盖 | n/a | 3+4+5 | redact lint + Vitest + 5 个 scenario verify |
| Privacy / security · window.open noopener,noreferrer | Open source 调用必须带这两个 flags 防止 reverse tabnabbing | `JDMatchScreen::openSource` line 213 | 3 | Vitest spy |
| Privacy / security · auth gate | 未登录 API side-effect 操作触发 `requestAuth` + opaque pendingAction；Source button 与 Watchlist chevron 不触发 pendingAction；query / label 等不进 pendingAction params | D1 `requestAuth` + plan 001 模式 | 3+4 | Vitest + E2E.P0.025 |
| Observability | mockTransport call audit；只允许 generated `JobMatch` client + existing shell runtime/auth；不允许前端直连 LLM/provider/外部招聘平台；进入 jd_match 调用次数与 fixture 期望一致 | n/a | 6 | scenario verify + Vitest spy |
| UX · loading state | Recommended/Search/Watchlist/MarketSignals/Profile/AgentStatus 6 处独立 loading；Search loading 是单一动态加载文案持续到 settle | n/a | 2-5 | Vitest |
| UX · empty state | empty recommendations / empty saved searches / empty watchlist / empty market signals 各自独立 empty 文案 | n/a | 2-5 | Vitest |
| UX · error state | 各 operation 4xx/5xx 各自独立 error 占位（保留入口、可重试）；revert 行为锁定 | n/a | 2-5 | Vitest |
| UX · i18n zh/en | 全文案通过 typed helper；切换立即重绘；删除 plan 001 placeholder key 并新增 5 子命名空间 `jdMatch.recommended.*` / `jdMatch.search.*` / `jdMatch.watchlist.*` / `jdMatch.profile.*` / `jdMatch.agent.*` | D1 typed locale helper | 2-5 | Vitest `i18n` namespaces test |
| UX · dark + customAccent | 三 tab 三态切换关键元素 computed 颜色变化（Hero / Profile chip / JobMatchCard score 三档着色 / JDDetail reasons-risks-INTEL 区配色 / chip filter active / WatchlistItem 3 tone / Market signals） | D2 `data-theme` / `data-mode` / `data-custom-accent` | 6 | Playwright + Vitest computed style |
| UX · responsive layout | mobile (390×844) viewport 下 Recommended 双列折单列、Saved searches 3 列折单列、Market signals 4 列折 2 列、Search results 2 列折单列；JDDetail sticky 在 mobile 改为流式渲染（不 sticky） | n/a | 6 | Playwright mobile project |
| UI source structure parity · JDMatchScreen 主壳 | Hero (label/title/sub) + Profile snapshot chip (avatar/SEARCHING AS/skills tags/PROFILE SOURCES/AGENT ACTIVE) + 三 tab 标签（带数量 badge） | `screen-jd-match.jsx::JDMatchScreen` line 4-170 | 2 | Vitest + testid `jdmatch-hero-{label,title,sub}` / `jdmatch-profile-chip-{avatar,searching-as,skills,sources}` / `jdmatch-agent-status-{badge,last-scan,next-scan}` / `jdmatch-tab-{recommended,search,watchlist}` |
| UI source structure parity · JobMatchCard | 左 3px 边框按 score 三档着色 + 标题 + 未看过点 + saved pin + company/companyTag + location/comp + 右上 score 32px + STRONG/GOOD/STRETCH label + Top reason + 底部 must X/Y · plus X/Y · posted · source | `screen-jd-match.jsx::JobMatchCard` line 335 | 3 | Vitest + testid `jdmatch-card-${id}` + `jdmatch-card-${id}-score` + `jdmatch-card-${id}-top-reason` 等 |
| UI source structure parity · JDDetail 6 段 | Header (company tag/title/level-location-comp tag/right 36px score) + + WHY IT MATCHES (绿 check) + ⚠ WHERE IT'S A STRETCH (橙 info) + ROLE SNAPSHOT (列表) + INTEL (条件渲染 networkNote/similarInterviewers，合规措辞「公开面经 / JD / 公司资料信号」) + Action bar (Confirm interview / Save toggle / Source / Not relevant) | `screen-jd-match.jsx::JDDetail` line 377 | 3 | Vitest + testid `jdmatch-detail-{header,why,risk,snapshot,intel,action-{confirm,save,source,dismiss}}` + ui-design-contract.test.mjs 不破坏 |
| UI source structure parity · SearchTab | 自然语言搜索框 + Run web search 按钮 + 数据源 tag 列 + 单一动态加载文案区（用户简化后；以届时 ui-design 真理源为准）+ Saved searches 3 列网格（list + create，无 delete）+ Results 2 列网格 cap 6 + 4 chip filter (all/strong/remote/unseen) | `screen-jd-match.jsx::SearchTab` line 480 | 4 | Vitest + testid `jdmatch-search-{input,run,sources,searching,saved-grid,results,filter-${k}}` |
| UI source structure parity · WatchlistTab | WatchlistItem 列表（3px 左边框按 tone 着色 ok/warn/muted、added 时间、change 信号、chevron）+ Market signals 4 列 grid（k/v/d/tone 4 卡）+ refresh footer 文案 | `screen-jd-match.jsx::WatchlistTab` line 592 | 5 | Vitest + testid `jdmatch-watchlist-item-${id}` + `jdmatch-market-signal-${k}` + `jdmatch-watchlist-refresh-footer` |
| UI visual geometry parity · desktop | 1440×900 三 tab 主区块 bounding box stays in viewport, no overlap；Recommended 1.1fr/1.4fr 双列 grid；Saved searches `repeat(3, 1fr)`；Market signals `repeat(4, 1fr)`；Search results 2 列 grid | n/a | 6 | Playwright desktop project |
| UI visual geometry parity · mobile | 390×844 折叠：Recommended 双列折单列；Saved searches 折单列；Market signals 折 2 列；Search results 折单列；JDDetail sticky 改流式 | n/a | 6 | Playwright mobile project |
| UI visual geometry parity · dark / customAccent | warm/light → dark → customAccent 三态切换：score 三档色、left border 着色、chip filter active 状态、AGENT badge tone、Watchlist 3 tone、Market signals tone 全部出现可见变化 | n/a | 6 | Playwright |
| UI visual geometry parity · screenshot regression | toHaveScreenshot baseline maxDiffPixels 阈值（沿用 D3 经验，不与 ui-design golden 跨字体源做硬 diff） | n/a | 6 | Playwright + baseline |
| UI stale-contract negative · plan 001 placeholder | `jdmatch-placeholder` / `jdmatch-placeholder-cta` / `placeholder.copy` / `placeholder.cta` / "Coming in P1" / "Coming Soon" / `jdMatch.placeholderTitle` 等 plan 001 placeholder 文案与 testid 在 `frontend/src/` runtime 0 命中（除负向断言文件） | n/a | 6 | grep negative |
| UI stale-contract negative · 5 步 AGENT panel DOM | ui-design 简化前的 5 步步骤文案（向量化你的画像 / 并行查询 LinkedIn 等）与 opacity 切换 DOM 在 frontend SearchTab 0 命中 | n/a | 4+6 | grep negative + Vitest |
| UI stale-contract negative · 旧 route alias | `welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star` / 独立 `voice` route 在 jd_match 新代码中不出现 | n/a | 全 phase | Vitest + scenario verify negative grep |
| Regression / legacy-negative · D1+D2+D3+plan001 现有 gate | E2E.P0.001/002/004/005/006/014/015/016/017 重跑通过 | n/a | 6 | scenario rerun |
| Regression / legacy-negative · 不直接 import prototype data | `frontend/src` 不 import `ui-design/src/data.jsx` 或 `window.EI_DATA`；JobMatchScreen 不读取硬编码 `decoratedJobs` / `savedSearches` mock | n/a | 全 phase | Vitest + tsc grep negative |
| Regression / legacy-negative · 不直接调用 LLM/provider/外部招聘平台 | `frontend/src` 不出现 LinkedIn / Boss / 脉脉 / 拉勾 SDK、AI provider key、provider registry、prompt registry、AIClient、LLM endpoint 或 ad hoc fetch；只允许 generated `JobMatch` client / fixture transport / D1 已有 generated client | n/a | 全 phase | Vitest + grep negative |

### 高风险类别 N/A 说明

- 无高风险类别整体 N/A；本 plan 覆盖 primary / alternate / failure / boundary / cross-layer / privacy / observability / UX / UI source / visual geometry / regression-negative 全部类别

## 3.6 Frontend / Backend Operation Matrix

本 plan 走 `docs/development.md` §2.2 Frontend-First Path：正式前端先对齐 `ui-design/` 并通过 generated `JobMatch` client + fixture-backed transport 完成 P0 UI/BDD；真实 handler / store / agent scan pipeline / 真实联网搜索 / 候选池抓取 / market signals 计算由独立未来 subspec `backend-jobs-recommendations` 落地前，以下 backend 状态保持 `not-yet-implemented`，不得把 fixture PASS 宣称为真实 backend 闭环。

| operationId | path / method | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getJobMatchProfile` | `GET /jd-match/profile` | `openapi/fixtures/JobMatch/getJobMatchProfile.json` scenarios: `default` / `unauthenticated` / `partial-profile` | `JDMatchScreen` `useJobMatchProfile()`；进入 jd_match 调一次；为空时 chip 显示 fallback | `not-yet-implemented`；owned by `backend-jobs-recommendations` | backend: `user_profiles` aggregated view | none in frontend；backend 可能聚合 resume + targetJob 历史 | E2E.P0.022, E2E.P0.025 |
| `getAgentScanStatus` | `GET /jd-match/agent-status` | `openapi/fixtures/JobMatch/getAgentScanStatus.json` scenarios: `idle` / `scanning` / `error` / `next-scan-soon` | `JDMatchScreen` `useAgentScanStatus()`；进入 jd_match + 切回 Recommended tab 各调一次；不引入 SSE/WebSocket/真实定时器 | `not-yet-implemented`；owned by `backend-jobs-recommendations` | backend: agent_scan_runs | none in frontend | E2E.P0.022, E2E.P0.025 |
| `listJobRecommendations` | `GET /jd-match/recommendations` | `openapi/fixtures/JobMatch/listJobRecommendations.json` scenarios: `empty` / `one` / `many` / `failed` | `RecommendedTab` `useJobMatchRecommendations()`；query 支持 cursor pagination；返回 `JobMatchRecommendation` 数组（含 score / reasons / risks / highlights / fit / posted / source / saved / seen / similarInterviewers / networkNote 等） | `not-yet-implemented`；owned by `backend-jobs-recommendations` | backend: job_match_recommendations + 关联表 | backend-only score / reasons / risks 生成（GenerationProvenance 必带）；frontend 仅展示 | E2E.P0.022 |
| `getJobRecommendation` | `GET /jd-match/recommendations/{jobMatchId}` | `openapi/fixtures/JobMatch/getJobRecommendation.json` scenarios: `default` / `network-intel-empty` / `failed` | `JDDetail` `useJobRecommendation(jobMatchId)`；用户切卡时必须调；`listJobRecommendations` 只作为列表摘要来源，不作为详情 operation 的替代证据 | `not-yet-implemented`；owned by `backend-jobs-recommendations` | backend: 同上 | 同上 | E2E.P0.022 |
| `addToWatchlist` | `POST /jd-match/watchlist` | `openapi/fixtures/JobMatch/addToWatchlist.json` scenarios: `default` / `4xx-validation` / `5xx-server-error` | `useToggleWatchlist()`；body `{ jobMatchId }`；带 `Idempotency-Key`；4xx → revert + error toast | `not-yet-implemented`；owned by `backend-jobs-recommendations` | backend: watchlist_items + idempotency_keys | none | E2E.P0.022 |
| `removeFromWatchlist` | `DELETE /jd-match/watchlist/{jobMatchId}` | `openapi/fixtures/JobMatch/removeFromWatchlist.json` scenarios: `default` / `4xx-not-found` | `useToggleWatchlist()`；带 `Idempotency-Key`；4xx → revert + error toast | `not-yet-implemented` | backend: 同上 | none | E2E.P0.022 |
| `markJobNotRelevant` | `POST /jd-match/recommendations/{jobMatchId}/dismiss` | `openapi/fixtures/JobMatch/markJobNotRelevant.json` scenarios: `default` / `4xx` | `useDismissRecommendation()`；body `{ reason: enum, freeNote?: string }`；reason enum 默认值由 frontend 透传，ui-design 不弹窗收集；带 `Idempotency-Key`；4xx → revert + error toast | `not-yet-implemented`；owned by `backend-jobs-recommendations` | backend: dismissed_recommendations | backend-only learning loop；frontend 仅发送 | E2E.P0.022 |
| `searchJobs` | `POST /jd-match/search` | `openapi/fixtures/JobMatch/searchJobs.json` scenarios: `default` / `empty` / `failed` / `slow-response` | `SearchTab` `useSearchJobs()`；body `{ query, filters?, profileSnapshot? }`；带 `Idempotency-Key`；in-flight 期间渲染单一动态加载文案；slow-response variant 验证文案持续渲染至 settle；query 不进 URL/localStorage/telemetry | `not-yet-implemented`；owned by `backend-jobs-recommendations` | backend: search_runs（agent scan 触发）；results pool | backend-only LLM 召回 + 排序 + 解释；frontend 仅展示；response 不再含 `progress` 字段（D-12） | E2E.P0.023 |
| `listSavedSearches` | `GET /jd-match/saved-searches` | `openapi/fixtures/JobMatch/listSavedSearches.json` scenarios: `default` / `empty` / `4xx` | `SearchTab` `useSavedSearches()`；切 Search tab 调一次 | `not-yet-implemented` | backend: saved_searches | none | E2E.P0.023 |
| `createSavedSearch` | `POST /jd-match/saved-searches` | `openapi/fixtures/JobMatch/createSavedSearch.json` scenarios: `default` / `4xx-validation` | `SearchTab` `useCreateSavedSearch()`；body `{ label, query, filters? }`；带 `Idempotency-Key`；label 不进 URL/localStorage/telemetry；ui-design 无 delete 入口故 plan 002 不实现 `deleteSavedSearch` | `not-yet-implemented` | backend: 同上 | none | E2E.P0.023 |
| `listWatchlist` | `GET /jd-match/watchlist` | `openapi/fixtures/JobMatch/listWatchlist.json` scenarios: `default` / `empty` / `few` / `4xx` | `WatchlistTab` `useWatchlist()`；切 Watchlist tab 调一次；返回 `WatchlistItem[]`，每项必含 `linkedJobMatchId`（chevron handoff 用） | `not-yet-implemented` | backend: watchlist_items | none | E2E.P0.024 |
| `getMarketSignals` | `GET /jd-match/market-signals?window=7d` | `openapi/fixtures/JobMatch/getMarketSignals.json` scenarios: `default` / `partial-data` / `failed` | `WatchlistTab` `useMarketSignals()`；切 Watchlist tab 调一次；返回 4 卡 `{ k, v, d, tone }`；partial-data variant → 缺失 fallback | `not-yet-implemented`；owned by `backend-jobs-recommendations` | backend: market_signal_snapshots | backend-only 计算 | E2E.P0.024 |

## 3.7 JobMatch Frontend View-Model Mapping

正式前端不得从 `ui-design/src/screen-jd-match.jsx` 内嵌的 `decoratedJobs` / `savedSearches` / `watchlist` mock 数据复制粘贴；所有数据均来自 generated `JobMatch` client。Recommended / Search / Watchlist 三 tab 与 chevron handoff 统一使用以下 mapping：

| UI slot / param | Source | Rule |
|-----------------|--------|------|
| Profile chip avatar | `JobMatchProfile.displayName` initials 或 `avatarUrl` | 优先 `avatarUrl`；缺失时 fallback initials；不读取本地图片 |
| Profile chip skills tags | `JobMatchProfile.skills[]` | 直接 map 渲染；超过 6 个折叠 + tooltip |
| Profile chip sources | `JobMatchProfile.sources` aggregated counts | "N resumes · N JDs · N mocks · N debriefs" 文案通过 typed i18n helper |
| AGENT badge status text + tone | `AgentScanStatus.status` enum + `lastScanAt` / `nextScanAt` | idle=neutral, scanning=accent, error=warn；时间用 D1 relative time helper；不在前端硬编码 4h |
| JobMatchCard score color | `JobMatchRecommendation.score` | ≥85 = ok token, ≥70 = warn token, else = ink3；左边框 + 右上 32px 数字同色 |
| JobMatchCard score label | 同 score | "STRONG FIT" / "GOOD FIT" / "STRETCH" 通过 i18n helper |
| JobMatchCard new dot | `JobMatchRecommendation.seen` | `!seen` 时显示 6×6 accent 点 |
| JobMatchCard saved pin | `JobMatchRecommendation.saved` | true 时显示 pin icon |
| JobMatchCard top reason | `JobMatchRecommendation.reasons[0]` | 第一个 reason；不展示其他 |
| JobMatchCard fit footer | `JobMatchRecommendation.fit` | "must X/Y · plus X/Y" |
| JDDetail header score | `JobMatchRecommendation.score` | 36px；"/100" 副文 |
| JDDetail level/location/comp tag | `JobMatchRecommendation.level/location/comp` | 直接 Tag 渲染；缺失时不渲染对应 tag |
| JDDetail Why matches | `JobMatchRecommendation.reasons[]` | 全部 reasons map；绿 check icon |
| JDDetail Where stretch | `JobMatchRecommendation.risks[]` | 全部 risks map；橙 info icon |
| JDDetail Role snapshot | `JobMatchRecommendation.highlights[]` | 项目列表 |
| JDDetail INTEL 区 | `networkNote && similarInterviewers` 条件渲染 | 任一存在则渲染整段；合规措辞按 `ui-design-contract.test.mjs` 锁定的「公开面经 / JD / 公司资料信号」 |
| Save button state | local optimistic + server response | optimistic toggle + 4xx revert |
| Open source button | `JobMatchRecommendation.sourceUrl` | `window.open(url, "_blank", "noopener,noreferrer")`；缺失时 button disabled |
| Confirm interview nav params | `JobMatchRecommendation.id` | `nav("parse", { source: "jd_match", sourceJobMatchId: id })` |
| Search results filter | client-side `JobMatchRecommendation.score / location / seen` | all = 所有；strong = score≥85；remote = location 含 "remote/远程"（不区分大小写 i18n match）；unseen = `!seen` |
| Saved search badge `+N` | `SavedSearch.newJobsCount` | >0 时显示 |
| WatchlistItem tone | `WatchlistItem.tone` enum | ok/warn/muted 各对应 D2 token；3px 左边框同色 |
| WatchlistItem chevron handoff | `WatchlistItem.linkedJobMatchId` | tab 切 Recommended + selected = linkedJobMatchId；list 中找不到时 fallback 第一张 + warning toast |
| Market signal v + d | `MarketSignal.v` + `MarketSignal.d` | v=主值，d=delta；tone 决定 d 颜色 |
| pendingAction params | `{ tab, selectedJobMatchId?, action }` enum | 仅用于 API side-effect：`save` / `unsave` / `dismiss` / `confirm_interview` / `run_search` / `create_saved_search`；不携带 query / label / freeNote / sourceUrl 等隐私字段；Source button 与 Watchlist chevron 不进入 pendingAction |

## 4 实施步骤

### Phase 0: UI truth-source + B2 owner contract preflight

#### 0.1 SearchTab UI truth-source confirmation

读取最新 `ui-design/src/screen-jd-match.jsx` 与 `docs/ui-design/` 对应文档，确认 `SearchTab` `searching=true` 区域已经从 5 步 AGENT panel 改为单一动态加载文案；若仍包含 `Embedding your profile` / `向量化你的画像` / `Querying LinkedIn` / `并行查询` 等步骤文案或 opacity step DOM，则暂停 frontend 实施，先通过 `/design` 或 UI 原型修订把 `ui-design/` 真理源改到 D-12，再继续本 plan。

#### 0.2 B2 OpenAPI owner additive inventory sync

在新增 `JobMatch` tag 前，复核并计划同步 B2 owner truth source：`docs/spec/openapi-v1-contract/spec.md` §2 / §3.1.1 / §6、`openapi/README.md`、`openapi/fixtures/README.md`、`scripts/lint/openapi_inventory.py`、`scripts/lint/validate_fixtures.py`、`docs/spec/mock-contract-suite/spec.md` 与 `docs/spec/engineering-roadmap/spec.md` 中的 12 tag / 34 endpoint / 34-operation coverage 口径。Phase 1 完成后必须 additive 升级为 13 tag / 46 endpoint，且 validator 单测证明旧 34-operation gate 不再误拒 `JobMatch` fixtures。

#### 0.3 Runtime action contract freeze

冻结本 plan 的 12 operationId、5 个 `Idempotency-Key` side-effect operation、`getJobRecommendation` 必调详情契约、pendingAction action enum 与隐私字段负向清单。Source button 只执行 `window.open(..., "noopener,noreferrer")`，Watchlist chevron 只做本地 tab handoff；二者不触发 auth pending action。

#### 0.4 Preflight Gate

- Preflight Gate: Phase 0 evidence log 记录 UI truth-source 当前形态、B2 owner inventory 待同步清单、12 operation matrix freeze 与 auth action enum freeze。

### Phase 1: JobMatch OpenAPI 契约 + fixture + codegen

#### 1.1 新增 `openapi/openapi.yaml` `JobMatch` tag

新增 12 operationId（路径、HTTP 方法、requestBody、responses、Idempotency-Key parameter 与 D-9/D-10 锁定），新增 schemas：`JobMatchProfile`、`AgentScanStatus`、`JobMatchRecommendation`、`WatchlistItem`、`SavedSearch`、`MarketSignal`、`MarkNotRelevantRequest`（含 reason enum）、`SearchJobsRequest`、`SearchJobsResponse`、`AddToWatchlistRequest`、`CreateSavedSearchRequest`、`MarketSignalsResponse`；AI-generated 字段（score / reasons / risks / highlights / interviewHypotheses）schema 必须引用现有 `GenerationProvenance` 必填字段；同步 B2 owner inventory / README / validator / mock-contract coverage 口径到 13 tag / 46 endpoint。

#### 1.2 新增 `openapi/fixtures/JobMatch/` 目录

每个 operation 至少 default + 一个 edge variant，按 §3.6 Operation Matrix 列出的 scenarios 配置；通过 `mock-contract-suite` per-operation per-scenario 切换约定。

#### 1.3 mock-contract-suite 升级

如 `mock-contract-suite` 当前不支持本 plan 需要的 per-operation per-scenario 组合切换，需在本 phase 与对应 spec 同步扩展（保留向后兼容；不破坏 plan 001 fixture variant 切换）。

#### 1.4 codegen 重建

运行 OpenAPI codegen → `frontend/src/api/generated/client.ts` 与 schema typings 重新生成；`pnpm typecheck` PASS；codegen drift check 0。

#### 1.5 Vitest 红灯 → 绿灯

新增 contract test 覆盖 12 operation request body / response schema 与 fixture variant parity；新增 `frontend/src/api/generated/jobMatch.contract.test.ts`（或同等）断言 generated client 可调用。

#### 1.6 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.022 / E2E.P0.023 / E2E.P0.024 / E2E.P0.025 / E2E.P0.026` 资产构建到 ready 态前置 — fixture 存在性、generated client 可调用

### Phase 2: JDMatchScreen 容器升级 + Hero / Profile chip / AGENT badge 数据驱动

#### 2.1 删除 plan 001 placeholder 文案与 testid

删除 `frontend/src/app/screens/jd_match/JDMatchScreen.tsx` 中所有 `jdmatch-placeholder` / `jdmatch-placeholder-cta` testid 与 `jdMatch.placeholderTitle` / `placeholderCopy` / `placeholderCta` i18n key；删除三 tab placeholder 文案 fallback。

#### 2.2 新增 `useJobMatchProfile` / `useAgentScanStatus` hooks

通过 D1 generated client 调 `getJobMatchProfile` 与 `getAgentScanStatus`；React state 跟踪 loading / data / error 三态；Vitest 断言：进入 jd_match 各调 1 次；切回 Recommended tab 时 `getAgentScanStatus` 再调 1 次（不重复调 profile）；切其他 tab 时不调。

#### 2.3 Hero / Profile snapshot chip / AGENT badge 数据驱动

按 `screen-jd-match.jsx::JDMatchScreen` line 4-170 源级复刻 Hero（label/title/sub）+ Profile snapshot chip（avatar / SEARCHING AS / skills tags / PROFILE SOURCES / AGENT ACTIVE 状态）+ 三 tab 标签（带数量 badge）；Profile chip 字段映射严格按 §3.7；AGENT badge tone 与文案按 D-10 + §3.7。

#### 2.4 i18n 命名空间扩展

在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 中：
- 删除 `jdMatch.placeholderTitle` / `placeholderCopy` / `placeholderCta` 三个 key
- 扩展为 5 子命名空间：`jdMatch.profile.*`（≥8 key）/ `jdMatch.agent.*`（≥6 key）/ `jdMatch.recommended.*`（≥10 key）/ `jdMatch.search.*`（≥12 key 含 `searching` 单一加载文案）/ `jdMatch.watchlist.*`（≥10 key）；保留 plan 001 已有的 `jdMatch.hero.*` / `jdMatch.tab.*`；`messages.ts` typed helper 同步
- Vitest `i18n` 套件断言 zh/en 同步无缺漏；删除的 placeholder key 不再被任何代码引用

#### 2.5 Vitest

新增 `jd_match/JDMatchScreen.test.tsx`：测 Hero / Profile chip / AGENT badge / 三 tab 标签 testid 全部命中；测 `getJobMatchProfile` / `getAgentScanStatus` 调用次数与 fixture variant 三态渲染；测 i18n zh/en 切换；测 plan 001 placeholder testid 在 DOM 0 命中；测 `topbar-nav-jd_match` 高亮。

#### 2.6 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.022` Recommended tab 主路径中 Hero/Profile/AGENT 阶段 + `E2E.P0.025` Profile/AGENT/auth 综合场景资产构建到 ready 态

### Phase 3: Recommended tab — JobMatchCard + JDDetail + Save / Mark not relevant / Confirm interview

#### 3.1 新增 `RecommendedTab.tsx` + `JobMatchCard.tsx` + `JDDetail.tsx`

按 `screen-jd-match.jsx::JobMatchCard` line 335 + `JDDetail` line 377 源级复刻；JobMatchCard 三档 score 着色 / 未看过点 / saved pin / Top reason / 底部 fit footer；JDDetail header 36px score + WHY MATCHES + WHERE STRETCH + ROLE SNAPSHOT + INTEL（条件渲染 + 合规措辞）+ Action bar 4 button；testid 命名按 §3.5 列出；Card 视图模型严格按 §3.7。

#### 3.2 `useJobMatchRecommendations()` hook

通过 generated `listJobRecommendations` 调；状态：loading / data / error；Vitest 断言 cursor pagination（如 backend 返回 cursor）+ failed variant → 错误占位 + retry button。

#### 3.3 `useToggleWatchlist()` hook + Save / Unsave 闭环

Save → optimistic toggle + `addToWatchlist({ jobMatchId }, { idempotencyKey })`；Unsave → `removeFromWatchlist(jobMatchId, { idempotencyKey })`；4xx → revert + error toast；window.eiToast 反馈；Vitest 断言 request body / Idempotency-Key / 4xx revert / toast 文案 zh/en。

#### 3.4 `useDismissRecommendation()` hook + Mark not relevant 闭环

Dismiss → optimistic 隐藏卡片 + 自动选下一张 + `markJobNotRelevant(jobMatchId, { reason: enum, freeNote?: optional }, { idempotencyKey })`；ui-design 不弹窗收集 reason，frontend 透传默认 enum 值；4xx → revert（卡片重新显示）+ error toast；Vitest 断言行为完整。

#### 3.5 Confirm interview → parse 出口

点击 JDDetail Action bar Confirm interview 按钮 → `nav("parse", { source: "jd_match", sourceJobMatchId: jobMatchId })`；不携带其他 jd_match 内部状态；Vitest 断言 nav 调用 + params 完整性。

#### 3.6 Open source new tab

点击 JDDetail Action bar Source 按钮 → `window.open(sourceUrl, "_blank", "noopener,noreferrer")`；sourceUrl 缺失时 button disabled；Vitest spy + 隐私反查（sourceUrl 不进 console / URL / localStorage / telemetry）。

#### 3.7 Auth gate

未登录用户对 Save / Unsave / Mark not relevant / Confirm interview 触发 `requestAuth({ type: "jd_match_action", route: "jd_match", params: { tab: "recommended", selectedJobMatchId, action: enum }, label })`；登录后回到 jd_match 自动重新触发；Source button 不进 auth pending action，只在已登录/未登录状态下统一执行安全 `window.open` 行为；Vitest `jd_match/JDMatchAuthGate.test.tsx` 断言 query / label / freeNote / sourceUrl 不进 pendingAction params。

#### 3.8 隐私反查

Vitest 断言 jobMatchId / sourceUrl / freeNote 不出现在 `console.log` / URL query / `localStorage` / telemetry payload；mockTransport spy 仅记录 status code + 调用次数，不记录 body。

#### 3.9 Vitest

新增 `jd_match/RecommendedTab.test.tsx`、`jd_match/JobMatchCard.test.tsx`、`jd_match/JDDetail.test.tsx`、`jd_match/JDMatchAuthGate.test.tsx`、`jd_match/RecommendedToggleWatchlist.test.tsx`、`jd_match/RecommendedDismiss.test.tsx`、`jd_match/RecommendedConfirmInterview.test.tsx`、`jd_match/RecommendedOpenSource.test.tsx`；总计 ≥10 测试文件，全 PASS。

#### 3.10 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.022` Recommended tab 主路径 + 4 button 闭环 + auth pending action + `E2E.P0.026` Confirm interview → parse 出口 params 完整性

### Phase 4: Search tab — 自然语言搜索 + Saved searches + Result filters + 单一加载文案

#### 4.1 新增 `SearchTab.tsx` + 子组件

按 `screen-jd-match.jsx::SearchTab` line 480 源级复刻（注意 SearchTab 的 `searching=true` 区域以届时 ui-design 真理源为准 — 用户已或将简化为单一动态加载文案）：搜索框 + Run web search 按钮 + 数据源 tag 列 + 单一动态加载文案区 + Saved searches 3 列网格（list + create，无 delete）+ Results 2 列网格 cap 6 + 4 chip filter (all/strong/remote/unseen)；testid 命名按 §3.5。

#### 4.2 `useSearchJobs()` hook

Run → `searchJobs({ query, filters?, profileSnapshot? }, { idempotencyKey })`；in-flight 期间 `searching=true` 渲染 `jdMatch.search.searching` i18n 文案 + 简单脉冲/省略号动画（视 ui-design 修订后形态）；slow-response variant 验证文案持续渲染至 settle，不超时；query 不进 URL/localStorage/telemetry；Vitest 断言 request body / Idempotency-Key / loading 行为。

#### 4.3 `useSavedSearches()` + `useCreateSavedSearch()` hooks

切 Search tab 调 `listSavedSearches` 1 次；点击 "Save current as watch" 按钮 → `createSavedSearch({ label, query, filters? }, { idempotencyKey })` + toast；ui-design 无 delete 入口故 plan 002 不实现 `deleteSavedSearch`；Vitest 断言 4xx → toast + 不写入 saved searches；label 不进 URL/localStorage/telemetry。

#### 4.4 4 chip filter 纯 client-side

filter 按 §3.7 规则纯 client-side 切换；切换不发新 search request；Vitest 断言。

#### 4.5 Auth gate

未登录用户点击 Run / Save current 触发 `requestAuth({ type: "jd_match_action", route: "jd_match", params: { tab: "search", action: enum }, label })`；登录后回到 jd_match search tab 自动重新触发；query / saved-search label 不进 pendingAction params；Vitest 断言。

#### 4.6 隐私反查

Vitest 断言 query / saved-search label / filter state / sourceJobUrl 不出现在 `console.log` / URL query / `localStorage` / telemetry payload；mockTransport spy 仅记 status code。

#### 4.7 失败 / 空 / boundary

- failed variant → inline error + 保留输入 + 不展示伪造 results
- empty variant → no-results 空态文案
- slow-response variant → 单一加载文案持续渲染直到 settle
- searching=true 期间 Run 按钮 disabled 防止重复提交
- 切 tab 时 abort in-flight search，Vitest fake timer 断言无 race

#### 4.8 Vitest

新增 `jd_match/SearchTab.test.tsx`、`jd_match/SearchTabRun.test.tsx`、`jd_match/SearchTabSavedSearches.test.tsx`、`jd_match/SearchTabFilter.test.tsx`、`jd_match/SearchTabFailure.test.tsx`、`jd_match/SearchTabPrivacy.test.tsx`、`jd_match/SearchTabAuthGate.test.tsx`；总计 ≥7 测试文件，全 PASS。

#### 4.9 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.023` Search tab 主路径 + Saved searches + filter + slow-response + failure + privacy

### Phase 5: Watchlist tab + Market signals

#### 5.1 新增 `WatchlistTab.tsx` + 子组件

按 `screen-jd-match.jsx::WatchlistTab` line 592 源级复刻：WatchlistItem 列表（3px 左边框按 tone 着色 ok/warn/muted、added 时间、change 信号、chevron）+ Market signals 4 列 grid（k/v/d/tone 4 卡）+ refresh footer 文案；testid 命名按 §3.5。

#### 5.2 `useWatchlist()` + `useMarketSignals()` hooks

切 Watchlist tab 调 `listWatchlist` + `getMarketSignals` 各 1 次；Vitest 断言不重复调；fixture variant 三态（empty / few / partial-data）渲染。

#### 5.3 chevron handoff

点击 WatchlistItem chevron → tab 切 Recommended + selected = `WatchlistItem.linkedJobMatchId`；当 linkedJobMatchId 不在当前 Recommended 列表时 fallback 第一张 + warning toast；linkedJobMatchId 不进 URL/localStorage/telemetry；Vitest 断言。

#### 5.4 Market signals partial-data

partial-data variant → 部分卡片渲染 + 缺失值 fallback；不展示 0 或硬编码占位；Vitest 断言。

#### 5.5 隐私反查

Vitest 断言 watchlist label / sourceJobUrl / linkedJobMatchId 不出现在 `console.log` / URL query / `localStorage` / telemetry payload。

#### 5.6 Vitest

新增 `jd_match/WatchlistTab.test.tsx`、`jd_match/WatchlistChevron.test.tsx`、`jd_match/WatchlistEmpty.test.tsx`、`jd_match/MarketSignals.test.tsx`、`jd_match/MarketSignalsPartial.test.tsx`、`jd_match/WatchlistPrivacy.test.tsx`；总计 ≥6 测试文件，全 PASS。

#### 5.7 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.024` Watchlist + Market signals + chevron handoff + empty + partial-data + privacy

### Phase 6: 验证收口（pixel parity + scenario + regression rerun + 文档同步）

#### 6.1 Playwright pixel parity 升级

升级 `frontend/tests/pixel-parity/jd_match.spec.ts` 覆盖三 tab × desktop (1440×900) + mobile (390×844) × warm/light + dark + customAccent：

- DOM 锚点存在性（三 tab + 子组件）
- 关键元素 bounding box stays in viewport, no overlap
- mobile 折叠：Recommended 双列折单列、Saved searches 折单列、Market signals 4→2 列、Search results 2→单列；JDDetail sticky 改流式
- 三主题态 computed background / color 可见变化（score 三档色、left border、chip filter active、AGENT badge tone、Watchlist 3 tone、Market signals tone）
- toHaveScreenshot baseline 区域：JDMatchScreen Hero + Profile chip + AGENT badge、Recommended 双列布局、JDDetail Action bar、SearchTab 输入区 + Saved searches、WatchlistTab 列表 + Market signals
- D3 字体子像素差异沿用 maxDiffPixels 阈值

`pnpm --filter @easyinterview/frontend test:pixel-parity` 在 plan 001 已升级 baseline 上累加；总数全 PASS。

#### 6.2 Scenario 资产新增

派生 5 个新 scenario 目录，按 `test/scenarios/README.md` + `test/scenarios/e2e/README.md` 规范实现（含 README.md §6 baseline + §7 离线限制 + scripts/{setup,trigger,verify,cleanup}.sh）：

- `test/scenarios/e2e/p0-022-jd-match-recommended-and-confirm/`
- `test/scenarios/e2e/p0-023-jd-match-search-and-saved/`
- `test/scenarios/e2e/p0-024-jd-match-watchlist-and-signals/`
- `test/scenarios/e2e/p0-025-jd-match-profile-and-agent-status/`
- `test/scenarios/e2e/p0-026-jd-match-confirm-interview-handoff/`

#### 6.3 升级 `p0-017-jd-match-placeholder` scenario

verify 脚本不再 grep `jdmatch-placeholder` 与 "Coming in P1" 文案；改为断言三 tab 数据驱动渲染（Profile chip + AGENT badge + Recommended 列表 + 三 tab 切换）；保留旧业务 testid 负向断言（确认 plan 002 落地后旧 prototype 业务 testid 仍 0 命中）。

#### 6.4 Scenario INDEX 更新

`test/scenarios/e2e/INDEX.md` P0 表追加 5 行（P0.022/023/024/025/026），关联需求列指向 `frontend-home-job-picks-and-parse C-12～C-16`，状态 Ready，执行方式 automated；同步 P0.017 行（描述更新为「升级到三 tab 数据驱动 smoke」）。

#### 6.5 Regression 重跑

`E2E.P0.001 / 002 / 004 / 005 / 006 / 014 / 015 / 016 / 017` 全部 setup→trigger→verify→cleanup PASS（D1+D2+D3+plan001 不被 plan 002 改动破坏）；`pnpm --filter @easyinterview/frontend test`（全量 Vitest，含本 plan 新增 ≥25 spec）+ `pnpm --filter @easyinterview/frontend typecheck` + `pnpm --filter @easyinterview/frontend build` + `make build` + `make validate-fixtures` 全 PASS。

#### 6.6 文档与索引同步

`/sync-doc-index --fix-index` 把 `docs/spec/INDEX.md` 与 `docs/spec/frontend-home-job-picks-and-parse/plans/INDEX.md` 同步到 Header 当前；`make docs-check` zero drift；`check_md_links` 双 OK。

#### 6.7 负向搜索

- `frontend/src/` 内不 import `ui-design/src/data.jsx` 或 `window.EI_DATA`；不复制粘贴 `ui-design/src/screen-jd-match.jsx` 内嵌的 `decoratedJobs` / `savedSearches` / `watchlist` mock 数据
- plan 001 placeholder 文案与 testid（`jdmatch-placeholder` / `jdmatch-placeholder-cta` / `jdMatch.placeholderTitle` / `placeholderCopy` / `placeholderCta` / "Coming in P1" / "Coming Soon"）grep 0 命中（除升级后的 P0.017 verify 脚本如有合规残留）
- 5 步 AGENT panel 文案（"向量化你的画像" / "并行查询 LinkedIn" / "按 JD 哈希去重" 等）grep 0 命中
- 旧 route alias（`welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star` / 独立 `voice`）grep 0 命中
- LinkedIn / Boss / 脉脉 / 拉勾 SDK + AI provider key + provider registry + prompt registry + AIClient + LLM endpoint + bypass generated client 的外部招聘平台 fetch grep 0 命中（除 fixture / i18n / 测试负向断言中的合规文案展示）
- query / saved-search label / watchlist label / sourceJobUrl / freeNote / linkedJobMatchId / jobMatchId 不在 console.log / URL / localStorage / telemetry 调用中出现 0 命中

#### 6.8 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.022 / E2E.P0.023 / E2E.P0.024 / E2E.P0.025 / E2E.P0.026` 全部 setup→trigger→verify→cleanup PASS + `E2E.P0.017` 升级后 PASS + D1+D2+D3+plan001 P0.001/002/004/005/006/014/015/016 regression 全 PASS

## 5 验收标准

- 本计划列出的 Phase 1-6 全部 checklist 项通过
- spec C-12 ~ C-16 全部覆盖且通过对应测试
- 关联 BDD-Gate（E2E.P0.022 / E2E.P0.023 / E2E.P0.024 / E2E.P0.025 / E2E.P0.026）全部通过；E2E.P0.017 升级后 PASS；D1+D2+D3+plan001 regression（P0.001/002/004/005/006/014/015/016）全部 PASS
- pixel parity 在 desktop + mobile 两 viewport 下 jd_match 三 tab × 三主题态 spec 全 PASS
- `make docs-check` zero drift；`check_md_links` 双 OK；`pnpm typecheck` 0 错；`pnpm build` + `make build` + `make validate-fixtures` PASS
- 负向搜索（plan 001 placeholder / 5 步 AGENT panel / 旧 route alias / prototype data 直接 import / LLM/provider/外部招聘平台 / 隐私字段泄漏）全部 0 命中
- OpenAPI `JobMatch` tag 12 operationId schema 与 fixture 双向对齐；B2 owner inventory additive 升级到 13 tag / 46 endpoint；codegen drift 0；contract test 全 PASS

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| ui-design `screen-jd-match.jsx` SearchTab `searching=true` 区域用户独立修订与 plan 002 frontend 实施时间点错位 | Phase 1 启动前在 implement 会话开头读取最新 ui-design 文件确认 SearchTab 当前形态；如未修订则暂停 frontend SearchTab 实现并回到用户处确认；plan 中 D-12 与 §3.7 已锁定语义形态（单一加载文案），不锁定具体 DOM 行号 |
| 12 operationId 一次性新增导致 OpenAPI review 工作量大 | Phase 1 内部按四组分批（Profile+Agent 2 / Recommendations 5 / Search+SavedSearches 3 / Watchlist+MarketSignals 2）各自小提交；每组完成后跑 contract test 与 codegen drift；`/plan-review` 时按组逐一审 |
| Save / Unsave / Dismiss optimistic 与 4xx revert 的 race condition | optimistic toggle 前缓存原状态；4xx revert 时按缓存恢复；并发多次点击时按最后一次 server 响应为准；Vitest fake timer 覆盖 race；E2E.P0.022 含连续点击子用例 |
| Auth pending action 在 jd_match side-effect 操作恢复时丢失 selectedJobMatchId 上下文 | pendingAction params 显式包含 `{ tab, selectedJobMatchId, action }`；登录恢复时按 selectedJobMatchId 重新选中卡片再触发 action；如卡片已被 dismiss 则 fallback 第一张 + 提示；E2E.P0.025 覆盖 |
| chevron handoff `linkedJobMatchId` 不在当前 Recommended 列表（如已被 dismiss / pagination 未加载） | 显式 fallback 第一张 + warning toast；不抛错；不强制重新调 `listJobRecommendations`；E2E.P0.024 含 unknown linkedJobMatchId 子用例 |
| Watchlist add/remove 与 Recommended 列表 saved 状态同步 | `useToggleWatchlist` 状态 hoist 到 JDMatchScreen 容器层（或 Zustand store），跨 tab 共享；切回 Recommended 时 saved 状态正确反映；Vitest 断言 |
| Search query 残留导致跨 session 隐私泄漏 | 切 tab / unmount 时 query reset；不写 URL/localStorage；登录后从 pendingAction params 恢复时 params 不含 query 文本，仅含 action 类型；Vitest 隐私反查 + E2E.P0.023 |
| `getAgentScanStatus` polling 与 D-10 锁定的「不引入定时器」冲突 | 严格按 D-10：仅在进入 jd_match + 切回 Recommended tab 时调用，不 setInterval；Vitest fake timer 断言无定时器调用 |
| Pixel parity baseline 在三 tab × 三主题 × 双 viewport 组合下 baseline 文件爆炸 | 沿用 D3 maxDiffPixels 阈值 + baseline 区域聚焦关键 DOM；不为每个 tab × 主题 × viewport 组合生成全屏 baseline；按需细分 baseline；CI cache baseline directory |
| OpenAPI 新增 12 operation 与 plan 001 已 generated client typings 冲突（如 method 名重复 / schema name 冲突） | Phase 1.4 codegen 后立即跑 `pnpm typecheck`；schema 命名加 `JobMatch` 前缀避免与 `TargetJob` 系列冲突；method 名严格按 operationId 不缩写 |
| Search tab `slow-response` variant 验证 loading 持续渲染时 Vitest fake timer 行为偏差 | 用 `vi.useFakeTimers({ shouldAdvanceTime: true })` + 显式 advance；不依赖 real timer；E2E.P0.023 提供真实 mockTransport 延迟验证 |
| plan 001 已 deploy 的 `jdMatch.placeholderTitle/copy/cta` i18n key 删除后引用残留 | Phase 2.4 完成后立即跑 typed helper test + grep negative；如有残留先修后再进 Phase 3 |
