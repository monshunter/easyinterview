# Historical Spec L2 Reconcile 完成验证

> **日期**: 2026-05-05
> **验证人**: Codex

> **说明**: 本报告记录本轮早期 L2 reconcile 快照。用户随后要求忽略历史 PASS / checklist 状态并执行 artifact-level deep reimplementation；最终验证证据以 [Historical Spec Deep Reimplementation Ledger](./2026-05-05-historical-spec-deep-reimplementation-ledger.md) 为准。

**关联计划**: [historical-spec-implementation-review](../spec/historical-spec-implementation-review/plans/001-implement-review-runway/plan.md)

## 1 验证范围

本报告记录本轮在新方案修订后的 fresh L1/L2 收口证据。`15/15 context validation PASS` 只作为 Scope Inventory 入口门禁，不作为 implementation / L2 完成证据。

实际审查范围按用户指定顺序覆盖：

1. `repo-scaffold/001-bootstrap`
2. `shared-conventions-codified/001-bootstrap`
3. `shared-conventions-codified/002-codegen-pipeline`
4. `openapi-v1-contract/001-bootstrap`
5. `openapi-v1-contract/002-fixtures-and-mock-source`
6. `openapi-v1-contract/003-breaking-change-gate`
7. `event-and-outbox-contract/001-bootstrap`
8. `db-migrations-baseline/001-bootstrap`
9. `secrets-and-config/001-bootstrap`
10. `ai-gateway-and-model-routing/001-aiclient-and-profile-bootstrap`
11. `local-dev-stack/001-bootstrap`
12. `ci-pipeline-baseline/001-local-quality-gates`
13. `engineering-roadmap/001-decompose-subspecs`

`ai-gateway-and-model-routing/002-tools-streaming-and-stt` 仍为 draft-gated，未纳入实施范围；`historical-spec-implementation-review/001` 是 completed docs-only orchestration，不作为新的 code implementation target。

## 2 验证结果

### 2.1 Per-plan fresh review

