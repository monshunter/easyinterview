# 001 Home + JD Import + Parse + JD Match Placeholder

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-24

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

在 D1+D2+D3 已交付的 App 壳、视觉系统与 pixel parity gate 之上，把 `home` 与 `parse` 两屏从 `ui-design/` 静态原型迁移到正式 frontend，端到端跑通 P0 主路径「粘贴/上传/URL 导入 JD → 解析确认 → 进入模拟面试规划」；同时把 `jd_match` 屏作为 P1 placeholder shell 接入路由，保留 TopBar 入口可达。

完成本计划后，用户在 frontend dev server 上能够：

1. 默认进入 home，看到 JD 导入卡片、Recent mock interviews 列表、Job Picks / Post-interview aux cards、empty state 与 resume create CTA
2. paste / upload / URL 三种 source variants 提交 JD，进入 parse 屏看到 4 步 loading → preview / confirm
3. 编辑 parse 屏的 basic fields、切换 hit toggle，点击 Confirm 进入 workspace（携带完整 interview context params）
4. 点击 TopBar Job Picks 或 home aux card 进入 jd_match P1 placeholder
5. UI variants 继续通过 generated client + fixture-backed mock transport 稳定覆盖；同时在 `VITE_EI_API_MODE=real` 下用 production generated client 证明 TargetJobs/upload/import/parse operations 指向真实 backend base URL；JD 原文不泄漏；i18n zh/en 完整切换；dark + customAccent 三态可见变化；desktop + mobile pixel parity 通过

## 2 背景

`frontend-shell` D1（001 app shell + auth + settings）+ D2（002 visual system）+ D3（003 pixel parity gate）已交付：App 默认进入 home、五入口 TopBar、route normalization、`requestAuth(pendingAction)` 与登录恢复、generated client + fixture transport bootstrap、warm/forest/ocean/plum 4 主题 + dark + customAccent、Vitest+jsdom smoke gate（E2E.P0.001/002/004/005）+ Playwright pixel parity gate（E2E.P0.006）。

`frontend-shell/spec.md` §2.1 显式让出 `parse` 业务内容；`engineering-roadmap/spec.md` 预占 `frontend-home-job-picks-and-parse` 子 spec。`openapi-v1-contract` 已声明 `TargetJobs` tag 4 个 operations + 完整 schema + fixtures（`importTargetJob.json` / `listTargetJobs.json` / `getTargetJob.json` / `updateTargetJob.json`）。`mock-contract-suite` 提供 generated client mock transport。2026-05-22 L2 remediation 复查时，`backend-targetjob/001-targetjob-import-and-parse-bootstrap` 与 `backend-upload/001-file-objects-and-presign-baseline` 已落地真实 handler；本 plan 原地补齐 real-mode generated-client gate，避免继续把 2026-05-08 的 fixture-first wording 当作当前 backend owner 事实。

本 plan 是新 subspec 的首个计划，覆盖 P0 用户首次接入闭环。`jd_match` 完整三 tab 业务由后续 plan `002-jd-match-recommendations` 在 backend recommendations API 落地后承接。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior（用户可感知 UI + API 行为 + 业务流程 + 端到端功能）
- **TDD 策略**: Red-Green-Refactor 入口为 `pnpm --filter @easyinterview/frontend test`（Vitest）；每个 Phase 在新增组件前先写失败测试，覆盖 DOM 锚点、控件类型、props/state、generated client 调用断言、URL/state 隐私反查；`pnpm --filter @easyinterview/frontend test:pixel-parity` 在 Phase 6 扩展为 home + parse 双屏 desktop + mobile 4 个 project；新增组件文件位于 `frontend/src/app/screens/home/`、`frontend/src/app/screens/parse/`、`frontend/src/app/screens/jd_match/`；测试文件与组件 colocate（`*.test.tsx`）。
- **BDD 策略**: Feature plan requires BDD；本 plan 在 `bdd-plan.md` 定义 4 个场景 `E2E.P0.014 / E2E.P0.015 / E2E.P0.016 / E2E.P0.017`，`bdd-checklist.md` 跟踪每个场景资产创建与执行；主 `checklist.md` 在每个 Phase 末尾保留 `BDD-Gate:` 项引用对应场景 ID。
- **替代验证 gate**: 不适用（feature plan，已有完整 BDD + TDD 双层覆盖）

## 3.5 Coverage Matrix

