# OpenAPI v1 Contract Breaking-Change Gate Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-04-29

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
- [x] 4.3 W2 implementation 准入 gate 解锁声明：工作日志声明 B2 的 W2 implementation 准入 gate 已闭合；C / D 域可直接消费 generated 类型；不修改 engineering-roadmap/001 父 checklist
- [x] 4.4 文档与 INDEX 同步：plans/INDEX.md 把 001 / 002 / 003 切到 completed；history.md 追加版本变更（视范围决定保持 1.3 还是递增）；`/sync-doc-index --check` 通过
