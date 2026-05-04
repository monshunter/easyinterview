# Historical Spec Implementation Review Runway 完成验证

> **日期**: 2026-05-04
> **验证人**: Codex

**关联计划**: [historical-spec-implementation-review/001-implement-review-runway](../spec/historical-spec-implementation-review/plans/001-implement-review-runway/plan.md)

## 1 验证范围

本报告记录 `historical-spec-implementation-review/001-implement-review-runway` 在目标 `docs` 下的执行证据。当前执行范围是历史 spec / plan 的治理跑道：先完成 scope inventory，再判断 L1/L2、implementation 和并行写入是否具备启动条件。

本报告不替代任何原始 subject 的实现计划；代码实现仍由原 subject 的 `/implement` -> `/tdd` 负责。

## 2 Phase 1 Scope Inventory

### 2.1 前置上下文

已读取：

- `docs/work-journal/INDEX.md`
- `docs/work-journal/2026-05-04.md`
- `docs/spec/INDEX.md`
- 所有真实存在的 `docs/spec/*/plans/INDEX.md`
- `docs/spec/product-scope/spec.md`
- `docs/spec/engineering-roadmap/spec.md`
- `docs/ui-design/INDEX.md`
- `docs/ui-design/*.md` 的目标模块、移除模块、route、BDD、report、practice、resume、debrief 相关约束
- `ui-design/src/app.jsx`
- `ui-design/src/data.jsx`
- `ui-design/ui-design-contract.test.mjs`

最近完成事项：

- 2026-05-04 08:50-08:52 已完成 roadmap/spec、conventions codegen、UI design route notes、README 入口与 report 收口。
- 2026-05-04 22:15 已创建本 runway subject / plan / checklist / context，但尚未执行具体 historical spec implementation。

强制校正参考：

- Product scope：`docs/spec/product-scope/spec.md` §§1.4、2.4、4.2、5.4。
- Engineering roadmap：`docs/spec/engineering-roadmap/spec.md` §§4.1、5.1、5.2、6。
- UI design：`docs/ui-design/INDEX.md` active 文档清单，特别是 `ui-architecture.md`、`user-flow.md`、`module-map.md`、`module-practice-review.md`、`report-dashboard.md`、`resume-module.md`、`review-module.md`、`removed-modules-and-scope.md`。
- Static UI truth source：`ui-design/src/app.jsx` route normalization / `TopBar` / active screens，`ui-design/ui-design-contract.test.mjs` 当前 15 项 UI contract assertions。

### 2.2 候选 subject / plan 表

