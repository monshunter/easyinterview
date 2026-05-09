# Bug 记录索引

> 本文件按模块组织所有 Bug 记录，便于快速检索和模式识别。

## Workspace

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0028](./BUG-0028.md) | targetjob review exposed error envelope and parse runtime drift | medium | resolved | 2026-05-08 | `fix(backend-targetjob): align targetjob review contracts` |
| [BUG-0027](./BUG-0027.md) | targetjob L2 review exposed runtime gate and SSRF gaps | medium | resolved | 2026-05-08 | `fix(backend-targetjob): harden targetjob L2 runtime gates` |
| [BUG-0026](./BUG-0026.md) | targetjob L2 review exposed import parse contract drift | medium | resolved | 2026-05-08 | `fix(backend-targetjob): remediate import parse L2 findings` |
| [BUG-0025](./BUG-0025.md) | analysis failed redline rejected documented provider secret code | medium | resolved | 2026-05-08 | `feat(backend-targetjob): complete import parse bootstrap handoff` |
| [BUG-0024](./BUG-0024.md) | targetjob detail omitted parsed summary provenance | medium | resolved | 2026-05-08 | `feat(backend-targetjob): complete import parse bootstrap handoff` |

## Practice

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0004](./BUG-0004.md) | voice interview surface was removed while unifying practice routes | medium | resolved | 2026-05-02 | `fix(ui-design): restore voice interview surface in practice shell` |

## Review

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0005](./BUG-0005.md) | report follow-up CTAs returned to setup instead of starting sessions | medium | resolved | 2026-05-02 | `fix(ui-design): start report follow-up sessions directly` |

## Materials

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|

## Debrief

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|

## Eval

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0030](./BUG-0030.md) | prompt registry L2 review exposed provenance and no-op gate drift | medium | resolved | 2026-05-09 | `fix(prompt-rubric-registry): remediate provenance L2 findings` |
| [BUG-0006](./BUG-0006.md) | openai-compatible adapter assumed provider-specific model naming | medium | resolved | 2026-05-05 | `fix(historical-spec): deep reconcile existing plans` |

## Frontend

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0032](./BUG-0032.md) | workspace follow-up review exposed synthetic id and fetch race drift | medium | resolved | 2026-05-09 | `fix(frontend-workspace): harden workspace review follow-up` |
| [BUG-0031](./BUG-0031.md) | workspace L2 review exposed route hydration and start flow drift | medium | resolved | 2026-05-09 | `fix(frontend-workspace): remediate workspace L2 findings` |
| [BUG-0029](./BUG-0029.md) | home JD import L2 review exposed privacy and gate drift | medium | resolved | 2026-05-08 | `fix(frontend-home): remediate jd import L2 findings` |
| [BUG-0021](./BUG-0021.md) | frontend shell TopBar drifted from ui-design source structure | medium | resolved | 2026-05-08 | `fix(frontend-shell): restore topbar ui-design source parity` |
| [BUG-0020](./BUG-0020.md) | frontend shell language switch was state-only | medium | resolved | 2026-05-07 | `fix(frontend-shell): restore app shell i18n` |
| [BUG-0019](./BUG-0019.md) | frontend shell review remediation missed build and auth edge gates | medium | resolved | 2026-05-07 | `fix(frontend-shell): harden app shell review remediation` |
| [BUG-0018](./BUG-0018.md) | frontend shell L2 review exposed route and auth wire drift | medium | resolved | 2026-05-07 | `fix(frontend-shell): remediate app shell L2 findings` |

## Platform

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0022](./BUG-0022.md) | speech adapters ignored profile timeouts | medium | resolved | 2026-05-08 | `fix(ai-provider): enforce speech adapter timeouts` |
| [BUG-0017](./BUG-0017.md) | runtime topology lint missed structured producer and owner handoff forms | medium | resolved | 2026-05-07 | `fix(runtime): harden worker topology structured gate` |
| [BUG-0016](./BUG-0016.md) | runtime topology lint missed scripts and raw producer fields | medium | resolved | 2026-05-07 | `fix(runtime): harden worker topology scripts gate` |
| [BUG-0015](./BUG-0015.md) | runtime topology lint missed false-negative worker handoff forms | medium | resolved | 2026-05-06 | `fix(runtime): harden worker topology false-negative gate` |
| [BUG-0014](./BUG-0014.md) | backend runtime topology L2 review exposed worker drift gate gap | medium | resolved | 2026-05-06 | `fix(runtime): harden worker topology drift gate` |
| [BUG-0013](./BUG-0013.md) | backend auth L2 review exposed session contract drift | medium | resolved | 2026-05-06 | `fix(backend-auth): remediate passwordless L2 findings` |
| [BUG-0012](./BUG-0012.md) | AI client L2 review exposed tools, streaming, and observability drift | medium | resolved | 2026-05-06 | `fix(ai-provider): remediate tools streaming stt L2 findings` |
| [BUG-0009](./BUG-0009.md) | dev-stack profile catalog drift escaped lint gates | medium | resolved | 2026-05-05 | `fix(ai-provider): harden profile catalog drift gates` |
| [BUG-0008](./BUG-0008.md) | provider registry runtime bootstrap was only test-wired | medium | resolved | 2026-05-05 | `fix(ai-provider): wire provider registry runtime bootstrap` |
| [BUG-0007](./BUG-0007.md) | AI provider contract retained gateway terminology | medium | resolved | 2026-05-05 | `fix(ai-provider): remove gateway terminology from provider contract` |
| [BUG-0002](./BUG-0002.md) | secrets-config completed plan missed runtime binding drift | medium | resolved | 2026-04-30 | `fix(secrets-config): remediate L2 review findings` |

## Schema

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0023](./BUG-0023.md) | jobs_test referenced removed embedding_upsert constant after capability cleanup | medium | resolved | 2026-05-08 | `fix(event-outbox): drop stale embedding job reference in jobs_test` |
| [BUG-0001](./BUG-0001.md) | OpenAPI breaking-change gate missed composition diffs | medium | resolved | 2026-04-29 | `fix(openapi): tighten breaking-change gate composition diff` |

## Test

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0011](./BUG-0011.md) | mock contract gate ignored empty retired fixture tag directories | medium | resolved | 2026-05-06 | `fix(mock-contract): reject retired fixture tag dirs` |
| [BUG-0010](./BUG-0010.md) | mock contract runtime gate missed registry and stale route count | medium | resolved | 2026-05-05 | `fix(mock-contract): harden runtime drift gates` |
| [BUG-0003](./BUG-0003.md) | local quality gates skipped real backend and frontend execution | medium | resolved | 2026-04-30 | `fix(ci-pipeline): remediate local quality gates` |