| 类别 | 覆盖描述 | UI Source Anchor | Phase | 验证入口 |
|------|----------|------------------|-------|---------|
| Primary path | Paste JD → importTargetJob → parse loading → analysisStatus=ready → preview → Confirm → workspace | `screen-home.jsx::HomeScreen` lines 49-90 + `screens-p0-complete.jsx::ParseScreen` lines 6-242 | 1+2+3+4 | E2E.P0.015 + Vitest `home/HomeImport.test.tsx` + `parse/ParseFlow.test.tsx` |
| Alternate path · upload source | upload modal → `createUploadPresign` `purpose=target_job_attachment` → `fileObjectId` → importTargetJob `source.type=file` | `screen-home.jsx::JDAssistModal` lines 218-262 | 3 | Vitest `home/JDAssistModal.test.tsx` + E2E.P0.015 variant |
| Alternate path · URL source | URL modal → url → importTargetJob `source.type=url` | `screen-home.jsx::JDAssistModal` lines 218-262 | 3 | Vitest 同上 + fixture variant |
| Alternate path · auth pending action | 未登录提交 → requestAuth → 登录后恢复 | `app.jsx::App.requestAuth` + `app.jsx::completeSignIn` | 3+4 | Vitest `home/HomeAuthGate.test.tsx` + `parse/ParseAuthGate.test.tsx` |
| Failure / recovery · import 4xx | importTargetJob 422 invalid source / 401 unauthenticated | n/a (error state) | 3 | Vitest fixture negative variant + inline error UI |
| Failure / recovery · parse failed | getTargetJob analysisStatus=failed | n/a (error state) | 4 | Vitest `parse/ParseFailedState.test.tsx` + fixture variant |
| Failure / recovery · re-parse | Re-parse 重置 stage=loading 重新轮询 | `screens-p0-complete.jsx::ParseScreen` line 234 | 4 | Vitest `parse/ParseFlow.test.tsx` |
| Failure / recovery · update 4xx | updateTargetJob 4xx 显示 inline 错误并保留编辑态 | n/a (error state) | 4 | Vitest |
| Boundary · empty textarea | textarea 空时 Parse 按钮 disabled | `screen-home.jsx::HomeScreen` line 79 | 1 | Vitest |
| Boundary · empty listTargetJobs | listTargetJobs 返回空数组 → HomeEmptyState | `screen-home.jsx::HomeEmptyState` lines 135-146 | 2 | Vitest `home/HomeEmptyState.test.tsx` |
| Boundary · listTargetJobs >12 items | 取最近 12 条按 updatedAt desc | `screen-home.jsx::HomeScreen` line 96-98 | 2 | Vitest + fixture variant |
| Boundary · 字段长度 | title/company 超长 ellipsis；location 折行 | `screen-home.jsx::MockInterviewCard` lines 165-169 | 2 | Vitest computed style |
| Cross-layer contract · UploadPresignRequest | file source 先调用 `createUploadPresign`，`purpose=target_job_attachment`，取返回 `fileObjectId`；测试不真实上传二进制 | OpenAPI `UploadPresignRequest` / `UploadPresign` | 3 | Vitest + fixture `Uploads/createUploadPresign.json` plan-added variant |
| Cross-layer contract · ImportTargetJobRequest discriminator | source oneOf 四 variant；`type` discriminator + 必填字段；side-effect 调用带 `Idempotency-Key` | OpenAPI `TargetJobImportSource` | 3 | mock-contract-suite parity test + Vitest 4 variant |
| Cross-layer contract · TargetJob schema | requirements / summary / fitSummary / analysisStatus 渲染 | OpenAPI `TargetJob` schema | 4 | Vitest fixture parity |
| Cross-layer contract · UpdateTargetJobRequest 部分字段 | 仅 supplied fields 写入；side-effect 调用带 `Idempotency-Key` | OpenAPI `UpdateTargetJobRequest` | 4 | Vitest request body 反查 |
| Cross-layer contract · provenance 渲染 | summary.interviewHypotheses / fitSummary.riskSignals 必带 provenance | OpenAPI `GenerationProvenance` | 4 | Vitest |
| Privacy / security · JD raw text | rawText / rawDescription / url 不进 console / URL / localStorage / telemetry | n/a | 3+4 | Vitest 反查 + redact lint |
| Privacy / security · auth | 未登录提交触发 requestAuth；登录后恢复 | `app.jsx::App.requestAuth` | 3+4 | Vitest |
| Privacy / security · provenance redact | provenance 字段中 promptTemplate / rubric id 不在 UI 中暴露完整 hash | `screens-p0-complete.jsx::ParseScreen` lines 100-104 | 4 | Vitest |
| Observability | 仅 fixture transport 调用次数 / latency / 4xx code 进 telemetry；不带 raw text | n/a | 3+4 | Vitest mockTransport spy |
| UX · loading state | Parse 4 步进度条节奏；listTargetJobs 加载占位 | `screens-p0-complete.jsx::ParseScreen` lines 68-107 | 2+4 | Vitest |
| UX · empty state | listTargetJobs 空 → HomeEmptyState；search/watchlist tab 空 → P1 placeholder | `screen-home.jsx::HomeEmptyState` + `screen-jd-match.jsx` placeholders | 2+5 | Vitest |
| UX · error state | importTargetJob 4xx 内联错误；getTargetJob failed 全屏错误 | n/a | 3+4 | Vitest |
| UX · i18n zh/en | 全文案通过 typed helper；切换立即重绘；新增 home / parse / jdMatch namespaces | D1 typed locale helper | 1-5 | Vitest `i18n` namespaces test |
| UX · dark + customAccent | home + parse 三态切换关键元素 computed 颜色变化 | D2 `data-theme` / `data-mode` / `data-custom-accent` | 1+4+6 | Playwright + Vitest computed style |
| UX · responsive layout | mobile 390×844 下不溢出；Requirements 双列折叠为单列；textarea card 不溢出 | n/a | 1+4+6 | Playwright mobile project |
| UI source structure parity · home Hero | Hero label / title / sub i18n + textarea card + upload/URL/Submit | `screen-home.jsx::HomeScreen` lines 49-90 | 1 | Vitest DOM + testid `home-hero-*` / `home-jd-textarea` / `home-jd-submit` |
| UI source structure parity · home aux cards | JOB PICKS + POST-INTERVIEW 2 张 aux 卡片 + 各自 Btn | `screen-home.jsx::HomeScreen` lines 105-128 | 1 | Vitest `home-aux-jobpicks` / `home-aux-debrief` |
| UI source structure parity · MockInterviewCard | company meta slot / title / location / status pill / MiniRoundRail 圆点 + 进度线；meta slot 数据按 §3.7 映射，不读取 OpenAPI 未声明 `level` 字段 | `screen-home.jsx::MockInterviewCard` + `MiniRoundRail` lines 148-216 | 2 | Vitest + testid `home-recent-mock-card-${id}` / `home-recent-mock-rail-${id}` |
| UI source structure parity · JDAssistModal | upload + URL 双模态、Continue / Cancel 按钮、关闭 X、外层遮罩点击关闭 | `screen-home.jsx::JDAssistModal` lines 218-262 | 3 | Vitest + testid `home-modal-{upload\|url}-*` |
| UI source structure parity · Parse loading | 4 步进度条 + footer model/rubric/prompt hash 作为 backend parse metadata / fixture metadata 展示；前端不调用 LLM | `screens-p0-complete.jsx::ParseScreen` lines 68-107 | 4 | Vitest + testid `parse-loading-step-${i}` |
| UI source structure parity · Parse Basic fields | Title / Company / Location editable；Level / Language 保留 `ui-design` 槽位但 read-only，避免写入 OpenAPI 未声明字段 | `screens-p0-complete.jsx::ParseScreen` lines 156-175 | 4 | Vitest + testid `parse-basics-${field}` |
| UI source structure parity · Requirements blocks | Must Have / Nice to Have 双列、hit toggle (true/partial/false 三态)、note | `screens-p0-complete.jsx::RequirementBlock` lines 244-264 | 4 | Vitest + testid `parse-requirement-${kind}-${idx}` |
| UI source structure parity · Hidden signals | sparkle icon list + confidence tag | `screens-p0-complete.jsx::ParseScreen` lines 184-206 | 4 | Vitest + testid `parse-hidden-signal-${idx}` |
| UI source structure parity · Round assumptions | 4 卡 R1-R4 grid | `screens-p0-complete.jsx::ParseScreen` lines 209-225 | 4 | Vitest + testid `parse-round-${idx}` |
| UI source structure parity · Parse footer | Cancel / Re-parse / Confirm 三 button | `screens-p0-complete.jsx::ParseScreen` lines 228-239 | 4 | Vitest + testid `parse-action-{cancel\|reparse\|confirm}` |
| UI source structure parity · jd_match shell | Hero + Profile snapshot chip + 三 tab 标签 + placeholder 内容 | `screen-jd-match.jsx::JDMatchScreen` lines 244-300 | 5 | Vitest + testid `jdmatch-hero` / `jdmatch-tab-${k}` / `jdmatch-placeholder` |
| UI visual geometry parity · desktop | 1440×900 home + parse + jd_match bounding box stays in viewport, no overlap | n/a | 6 | Playwright `tests/pixel-parity/home.spec.ts` + `parse.spec.ts` + `jd_match.spec.ts` desktop project |
| UI visual geometry parity · mobile | 390×844 Requirements 折单列、textarea/cards 不溢出 | n/a | 6 | Playwright mobile project |
| UI visual geometry parity · dark / customAccent | 三态切换关键元素 computed background / color 可见变化 | n/a | 6 | Playwright |
| UI visual geometry parity · screenshot regression | toHaveScreenshot baseline maxDiffPixels 阈值 | n/a | 6 | Playwright + `frontend/.gitignore` baseline |
| UI stale-contract negative · 旧 jd_match 占位 | 旧 prototype 的 search / watchlist 业务字段（saved-search-*, watchlist-*, market-signal-*）testid 在本 plan 不能命中 | `screen-jd-match.jsx` 旧 prototype | 5+6 | Vitest negative grep + scenario verify |
| UI stale-contract negative · 旧 route alias | 旧 `welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star` / 独立 `voice` route 在 Home / Parse / jd_match 新代码中不出现 | n/a | 全 phase | Vitest + scenario verify negative grep |
| Regression / legacy-negative · D1+D2+D3 现有 gate | E2E.P0.001/002/004/005/006 重跑通过 | n/a | 6 | scenario rerun |
| Regression / legacy-negative · 不直接 import prototype data | `frontend/src` 不 import `ui-design/src/data.jsx` 或 `window.EI_DATA` | n/a | 全 phase | Vitest + tsc grep negative |
| Regression / legacy-negative · 不直接调用 LLM/provider | `frontend/src` 不出现 AI provider key、provider registry、prompt registry、LLM endpoint 或 ad hoc parse fetch；只允许 generated TargetJobs client / fixture transport | n/a | 全 phase | Vitest + grep negative |

