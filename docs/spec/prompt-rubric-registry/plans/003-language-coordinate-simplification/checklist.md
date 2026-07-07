# 003 - Language Coordinate Simplification Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Completed Owner Gates

- [x] Current F3 spec and config README describe canonical `multi` baseline storage and language override policy.
  <!-- verified: 2026-05-24 method=docs-contract evidence="prompt-rubric-registry spec and config prompt/rubric README updated; sync-doc-index and docs-check passed." -->
- [x] Prompt and rubric lint enforce current baseline storage, output-language instruction, override pairing, hash drift, schema drift, and rubric weight rules.
  <!-- verified: 2026-05-24 method=lint-tests evidence="prompt_lint_test.py, rubric_lint_test.py, prompt_lint.py, rubric_lint.py, make lint-prompts, and make lint-rubrics passed." -->
- [x] Registry loader/resolver and seed coverage align with current canonical `multi` truth source.
  <!-- verified: 2026-05-24 method=go-tests evidence="go test ./backend/internal/ai/registry -count=1 passed with loader, resolver, snapshot, and seed coverage tests." -->
- [x] Current config inventory has 9 prompt YAML, 9 prompt Markdown, 9 prompt schema JSON, and 9 rubric YAML files.
  <!-- verified: 2026-07-07 method=config-counts evidence="find config/prompts -mindepth 2 -name 'v0.1.0.yaml' | wc -l => 9; v0.1.0.md => 9; v0.1.0.schema.json => 9; find config/rubrics -mindepth 2 -name 'v0.1.0.yaml' | wc -l => 9." -->
- [x] Current owner docs only carry the 9-key canonical `multi` implementation inventory.
  <!-- verified: 2026-07-07 method=prompt-rubric-003-owner-compression evidence="Updated prompt-rubric-registry/003 owner docs to v1.1 completed. PASS: targeted stale-wording grep returned no matches; validate_context.py prompt-rubric-registry/003 backend PASS; sync-doc-index --fix-index updated prompt-rubric plans INDEX; make lint-prompts PASS (9 files); make lint-rubrics PASS (9 files); make lint-ai-profile-coverage PASS; sync-doc-index --check PASS; make docs-check PASS; git diff --check PASS; make lint-core-loop-pruning-surface PASS real_residuals=0." -->

## BDD-Gate

> **BDD 不适用**: 本 plan 收敛内部 registry truth source、lint、loader/resolver 和 seed semantics，不新增用户可见 UI、新 HTTP API 行为或端到端业务工作流。多语言用户体验由 runtime `language`、prompt interpolation、provenance、业务 owner 和前端 i18n/BDD gate 承接。

## Evidence Commands

```bash
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
