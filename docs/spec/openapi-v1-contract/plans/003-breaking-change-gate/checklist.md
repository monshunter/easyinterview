# OpenAPI v1 Contract Breaking-Change Gate Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-03

**关联计划**: [plan](./plan.md)

## Phase 1: baseline 锁定与 diff 工具入口

- [x] 1.1 落地 `openapi/baseline/openapi-v1.0.0.yaml`：从 [001-bootstrap](../001-bootstrap/plan.md) 末态 `openapi/openapi.yaml` 按位拷贝；`info.description` 标注 `BASELINE — DO NOT EDIT`，并说明任何 breaking change 必须走 ADR + spec 修订
- [x] 1.2 落地 `make openapi-diff`：调用 `openapitools/openapi-diff` CLI（或等价）固定版本；`--fail-on-incompatible`；JSON 摘要 stdout；breaking 时 exit ≥1；接入根 `Makefile` 与 `make help`
- [x] 1.3 baseline 选择策略：默认按 SemVer 取 `openapi/baseline/` 下最大版本；允许 `BASELINE_VERSION=` 覆盖；当前只锁 `v1.0.0`，后续 baseline 递增由 spec 修订附带本 plan 修订

## Phase 2: Ruleset 与 privacy-export 白名单

- [x] 2.1 落地 `openapi/diff-config.yaml`（或 wrapper 脚本）：禁止 delete-field / change-type / required-add / enum-remove / endpoint-delete / path-rename / method-change；允许 additive（new endpoint / tag / optional field / new enum value / new optional query / example）
- [x] 2.2 配置 privacy export 白名单：`POST /api/v1/privacy/exports` 从 `501` 切到 `202` 视为 additive；命中时输出 informational 但不报 breaking；wrapper 必须额外检查同 PR `history.md` 是否含对应增量行，缺则 fail；白名单不可扩展到其它 endpoint
- [x] 2.3 落地 wrapper 脚本 `scripts/lint/openapi_diff.py`（如工具配置不够强）：reclassify 工具输出 → spec §4.4 口径；最终退出码以 wrapper 为准；wrapper 启动时打印 `openapi-diff --version`
- [x] 2.4 Phase 2 自检：删字段 → fail（C-4）；加 optional → pass（C-5）；privacy export `501→202` + `history.md` 增量 → pass，缺 history 增量 → fail；revert 后恢复
- [x] 2.5 L2 remediation：补齐 wrapper 对 `oneOf` / `allOf` / `anyOf` composition schema 的 breaking diff 检测；修复 privacy export 白名单 history gate 的默认 base-ref 语义，并允许 `Makefile` 显式覆盖 history ref

## Phase 3: ADR 模板与升级流程

- [x] 3.1 落地 `docs/spec/openapi-v1-contract/decisions/TEMPLATE.md`：与 engineering-roadmap/decisions 风格一致；字段 `ID` (`OPENAPI-NNN`) / `状态` / `日期` / `背景` / `决策` / `影响` / `迁移与回滚` / `相关` / `审计`；首段强调 breaking change 必须先 accepted ADR 再改 yaml
- [x] 3.2 落地 `openapi/baseline/README.md`：v1.0.x patch / v1.x.0 minor（accumulated additive ≥ 5 个新 endpoint）/ v2.0.0 major（任何 breaking）阈值；标注 `openapi-diff` 最低工具版本；声明阈值默认值，后续由 spec §3.2 决策
- [x] 3.3 在 `openapi-v1-contract/history.md` 追加「修订规则」章节：任何 schema / endpoint / response status 变更必须递增 history；privacy export 白名单切换显式标记；白名单外 breaking 必须引用 ADR id

## Phase 4: Verification + B2 freeze handoff

- [x] 4.1 spec C-4 / C-5 复跑自检：删字段 / 加 optional / privacy export 白名单切换 + history 增量；每段命令 + 退出码贴入工作日志
- [x] 4.2 spec C-10 B2 freeze handoff：确认 001 / 002 全部勾选；本 plan Phase 1–3 全部勾选；`make codegen-check && make validate-fixtures && make openapi-diff` 一键全绿；命令日志贴入工作日志
- [x] 4.3 implementation 准入 gate 解锁声明：工作日志声明 B2 的 implementation 准入 gate 已闭合；C / D 域可直接消费 generated 类型；不修改 engineering-roadmap/001 父 checklist
- [x] 4.4 文档与 INDEX 同步：plans/INDEX.md 把 001 / 002 / 003 切到 completed；history.md 追加版本变更（视范围决定保持 1.3 还是递增）；`/sync-doc-index --check` 通过

## Phase 5: v1.8 baseline remediation

- [x] 5.1 在 001 / 002 完成历史 v1.8 remediation 后，重新冻结 `openapi/baseline/openapi-v1.0.0.yaml`，确保 baseline 含当时 v1.8 freeze 清单与 `DELETE /api/v1/me`；不得错误创建 `openapi-v1.0.1.yaml` 来掩盖当时的 v1.0.0 baseline 漂移；当前 freeze 已由 Phase 6 收敛为 12 tag / 34 operation
- [x] 5.2 更新 `scripts/lint/openapi_diff.py`、`openapi/diff-config.yaml` 与 baseline README 中 endpoint inventory 到当时 v1.8 freeze 清单；privacy export `501→202` 白名单仍仅作用于 `POST /api/v1/privacy/exports`；当前 endpoint inventory 以 Phase 6 的 34 为准
- [x] 5.3 修正 checklist / context specVersion 到 current spec v1.8；本 remediation 未完成前 plan 保持 active
- [x] 5.4 复跑 `make openapi-diff`、`make codegen-check`、`make validate-fixtures`，确认当时 v1.8 baseline / diff / fixtures gate 均通过；Phase 6 后当前 gate 已确认 34 operation

## Phase 6: product-scope v1.2 baseline remediation

- [x] 6.1 Red: 在 001 / 002 未完成前运行 `make openapi-diff`，确认 baseline/current 或 inventory 仍能暴露旧 37 endpoint 漂移
  - 2026-05-03: `make openapi-diff` exit 2，summary `baselineOperations=37` / `currentOperations=34`，breaking findings 包含 `tag-removed` Mistakes/Growth、3 个旧 endpoint removed、旧 PracticeMode / PracticeGoal enum、`MistakeStatus` / `MistakeEntry` / `GrowthOverview*` schemas 与旧字段删除。
- [x] 6.2 Green: 重新冻结 `openapi/baseline/openapi-v1.0.0.yaml` 到 12 tag / 34 endpoint，并更新 baseline README / diff 说明；不创建 v1.0.1
  - 2026-05-03: 用当前 `openapi/openapi.yaml` 重新冻结 `openapi/baseline/openapi-v1.0.0.yaml`，保留 `BASELINE — DO NOT EDIT` 描述；更新 `openapi/diff-config.yaml` `contractInventory.endpointCount=34` 与 baseline README 12 tag / 34 operation 说明，未创建 `openapi-v1.0.1.yaml`。
- [x] 6.3 Verify: `make openapi-diff` 通过；privacy export `501→202` 白名单仍只作用于 `POST /api/v1/privacy/exports`
  - 2026-05-03: `make openapi-diff` exit 0，payload `expectedOperations=34`、`baselineOperations=34`、`currentOperations=34`、`breaking=0`；`python3 -m unittest scripts/lint/openapi_diff_test.py scripts/lint/openapi_inventory_test.py scripts/lint/validate_fixtures_cli_test.py` 35 tests OK，继续覆盖 privacy export 501→202 whitelist history gate。
