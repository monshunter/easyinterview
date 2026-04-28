# OpenAPI v1 Contract Fixtures & Mock Source Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-28

**关联计划**: [plan](./plan.md)

## Phase 1: default fixtures + 校验工具

- [x] 1.1 落地 `openapi/fixtures/<tag>/<operationId>.json` 目录骨架（14 tag 子目录、36 文件）+ 文件结构 `{operationId, scenarios: {default: {request?, response: {status, headers?, body}}}}`，第一项必须是 `default`
- [x] 1.2 写入 36 份 default fixture 内容：列表 endpoint 1–3 条 + `pageInfo.nextCursor: null`；长耗时 operation 走 `202 + *WithJob`；AI schema 含 `provenance` 6 字段（`rubricVersion` 非评分场景填 `not_applicable`）；`POST /privacy/exports` 必须 `501 + error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`；`POST /privacy/deletions` 保持 `202 + PrivacyRequestWithJob`；隐私字段只使用 `Acme` / 保留 example 域名邮箱 / `+1-555-0100`..`+1-555-0199` 占位；id 用 UUIDv7 字面量且不出现 `tmp_`
- [x] 1.3 落地 `scripts/lint/validate_fixtures.py`（或等价 Go 实现）：schema 校验对应 `openapi.yaml` operation 的 requestBody 与 2xx/4xx/5xx response 分支；强制 6 个 AI schema 含非空 provenance；隐私 allowlist + 黑名单扫描；UUIDv7 / `tmp_` id 扫描；36 operation 全覆盖；接入 `make validate-fixtures`
- [x] 1.4 Phase 1 自检：`make validate-fixtures` exit 0；删除任一 AI schema 的 `provenance` / 改 privacy export 为 202 / 写入真实邮箱 / 写入 `tmp_` id → fail 且错误指向 operationId，revert 后恢复

## Phase 2: prototype-baseline scenario 同步工具

- [x] 2.1 落地 `openapi/fixtures/PROTOTYPE_MAPPING.md`：把 `easyinterview-ui/src/data.jsx` 的 mock 数据节映射到 operationId（一对多 / 多对一显式标注）
- [x] 2.2 落地 `scripts/codegen/sync_fixtures_from_prototype.{py,ts}`：按 mapping 把数据写入每个 fixture 的 `scenarios.prototype-baseline` 节；schema 不通过 fail-fast；接入 `make sync-fixtures-from-prototype`；幂等（再跑 `git diff --exit-code` 不变）
- [x] 2.3 更新 `openapi/fixtures/README.md`：scenario 命名规则（`default` 必填、`prototype-baseline` 来自 ui 原型、其它 `<purpose>-<variant>`）+ consumer 选择 scenario 的契约（默认 fallback `default`）
- [x] 2.4 Phase 2 自检：`make sync-fixtures-from-prototype` 幂等；P0 闭环关键 8 个 endpoint 的 `prototype-baseline` 节非空；`make validate-fixtures` 同时通过 `default` 与 `prototype-baseline`

## Phase 3: Mock parity 接口预演（E1 handoff）

- [x] 3.1 落地 fixtures → OpenAPI named examples 投影工具：读取 `openapi/openapi.yaml` + `openapi/fixtures/`，输出 `openapi/.generated/openapi-with-fixtures.yaml` 或临时等价产物；36 个 default example 全覆盖；生成 example body 与 fixture body 字节级一致；重复运行幂等
- [x] 3.2 在 `openapi/README.md` / `openapi/fixtures/README.md` 写入 Prism 启动方式（`prism mock openapi/.generated/openapi-with-fixtures.yaml -p 4010`）+ 固定 5 个 operation（`getMe` / `listTargetJobs` / `getPracticeSession` / `getFeedbackReport` / `requestPrivacyExport`）用 curl `Prefer: example=default` 验证返回 body 与 fixture 字节级一致；不落正式 mock server 入口（归 E1）
- [x] 3.3 在 `openapi/fixtures/README.md` 明确 frontend `msw` 与 backend `mock-server` / Prism 必须共享 `openapi/fixtures/`，前端禁止 hardcode mock；该约束在 E1 / D1 后续 plan 落实，本 plan 只声明真理源位置
- [x] 3.4 工作日志记录：spec C-9 中「fixture 唯一真理源」与「default scenario → OpenAPI example → Prism response 字节级一致」由本 plan 关闭；「真实 msw / 后端 mock-server 同字节」由 E1 / D1 在 W2 闭合

## Phase 4: Verification + handoff

- [ ] 4.1 spec C-6 / C-7 / C-9 partial / C-11 自检：`make validate-fixtures` exit 0；删除 fixture / 临时改 request 或 response schema / 临时去 provenance / 临时使用真实邮箱 / 临时使用 `tmp_` id → 各 fail；examples 投影工具通过；Prism 跑 `POST /privacy/exports` 返回 `501 + error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`；固定 5 个 operation Prism + curl 字节级一致；至少 8 个 P0 关键 endpoint `prototype-baseline` 非空且 schema-valid
- [ ] 4.2 文档与 INDEX 同步：仅本 plan 切 completed；001 必须已 completed，003 保持 active 并由自身 Phase 4 关闭 B2 freeze handoff；`openapi/fixtures/README.md` 与 `openapi/README.md` Header 完整；`/sync-doc-index --check` 通过
- [ ] 4.3 E1 handoff：工作日志声明 E1 mock-contract-suite 在 W2 直接消费 `openapi/fixtures/` 与 `openapi/openapi.yaml`，不重建 fixture 真理源；本 plan 不修改 E1 spec / plan
