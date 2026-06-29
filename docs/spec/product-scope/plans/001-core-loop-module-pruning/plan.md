# Core Loop Module Pruning Plan

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-06-29

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

按用户已确认的方案 B，硬清理当前 P0 中的真实面试复盘模块和候选人画像模块，使产品核心回到：

```text
上传 / 粘贴简历 + JD
  -> 模拟面试
  -> 报告
  -> 复练当前轮 / 进入下一轮
```

完成后，`debrief` / `profile` / `CandidateProfile` / `ExperienceCard` 不再作为用户可见入口、OpenAPI tag、后端领域、DB 表、AI feature key、shared event/job 或场景验收对象存在。账号设置、邮箱认证、首次资料补全与隐私删除保留，但不得继续承担“用户画像”产品语义。

## 2 背景

当前产品 scope、engineering roadmap、UI 文档、静态原型、正式前端、OpenAPI、backend、migrations、shared、config 和 E2E 场景都仍包含复盘和候选人画像。用户已明确选择硬删除方案，而不是降级隐藏或保留兼容层。

由于本项目尚未上线，不要求保留历史 route / API / schema 兼容。删除必须以当前 active spec、`docs/ui-design/`、`ui-design/` 和编码 truth source 为准，避免文档仍把旧模块作为后续 workstream 自动恢复。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `contract` + `migration` + `code-internal` + `tooling`
- **TDD 策略**: 通过 `/implement product-scope/001-core-loop-module-pruning cross-layer` 进入 `/tdd`。每个 checklist item 在改实现前先补或改对应 red test：route/topbar/i18n/pixel parity、OpenAPI inventory/codegen/fixture validation、Go handler/store/service tests、migration/schema lint、scenario wrapper negative gate。
- **BDD 策略**: 用户可见入口和端到端流程会变化，必须维护 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md)。主 checklist 以 `E2E.P0.001`、`E2E.P0.088`、`E2E.P0.090`、`E2E.P0.098`、`E2E.P0.099`、`E2E.P0.102` 作为更新后的行为 gate，删除 `E2E.P0.060`-`E2E.P0.069`、`E2E.P0.071`、`E2E.P0.073`、`E2E.P0.091`-`E2E.P0.093` 的复盘 / 画像正向场景。
- **替代验证 gate**: API、DB、shared、config、prompt/rubric、generated artifacts 通过 `make codegen-check`、`make validate-fixtures`、migration lint / focused Go tests、repo-wide legacy-negative grep 和 `git diff --check` 验证；不以历史 PASS 或 checklist 状态作为当前完成证据。

## 4 实施步骤

### Phase 1: 产品和 UI 真理源改写

#### 1.1 product-scope 与 engineering-roadmap 收敛

把 P0 闭环、一级导航、用户菜单、模块边界、阶段路线、验收标准和已丢弃能力改为三入口核心链路。`复盘`、`用户画像`、`Progressive Profile` 对用户可见模块和候选人画像数据模型从当前 scope 中删除，并登记为已丢弃能力。

#### 1.2 UI 文档与静态原型改写

更新 `docs/ui-design/` 与 `ui-design/src/app.jsx`、相关 screen 源码，使一级导航只保留 `首页 / 模拟面试 / 简历`，用户菜单只保留 `设置与隐私 / 退出登录`，`debrief`、`debrief_full`、`profile` 不再是目标 route 或用户入口。

### Phase 2: 前端实现清理

#### 2.1 App shell route 和用户菜单清理

更新 `frontend/src/app/routes.ts`、`normalizeRoute.ts`、`routeUrl.ts`、`App.tsx`、TopBar 和 URL fallback，使 `debrief` / `profile` 不再进入 RouteName、primary nav、user menu、canonical path 或 legacy live route。未知或旧路径必须归一到当前核心入口，且不暴露旧模块页面。

#### 2.2 复盘和画像 screen / hook / i18n / tests 删除

删除 `frontend/src/app/screens/debrief/`、`ProfileScreen`、复盘 / 用户画像 i18n key、dev mock debrief special cases、frontend owner tests 和 pixel parity spec 中的正向复盘 / 画像断言，补充负向测试证明旧 UI contract 不会回流。

### Phase 3: OpenAPI、shared、generated 和 fixture 清理

#### 3.1 API contract 删除

