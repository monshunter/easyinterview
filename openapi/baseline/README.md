# openapi/baseline

Frozen OpenAPI snapshots that anchor the [breaking-change gate](../../scripts/lint/openapi_diff.py).
Each `openapi-v<MAJOR>.<MINOR>.<PATCH>.yaml` here is a byte-equivalent freeze
of `../openapi.yaml` at the moment the corresponding SemVer was published,
with only the OpenAPI `info.description` field overwritten to carry the
`BASELINE — DO NOT EDIT` marker.

`make openapi-diff` (delivered by [plan 003-breaking-change-gate](../../docs/spec/openapi-v1-contract/plans/003-breaking-change-gate/plan.md))
compares `../openapi.yaml` against the SemVer-max baseline by default, or
against a pinned `BASELINE_VERSION=vX.Y.Z`.

## DO NOT EDIT

Baseline files exist so the gate can prove that `openapi.yaml` did not
introduce a breaking change. Editing a baseline to make the gate pass is a
governance violation — it bypasses the audit trail and produces silent
incompatibility for downstream consumers (Go DTOs, TS client, Prism mock,
docs site).

Current `openapi-v1.0.0.yaml` inventory: 38 operations across 10 tags:
Auth, Uploads, Resumes, TargetJobs, PracticePlans, PracticeSessions, Reports,
ResumeTailor, Jobs, and Privacy. It includes `DELETE /api/v1/me`
(`operationId=deleteMe`), flat Resume operations including PDF source preview,
TargetJob archive (`operationId=archiveTargetJob`), Practice session / voice
contracts, and Auth `updateMe`, which owns both profile completion and account
display-preference updates. The
freeze also includes protected failed-report same-ID regeneration
(`operationId=regenerateFeedbackReport`). The
project is still in a pre-launch P0 phase. A `v1.0.0 pre-release correction`
may re-freeze this file in place only when all of the following hold: explicit
product-owner authorization, an accepted OPENAPI ADR, a merge-base old-baseline
finding set that exactly matches the ADR, same-change spec/history/fixtures/
codegen/consumer migration, and a final current-baseline clean diff. Editing the
baseline first or using the resulting zero diff as authorization is forbidden.
The preserved OPENAPI-001 merge-base audit is tracked at
[`audits/OPENAPI-001-report-direct-semantics.json`](./audits/OPENAPI-001-report-direct-semantics.json);
it records the old-baseline Git source and the exact 33 breaking + 3 additive
finding set before any baseline re-freeze.
The current runtime content-limit correction is preserved at
[`audits/OPENAPI-006-runtime-content-limits.json`](./audits/OPENAPI-006-runtime-content-limits.json);
it records the old-baseline Git source and exact 1 breaking + 8 additive finding
set before the guarded in-place re-freeze.
The OPENAPI-008 generic current-user update correction is preserved at
[`audits/OPENAPI-008-account-theme-update-me.json`](./audits/OPENAPI-008-account-theme-update-me.json);
it records the unchanged old baseline, exact 3 breaking + 4 additive wrapper findings,
and the separately locked `completeMyProfile -> updateMe` operationId invariant.

After a baseline is release-ready or published, follow [the SemVer upgrade
flow](#semver-upgrade-flow) below and never modify that existing baseline file.

## SemVer upgrade flow

These thresholds are the **default values** locked by plan 003 Phase 3.2.
The [openapi-v1-contract spec §3.2](../../docs/spec/openapi-v1-contract/spec.md#32-待确认事项)
lists "v1.0.1 / v1.1.0 升级阈值" as a 待确认 item; the values below stand
until that decision is recorded.

| 升级类型 | 触发条件 | 是否新增 baseline 文件 | history.md / spec / ADR 要求 |
|----------|----------|----------------------|------------------------------|
| **v1.0.0 pre-release correction** | 未上线且未发布；product owner 明确授权；accepted ADR；所有 consumer 同批迁移 | 原地 re-freeze，但必须先保存 merge-base finding artifact | spec/history/ADR + exact finding + fixtures/codegen/consumer gates 全部必需 |
| **v1.0.x patch** | 仅 fixture / example / 文案修订；schema 与 endpoint 集合不变 | 不强制递增；如要递增需在 PR 描述中说明动机 | 仍需 `history.md` 增量记录；不需要 ADR；spec 通常不升版 |
| **v1.x.0 minor** | release-ready baseline 已发布后，additive 累积 ≥ 5 个新 endpoint，或显著新 tag / 新业务领域 | **必须**新增 `openapi-v1.<X>.0.yaml`；既有 baseline 保留 | `history.md` 递增 + 相关 plan 增量；不强制 ADR（仅 additive 时） |
| **v2.0.0 major** | release-ready/published baseline 后的任何 breaking change：删字段 / 改字段类型 / required 新增 / 删 endpoint / 删 enum 值 / 改 method / 重命名 path（除已纳入白名单的 P0 例外） | **必须**新增 `openapi-v2.0.0.yaml` | `history.md` 递增 + spec 修订 + **必须**有 `状态: accepted` 的 ADR（[OPENAPI-NNN-...](../../docs/spec/openapi-v1-contract/decisions/TEMPLATE.md)） |

阈值校准触发条件：每次实际触发 minor / major 时，spec §3.2 owner 把当时的执行
理由回填到 §3.2，并在本 README 调整阈值默认值（如有调整）。

## Tooling

| 工具 | 最低版本 | 备注 |
|------|---------|------|
| `scripts/lint/openapi_diff.py` (wrapper) | `wrapper-1.0.0` | 由本 plan 锁定；`make openapi-diff` 启动时打印于 stderr，与本表不一致即报警 |
| `python3` | 3.11+ | 与 [openapi/README.md](../README.md) tooling baseline 一致 |
| `PyYAML` | 与 repo `requirements*` 一致 | wrapper 仅依赖 `yaml.safe_load` |
| `OpenAPITools/openapi-diff` (可选) | 暂未启用 | wrapper 直接实现 spec §4.4 规则；如未来引入 OpenAPITools CLI，需在 [openapi/diff-config.yaml](../diff-config.yaml) `tooling` 中固定版本，且 wrapper 仍持有最终退出码 |

`make openapi-diff` 默认使用 [openapi/diff-config.yaml](../diff-config.yaml)
中的 `tooling.historyDiffBase`（当前为 `main`）与 `HEAD` 的 merge-base 作为
privacy export 白名单 history 增量比较基准；如果该 ref 不存在，会回退到
`main` / `master` 候选，最后才回退 `HEAD`。需要复现既有自检或临时指定基准时，
使用 `HISTORY_REF=<git-ref> make openapi-diff`。

## Whitelist

[openapi/diff-config.yaml](../diff-config.yaml) 维护唯一的状态码切换白名单：
`POST /api/v1/privacy/exports` 从 `501` 切到 `202`（spec §3.1 D-12 / §4.4 P0
例外）。命中白名单时 wrapper 把对应 finding 降级为 informational，但同 PR
必须递增 [history.md](../../docs/spec/openapi-v1-contract/history.md) 表中的
对应行；缺增量则 wrapper 重新升级为 breaking 并退出码 1。该检查按 base
branch merge-base 比较，因此 history 行随 feature branch commit 一起提交后仍应通过。

任何对白名单的扩展（新 path / method / 状态码组合）都必须先有 `状态: accepted`
的 ADR + 本 spec 修订 + 本 README 阈值表更新。