### 高风险类别 N/A 说明

- 无高风险类别整体 N/A；本 plan 覆盖 primary / alternate / failure / boundary / cross-layer / privacy / UX / UI source / visual geometry / regression-negative 类别

## 3.6 Frontend / Backend Operation Matrix

本 plan 最初走 `docs/development.md` §2.2 Frontend-First Path：正式前端先对齐 `ui-design/` 并通过 generated client + fixture-backed transport 完成 P0 UI/BDD。自 2026-05-22 起，真实 TargetJob handler / store / parse runner 已由 `backend-targetjob/001-targetjob-import-and-parse-bootstrap` 落地，上传 presign handler 已由 `backend-upload/001-file-objects-and-presign-baseline` 落地；frontend 仍保留 fixture-backed UI variants，同时必须用 `VITE_EI_API_MODE=real` generated-client gate 证明 production bootstrap 指向真实 backend base URL，并与 backend live scenarios / focused handler tests 配对。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` scenarios: `default` / `prototype-baseline` / plan-added `empty` / `one-job` / `12-plus` | `HomeScreen` `useRecentTargetJobs()` via generated client; query uses `RequestOptions.query.pageSize=12` and optional cursor/status filters only when exposed by generated method; card view model uses only generated `TargetJob` fields plus explicit fallback mapping in §3.7 | real: `backend/cmd/api/main.go` mounts `GET /api/v1/targets` to `backend/internal/targetjob.Handler.ListTargetJobs`; owned by `backend-targetjob/001-targetjob-import-and-parse-bootstrap` | backend: `target_jobs` + `target_job_requirements`; frontend: none | none in frontend; backend may return AI-generated summary already persisted | E2E.P0.014 fixture UI + real gate; backend E2E.P0.010 / P0.013 |
| `createUploadPresign` | `openapi/fixtures/Uploads/createUploadPresign.json` scenarios: `default` / plan-added `target-job-attachment-success` / `validation-4xx` | `JDAssistModal` upload Continue via generated `createUploadPresign`; request body uses `purpose=target_job_attachment`, fileName/contentType/byteSize from selected placeholder file; response `fileObjectId` feeds `importTargetJob`; tests do not PUT binary to `uploadUrl` | real: `backend/cmd/api/main.go` mounts `POST /api/v1/uploads/presign` to `backend/internal/upload/handler.Handler.CreateUploadPresign`; owned by `backend-upload/001-file-objects-and-presign-baseline` | backend: `file_objects`; frontend: selected File metadata in React state only | none | E2E.P0.015 fixture UI + real gate; backend upload route/handler focused tests |
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json` scenarios: `manual_text-success` / `file-success` / `url-success` / `invalid-source-422` / `server-error-500` | `HomeScreen` import submit via generated `importTargetJob`; source variants are `manual_text` / `file` / `url`; request body carries JD raw content only through generated body + React state; side-effect call supplies `idempotencyKey` | real: `backend/cmd/api/main.go` mounts `POST /api/v1/targets/import` to `backend/internal/targetjob.Handler.ImportTargetJob`; frontend must not call URL fetcher, file parser, LLM, prompt registry, or provider-specific endpoint | backend: `target_jobs`, `target_job_sources`, `target_import` job / outbox; frontend: none | backend-only `target.import.parse` through F3/A3 after import job; frontend fixture/stub only for UI variants | E2E.P0.015 fixture UI + real gate; backend E2E.P0.010 / P0.011 / P0.012 / P0.013 |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` scenarios: `default` / `prototype-baseline` / plan-added `queued` / `processing` / `ready` / `failed` / `hidden-signal-rich` | `ParseScreen` polling via generated `getTargetJob(targetJobId)`; `analysisStatus` drives loading/preview/failed; Hidden signals render only backend/API `summary.interviewHypotheses` + `coreThemes` + `fitSummary.riskSignals` with `GenerationProvenance` present | real: `backend/cmd/api/main.go` mounts `GET /api/v1/targets/{targetJobId}` to `backend/internal/targetjob.Handler.GetTargetJob`; no frontend LLM interaction | backend: `target_jobs`, `target_job_requirements`; frontend: ephemeral UI state only | backend-only parse result; frontend displays returned AI-generated fields without inference or regeneration | E2E.P0.015 fixture UI + real gate; backend E2E.P0.010 / P0.011 / P0.012 / P0.013 |
| `updateTargetJob` | `openapi/fixtures/TargetJobs/updateTargetJob.json` scenarios: `success` / `validation-4xx` | `ParseScreen` Confirm via generated `updateTargetJob`; request body includes only supplied editable fields (`titleHint`, `companyNameHint`, `locationText`, `notes`); hit toggle state is not sent; side-effect call supplies `idempotencyKey` | real: `backend/cmd/api/main.go` mounts `PATCH /api/v1/targets/{targetJobId}` to `backend/internal/targetjob.Handler.UpdateTargetJob`; owned by `backend-targetjob/001-targetjob-import-and-parse-bootstrap` | backend: `target_jobs` metadata columns; frontend: none | none in frontend; no parse regeneration | E2E.P0.016 fixture UI + real gate; backend E2E.P0.010 |
| `N/A` UI-only `jd_match` placeholder | no new fixture; generated client must not be called by `JDMatchScreen` placeholder except existing shell auth/runtime calls | `JDMatchScreen` hero/profile-chip/tabs/placeholder shell; TopBar + Home aux card route only | N/A until future recommendations API / plan `002-jd-match-recommendations` | none | none | E2E.P0.017 |

## 3.7 TargetJob Frontend View-Model Mapping

正式前端不得从 `ui-design/src/data.jsx` 或未声明 fixture 字段补齐 `TargetJob` 之外的数据。Home recent cards 与 workspace 跳转统一使用以下 mapping：

| UI slot / param | Source | Rule |
|-----------------|--------|------|
| card title | `TargetJob.title` | 直接展示；超长 ellipsis / wrap 由 CSS gate 覆盖 |
| company meta slot | `TargetJob.companyName` + `TargetJob.status` | 显示为 `companyName · status label`，保留 `screen-home.jsx` 的 meta DOM slot；不读取不存在的 `level` 字段 |
| location | `TargetJob.locationText` | 为空时显示 locale fallback `Remote / TBD`，不得写死真实地点 |
| status pill text / tone | `TargetJob.status` | `draft/preparing=muted`，`applied/interviewing=amber`，`offer=neutral`，`rejected/archived=neutral`；token 来源 D2，后续若需要 success tone 先扩展 `ui-design` / D2 token |
| MiniRoundRail | P0 default rounds + `TargetJob.status` | `draft/preparing` currentIndex=0，`applied/interviewing` currentIndex=1，`offer/rejected/archived` currentIndex=last；后续真实 round contract 落地后由 owner spec 修订 |
| workspace params | `TargetJob.id` + deterministic defaults | `targetJobId=id`、`jobId=id`、`jdId=jd-${id}`、`planId=plan-${id}`、`resumeVersionId=resume-unbound`、`roundId=round-technical-1`、`roundName` locale fallback；不得依赖 OpenAPI 未声明字段 |

## 3.8 修订记录

| 日期 | 版本 | 类型 | 说明 |
|------|------|------|------|
| 2026-05-24 | 1.3 | regression remediation | 修复 Phase 4 ready 响应直接进入 preview 的 implementation drift：`ParseScreen` 必须先展示并完成 `ui-design/src/screens-p0-complete.jsx::ParseScreen` 4 步 loading 演示，再进入 parsed preview；`ParseFlow` 与 E2E.P0.015 gate 固化该行为。 |

## 4 实施步骤

### Phase 1: Home shell 静态壳 + 路由壳 + i18n（无数据）

#### 1.1 新增 `frontend/src/app/screens/home/` 目录与 `HomeScreen.tsx`

按 `ui-design/src/screen-home.jsx::HomeScreen` 源级复刻渲染 Hero（label / title / sub）、JD textarea card（含 upload / URL / Submit 按钮）、Resume create CTA、aux cards（JOB PICKS + POST-INTERVIEW），不接入数据（recent mocks 用 placeholder）。`onSubmit` / `onUpload` / `onUrl` 仅记录调用次数，不发起 API。

#### 1.2 路由壳接入

在 `frontend/src/app/App.tsx` route table 中绑定 `home` → `<HomeScreen />`（替换 D1 `PlaceholderScreen`）。保留 D1 `eiCreateInterviewContext` 等价契约。

#### 1.3 i18n locale 文件扩展

在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 中新增 `home.*` 命名空间（至少 14 个 key 对应 `screen-home.jsx::L` 字典）；通过 D1 typed helper 渲染。

#### 1.4 Vitest 红灯 → 绿灯

新增 `home/HomeScreen.test.tsx`：测 Hero label/title/sub 渲染、5 个主要 testid 存在、空 textarea 时 Parse 按钮 disabled、aux cards 点击调用 nav stub、i18n zh/en 切换；DOM 控件类型断言（textarea / button / 自定义 Btn）；负向断言旧 prototype testid 不命中。

#### 1.5 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.014` Home 默认渲染场景中 home 静态部分的资产构建到 ready 态