| Subject / Plan | 状态 | Checklist | Target | Context validation | 主要写入范围 | 并行安全级别 | 执行判断 |
|----------------|------|-----------|--------|--------------------|--------------|--------------|----------|
| `product-scope/(no plan)` | spec `active` | n/a | n/a | n/a | `docs/spec/product-scope` | `read-only-parallel-safe` | 产品范围参考，不直接实施 |
| `engineering-roadmap/001-decompose-subspecs` | `active` / `active` | 9/13 | `docs` | PASS | `docs/spec/engineering-roadmap`, `docs/spec` | `docs-write-serial` | L1 候选；剩余项是 child 创建治理，不是当前 code implementation |
| `historical-spec-implementation-review/001-implement-review-runway` | `active` / `active` | 0/25 | `docs` | PASS | `docs/spec/historical-spec-implementation-review`, `docs/reports`, `docs/work-journal` | `docs-write-serial` | 当前 owner plan |
| `repo-scaffold/001-bootstrap` | `active` / `active` | 16/16 | `repo` | PASS | root containers, `Makefile`, `scripts`, `docs/spec/repo-scaffold` | `docs-write-serial` for lifecycle; implementation otherwise serial | L1 lifecycle drift candidate：checklist 全勾选但 Header 仍 active |
| `ai-gateway-and-model-routing/002-tools-streaming-and-stt` | `draft` / `draft` | 0/19 | `backend` | PASS | `backend/internal/ai/aiclient`, A3/F1/F3/B1 docs | `truth-source-serial` if activated | Draft-gated；不得直接 implement |
| `ai-gateway-and-model-routing/001-aiclient-and-profile-bootstrap` | `completed` / `completed` | 26/26 | `backend` | PASS | `backend/internal/ai/aiclient`, `config/ai-profiles` | `implementation-serial` | Completed; L2 evidence exists in reports/work journal |
| `ci-pipeline-baseline/001-local-quality-gates` | `completed` / `completed` | 17/17 | `repo` | PASS | `Makefile`, `scripts`, `.agent-skills/sync-doc-index` | `implementation-serial` | Completed; no current implementation delta |
| `db-migrations-baseline/001-bootstrap` | `completed` / `completed` | 27/27 | `backend` | PASS | `migrations`, `backend/cmd/migrate`, migration lint | `truth-source-serial` | Completed; DB truth source serial zone |
| `event-and-outbox-contract/001-bootstrap` | `completed` / `completed` | 32/32 | `backend` | PASS | `shared/events.yaml`, `shared/jobs.yaml`, event/job generated outputs | `truth-source-serial` | Completed; event/job truth source serial zone |
| `local-dev-stack/001-bootstrap` | `completed` / `completed` | 20/20 | `repo` | PASS | `deploy/dev-stack`, `Makefile` | `implementation-serial` | Completed; no current implementation delta |
| `openapi-v1-contract/001-bootstrap` | `completed` / `completed` | 28/28 | `contract` | PASS | `openapi`, generated clients, docs renderer | `truth-source-serial` | Completed; OpenAPI truth source serial zone |
| `openapi-v1-contract/002-fixtures-and-mock-source` | `completed` / `completed` | 22/22 | `contract` | PASS | `openapi/fixtures`, fixture projection, `ui-design/src` | `truth-source-serial` | Completed; fixtures/mock truth source serial zone |
| `openapi-v1-contract/003-breaking-change-gate` | `completed` / `completed` | 22/22 | `contract` | PASS | `openapi/baseline`, breaking-change scripts | `truth-source-serial` | Completed; OpenAPI gate serial zone |
| `secrets-and-config/001-bootstrap` | `completed` / `completed` | 40/40 | `platform-config` | PASS | `config`, runtime config, feature flags, secrets | `truth-source-serial` | Completed; config truth source serial zone |
| `shared-conventions-codified/001-bootstrap` | `completed` / `completed` | 20/20 | `docs` | PASS | `docs/spec/shared-conventions-codified` | `docs-write-serial` | Completed docs bootstrap |
| `shared-conventions-codified/002-codegen-pipeline` | `completed` / `completed` | 15/15 | `backend` | PASS | `shared/conventions.yaml`, Go/TS generated helpers | `truth-source-serial` | Completed; shared conventions truth source serial zone |
| `observability-stack/(no plan)` | spec `active` | n/a | n/a | n/a | `docs/spec/observability-stack` | `read-only-parallel-safe` | Active spec without implementation plan; create plan only on demand |
| `prompt-rubric-registry/(no plan)` | spec `active` | n/a | n/a | n/a | `docs/spec/prompt-rubric-registry` | `read-only-parallel-safe` | Active spec without implementation plan; create plan only on demand |

### 2.3 Context validation evidence

Command executed:

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py --context <context.yaml> --docs-root docs --target <defaultTarget>
```

Result: all 15 existing `context.yaml` manifests passed. Normalized roles were `plan,checklist,spec` for every target.

### 2.4 Parallelism classification

Read-only scans are parallel-safe for all listed subjects. Writes are not globally parallel-safe because several candidates share `docs/spec/INDEX.md`, subject `plans/INDEX.md`, generated outputs, or repository truth sources.

Serial zones:

- `shared/conventions.yaml` and generated Go/TS shared helpers
- `openapi/openapi.yaml`, fixtures, baselines, generated clients/docs
- `shared/events.yaml`, `shared/jobs.yaml`, event/job schemas and generated outputs
- `migrations/` and migration lint truth source
- `config/` runtime config, feature flags, env dictionary, AI profiles
- global `docs/spec/INDEX.md`, `docs/reports/INDEX.md`, `docs/work-journal/INDEX.md`

## 3 当前结论

Phase 1 inventory 完成。下一步进入 L1 Reconcile，优先检查：

1. `repo-scaffold/001-bootstrap` lifecycle drift：checklist 16/16 但 plan/checklist Header 仍为 `active`。
2. `engineering-roadmap/001-decompose-subspecs` remaining Phase 3 items 是否是当前执行项，或是否只是 future child 创建规则。
3. `ai-gateway-and-model-routing/002-tools-streaming-and-stt` draft gate 是否仍阻止直接 implementation。
4. Completed truth-source plans 是否已有足够 L2/assessment evidence；若没有新 implementation delta，本 runway 不应无条件重跑 code remediation。

## 4 Phase 2 L1 Reconcile

### 4.1 Findings

#### P1: 历史 plan 缺少当前强制质量门禁分类

影响：当前 `/implement` Step 4.2 要求每个 plan 的 `## 3 质量门禁分类` 明确 Plan 类型、TDD 策略、BDD 策略和替代验证 gate。缺失时，后续重进 `/implement` 会被门禁阻止或要求先路由 `/plan-review --fix`。

证据：以下 9 个 plan 缺少 `## 3 质量门禁分类`。

