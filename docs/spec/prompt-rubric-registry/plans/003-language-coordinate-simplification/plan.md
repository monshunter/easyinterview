# 003 - Language Coordinate Simplification

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 F3 prompt / rubric registry 的 baseline storage 收敛为当前 9 个 feature_key 的 canonical `multi` truth source。运行时 `language` 继续作为目标输出语言和 provenance 字段；`ResolveActive(featureKey, requestedLanguage)` 在没有 exact override 时 fallback 到 `multi`。

本 plan 不修改用户可见 UI、OpenAPI、A3 provider contract 或业务流程。它只约束 F3 config truth source、loader/resolver、seed rows、lint 和 migration gates。

## 2 当前合同

| Surface | 当前合同 | Gate |
|---------|----------|------|
| Prompt files | `config/prompts/<feature_key>/v0.1.0.{yaml,md,schema.json}` only for current 9 feature keys | `make lint-prompts`, prompt lint tests |
| Rubric files | `config/rubrics/<feature_key>/v0.1.0.yaml` only for current 9 feature keys | `make lint-rubrics`, rubric lint tests |
| Resolver | exact language match if an approved override exists; otherwise fallback to `multi` | registry Go tests |
| Output schema | one language-independent schema per current feature key | prompt lint and registry tests |
| Seed rows | migration seed rows align to current 9 canonical prompt/rubric coordinates | migration lint and registry seed tests |
| Override policy | language override requires real semantic rationale and paired prompt/rubric coverage | prompt/rubric lint negative tests |

## 3 Current Feature Key Inventory

| feature_key | Default model profile |
|-------------|-----------------------|
| `target.import.parse` | `target.import.default` |
| `practice.session.first_question` | `practice.first_question.default` |
| `practice.session.follow_up` | `practice.followup.default` |
| `practice.turn.lightweight_observe` | `practice.turn_observe.default` |
| `report.generate` | `report.generate.default` |
| `report.question_assessment` | `report.assessment.default` |
| `resume.parse` | `resume.parse.default` |
| `resume.tailor.gap_review` | `resume.tailor.default` |
| `resume.tailor.bullet_suggestions` | `resume.tailor.default` |

## 4 Completed Implementation Scope

- F3 spec and config README describe canonical `multi` baseline storage and exact-language override policy.
- Prompt/rubric lint accepts current `multi` baseline files and rejects unpaired or unsupported language overrides.
- Registry loader requires canonical `multi` prompt/rubric presence for each current feature key.
- Registry resolver falls back from runtime languages to `multi` while preserving requested language for interpolation/provenance.
- Seed coverage tests align migration rows with current config truth source.
- Current config counts are 9 prompt YAML, 9 prompt Markdown, 9 prompt schema JSON, and 9 rubric YAML.

## 5 Verification Commands

```bash
find config/prompts -name 'v0.1.0.*.*' -print
find config/rubrics -name 'v0.1.0.*.yaml' -print
find config/prompts -mindepth 2 -name 'v0.1.0.yaml' | wc -l
find config/prompts -mindepth 2 -name 'v0.1.0.md' | wc -l
find config/prompts -mindepth 2 -name 'v0.1.0.schema.json' | wc -l
find config/rubrics -mindepth 2 -name 'v0.1.0.yaml' | wc -l
python3 -m pytest scripts/lint/prompt_lint_test.py scripts/lint/rubric_lint_test.py -q
python3 scripts/lint/prompt_lint.py
python3 scripts/lint/rubric_lint.py
cd backend && go test ./internal/ai/registry -count=1
python3 scripts/lint/migrations_lint.py
make lint-prompts
make lint-rubrics
make lint-ai-profile-coverage
python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/prompt-rubric-registry/plans/003-language-coordinate-simplification/context.yaml --target backend
python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check
make docs-check
git diff --check
```

## 6 BDD Applicability

BDD is not applicable. This plan changes internal F3 registry storage, lint, loader/resolver, and seed semantics. User-visible language behavior remains covered by runtime `language`, prompt interpolation, provenance, and each business/frontend owner gate.

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.1 | Compress owner plan to current 9-key canonical multi contract and executable evidence index. |
| 2026-05-24 | 1.0 | Complete language coordinate simplification. |