删除 OpenAPI `Profile` / `Debriefs` tags、paths、schemas、fixtures 和 generated Go/TS client/server artifacts。更新 inventory、fixtures、mock transport allowlist 和 frontend/backend API tests。

#### 3.2 shared event/job/enum 清理

删除 `debrief.created`、`debrief.completed`、`debrief.generate`、`PracticeGoal=debrief`、`source_debrief_id` API/shared surface，以及候选人画像 / experience card 相关共享契约。保留账号资料补全和隐私删除必要的账号数据，不把它们命名为用户画像。

### Phase 4: 后端、迁移、AI config 清理

#### 4.1 backend debrief/profile 领域删除

删除 `backend/internal/debrief`、`backend/internal/api/debriefs`、`backend/internal/profile` 的 candidate profile / experience card 领域代码和 cmd/api wiring。更新 practice/report/resume/auth 直接消费者，确保核心链路不依赖 debrief/profile。

#### 4.2 migration 和 seed 清理

在当前未上线前提下修订 baseline/seed migrations，移除 `debriefs`、`candidate_profiles`、`experience_cards`、`practice_plans.source_debrief_id`、`goal='debrief'`、debrief AI prompt/rubric seeds、profile update seed 和相关 enum source。

#### 4.3 AI prompt/rubric/profile 清理

删除 `config/prompts/debrief.*`、`config/rubrics/debrief.*`、`config/evals/debrief.*`、`config/ai-profiles.yaml` 中的 debrief / profile feature key，并更新 prompt/rubric registry gate。

### Phase 5: 场景、文档索引和验收收口

#### 5.1 E2E 场景目录与索引清理

删除复盘 / 画像正向场景目录和索引项，更新核心闭环场景，使它们验证三入口 app shell、旧路径负向归一、JD 到报告再到复练 / 下一轮的闭环仍通过。

#### 5.2 零残留和质量门禁

运行文档、contract、backend、frontend、migration、scenario 和 legacy-negative gate。任何残留 `debrief` / `CandidateProfile` / `ExperienceCard` / `用户画像` / `复盘` 命中必须分类为允许的历史记录、已删除文件路径、报告 / work-journal 历史，或继续清理。

## 5 Operation Matrix