- `docs/spec/ai-gateway-and-model-routing/plans/001-aiclient-and-profile-bootstrap/plan.md`
- `docs/spec/ai-gateway-and-model-routing/plans/002-tools-streaming-and-stt/plan.md`
- `docs/spec/local-dev-stack/plans/001-bootstrap/plan.md`
- `docs/spec/openapi-v1-contract/plans/001-bootstrap/plan.md`
- `docs/spec/openapi-v1-contract/plans/002-fixtures-and-mock-source/plan.md`
- `docs/spec/openapi-v1-contract/plans/003-breaking-change-gate/plan.md`
- `docs/spec/repo-scaffold/plans/001-bootstrap/plan.md`
- `docs/spec/secrets-and-config/plans/001-bootstrap/plan.md`
- `docs/spec/shared-conventions-codified/plans/001-bootstrap/plan.md`

修复方向：按当前 `docs/spec/README.md` 和 `docs/spec/TEMPLATES.md` 在每个 plan 的背景之后补入质量门禁分类，保持原计划范围不扩张。Completed code/internal/tooling/contract plans 写明历史执行已由原 checklist 的验证项和既有测试证据承担；draft plan 写明激活前不可实施，激活后才进入 `/implement` -> `/tdd`。

#### P2: `repo-scaffold/001-bootstrap` lifecycle 与 checklist 完成状态不一致

影响：plan 和 checklist Header 仍为 `active`，但 checklist 16/16 已完成，`plans/INDEX.md` 也仍投影 active。后续 inventory 会反复把它误判为待实施 plan。

证据：

- `docs/spec/repo-scaffold/plans/001-bootstrap/checklist.md`：16/16 checked。
- `docs/spec/repo-scaffold/plans/001-bootstrap/plan.md`：Header `状态: active`。
- `docs/spec/repo-scaffold/plans/INDEX.md`：该 plan 位于 active row。

修复方向：把 plan/checklist Header 切到 `completed`，更新日期为 `2026-05-04`，同步 `docs/spec/repo-scaffold/plans/INDEX.md`。不修改已完成 checklist 内容，不 reopen 代码实现。

#### P3: `engineering-roadmap/001-decompose-subspecs` 剩余 4 项是 future child 创建规则，不应由本 runway 直接勾选

影响：该 plan 仍有 3.1-3.4 未完成，但内容是“创建任一 P0 workstream child spec/plan 前”的治理规则，不对应当前已选 implementation owner。直接勾选会形成虚假完成记录。

证据：`docs/spec/engineering-roadmap/plans/001-decompose-subspecs/checklist.md` Phase 3 的 4 个未勾选项均以 future child 创建为条件。

修复方向：本 runway 记录为 L1 审查结论，不直接修改或勾选该 roadmap checklist。后续只有在实际创建 child spec/plan 时才由该 owner 原地推进。

#### P3: `ai-gateway-and-model-routing/002-tools-streaming-and-stt` draft gate 仍有效

影响：该 plan 是 `draft`，且正文 `Activation governance` 明确禁止直接进入 `/implement`。本 runway 不应把它作为当前 implementation target。

证据：plan Header 为 `状态: draft`；checklist 顶部声明任何 phase 仅在对应 trigger 出现并完成 ADR / spec 修订后才可勾选。

修复方向：只补质量门禁分类；不切 active、不执行 implementation。

### 4.2 修复预览

待用户确认后执行 `/plan-review --fix` 等价修复：

1. 在 9 个缺失 plan 中补 `## 3 质量门禁分类`，并顺延后续章节编号或保留原修订记录编号修复已有错序。
2. 将 `repo-scaffold/001-bootstrap` plan/checklist Header 切到 `completed` 并同步 `plans/INDEX.md`。
3. 不修改 `engineering-roadmap/001-decompose-subspecs` 未勾选 Phase 3 项。
4. 不激活 `ai-gateway-and-model-routing/002-tools-streaming-and-stt`。

### 4.3 Applied fixes

用户确认方案 A 后已执行 L1 fix：

- 补齐 9 个历史 plan 的 `## 3 质量门禁分类`：
  - `ai-gateway-and-model-routing/001-aiclient-and-profile-bootstrap`
  - `ai-gateway-and-model-routing/002-tools-streaming-and-stt`
  - `local-dev-stack/001-bootstrap`
  - `openapi-v1-contract/001-bootstrap`
  - `openapi-v1-contract/002-fixtures-and-mock-source`
  - `openapi-v1-contract/003-breaking-change-gate`
  - `repo-scaffold/001-bootstrap`
  - `secrets-and-config/001-bootstrap`
  - `shared-conventions-codified/001-bootstrap`
- 将上述 plan 的 Header 版本和更新日期同步到 `2026-05-04`。
- 将 `repo-scaffold/001-bootstrap` plan/checklist Header 切到 `completed`，并由 `sync-doc-index --fix-index` 将该 plan 从 Active 迁移到 Completed。
- 保持 `engineering-roadmap/001-decompose-subspecs` Phase 3 未勾选，原因是这些项是 future child 创建规则。
- 保持 `ai-gateway-and-model-routing/002-tools-streaming-and-stt` 为 `draft`，原因是 `Activation governance` 仍禁止直接 implementation。