### Phase 2: Recent mock interviews（消费 listTargetJobs）

#### 2.1 generated client 接入

通过 D1 `frontend/src/api/generated/client.ts` 调用 `listTargetJobs`；在 `HomeScreen` 中通过 React state + effect 拉取并渲染。Loading 占位、empty state（`HomeEmptyState`）、错误占位三态。

#### 2.2 `MockInterviewCard` + `MiniRoundRail` 组件

按 `screen-home.jsx::MockInterviewCard` lines 148-216 + `MiniRoundRail` lines 188-216 源级复刻；字段与状态映射必须遵守 §3.7，只从 generated `TargetJob` schema 派生 `company meta slot`、status tone、round fallback 和 workspace params；点击调 `nav("workspace", interviewContextFromTargetJob(targetJob))`。

#### 2.3 排序与限制

按 `updatedAt desc` 排序；最多展示 12 条；超出由 fixture 与生产 API 端 cursor pagination（不在本 plan 内消费 cursor）。

#### 2.4 fixture variant

新增 `openapi/fixtures/TargetJobs/listTargetJobs.json` variants（empty / 1-3 条 / 12+ 条）— 若现有 fixture 已经是单 variant，扩展为 variant 字典或新增对应 fixture sibling 文件，按 `mock-contract-suite` 规则配置。