| operationId / contract | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|------------------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getMyProfile` | delete `openapi/fixtures/Profile/getMyProfile.json` | remove `ProfileScreen` and generated client calls | delete profile handler | delete `candidate_profiles` | none after cleanup | delete `E2E.P0.091` |
| `updateMyProfile` | delete `openapi/fixtures/Profile/updateMyProfile.json` | remove profile correction UI | delete profile handler | delete `candidate_profiles` | remove `profile.update` if present | delete `E2E.P0.091` |
| `listExperienceCards` | delete `openapi/fixtures/Profile/listExperienceCards.json` | no consumer after cleanup | delete profile handler | delete `experience_cards` | none | delete `E2E.P0.092` |
| `createExperienceCard` | delete `openapi/fixtures/Profile/createExperienceCard.json` | no consumer after cleanup | delete profile handler | delete `experience_cards` + idempotency consumer | none | delete `E2E.P0.092` |
| `updateExperienceCard` | delete `openapi/fixtures/Profile/updateExperienceCard.json` | no consumer after cleanup | delete profile handler | delete `experience_cards` + idempotency consumer | none | delete `E2E.P0.092` |
| `createDebrief` | delete `openapi/fixtures/Debriefs/createDebrief.json` | delete DebriefScreen submit hook | delete debrief handler/job | delete `debriefs`, `task_runs` debrief usage, `source_debrief_id` | delete `debrief.generate` | delete `E2E.P0.060`, `E2E.P0.066` |
| `suggestDebriefQuestions` | delete `openapi/fixtures/Debriefs/suggestDebriefQuestions.json` | delete DebriefScreen suggestions hook | delete debrief handler | none after cleanup | delete `debrief.suggest_questions` | delete `E2E.P0.063`, `E2E.P0.066` |
| `getDebrief` | delete `openapi/fixtures/Debriefs/getDebrief.json` | delete DebriefScreen polling hook | delete debrief handler | delete `debriefs` | none after cleanup | delete `E2E.P0.061`, `E2E.P0.067` |
| `createPracticePlan goal=debrief` | update existing fixture scenarios only if present | remove debrief replay launcher | update practice handler/store to reject/omit debrief source goal | delete `practice_plans.source_debrief_id`, `goal='debrief'` | none | update `E2E.P0.070`, delete `E2E.P0.071`, `E2E.P0.073` |

## 6 Coverage Matrix

| Source | Category | Plan phase | Verification | Negative scope |
|--------|----------|------------|--------------|----------------|
| 用户确认方案 B | Primary path | Phase 1-5 | `E2E.P0.098`, `E2E.P0.099` | 核心链路不得依赖复盘或画像 |
| product-scope P0 改写 | Cross-layer contract | Phase 1 | `make docs-check`, sync-doc-index | `复盘` / `用户画像` 不得作为 active 产品能力出现 |
| UI 真理源 | UI source structure parity | Phase 1-2 | frontend route/topbar tests, pixel parity smoke, `ui-design/canvas.html` retired-artboard negative gate | `debrief` primary nav, `profile` user menu, positive retired design-canvas artboards |
| Old route behavior | Regression / legacy-negative | Phase 2 | `E2E.P0.088`, `E2E.P0.090`, routeUrl tests | `/debrief`, `#route=debrief_full`, `/profile` live page |
| API removal | Cross-layer contract | Phase 3 | `make codegen-check`, `make validate-fixtures` | `Profile` / `Debriefs` tags, generated methods |
| DB removal | Migration | Phase 4 | migration lint / focused DB tests | `debriefs`, `candidate_profiles`, `experience_cards`, `source_debrief_id` |
| Async/event removal | Cross-layer contract | Phase 3-4 | shared jobs/events codegen/lint | `debrief.created`, `debrief.completed`, `debrief.generate` |
| AI config removal | Cross-layer contract | Phase 4 | prompt/rubric/profile lint/eval inventory | `debrief.*`, `profile.update` feature keys |
| Privacy boundary | Privacy / security | Phase 4-5 | backend privacy delete tests, legacy grep for retired profile cleanup hooks | account delete must still clean retained core data without candidate-profile runtime hooks |
| Scenario scope | Regression / legacy-negative | Phase 5 | scenario INDEX and script verification | retired P0 debrief/profile scenarios must not remain Ready |

## 7 验收标准

- Product scope、engineering roadmap、UI 文档、静态原型和正式前端都只把 `首页 / 模拟面试 / 简历` 作为一级业务入口。
- 用户菜单不再出现 `用户画像`，设置与隐私仍可进入。
- OpenAPI inventory、fixtures、generated Go/TS artifacts 不再包含 `Profile` / `Debriefs` tags 或对应 operationId。
- Backend、migrations、shared、config 不再包含运行时复盘或候选人画像领域。
- 核心 BDD 场景仍能证明 JD / 简历 -> 模拟面试 -> 报告 -> 复练 / 下一轮闭环。
- 复盘 / 画像旧 route、testid、table、event、job、feature key、prompt/rubric、scenario 通过负向搜索归零；历史 work-journal、bug、report 记录可保留为历史上下文，但不得作为 active truth source。

## 8 风险与应对

| 风险 | 应对措施 |
|------|----------|
| `profile` 与认证资料补全命名混用 | 只删除候选人画像 / experience card 领域；账号资料补全保留，并在文档中改称账号资料或资料补全 |
| 复盘 goal 已进入 practice 派生链路 | Phase 3/4 同步删除 `goal=debrief`、`source_debrief_id` 和 debrief first-question reservation，并补充 practice focused tests |
| OpenAPI 删除导致 generated consumer 大面积编译失败 | 先写 contract red tests / inventory gate，再代码删除并运行 codegen，最后修 frontend/backend consumers |
| 历史场景索引保留 Ready 状态 | Phase 5 删除正向场景目录和 INDEX 行，保留核心闭环场景的替代覆盖 |
| 文档仍把旧模块写为 P1/P2 自动恢复对象 | Phase 1 和 zero-reference gate 覆盖 product-scope、engineering-roadmap、docs/ui-design、docs/spec/INDEX |

## 9 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-06-29 | 1.1 | L2 review remediation: reopen cross-layer cleanup to remove stale design-canvas Profile/Debrief artboards and privacy runner profile cleanup hook drift. |