### 4.4 L1 verification evidence

已执行并通过：

- `validate_context.py`：15/15 existing `context.yaml` manifests passed for their default targets。
- `sync-doc-index --fix-index`：10 个 `plans/INDEX.md` 投影更新，Post-fix Verification 为 zero drift。
- `sync-doc-index --check`：zero drift。
- `python3 scripts/lint/check_md_links.py docs`：OK。
- docs/spec heading anchor audit：`TOTAL 0`。
- `make docs-check`：Header/INDEX zero drift + Markdown links OK。
- `git diff --check`：通过。

## 5 Phase 3 Implementation Runway

### 5.1 Selection decision

L1 后未启动新的 historical plan implementation。判定如下：

| Candidate | Decision | Reason |
|-----------|----------|--------|
| `engineering-roadmap/001-decompose-subspecs` | 不启动 implementation | 剩余 3.1-3.4 是 future child 创建规则；当前未创建 child spec/plan，不应虚假勾选 |
| `repo-scaffold/001-bootstrap` | 不启动 implementation | checklist 16/16；本轮仅修正 lifecycle drift 并收口为 completed |
| `ai-gateway-and-model-routing/002-tools-streaming-and-stt` | 不启动 implementation | plan/checklist 仍为 draft；Activation governance 禁止直接进入 `/implement` |
| completed truth-source plans | 不启动 implementation | 本轮只补文档质量门禁分类，没有新的 code / generated artifact delta |
| product-scope / observability-stack / prompt-rubric-registry | 不启动 implementation | 当前为 active spec reference 或 no-plan subject；需要设计/实现时另由 owner 创建或修订 plan |

### 5.2 Implementation evidence

本轮没有通过其它 subject 的 `/implement` -> `/tdd` 启动代码实现。原因不是跳过实施，而是 L1 inventory 没有发现可立即启动且安全的 implementation owner。后续若用户指定某个 P0 workstream，应先按 product-scope / engineering-roadmap / docs-ui-design 创建或修订 owner plan，再由该 plan 自己进入 `/implement`。

## 6 Phase 4 L2 Plan Code Review

本轮没有新增代码实现或 generated artifact 变更，因此没有新的 `/plan-code-review --fix` 目标。已有 completed truth-source plans 的历史 L2 / assessment evidence 继续作为当前事实：

- `docs/reports/2026-04-30-a-b-l2-code-review-remediation-assessment.md`
- `docs/reports/2026-05-03-product-contract-ui-scope-alignment-assessment.md`
- `docs/reports/2026-05-04-engineering-roadmap-plan-code-review-alignment-assessment.md`

如果后续启动某个 concrete implementation plan，该 plan 完成后再按本 runway Phase 4 执行对应 `/plan-code-review --fix`。

## 7 Phase 5 Parallelism Gate

本轮写入只由当前 owner 串行完成，没有使用多 agent 或并行写入。涉及的 shared INDEX / report INDEX / plan docs 均由当前 session owner 集成并验证。Truth source serial zones（`shared/`、`openapi/`、events/jobs、migrations、config、generated artifacts）没有代码或生成物写入。

## 8 Phase 6 Final Reconcile

### 8.1 Global verification

已执行并通过：

- `make docs-check`
- docs/spec heading anchor audit：`TOTAL 0`
- `make codegen-check`
- `make test`
- `make build`
- `git diff --check`

`make build` 当前通过的是 D1 `frontend-shell` 前的占位 build target；该 target 输出 `TODO: build implemented by D1 frontend-shell` 并返回 0。

### 8.2 Close-out skills

- `/retrospective`：已创建 [交付复盘报告](./2026-05-04-historical-spec-implementation-review-runway-assessment.md) 并更新 reports INDEX。
- `/bug-report`：不适用。本轮是 docs/governance remediation，不是 bugfix 或回归修复。
- `/work-journal`：将随本次提交记录 `docs(historical-spec): execute implementation review runway`。

### 8.3 Handoff

本 runway checklist 已全部完成。后续新对话若要进入具体 historical implementation，应从以下入口开始：

1. 明确要启动的 owner subject / plan，而不是从本 runway 重新选择全部 historical plans。
2. 若目标是 `ai-gateway-and-model-routing/002-tools-streaming-and-stt`，必须先满足其 Activation governance，把 plan 从 draft 切到 active。
3. 若目标是 P0 UI/API/backend workstream，应先按 product-scope、engineering-roadmap 和 docs-ui-design 创建或修订对应 owner plan，再由该 owner 进入 `/implement`。
