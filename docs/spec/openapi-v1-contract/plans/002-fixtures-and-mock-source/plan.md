# OpenAPI v1 Contract Fixtures & Mock Source

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-04-29

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [openapi-v1-contract spec](../../spec.md) §2.1 / §4.7 / §4.6 锁定的「fixtures 同源 + provenance + 隐私脱敏」契约落到 `openapi/fixtures/` 目录：为 [001-bootstrap](../001-bootstrap/plan.md) 落地的 37 个 operationId 生成默认 fixture（`scenario: default`）+ 来源于 `easyinterview-ui/src/data.jsx` 的 `scenario: prototype-baseline`；落地 `make validate-fixtures`（schema-valid + provenance + 隐私脱敏）与 `make sync-fixtures-from-prototype`（前端原型数据折叠工具）；将 fixtures 投影为 Prism / 文档站可消费的 OpenAPI named examples，避免手写第二份 example；为 [E1 mock-contract-suite](../../../engineering-roadmap/spec.md#55-layer-e--integration4-份) 提供唯一可消费的 fixture 真理源；通过本 plan Phase 4 的本地命令证明 spec §6 中 C-6 / C-7 / C-9（partial）/ C-11（fixture 级）已成立。

本 plan 不实现 mock server 运行壳（归 E1）、不修改 `openapi/openapi.yaml` schema（归 001 / 003）、不引入 breaking-change linter（归 003）。

## 2 背景

[engineering-roadmap §4.3 mock-first 集成策略](../../../engineering-roadmap/spec.md#43-mock-first-集成策略) 把 fixtures 列为 `frontend/msw` 与 `backend mock-server` / E1 三处共享的唯一数据来源；[spec §3.1 D-9 / §4.7](../../spec.md) 把「每个 operationId 一份默认 fixture + `prototype-baseline` 同步」绑死在 v1.0.0 freeze 范围内。

执行本 plan 前必须确认 [001-bootstrap](../001-bootstrap/plan.md) Phase 7 已完成：`openapi/openapi.yaml` 中 37 个 operation 的 schema、`ApiError`、`GenerationProvenance`、`deleteMe` 与 privacy export 501 example 必须已锁定，作为 `make validate-fixtures` 的 schema 校验源。若 001 未完成，先暂停本 plan。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有 37 份 default fixtures；Phase 2 起来就有 prototype-baseline scenario；Phase 3 起来就能用 Prism / 自建 mock server 消费；Phase 4 收口 4 项 AC + handoff；Phase 5 做 v1.8 fixture remediation。本 plan 不引入 BDD 资产（`test/scenarios/` 由 E2 在 W4 spawn）。

## 3 实施步骤

### Phase 1: default fixtures + 校验工具

#### 1.1 fixture 目录骨架

按 spec §2.1 落地 `openapi/fixtures/<tag>/<operationId>.json`：14 tag 子目录，37 个 fixture 文件。每份 fixture 文件结构：

```json
{
  "operationId": "...",
  "scenarios": {
    "default": {
      "request": { "headers": {}, "body": {} },
      "response": { "status": 200, "headers": {}, "body": {} }
    }
  }
}
```

`request` 字段在无 body 的 GET / DELETE operation 中可省略；`scenarios` 是有序键，第一项必须是 `default`。

#### 1.2 37 份 default fixture 内容

按 [02-api-definition.md §4–§17](../../../../../easyinterview-tech-docs/02-api-definition.md) 与 spec §3.1.1 / §4.2 schema inventory 写入合理的 example：

- 列表 endpoint 默认返回 1–3 条记录 + `pageInfo.nextCursor: null` / `hasMore: false`。
- 长耗时 operation 返回 `202 + *WithJob`，`job.status` 为 `pending` 或 `running`，`job.kind` 与 endpoint 对应。
- AI 生成 schema（spec §4.6）必须显式包含 `provenance` 对象（6 字段全部填合理值；`rubricVersion` 在非评分场景填 `not_applicable`）。
- `POST /api/v1/privacy/exports` fixture 必须返回 `501` + `error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`（spec C-7 / D-12）；`POST /api/v1/privacy/deletions` 与 `GET /api/v1/privacy/requests/{privacyRequestId}` 保持 `202 + PrivacyRequestWithJob` / `200 + PrivacyRequest`。
- 隐私敏感字段统一使用保留占位：`Acme` / `acme.example` / `alice@example.com` / `+1-555-0100`（或 `+1-555-0100`..`+1-555-0199`）等；允许域名必须是 `example.com` / `example.org` / `example.net` 或 `.example` 保留域（spec §4.7）；不出现真实邮箱 / 电话 / 公司名。
- 时间字段统一 `2026-04-28T...Z` 风格（spec D-3）；id 字段使用 UUIDv7 字面量（spec §4.3），不允许 `tmp_` 前缀。

#### 1.3 `make validate-fixtures`

落地 `scripts/lint/validate_fixtures.py`（或 Go 实现，由 `make validate-fixtures` 调用）：

1. 对每个 `openapi/fixtures/<tag>/<operationId>.json`：
   - 必须存在且 `operationId` 与文件名一致；
   - `scenarios.default` 必填；其它 scenario 可选（`prototype-baseline` 在 Phase 2 接入）；
   - 对每个 scenario：`request.body`（若 operation 声明 requestBody）和 `response.body` schema-valid against `openapi/openapi.yaml` 中对应 operation 的 request / response schema；`response.status` 必须是该 operation 已声明状态码或 `default` 可覆盖的错误状态，并按实际状态选择 `2xx` / `4xx` / `5xx` 分支。
2. AI schema 强制 provenance：扫描固定列表（`TargetJob.summary` / `TargetJob.fitSummary` / `AssistantAction` / `FeedbackReport` / AI-created `MistakeEntry` / `ResumeTailorRun` / `Debrief`）出现的字段，必须含 `provenance` 对象且 6 字段非空（spec §4.6 / C-11）。
3. 隐私敏感字段扫描：拒绝真实邮箱模式（允许 `example.com` / `example.org` / `example.net` / `.example` 保留域）、真实电话区号（允许 `+1-555-0100`..`+1-555-0199` 保留号码）、真实公司名（黑名单可放在 `scripts/lint/fixtures_privacy_blacklist.txt`）。命中即报错。
4. ID 扫描：所有字段名以 `id` / `Id` 结尾或 schema 标记为 `format: uuid` 的值必须是 UUIDv7 字面量；任何 `tmp_` 前缀直接 fail。
5. 37 个 operation 必须全部存在 fixture；缺失 operationId 直接 fail。
6. 接入根 `Makefile` 的 `make validate-fixtures` target，`make help` 自动包含。

#### 1.4 Phase 1 自检

- `make validate-fixtures` exit 0。
- 临时把任一 AI schema fixture 的 `provenance` 字段删掉：`make validate-fixtures` fail，错误日志列出 operationId 与缺失字段。
- 临时把 privacy export fixture 改成 `202` / 写入真实邮箱 / 写入 `tmp_` id：`make validate-fixtures` fail，错误日志列出 operationId 与触发规则；revert 后恢复。

### Phase 2: prototype-baseline scenario 同步工具

#### 2.1 数据源映射表

落地 `openapi/fixtures/PROTOTYPE_MAPPING.md`：把 `easyinterview-ui/src/data.jsx` 中的核心 mock 数据节（welcome / home / parse / onboarding / jd_match / workspace / practice / report / mistakes / resume / debrief / growth / ...）映射到对应 operationId。一对多 / 多对一关系在表中显式标注。

#### 2.2 `make sync-fixtures-from-prototype`

落地 `scripts/codegen/sync_fixtures_from_prototype.{py,ts}`（B2 owner；语言可与 generator 一致）：

- 输入：`easyinterview-ui/src/data.jsx`（按 §2.1 mapping 提取节）+ `openapi/fixtures/PROTOTYPE_MAPPING.md`。
- 输出：在每个相关 fixture 文件的 `scenarios.prototype-baseline` 节写入数据；缺失数据节的 fixture 不写入该 scenario（不强制 37/37 覆盖）。
- 写入字段必须满足 schema：脚本内部跑一次 schema 校验，不通过的字段 fail-fast 并打印映射缺口（让人补 mapping，不静默兜底）。
- 接入根 `Makefile` 的 `make sync-fixtures-from-prototype` target；执行幂等（再跑一次 `git diff --exit-code` 不变）。

#### 2.3 Scenario 渲染规则

更新 `openapi/fixtures/README.md`：

- scenario 命名规则：`default` 必填；`prototype-baseline` 来自 ui 原型；其它 scenario 命名按 `<purpose>-<variant>`（如 `error-conflict`），但本 plan 不强制接入。
- consumer 选择 scenario 的契约：默认 fallback 到 `default`；mock server / msw 通过 query / header / 配置覆盖。

#### 2.4 Phase 2 自检

- `make sync-fixtures-from-prototype` 后 `git diff --exit-code` 干净（再跑一次幂等）。
- 至少 P0 闭环关键 endpoint（`getMe` / `listExperienceCards` / `listTargetJobs` / `getTargetJob` / `getPracticeSession` / `getFeedbackReport` / `listMistakes` / `getGrowthOverview`）的 `prototype-baseline` 节非空。
- `make validate-fixtures` 同时通过 `default` 与 `prototype-baseline`。

### Phase 3: Mock parity 接口预演（E1 handoff）

#### 3.1 fixtures → OpenAPI examples 投影

落地 `scripts/codegen/render_openapi_fixture_examples.py`（或与 codegen 语言一致的等价实现）+ 根 `Makefile` target：读取 `openapi/openapi.yaml` 与 `openapi/fixtures/`，把每个 operation 的 `scenarios.default.response.body` 投影为 OpenAPI named example `default`，输出到 `openapi/.generated/openapi-with-fixtures.yaml`。该文件用于 Prism smoke / 文档站预览，不作为 schema 真理源；不得人工手写 OpenAPI examples。

投影工具必须校验：37 个 operation 均存在 default example；生成的 OpenAPI example body 与对应 fixture response body 字节级一致；再次运行幂等（`git diff --exit-code -- openapi/.generated/openapi-with-fixtures.yaml` 干净，若该文件选择不入库则用临时目录前后哈希一致替代）。

#### 3.2 本地 Prism smoke

在 `openapi/README.md` 与 `openapi/fixtures/README.md` 中写入 Prism 启动方式（不写正式 mock server 入口；E1 落地正式 mock server 时收口）：

```bash
npx @stoplight/prism-cli mock openapi/.generated/openapi-with-fixtures.yaml -p 4010
```

跑一次固定 5 个 operation（`getMe` / `listTargetJobs` / `getPracticeSession` / `getFeedbackReport` / `requestPrivacyExport`）用 curl 拉 fixture（`Prefer: example=default`）；返回 body 与 `openapi/fixtures/<tag>/<operationId>.json#/scenarios/default/response/body` 字节级一致。

#### 3.3 前端 msw / 后端 mock-server 同源声明

更新 `openapi/fixtures/README.md`：明确 frontend `msw` 与 backend `mock-server` / Prism / 自建 mock 在 P0 范围必须共享 `openapi/fixtures/`；前端禁止 hardcode mock（spec §2.1 / §4.7）。该约束落到 [E1 mock-contract-suite](../../../engineering-roadmap/spec.md#55-layer-e--integration4-份) 与 [D 域 frontend-shell](../../../engineering-roadmap/spec.md#54-layer-d--frontend4-份p0) plan，本 plan 只声明真理源位置，不实现 mock server 运行壳。

#### 3.4 C-9 设置完成确认

- 当前 plan 阶段：spec C-9（mock 同源）的「fixture 唯一真理源」与「default scenario → OpenAPI example → Prism response 字节级一致」由本 plan Phase 3.1 / 3.2 / 3.3 关闭。
- 「前端 msw + 后端 mock-server 真实运行时同字节」由 E1 / D1 后续 plan 在 W2 闭合，本 plan 不替代。

### Phase 4: Verification + handoff

#### 4.1 spec C-6 / C-7 / C-9（partial）/ C-11 自检

- `make validate-fixtures` exit 0；删除任一 fixture / 临时构造 request 或 response schema 不符 / 临时去掉 provenance / 临时使用真实邮箱 / 临时使用 `tmp_` id：每种均 fail，错误日志精确指向问题（C-6 / C-11 / 隐私边界 / ID 边界）。
- examples 投影工具通过；固定 5 个 operation 生成的 OpenAPI examples 与 fixtures 字节级一致。
- `POST /api/v1/privacy/exports` fixture 通过 `openapi/.generated/openapi-with-fixtures.yaml` 实际跑 Prism 返回 `501 + error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`（C-7）。
- 固定 5 个 operation 用 Prism mock + curl 验证字节级一致（C-9 partial）。
- 至少 8 个 P0 关键 endpoint 的 `prototype-baseline` scenario 非空且 schema-valid。

#### 4.2 文档与 INDEX 同步

- 本 plan checklist 全部勾选；Phase 4 关键命令日志贴入工作日志。
- 更新 `openapi/fixtures/README.md` 与 `openapi/README.md`（fixtures 用法、scenario 规则、Prism smoke、E1 handoff 说明）。
- plans/INDEX.md 中仅本 plan 状态切到 completed；001 必须已 completed 才能执行本 plan 收口，003 保持 active 并在自身 Phase 4 关闭 B2 freeze handoff；`/sync-doc-index --check` 通过。

#### 4.3 E1 handoff

- 在工作日志中明确 [E1 mock-contract-suite](../../../engineering-roadmap/spec.md#55-layer-e--integration4-份) 在 W2 启动时直接消费 `openapi/fixtures/` 与 `openapi/openapi.yaml`，不重建 fixture 真理源。
- 本 plan 不修改 E1 spec 或 plan；E1 spawn 时由 roadmap owner 触发。

### Phase 5: v1.8 fixture remediation

#### 5.1 `Auth/deleteMe` fixture

新增 `openapi/fixtures/auth/deleteMe.json` default fixture：request 带 `Idempotency-Key`，response `202 + PrivacyRequestWithJob`，`job.jobType="privacy_delete"`，语义与 `requestPrivacyDelete` 保持一致。

#### 5.2 37 operation fixture / example coverage

更新 `make validate-fixtures`、fixtures → examples 投影工具与 README 中的 operation count 到 37；缺 `deleteMe` fixture 或 example 必须 fail。

#### 5.3 P0 debrief fixture 收口

`Debrief` / `DebriefWithJob` default fixture 不包含 P1 感谢信草稿或完整跟进建议 required 字段；如果 schema 保留这些字段，fixture 中必须体现 optional / hidden 口径，不阻塞 P0。

## 4 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-6 / C-7 / C-11 全部成立；C-9 中「fixture 同源 + default scenario → OpenAPI example → Prism response 字节级一致」部分成立，剩余「真实 msw / 后端 mock-server 同字节」由 E1 / D1 后续 plan 在 W2 闭合。
- 本 plan checklist 全部勾选；Phase 4 关键命令日志贴入工作日志。

## 5 风险与应对

| 风险 | 应对措施 |
|------|----------|
| `easyinterview-ui/src/data.jsx` 字段命名与 OpenAPI schema 不一致（如 snake_case / 旧字段） | Phase 2.2 同步脚本做 schema 校验且 fail-fast 而非静默兜底；mapping 缺口必须人工补 `PROTOTYPE_MAPPING.md`；不允许 sync 工具自动重命名 |
| 37 份 fixture 手写量大且容易漂出 schema | Phase 1.3 强制 schema 校验；建议先用 generator / Prism `--seed-fixture` 工具生成最小骨架再人工补字段；本 plan 验证 idempotency，不引入二次手写 |
| 37 份 fixture 中遗漏 `deleteMe` 或与 `requestPrivacyDelete` 语义不一致 | Phase 5.1 / 5.2 强制 `Auth/deleteMe` fixture + operation count gate；Prism examples 从 fixture 投影，避免单独手写 |

## 6 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-04-29 | 1.1 | 原地 reopen，新增 Phase 5 remediation：补齐 `Auth/deleteMe` fixture、37 operation coverage 与 P0 debrief fixture 口径。 | plan-review remediation |
| privacy export fixture 被误改成 202（被「正常成功」习惯覆盖） | Phase 1.3 校验脚本对 `POST /api/v1/privacy/exports` 单独走白名单：必须 status=501 + error.code=PRIVACY_EXPORT_NOT_AVAILABLE，否则 fail；Phase 4.1 复跑确认 |
| AI schema provenance 字段被 stub 成空字符串 | Phase 1.3 校验 6 字段非空；`rubricVersion` 在非评分场景必须显式写 `not_applicable` 而非空串；脚本拒绝空白 |
| 隐私敏感字段黑名单遗漏导致真实信息漏入 | Phase 1.3 黑名单 `scripts/lint/fixtures_privacy_blacklist.txt` 持续维护；遇到漏报由 plan 修订（递增本 plan 版本）补充 |
| Prism / msw 对 `examples` vs `example` 解读差异导致字节不一致 | Phase 3.1 先从 fixtures 投影 OpenAPI named examples，再用 Phase 3.2 Prism 命令固定 `Prefer: example=default`；`openapi/fixtures/<tag>/<operationId>.json` 是真理源，OpenAPI yaml 中 example 由工具同步（不手写两份） |
