# Event and Outbox Contract Resume Tailor Mode Drift Fix

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [event-and-outbox-contract spec](../../spec.md) §3.1 D-14 声明的 `ResumeTailorMode` 漂移修复落到可执行 artifact：

- 修订 `shared/events.yaml` `eventLocalEnums.ResumeTailorMode` 字面量集合：`[inline, rewrite, mirror]` → `[gap_review, bullet_suggestions]`，对齐 [B2 OpenAPI `RequestResumeTailorRequest.mode`](../../../openapi-v1-contract/spec.md#42-schema-inventory-约束) 与 [B4 `resume_tailor_runs.mode`](../../../db-migrations-baseline/spec.md) check constraint；
- 同步重新生成 `backend/internal/shared/events/` Go 类型与 `frontend/src/lib/events/` TS 类型，验证 `make codegen-events && make codegen-check` 无漂移；
- 同步 `shared/events/baseline/events.v1.json` baseline manifest，将 `ResumeTailorMode` baseline 字面量改为新值，并由 `make lint-events` 验证不触发 breaking 报警（因 baseline 期 `resume.tailor.completed` 无真实 producer/consumer，依本 history 写作规则归属 fixture/docs-only 路径）；
- Out-of-scope literal negative search 仅扫描 executable/generated/source truth 中的事件模式 artifacts；`history.md` 与本 plan 的 diff 表述不纳入零残留 gate；
- 同步 B3 spec §3.1.4 `resume.tailor.completed.mode` 列 enum 值描述（移除"声明阶段 → 落地阶段"措辞，留下 `[gap_review, bullet_suggestions]` 唯一表述）；
- 通过 B3 spec §6 验收（C-1 envelope freeze + C-6 breaking-change 拦截 + C-12 `ResumeTailorMode` 对齐）。

本 plan 不修订 B2 OpenAPI `RequestResumeTailorRequest.mode`（已就位）；不修订 B4 `resume_tailor_runs.mode` check constraint（已就位）；不创建 backend internal runner 与 consumer（归后续 `backend-async-runner` / `backend-resume`）。

## 2 背景

B3 spec §3.1 D-14 已声明：当前 `shared/events.yaml` `eventLocalEnums.ResumeTailorMode` 字面量与 B2/B4 不同步，是 baseline 期既有契约漂移。Resume Workshop 阶段 0 contract additive 升级（roadmap 3.10 / B2 D-18 / B4 D-N 同步推进）要求在 spawn `backend-resume` / `backend-upload` / `frontend-resume-workshop` 三个新 subspec 之前消除这一漂移，避免后续业务 plan 启动时立即遭遇双 enum source-of-truth 冲突。

本 plan 是 [event-and-outbox-contract spec §7 关联计划](../../spec.md#7-关联计划) 列出的第 2 个，承担 D-14 落地：把 B3 spec 2.3 声明阶段实际投影到 `shared/events.yaml`、`shared/events/baseline/events.v1.json`、generated Go/TS 类型与 B3 spec §3.1.4 描述。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有 yaml 修订 + codegen drift gate；Phase 2 起来就有 baseline manifest 同步；Phase 3 起来就有 grep negative search 证明零残留；Phase 4 收口验收 + 解锁 B2 D-18 / B4 002 同步落地。

执行本 plan 前必须确认：

- [001-bootstrap](../001-bootstrap/plan.md) Phase 已完成：`shared/events.yaml` / `shared/events/baseline/events.v1.json` / generated 类型与 `make codegen-events && make lint-events` 入口可用。
- [B2 D-18 声明](../../../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) 已声明阶段（spec 1.16）；本 plan 不依赖 B2 plan 004 落地，可独立推进，但建议同 sprint 落地以避免 cross-spec drift window。

## 3 质量门禁分类

- **Plan 类型**: `contract + tooling + drift-fix`。本 plan 修订 `shared/events.yaml` enum、baseline manifest、generated Go/TS 类型；不实现 backend internal runner、producer、consumer 或业务逻辑。
- **TDD 策略**: 适用（Code plan requires TDD）。Red-Green-Refactor 入口：
  1. 修订前先跑 `make codegen-events && make codegen-check` 确认 baseline pass；
  2. 修订 yaml 后跑 `make codegen-events`，verify generated Go `events.ResumeTailorModeGapReview` / TS `as const` 字面量已对齐；
  3. `make lint-events` 验证 baseline manifest 同步（由 Phase 2 同步）；
  4. 事件契约 artifacts 中 out-of-scope literals 精准 grep 为空；当前 B3 spec 文本不再保留"声明阶段 → 落地阶段"out-of-scope 描述；
  5. 新增或更新 Go / TS type-narrowing 测试，先断言 out-of-scope 值不可回流、新值必须导出，再确认 Red → Green。
  执行入口：`/implement event-and-outbox-contract/002-resume-tailor-mode-drift-fix` → `/tdd`。
- **BDD 策略**: 不适用。本 plan 是内部 contract drift-fix，无用户可感知 UI / API 行为变化（`resume.tailor.completed` 事件未有真实 producer/consumer）；后续用户可见 Resume Tailor 流程由 `frontend-resume-workshop` / `backend-resume` 维护 BDD gate。
- **替代验证 gate**:
  - `make codegen-events && make codegen-check`
  - `make lint-events`（baseline drift gate）
  - `git grep -nE 'ResumeTailorMode(Inline|Rewrite|Mirror)|"(inline|rewrite|mirror)"' -- shared/events.yaml shared/events/refs/ResumeTailorMode.json shared/events/baseline/events.v1.json shared/events/schemas/resume.tailor.completed.v1.json backend/internal/shared/events frontend/src/lib/events openapi/openapi.yaml openapi/fixtures`（断言事件契约 artifacts out-of-scope literals 0 命中）
  - `git grep -nE '\[inline, rewrite, mirror\]|声明阶段' -- docs/spec/event-and-outbox-contract/spec.md`（断言当前 B3 spec 不再保留 out-of-scope 声明措辞；`docs/spec/event-and-outbox-contract/history.md` 与本 plan 的 diff 文本不纳入该 gate）
  - `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml`（间接断言 B2 mode enum 仍是 `gap_review / bullet_suggestions`，与 events 对齐）
  - `sync-doc-index --check`

## 4 实施步骤

### Phase 1: `shared/events.yaml` 修订与 codegen drift 验证

#### 1.1 修订 `eventLocalEnums.ResumeTailorMode`

定位 `shared/events.yaml` 中 `eventLocalEnums.ResumeTailorMode`（约 line 53-58），将字面量改为：

```yaml
eventLocalEnums:
  ResumeTailorMode:
    - gap_review
    - bullet_suggestions
```

#### 1.2 重新生成 Go/TS 类型并验证

运行 `make codegen-events && make codegen-check`：
- `backend/internal/shared/events/` Go 类型新增 `ResumeTailorModeGapReview` / `ResumeTailorModeBulletSuggestions`；out-of-scope `ResumeTailorModeInline` / `Rewrite` / `Mirror` 删除（generator 不应留残留）。
- `frontend/src/lib/events/` TS 类型 `as const` 字面量同步。
- `git diff --exit-code` PASS（generated artifact 已对齐）。

#### 1.3 跑 B3 unit test 验证 Red→Green

新增或更新：
- `backend/internal/shared/events/resume_tailor_mode_test.go`：断言 `ResumeTailorModeGapReview` / `ResumeTailorModeBulletSuggestions` 存在，out-of-scope `Inline` / `Rewrite` / `Mirror` 不可作为允许集合回流；
- `frontend/src/lib/events/events.test.ts`：断言 `ResumeTailorMode` typed payload 接受 `gap_review` / `bullet_suggestions`，并通过 `// @ts-expect-error` 或等价类型断言拒绝 out-of-scope `inline` / `rewrite` / `mirror`。

运行 `cd backend && go test ./internal/shared/events/...` 与 `pnpm --filter @easyinterview/frontend test src/lib/events/events.test.ts`，验证 `ResumeTailorMode` 类型相关测试通过。

### Phase 2: Baseline manifest 同步

#### 2.1 修订 `shared/events/baseline/events.v1.json`

把 baseline manifest 中 `eventLocalEnums.ResumeTailorMode` 字面量同步为新值。

#### 2.2 `make lint-events` 验证

运行 `make lint-events`：
- B3 spec D-4 additive 语义对漂移修复的处理：因 baseline 期 `resume.tailor.completed` 无真实 producer/consumer，baseline 修订归 fixture/docs-only 路径（参考 B2 history.md §1 fixture-only 行约定），不触发 breaking 报警；如 lint 报 breaking，验证 lint 是否正确分类 event-local enum 与 envelope 字段集合。
- 必要时在 lint 配置中加入 `ResumeTailorMode drift-fix` 白名单条目（参照 B2 `openapi/diff-config.yaml` privacy export 白名单模式）。

### Phase 3: 跨仓库 grep negative search

#### 3.1 事件契约 artifact out-of-scope literal 扫描

运行：

```bash
git grep -nE 'ResumeTailorMode(Inline|Rewrite|Mirror)|"(inline|rewrite|mirror)"' -- \
  shared/events.yaml \
  shared/events/refs/ResumeTailorMode.json \
  shared/events/baseline/events.v1.json \
  shared/events/schemas/resume.tailor.completed.v1.json \
  backend/internal/shared/events \
  frontend/src/lib/events \
  openapi/openapi.yaml \
  openapi/fixtures
```

预期：0 命中。该 gate 只覆盖 executable/generated/source truth，避免被 CSS `inline`、普通文案 `rewrite`、以及本 plan / history 中的 diff 说明污染。若有命中，必须同步清理（包括 fixture 文件、生成物缓存、测试 dataset）。

随后运行：

```bash
git grep -nE '\[inline, rewrite, mirror\]|声明阶段' -- \
  docs/spec/event-and-outbox-contract/spec.md
```

预期：0 命中。`docs/spec/event-and-outbox-contract/history.md` 与本 plan 的 diff 表述不计入 executable/generated/source truth gate。

#### 3.2 同步 B3 spec §3.1.4 描述

修订 `docs/spec/event-and-outbox-contract/spec.md` §3.1.4 `resume.tailor.completed` 行：

- `mode` 列描述从 "（当前字面量 `[inline, rewrite, mirror]`；D-14 声明阶段 → 002 plan 落地阶段对齐为 `[gap_review, bullet_suggestions]`，与 B2 OpenAPI / B4 DB 同步）" 改为 "`[gap_review, bullet_suggestions]`，与 B2 OpenAPI `RequestResumeTailorRequest.mode` / B4 `resume_tailor_runs.mode` 同步"。
- spec.md 2.3 → 2.4，history.md 追加 2.4 行（"D-14 落地：`ResumeTailorMode` `[inline, rewrite, mirror]` → `[gap_review, bullet_suggestions]`，baseline manifest 同步，零残留"）。

### Phase 4: 验收与下游同步

#### 4.1 跨 gate 收口

按 §3 替代验证 gate 依序运行：
- `make codegen-events && make codegen-check` PASS
- `make lint-events` PASS（无 breaking）
- 事件契约 artifact out-of-scope literals grep 0 命中；当前 B3 spec out-of-scope 声明措辞 grep 0 命中
- `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` PASS（间接断言）
- `sync-doc-index --check` PASS

#### 4.2 同步通知 B2 / B4 owner

通知 [openapi-v1-contract/004-resume-additive-coverage](../../../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) 与 [db-migrations-baseline/002-flat-resume-migration](../../../db-migrations-baseline/plans/002-flat-resume-migration/plan.md) owner：events `ResumeTailorMode` 已对齐，可在各自 plan 中 fully reference `gap_review / bullet_suggestions` 字面量。

#### 4.3 修订 `docs/spec/INDEX.md`

`docs/spec/INDEX.md` 中 event-and-outbox-contract 版本与日期同步到 2.4。

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成
- §3 替代验证 gate 全部通过
- `shared/events.yaml` `eventLocalEnums.ResumeTailorMode` 字面量 = `[gap_review, bullet_suggestions]`
- `shared/events/baseline/events.v1.json` 与 yaml 同步
- executable/generated/source truth 中 out-of-scope `ResumeTailorMode` 字面量为 0；B3 当前 spec 不再保留 out-of-scope 声明措辞
- B3 spec.md 2.3 → 2.4，§3.1.4 描述简洁化
- B2 plan 004 / B4 plan 002 收到同步信号

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: `make lint-events` 将 enum 集合替换识别为 breaking | 类比 B2 D-12 privacy export whitelist 模式，在 lint 配置中加入 `ResumeTailorMode drift-fix` 白名单；baseline manifest 与 yaml 同 PR 同步避免 inter-commit drift |
| R2: out-of-scope literal 在 generated artifact 缓存或测试 fixture 中残留 | Phase 3.1 精准 grep 必须覆盖 shared event source、baseline manifest、schema refs、Go/TS generated event 类型、OpenAPI 与 fixtures；如有 cache 文件，需要 `make codegen-clean` 后重新生成 |
| R3: B2 plan 004 / B4 plan 002 与本 plan 同 sprint 落地时序冲突 | 本 plan 不依赖 B2/B4 plan 落地；B2/B4 plan 也不依赖本 plan 完成；三 plan 可并行推进，每个 plan 在自己 spec 内独立验证；最终 cross-spec drift gate 在 Phase 4 同时验证 |
| R4: 如果未来需要扩展为四值 enum（包含 `inline` 等新产品决策） | 走 B3 spec D-N additive 升级并重新定义 event-local enum；必须通过 current spec/owner gate，且 event-and-outbox-contract spec §4 schema 仍约束 enum additive only |
