# Bug 记录索引

> 本文件按模块组织所有 Bug 记录，便于快速检索和模式识别。

## Workspace

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|

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
| [BUG-0006](./BUG-0006.md) | openai-compatible adapter assumed provider-specific model naming | medium | resolved | 2026-05-05 | `fix(historical-spec): deep reconcile existing plans` |

## Frontend

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|

## Platform

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0009](./BUG-0009.md) | dev-stack profile catalog drift escaped lint gates | medium | resolved | 2026-05-05 | `fix(ai-provider): harden profile catalog drift gates` |
| [BUG-0008](./BUG-0008.md) | provider registry runtime bootstrap was only test-wired | medium | resolved | 2026-05-05 | `fix(ai-provider): wire provider registry runtime bootstrap` |
| [BUG-0007](./BUG-0007.md) | AI provider contract retained gateway terminology | medium | resolved | 2026-05-05 | `fix(ai-provider): remove gateway terminology from provider contract` |
| [BUG-0002](./BUG-0002.md) | secrets-config completed plan missed runtime binding drift | medium | resolved | 2026-04-30 | `fix(secrets-config): remediate L2 review findings` |

## Schema

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0001](./BUG-0001.md) | OpenAPI breaking-change gate missed composition diffs | medium | resolved | 2026-04-29 | `fix(openapi): tighten breaking-change gate composition diff` |

## Test

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0010](./BUG-0010.md) | mock contract runtime gate missed registry and stale route count | medium | resolved | 2026-05-05 | `fix(mock-contract): harden runtime drift gates` |
| [BUG-0003](./BUG-0003.md) | local quality gates skipped real backend and frontend execution | medium | resolved | 2026-04-30 | `fix(ci-pipeline): remediate local quality gates` |