#### 2.5 Vitest

新增 `home/HomeRecentMocks.test.tsx`：测 fixture variant 三态渲染、排序、点击 nav 携带正确 params、status pill 三档 computed style、MiniRoundRail 当前轮次圆点；空数组 → `HomeEmptyState` 渲染并 focus textarea；错误响应 → 错误占位。

#### 2.6 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.014` Home + Recent mocks 综合场景

### Phase 3: JD 导入（textarea + upload + URL → importTargetJob）

#### 3.1 `JDAssistModal` 组件

按 `screen-home.jsx::JDAssistModal` lines 218-262 源级复刻 upload / URL 双模态；外层遮罩点击关闭、ESC 关闭、Continue / Cancel 按钮；upload 模式显示拖拽区，URL 模式显示 URL input。

#### 3.2 三种 source 提交逻辑

textarea paste 提交 → `source.type=manual_text`；upload 模态 Continue 先调用 generated `createUploadPresign`（`purpose=target_job_attachment`，带 `idempotencyKey`），取返回 `fileObjectId` 后提交 `source.type=file`；URL 模态 Continue → `source.type=url` + `url`。所有 variants 都通过 generated `ImportTargetJobRequest` schema 提交并带 `idempotencyKey`。`targetLanguage` 取当前 UI locale（`zh` / `en`）。