| 范围 | 本轮结论 | Fresh evidence |
|------|----------|----------------|
| `repo-scaffold/001-bootstrap` | 修复 1 个 L2 文档事实漂移：根 README 不再暗示 `test/scenarios/` 已落地 | `make help`、`bash scripts/bootstrap.sh`、`make -n install-hooks`、根 README diff |
| `shared-conventions-codified/001-bootstrap` | 无新增 L2 finding；B1 truth source、Go/TS generated output、lint/parity 当前一致 | `make lint-conventions`、`go test ./internal/shared/... ./cmd/codegen/conventions -count=1`、`pnpm --dir frontend test src/lib/conventions src/lib/ids` |
| `shared-conventions-codified/002-codegen-pipeline` | 无新增 L2 finding；AI shared vocabulary / drift gate 仍对齐当前 generated output | `make lint-conventions`、`make codegen-check` |
| `openapi-v1-contract/001-bootstrap` | 无新增 L2 finding；OpenAPI 当前仍为 12 tag / 34 operation，B1 enum 与 ApiError envelope 对齐 | `make lint-openapi`、`openapi_inventory.py`、`make codegen-check` |
| `openapi-v1-contract/002-fixtures-and-mock-source` | 无新增 L2 finding；34 fixtures、prototype projection tests 和 provenance gate 通过 | `make validate-fixtures`、`python3 -m unittest scripts/codegen/render_openapi_fixture_examples_test.py scripts/lint/validate_fixtures_test.py scripts/codegen/sync_fixtures_from_prototype_test.py` |
| `openapi-v1-contract/003-breaking-change-gate` | 无新增 L2 finding；baseline/current 都是 34 operation，breaking/additive/informational 均 0 | `make openapi-diff` |
| `event-and-outbox-contract/001-bootstrap` | 无新增 L2 finding；16 events / 10 jobs 与 generated Go/TS/schema/baseline 一致 | `make codegen-events-check`、`go test ./cmd/codegen/events ./internal/shared/events ./internal/shared/jobs -count=1`、`pnpm --dir frontend test src/lib/events src/lib/jobs` |
| `db-migrations-baseline/001-bootstrap` | 无新增 L2 finding；baseline DDL、enum/check constraints 与 privacy dry-run 当前一致 | `python3 scripts/lint/migrations_lint.py --repo-root .`、`go test ./internal/migrations ./cmd/migrate -count=1`、`make privacy-delete-dry-run` |
| `secrets-and-config/001-bootstrap` | 无新增 L2 finding；24 env keys、6 feature flags、runtime-config fetcher 与 frontend type surface 对齐 | `make lint-config`、`go test ./internal/platform/config ./internal/platform/featureflag -count=1`、`pnpm --dir frontend test src/lib/runtime-config` |
| `ai-gateway-and-model-routing/001-aiclient-and-profile-bootstrap` | 修复 1 个 L2 code finding：OpenAI-compatible adapter 的 model family 旧命名假设；记录 [BUG-0006](../bugs/BUG-0006.md) | Red/Green focused test、`go test ./internal/ai/aiclient/... -count=1`、vendor-model grep |
| `local-dev-stack/001-bootstrap` | 无新增 L2 finding；compose config、doctor/init shell 语法、AI provider fail-fast 文档边界有效 | `bash -n deploy/dev-stack/scripts/dev-doctor.sh deploy/dev-stack/init/minio/create-buckets.sh`、`docker compose -f deploy/dev-stack/docker-compose.yaml --env-file deploy/dev-stack/.env.example config` |
| `ci-pipeline-baseline/001-local-quality-gates` | 无新增 L2 finding；docs/codegen/test/build gate 重新跑通，docs anchor gate 仍有效 | `make docs-check`、`make codegen-check`、`make test`、`make build` |
| `engineering-roadmap/001-decompose-subspecs` | 修复 L1 owner drift：Phase 3 明确为 future creation rule，plan/checklist 收口为 completed 13/13；未创建 child | `sync-doc-index --fix-index`、`validate_context.py`、`list_context_candidates.py` 不再推荐该 active plan |

### 2.2 全局验证

| Gate | 结果 | 备注 |
|------|------|------|
| `make docs-check` | PASS | Header/INDEX、docs links、docs/spec heading fragments 全部通过 |
| `make codegen-check` | PASS | B1 / B2 / B3 generated output 无 drift |
| `make test` | PASS | Backend Go packages + frontend 10 test files / 49 tests 通过 |
| `make build` | PASS | Backend build 通过；frontend build 当前按 D1 placeholder 输出 `TODO: build implemented by D1 frontend-shell` 并 exit 0 |
| `git diff --check` | PASS | 修复 `plans/INDEX.md` 尾部空行后通过 |

### 2.3 限制与不适用说明

- `make lint-config` 中 gitleaks 二层扫描因本机未安装而按既有脚本跳过；主 pre-commit secret hook、env dictionary、os.Getenv boundary 均已通过。
- 未启动 `ai-gateway-and-model-routing/002-tools-streaming-and-stt`，原因是 plan Header 仍为 `draft` 且 checklist 0/19。
- 未创建任何新的 P0 workstream child spec / plan；`engineering-roadmap/001` Phase 3 仅记录后续创建规则。

## 3 遗留问题

- `list_context_candidates.py` 仍会在没有 active implementation candidate 时显示唯一 draft plan `ai-gateway-and-model-routing/002-tools-streaming-and-stt`。本轮未把它纳入可实施范围；后续可优化该脚本，把 draft-gated plan 与 active candidate 分区展示。

## 4 结论

本轮不以旧的 context validation PASS 作为完成依据，而是对每个现存 historical plan 做了 fresh L1/L2 复查、必要修复和重新验证。除 `ai-gateway-and-model-routing/002-tools-streaming-and-stt` 保持 draft-gated 外，当前 historical spec/plan 与代码事实已闭环。
