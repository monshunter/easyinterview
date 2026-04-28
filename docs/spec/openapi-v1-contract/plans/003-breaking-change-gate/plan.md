# OpenAPI v1 Contract Breaking-Change Gate

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-28

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [openapi-v1-contract spec](../../spec.md) §3.1 D-10 / §4.4 / §6 C-4 / C-5 / C-10 锁定的 breaking-change linter 与 v1.0.0 freeze 治理流程落到仓库：

- 锁定 `openapi/baseline/openapi-v1.0.0.yaml` 作为 [001-bootstrap](../001-bootstrap/plan.md) 末态产出的 v1.0.0 freeze 快照；
- 落地 `make openapi-diff`：调用 `openapi-diff`（OpenAPITools，或等价）比较当前 `openapi/openapi.yaml` 与 baseline；
- 配置规则集：禁止 delete-field / change-type / required-add / enum-remove / endpoint-delete / path-rename / method-change；允许 additive；privacy export `501 → 202` 切换显式白名单（spec D-12）；
- 落地破坏性变更走 ADR 的工作流模板（`docs/spec/openapi-v1-contract/decisions/TEMPLATE.md` + `history.md` 增量约定）；
- 通过本 plan Phase 4 的本地命令证明 spec §6 中 C-4 / C-5 / C-10 已成立，并把 B2 三个 child 的 executable freeze handoff 收口为「W2 implementation 准入 gate」可放行状态。

本 plan 不实现远端 CI required check / label workflow（归 [A5 ci-pipeline-baseline](../../../ci-pipeline-baseline/spec.md) 后续触发）；不修改 fixture 内容（归 002）；不修改 schema / endpoint 范围（归 001 + 后续修订）。

## 2 背景