#### 3.3 提交后路由

成功响应 `TargetJobWithJob.targetJobId` → `nav("parse", { targetJobId, source })`；4xx 显示内联错误（textarea 下方 / modal 内部）并保留输入；5xx 显示通用错误并允许重试。

#### 3.4 Auth pending action

未登录提交时调 `requestAuth({ type: "import_jd", route: "home", params: { source, pendingImportId }, label: "继续解析 JD" })`；`pendingImportId` 只引用当前 SPA 会话内存中的待提交 source payload，不把 JD 原文 / source URL / rawDescription 写入 URL 或 localStorage；登录成功后回到 home 并自动以保留的 form state 重新发起 importTargetJob，成功后跳 parse。

#### 3.5 Vitest

新增 `home/JDAssistModal.test.tsx`：测 upload / URL 模态 DOM、Continue 触发 `onConfirm`、关闭路径（X / 遮罩 / Cancel / ESC）；新增 `home/HomeImport.test.tsx`：测三种 source 提交对应 generated client request body schema、targetJobId 跳转、4xx/5xx 内联错误、auth pending action 触发与恢复；隐私反查 — JD 原文不写入 console/URL/localStorage/telemetry。

#### 3.6 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.015` Paste→Import→Parse 主路径已具备 home/import 阶段

### Phase 4: Parse 屏（loading + preview + confirm）

#### 4.1 新增 `frontend/src/app/screens/parse/` 目录与 `ParseScreen.tsx`

按 `screens-p0-complete.jsx::ParseScreen` lines 6-242 源级复刻：loading 阶段 4 步进度条 + footer backend parse metadata 占位文案（DOM / copy / rhythm 与 `ui-design` 一致，但不代表前端调用 LLM）；preview 阶段 Basic fields DOM 保持 5 槽位，其中 title / company / location / notes 可保存、level / language read-only（OpenAPI 当前不允许 update），Requirements 双列 + hit toggle 三态、Hidden signals + confidence tag、Round assumptions 4 卡、footer Cancel / Re-parse / Confirm。

#### 4.2 状态机驱动 loading→preview

进入 parse 屏后立即通过 generated client 调 `getTargetJob(targetJobId)` 并按 backend/API 返回的 `analysisStatus` 决定渲染；前端不得直接接入 AI provider、prompt registry、provider secret、LLM endpoint 或 ad hoc parse fetch：

- `queued` / `processing`：显示 loading；以可观察节奏（≥600ms 间隔）轮询 `getTargetJob` 直到非 queued/processing；progress 步骤随轮询次数递进，但不假装代表真实模型调用步骤
- `ready`：切到 preview 渲染 fixture/backend response 中 title / companyName / locationText / requirements / summary.interviewHypotheses / summary.coreThemes / fitSummary.riskSignals；summary / fitSummary 的 `GenerationProvenance` 缺失时必须进入错误或降级展示，不允许前端本地推断 hidden signals
- `failed`：切到错误态（C-6）

#### 4.3 Preview 编辑

Basic fields 中 title / company / location / notes onChange 更新 React state 并可在 Confirm 保存；level / language 按 `ui-design` 槽位展示但 read-only，避免用户误以为 `UpdateTargetJobRequest` 能持久化未声明字段；hit toggle 同样 ephemeral（不写后端）。Requirements label / evidenceLevel 只读展示。

#### 4.4 Confirm 调用

Confirm 时调 `updateTargetJob(targetJobId, body)`，body 仅包含 supplied fields（titleHint / companyNameHint / locationText / notes）。成功后 `nav("workspace", { targetJobId, jdId, planId, resumeVersionId, roundId })`，使用 D1 `eiCreateInterviewContext` 等价契约推导默认值。4xx 显示 inline 错误并保留编辑态。

#### 4.5 Re-parse / Cancel

Re-parse 重置 `stage=loading` 并重新调 `getTargetJob` 触发 polling；Cancel 跳 `home`。

#### 4.6 Auth pending action

未登录用户进入 parse 屏不直接挂壁；点击 Confirm 时触发 `requestAuth({ type: "confirm_interview", route: "workspace", params: { targetJobId, jdId, planId, resumeVersionId, roundId } })`；登录后回到 workspace。

#### 4.7 i18n

`parse.*` 命名空间（≥30 key 覆盖 4 步 loading 文案、Basic fields label、Must Have / Nice to Have、Hidden signals、Round assumptions、footer actions、错误态文案）。

#### 4.8 Vitest

