# Spec 索引

> 本索引按 subject 分组记录 `docs/spec/*/spec.md` 的版本和状态。
>
> 状态列标记 `_pending_`（斜体）的行是顶层 [`engineering-roadmap`](./engineering-roadmap/spec.md) 已规划但**尚未 spawn** 的 child subspec 占位；spawn 后第一列将变为指向真实 `spec.md` 的链接，状态从对应 Header 投影。占位行因没有链接，会被 `/sync-doc-index` 标记为 `non-standard entry` 警告，这是预期行为。

---

## 1 顶层规划

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| [product-scope](./product-scope/spec.md) | 1.3 | active | 2026-05-03 | [plans](./product-scope/plans/) |
| [engineering-roadmap](./engineering-roadmap/spec.md) | 2.3 | active | 2026-05-03 | [plans](./engineering-roadmap/plans/) |

## 2 P0 MVP（Wave 0–5）

### 2.1 Layer A · Foundation

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| [repo-scaffold](./repo-scaffold/spec.md) | 1.1 | active | 2026-04-29 | [plans](./repo-scaffold/plans/) |
| [local-dev-stack](./local-dev-stack/spec.md) | 1.5 | active | 2026-04-29 | [plans](./local-dev-stack/plans/) |
| [ai-gateway-and-model-routing](./ai-gateway-and-model-routing/spec.md) | 1.7 | active | 2026-04-29 | [plans](./ai-gateway-and-model-routing/plans/) |
| [secrets-and-config](./secrets-and-config/spec.md) | 1.9 | active | 2026-05-03 | [plans](./secrets-and-config/plans/) |
| [ci-pipeline-baseline](./ci-pipeline-baseline/spec.md) | 1.3 | active | 2026-04-29 | [plans](./ci-pipeline-baseline/plans/) |

### 2.2 Layer B · Contract

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| [shared-conventions-codified](./shared-conventions-codified/spec.md) | 1.7 | active | 2026-05-03 | [plans](./shared-conventions-codified/plans/) |
| [openapi-v1-contract](./openapi-v1-contract/spec.md) | 1.9 | active | 2026-05-03 | [plans](./openapi-v1-contract/plans/) |
| [event-and-outbox-contract](./event-and-outbox-contract/spec.md) | 1.5 | active | 2026-05-03 | [plans](./event-and-outbox-contract/plans/) |
| [db-migrations-baseline](./db-migrations-baseline/spec.md) | 1.7 | active | 2026-05-03 | [plans](./db-migrations-baseline/plans/) |

### 2.3 Layer C · Backend（P0）

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| backend-auth（待 W2 spawn） | – | _pending_ | – | – |
| backend-upload（待 W2 spawn） | – | _pending_ | – | – |
| backend-profile（待 W2 spawn） | – | _pending_ | – | – |
| backend-async-runtime（待 W2 spawn） | – | _pending_ | – | – |
| backend-targetjob（待 W3 spawn） | – | _pending_ | – | – |
| backend-practice（待 W3 spawn） | – | _pending_ | – | – |
| backend-review（待 W3 spawn） | – | _pending_ | – | – |
| backend-resume（待 W3 spawn） | – | _pending_ | – | – |
| backend-debrief（待 W3 spawn） | – | _pending_ | – | – |

### 2.4 Layer D · Frontend（P0）

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| frontend-shell（待 W2 spawn） | – | _pending_ | – | – |
| frontend-home-job-picks-and-parse（待 W2 spawn） | – | _pending_ | – | – |
| frontend-workspace-and-practice（待 W2 spawn） | – | _pending_ | – | – |
| frontend-report-dashboard（待 W2 spawn） | – | _pending_ | – | – |
| frontend-resume-workshop（待 W2 spawn） | – | _pending_ | – | – |
| frontend-debrief（待 W2 spawn） | – | _pending_ | – | – |

### 2.5 Layer E · Integration（P0）

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| mock-contract-suite（待 W2 spawn） | – | _pending_ | – | – |
| e2e-scenarios-p0（待 W4 spawn） | – | _pending_ | – | – |
| release-gate-and-rollout（待 W4 spawn） | – | _pending_ | – | – |

### 2.6 Layer F · Quality 横切（P0）

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| [observability-stack](./observability-stack/spec.md) | 1.3 | active | 2026-05-03 | [plans](./observability-stack/plans/) |
| analytics-funnel（待 W2 spawn） | – | _pending_ | – | – |
| [prompt-rubric-registry](./prompt-rubric-registry/spec.md) | 1.3 | active | 2026-04-29 | [plans](./prompt-rubric-registry/plans/) |

## 3 P1 Beta

### 3.1 Layer C · Backend（P1）

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| backend-readiness-signals（待 P1 spawn） | – | _pending_ | – | – |
| backend-retrieval（待 P1 spawn） | – | _pending_ | – | – |
| backend-privacy（待 P1 spawn） | – | _pending_ | – | – |

### 3.2 Layer D · Frontend（P1）

P1 前端增强不单独创建恢复旧模块的 child；后续只挂到 D5 `frontend-resume-workshop`、D6 `frontend-debrief`、D1 `frontend-shell` 或其他已保留 P0 前端 child 的新 plan。

### 3.3 Layer E · Integration（P1）

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| e2e-scenarios-p1（待 P1 spawn） | – | _pending_ | – | – |

### 3.4 Layer F · Quality 横切（P1）

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| privacy-and-audit-runtime（待 P1 spawn） | – | _pending_ | – | – |

## 4 P2 Deferred Capabilities

### 4.1 Layer C · Backend（P2）

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| backend-source-intel（待 P2 spawn） | – | _pending_ | – | – |
| backend-voice-stt（待 P2 spawn） | – | _pending_ | – | – |

### 4.2 Layer D · Frontend（P2）

| Subject | 版本 | 状态 | 更新日期 | Plans |
|---------|------|------|----------|-------|
| frontend-voice-production（待 P2 spawn） | – | _pending_ | – | – |