[engineering-roadmap §6 关键路径](../../../engineering-roadmap/spec.md#6-关键路径与并行机会) 把 B2 列为 DAG 瓶颈节点：「一旦 codegen 投产，破坏性变更会触发跨 spec 雪球」。spec §3.1 D-10 / §4.4 / §6 C-10 把 v1.0.0 freeze 后的演进路径绑死在 additive-only + ADR 例外两条规则上。本 plan 是 §7 关联计划列出的 3 个 child 中第三个，承担把 freeze gate 从「书面约束」变为「本地可执行 gate」的最后一公里。

执行本 plan 前必须确认：

- [001-bootstrap](../001-bootstrap/plan.md) Phase 4 已完成：`openapi/openapi.yaml` 已是 v1.0.0 锁定形态。
- [002-fixtures-and-mock-source](../002-fixtures-and-mock-source/plan.md) Phase 4 已完成：privacy export 501 fixture 与 provenance fixture 已就绪，作为白名单与回归测试样本。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有 baseline 锁定与 diff 工具入口；Phase 2 起来就有 ruleset 与 privacy export 白名单；Phase 3 起来就有 ADR 模板与升级流程；Phase 4 收口 3 项 AC + B2 freeze handoff。本 plan 不引入 BDD 资产。

## 3 实施步骤

### Phase 1: baseline 锁定与 diff 工具入口

#### 1.1 baseline 文件落地

把 [001-bootstrap](../001-bootstrap/plan.md) 末态 `openapi/openapi.yaml` 拷贝到 `openapi/baseline/openapi-v1.0.0.yaml`（按位拷贝；不再生成）。在文件头部以 OpenAPI `info.description` 注明：

```yaml
info:
  description: |-
    BASELINE — DO NOT EDIT.
    本文件是 openapi-v1-contract spec v1.3 锁定的 v1.0.0 freeze 快照。
    任何 breaking change 必须走 ADR + 本 spec 修订流程，新版本写入 openapi/baseline/openapi-v<MAJOR>.<MINOR>.<PATCH>.yaml。
```

`openapi/baseline/` 目录下保留每个已发布版本的快照；不删除历史 baseline。

#### 1.2 `make openapi-diff`

落地 `make openapi-diff`：

- 工具默认采用 [OpenAPITools/openapi-diff](https://github.com/OpenAPITools/openapi-diff) CLI（npm `openapi-diff`，或 java jar）；版本固定（README 中标注最低版本）。
- 调用方式：`openapi-diff openapi/baseline/openapi-v1.0.0.yaml openapi/openapi.yaml --fail-on-incompatible`。
- 输出：JSON 摘要写入 stdout；breaking 项归类到 `removed-paths` / `changed-types` / `required-fields-added` / `enum-values-removed` / `path-renamed` / `method-changed` 等。
- 退出码：detect 到 breaking change 时 exit ≥1；只有 additive 时 exit 0（允许 stdout 输出 informational 警告）。
- 接入根 `Makefile`；`make help` 自动包含。

#### 1.3 baseline 选择策略

`make openapi-diff` 默认对最新 baseline 比对（按 SemVer 取 `openapi/baseline/` 下最大版本）。允许通过 `BASELINE_VERSION=v1.0.0` 显式指定。本 plan 当前只锁 `v1.0.0`；后续如发布 `v1.1.0` 由 spec 修订附带本 plan 修订。

### Phase 2: Ruleset 与 privacy-export 白名单

#### 2.1 默认规则配置

落地 `openapi/diff-config.yaml`（如工具支持配置文件；否则用 wrapper 脚本 `scripts/lint/openapi_diff.py` 包装 `openapi-diff` 输出再 enforce）：

- **禁止**：删除 endpoint / 重命名 path / 改 method / 删除 schema 字段 / 改字段类型 / 把 optional 改成 required / 删除 enum 值（spec §4.4）。
- **允许（additive）**：新增 endpoint / 新增 tag / 新增 optional 字段 / 新增 string-typed enum 值 / 新增可选 query 参数 / 新增 example。

如 `openapi-diff` 内置规则集与上面口径不一致，wrapper 脚本必须在调用后做一次 reclassification：把工具误报的 informational 升级为 breaking、或把 spec §4.4 的 additive 项降级为 informational，确保最终退出码与 spec 一致。

#### 2.2 privacy export 白名单（spec D-12 / §4.4）

`POST /api/v1/privacy/exports` 从 P0 `501 ApiErrorResponse` 切到 P1 `202 PrivacyRequestWithJob` 是「已预留能力变为可用」，必须通过白名单识别为 additive：

- 在 `openapi/diff-config.yaml`（或 wrapper 脚本）中维护 `whitelist.responseStatusTransitions`：`{ path: "/privacy/exports", method: "POST", from: "501", to: "202" }`。
- 命中白名单时，`make openapi-diff` 输出 informational 警告而非 breaking；wrapper 脚本必须额外要求 `history.md` 在同一 PR 中递增一行（通过 git diff 跨文件检查）；缺 history 增量则 fail。
- 白名单仅对 privacy export 生效，不可扩展到其它 endpoint；扩展必须 ADR + 本 spec 修订。

#### 2.3 wrapper 脚本（必要时）

如 `openapi-diff` 工具自身配置不足以覆盖 §2.1 / §2.2，落地 `scripts/lint/openapi_diff.py`（或 Go 实现）：

- 输入：原始 `openapi-diff` JSON 输出 + `openapi/diff-config.yaml`。
- 输出：reclassified JSON + 最终退出码（按 spec §4.4 规则）。
- 接入 `make openapi-diff`：先调用工具拿原始结果，再调用 wrapper enforce；最终退出码以 wrapper 为准。

#### 2.4 Phase 2 自检

- 在分支上临时删除 `target_jobs.title` 字段（未提交）：`make openapi-diff` 失败，stdout 列出 `removed-fields: [TargetJob.title]`，exit ≥1（spec C-4）。revert 后恢复。
- 临时给 `practice_plans` schema 加一个 optional `metadata` 字段：`make openapi-diff` 仅 informational，exit 0（spec C-5）。revert 后恢复。
- 临时把 `POST /api/v1/privacy/exports` 响应从 501 改成 202（保持其它字段不动）+ 在 `history.md` 写一行升级记录：`make openapi-diff` 通过；不写 `history.md` 增量时 fail。revert 后恢复。

### Phase 3: ADR 模板与升级流程

#### 3.1 ADR 模板

落地 `docs/spec/openapi-v1-contract/decisions/TEMPLATE.md`（与 [engineering-roadmap/decisions](../../../engineering-roadmap/decisions/) 风格一致）：

- 字段：`ID`（`OPENAPI-NNN`）/ `状态`（draft / accepted / superseded）/ `日期` / `背景` / `决策` / `影响` / `迁移与回滚` / `相关` / `审计`。
- 模板首段强调：任何 breaking change（`make openapi-diff` 报 breaking 且不在白名单）必须先有一份 accepted 状态的 ADR，再修改 `openapi/openapi.yaml`，再递增 baseline 与 history。
- ADR 文件命名：`docs/spec/openapi-v1-contract/decisions/OPENAPI-NNN-<short>.md`。

#### 3.2 SemVer 升级流程文档

落地 `openapi/baseline/README.md`：

- v1.0.x patch：仅 fixture / example / 文案修订，不改 schema；不强制递增 baseline 文件，但需记 history。
- v1.x.0 minor：accumulated additive change ≥ 5 个新 endpoint 或显著新 tag 时，触发 baseline 递增；新 baseline 文件 `openapi-v1.<X>.0.yaml`，旧 baseline 保留。
- v2.0.0 major：任何 breaking change（含白名单外的字段变更），必须 ADR + 本 spec 修订 + 新 baseline。
- 上述阈值默认值，后续由 spec §3.2 待确认事项决策。

#### 3.3 history.md 增量约定

更新 [openapi-v1-contract/history.md](../../history.md) 的写作规则（在 Header 下追加一段「修订规则」章节，本 plan 完结时一并更新）：

- 任何 schema / endpoint / response status 变更都必须递增 history。
- privacy export 白名单切换（501 → 202）必须显式标记「白名单 additive」与对应 spec / plan 版本号。
- 白名单外的 breaking change 必须引用 ADR id。

### Phase 4: Verification + B2 freeze handoff

#### 4.1 spec C-4 / C-5 自检（复跑）

复跑 Phase 2.4 三段构造：删字段、加 optional、privacy export 501→202 + history。每段命令 + 退出码贴入工作日志。

#### 4.2 spec C-10 — B2 executable freeze handoff

依次确认：

- [001-bootstrap](../001-bootstrap/plan.md) Phase 4 全部勾选（`openapi/openapi.yaml` v1.0.0、codegen、drift check）；
- [002-fixtures-and-mock-source](../002-fixtures-and-mock-source/plan.md) Phase 4 全部勾选（fixtures、provenance、隐私脱敏、Prism smoke）；
- 本 plan Phase 1 / 2 / 3 全部勾选（baseline、ruleset、ADR 模板）；
- `make codegen-check && make validate-fixtures && make openapi-diff` 一键全绿。

把上述命令日志贴入工作日志；spec §6 C-10 成立。

#### 4.3 W2 implementation 准入 gate 解锁声明

- 在工作日志中明确声明：本 plan 完结后 [engineering-roadmap §5.2 / §5.7](../../../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0) W2 准入 gate 关于 B2 的部分已闭合；C 全域与 D 全域 child 在 W2 启动时可直接消费 `backend/internal/api/generated/` 与 `frontend/src/api/generated/`。
- 不修改 [engineering-roadmap/001-decompose-subspecs](../../../engineering-roadmap/plans/001-decompose-subspecs/checklist.md) 父 checklist；C-10 成立证据由本 plan 持有。

#### 4.4 文档与 INDEX 同步

- 本 plan checklist 全部勾选；Phase 4 关键命令日志贴入工作日志。
- 更新 [openapi-v1-contract/plans/INDEX.md](../INDEX.md)：001 / 002 / 003 三个 plan 状态切到 completed。
- 更新 [openapi-v1-contract/history.md](../../history.md)：追加一行版本变更（版本可保持 1.3 不变；如 ADR 模板与 baseline 治理引发约束补充，按需递增到 1.4）。
- `/sync-doc-index --check` 通过。

## 4 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-4 / C-5 / C-10 全部成立，证据贴入工作日志。
- 本 plan checklist 全部勾选；Phase 4 复跑日志贴入工作日志。
- B2 三个 child 的 executable freeze handoff 收口；W2 implementation 准入 gate 关于 B2 的部分解锁。

## 5 风险与应对

| 风险 | 应对措施 |
|------|----------|
| `openapi-diff` 工具内置规则集与 spec §4.4 口径不一致（误报或漏报） | Phase 2.3 落 wrapper 脚本 reclassify；wrapper 单元测试覆盖 §4.4 全部规则；任何分歧以 spec §4.4 为准，工具版本必须固定 |
| privacy export 白名单被滥用扩展到其它 endpoint | Phase 2.2 在 `openapi/diff-config.yaml` 显式只对 `POST /api/v1/privacy/exports` 生效；wrapper 校验匹配 path / method / from-status / to-status 全部一致；扩展必须 ADR + 本 spec 修订 |
| baseline 文件被误编辑（直接改 baseline 让 diff 通过） | Phase 1.1 在 baseline `info.description` 显式写 `DO NOT EDIT`；建议在仓库根 `.gitattributes` 标 baseline 为 `linguist-generated=true`；后续如接入 A5 远端 CI，可加 codeowners 二次校验，但本 plan 不强制 |
| ADR 与 baseline / history 顺序错位（先改 yaml 后写 ADR / history） | Phase 2.2 wrapper 在 privacy export 白名单切换时强制 `history.md` 同 PR 增量；非白名单 breaking 必须由 wrapper 输出 reclassify 错误并提示先写 ADR；流程靠人工 review，自动化只做闸门 |
| 工具版本随时间漂移导致 `make openapi-diff` 行为变化 | `openapi/baseline/README.md` 标注最低工具版本；`scripts/lint/openapi_diff.py` 启动时打印 `openapi-diff --version`，与预期不一致时报警 |
| W2 启动时 C / D 域 plan 误以为 codegen drift 由远端 CI 拦截 | Phase 4.3 工作日志显式声明：当前 P0 阶段 codegen / fixtures / breaking-change 三道 gate 都靠本地 `make` + owner self-review；远端 CI 接入由 A5 后续触发 |

## 6 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-04-28 | 1.1 | 对齐 B2 spec v1.4：privacy export P0 例外类型从旧称 `ApiError` 修正为 wire envelope `ApiErrorResponse`。 | 001-bootstrap assessment remediation |