新增 `parse/ParseScreen.test.tsx`：测 loading 4 步进度条 + footer 文案；新增 `parse/ParseFlow.test.tsx`：测 polling 三态切换；新增 `parse/ParseEdit.test.tsx`：测 inline 编辑、hit toggle 三态、Confirm body schema 反查、4xx inline 错误；新增 `parse/ParseFailedState.test.tsx`：测 failed 态 UI；新增 `parse/ParseAuthGate.test.tsx`：测 Confirm 未登录 → requestAuth → 登录恢复；隐私反查 — JD 原文 / hash 不出现在 URL/localStorage/telemetry。

#### 4.9 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.015`（主路径完整）+ `E2E.P0.016`（编辑 + Confirm → workspace）

### Phase 5: jd_match P1 Placeholder Shell

#### 5.1 新增 `frontend/src/app/screens/jd_match/` 目录与 `JDMatchScreen.tsx`

按 `ui-design/src/screen-jd-match.jsx::JDMatchScreen` lines 244-300 复刻 hero（label / title / sub）+ profile snapshot chip 静态版本（不连接真实 profile）+ 三 tab 标签（Recommended / Search / Watchlist）；tab 内容区固定渲染 P1 placeholder 文案 + 引用本 subspec spec §7 中的 plan `002-jd-match-recommendations`。不渲染 JobMatchCard / JDDetail / SearchTab / WatchlistTab。

#### 5.2 路由壳接入

在 `frontend/src/app/App.tsx` route table 中绑定 `jd_match` → `<JDMatchScreen />`（替换 D1 `PlaceholderScreen`）；TopBar 高亮 jd_match。

#### 5.3 i18n

`jdMatch.*` 命名空间（≤10 key：hero label / title / sub、tab 标签、placeholder 文案 zh/en）。

#### 5.4 Vitest

新增 `jd_match/JDMatchPlaceholder.test.tsx`：测 hero / profile chip / 三 tab 标签 DOM；测 placeholder 文案 zh/en；负向断言 — 旧 prototype 的 `JobMatchCard` / `JDDetail` / `SearchTab` / `WatchlistTab` testid（如 `jdmatch-card-*` / `jdmatch-saved-search-*` / `jdmatch-watchlist-*` / `jdmatch-market-signal-*`）不命中；TopBar `topbar-nav-jd_match` 高亮。

#### 5.5 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.017` jd_match P1 placeholder smoke

### Phase 6: 验证收口（pixel parity + scenario + regression rerun）

#### 6.1 Playwright pixel parity 扩展

新增 `frontend/tests/pixel-parity/home.spec.ts`、`parse.spec.ts`、`jd_match.spec.ts`，覆盖 desktop (1440×900) + mobile (390×844) 两个 chromium project：

- DOM 锚点存在性
- 关键元素 bounding box stays in viewport, no overlap
- mobile Requirements 折单列、textarea card 不溢出
- warm/light → dark → customAccent 三态切换 computed background / color 可见变化
- toHaveScreenshot baseline 区域：home Hero、Recent mocks 网格、parse loading 与 preview 主区块、jd_match hero 与 placeholder

`pnpm --filter @easyinterview/frontend test:pixel-parity` 全 PASS（在 D2/D3 现有 21 个 spec × 2 viewport = 42 项基础上累加）。

#### 6.2 Scenario 资产

派生 4 个新 scenario 目录：

- `test/scenarios/e2e/p0-014-home-default-render/`
- `test/scenarios/e2e/p0-015-jd-import-and-parse/`
- `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/`
- `test/scenarios/e2e/p0-017-jd-match-placeholder/`

每个目录包含 `README.md`（§6 文档 baseline 维护、§7 离线运行限制）+ `scripts/{setup,trigger,verify,cleanup}.sh`（按 `test/scenarios/README.md` + `test/scenarios/e2e/README.md` 规范）。

#### 6.3 Scenario INDEX 更新

`test/scenarios/e2e/INDEX.md` P0 表追加 4 行（`E2E.P0.014` / `E2E.P0.015` / `E2E.P0.016` / `E2E.P0.017`），关联需求列指向 `frontend-home-job-picks-and-parse C-1` … `C-10`，状态 Ready，执行方式 automated。

#### 6.4 Regression 重跑

`E2E.P0.001 / 002 / 004 / 005 / 006` 五个 scenario 的 `setup → trigger → verify → cleanup` 全部 PASS；`pnpm --filter @easyinterview/frontend test`（全量 Vitest）+ `pnpm --filter @easyinterview/frontend typecheck` + `pnpm --filter @easyinterview/frontend build` + `make build` 全 PASS。

#### 6.5 文档与索引同步

`/sync-doc-index --fix-index` 把 `docs/spec/INDEX.md` 与各 plans/INDEX 同步到当前 Header；`make docs-check` zero drift；`check_md_links` 双 OK。

#### 6.6 负向搜索

- `frontend/src/` 内不 import `ui-design/src/data.jsx` 或 `window.EI_DATA`
- 旧 prototype jd_match 业务 testid（`jdmatch-card-*` / `jdmatch-saved-search-*` / `jdmatch-watchlist-*` / `jdmatch-market-signal-*` / `jdmatch-search-bar` 等）grep 0 命中（除负向断言文件）
- 旧 route alias（`welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star` / 独立 `voice`）grep 0 命中（除 `app/normalizeRoute.ts` 的 alias map 与对应负向 D1 测试）
- JD raw text grep — 仅出现在 React state / generated client request body 与 fixture，不出现在 console.log / URL / localStorage / telemetry 调用
- LLM/provider grep — `frontend/src` 不出现 provider key、provider registry、prompt registry、AIClient、LLM endpoint 或任意 bypass generated client 的 parse fetch；Parse loading footer 仅允许作为 UI 文案 / fixture metadata 展示

#### 6.7 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.014` / `E2E.P0.015` / `E2E.P0.016` / `E2E.P0.017` 全部通过 + D1+D2+D3 现有 P0.001-006 regression PASS

#### 6.8 L2 remediation — real backend integration closure

原地关闭 plan 001 与 backend real implementation 的语义漂移：新增 `frontend/src/api/targetJob.realApiMode.test.ts`，在 `VITE_EI_API_MODE=real` 下证明 `createAppClient` 使用 production generated `EasyInterviewClient` 和真实 backend base URL 调用 `listTargetJobs`、`createUploadPresign`、`importTargetJob`、`getTargetJob`、`updateTargetJob`；断言每个 request `credentials: "include"`、默认不带 fixture `Prefer` header、side-effect operation 带 `Idempotency-Key`、TargetJob summary / fitSummary `GenerationProvenance` roundtrip，且 JD 原文只进入 POST body、不进入 URL。

同步 P0.014-P0.016 trigger/verify/README，让每个 frontend scenario 都先跑 real-mode generated-client gate，再跑 fixture-backed UI variants；与 backend E2E.P0.010-P0.013 的 live HTTP TargetJob scenarios 配对，并用 backend-upload focused tests 证明 `POST /api/v1/uploads/presign` route/handler 已真实落地。P0.017 仍是 jd_match UI-only smoke，不纳入 TargetJobs real backend overlay。

## 5 验收标准

- 本计划列出的 Phase 1-6 全部 checklist 项通过
- spec C-1 ~ C-11 全部覆盖且通过对应测试
- 关联 BDD-Gate（E2E.P0.014 / E2E.P0.015 / E2E.P0.016 / E2E.P0.017）全部通过；P0.014-P0.016 trigger log 必须包含 `VITE_EI_API_MODE=real` 与 `targetJob.realApiMode.test.ts` PASS；backend E2E.P0.010-P0.013 live TargetJob scenarios 全部 PASS；D1+D2+D3 regression（P0.001/002/004/005/006）全部 PASS
- pixel parity 在 desktop + mobile 两 viewport 下 home / parse / jd_match 三屏新增 spec 全 PASS
- `make docs-check` zero drift；`check_md_links` 双 OK；`pnpm typecheck` 0 错；`pnpm build` + `make build` PASS
- 负向搜索（旧 prototype 业务 testid、旧 route alias、prototype data 直接 import、JD raw text 泄漏）全部 0 命中

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| listTargetJobs / getTargetJob fixture 缺少 variants 导致 boundary / failure 路径无法测试 | Phase 2.4 / 4.2 先扩展 fixture variants（empty / 12+ / processing / failed / hidden-signal-rich），通过 `mock-contract-suite` parity test 校验 |
| polling 节奏与 fixture transport 同步立即返回 ready 导致 4 步进度条无法观察 | 在 fixture transport 引入可观察 latency 占位，或在前端组件内显式分步 ≥600ms 节奏并锁定为 acceptance criteria（避免后续优化为 0ms 跳过 loading 阶段） |
| ParseScreen Re-parse 与 polling 的 race condition | Re-parse 时 abort 当前 polling effect 并重置 `analysisStatus` 局部状态；Vitest fake timer 下断言 polling 不会泄漏到下一次 |
| `eiCreateInterviewContext` 等价契约不 stable 导致 workspace 携带的 params 漂移 | 在 D1 已有 helper 之上抽 `frontend/src/app/navigation/interviewContext.ts` 集中契约；新增 unit test 锁定字段集合与回退默认值 |
| jd_match P1 placeholder 后续被 plan 002 替换时 testid 漂移 | placeholder testid 命名锁定为 `jdmatch-placeholder-*`（不与未来 `jdmatch-recommendations-*` 冲突），plan 002 不需要改 D1 path |
| Pixel parity 跨 fontsource 字体子像素差异（D3 retrospective 已识别） | 沿用 D3 经验：home / parse / jd_match 的 toHaveScreenshot 仅作 frontend 内部 regression（含 maxDiffPixels 阈值），不与 ui-design golden 跨字体源做硬 diff |
| Auth pending action 在 paste 流恢复时表单 state 丢失 | 把待提交 source payload 存入当前 SPA 会话内存，并只通过 D1 `pendingAction.params` 序列化 opaque `pendingImportId`；登录恢复时消费内存 payload 并自动重新发起 importTargetJob；新增 Vitest 测试锁定行为，同时负向断言 JD 原文 / source URL 不进入 URL 或 localStorage |
| 旧 prototype data 渗透（开发者从 `ui-design/src/screen-home.jsx` 复制粘贴时把 `D.targetJobs` / `D.jdSample` 一并带过来） | Vitest negative grep + `eslint-rules` 反查（`no-restricted-imports` 限制 `ui-design/`）；scenario verify 阶段 grep `EI_DATA` / `targetJobs` literal |
