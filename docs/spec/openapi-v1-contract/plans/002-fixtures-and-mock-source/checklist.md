# OpenAPI v1 Contract Fixtures & Mock Source Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-03

**关联计划**: [plan](./plan.md)

## Phase 1: default fixtures + 校验工具

- [x] 1.1 落地 `openapi/fixtures/<tag>/<operationId>.json` 初始目录骨架；Phase 6 后当前为 12 tag 子目录、34 个 fixture 文件。文件结构 `{operationId, scenarios: {default: {request?, response: {status, headers?, body}}}}`，第一项必须是 `default`
- [x] 1.2 写入当前 34 份 default fixture 内容。列表 endpoint 1–3 条 + `pageInfo.nextCursor: null`；长耗时 operation 走 `202 + *WithJob`；AI schema 含 `provenance` 6 字段（`rubricVersion` 非评分场景填 `not_applicable`）；`POST /privacy/exports` 必须 `501 + error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`；`POST /privacy/deletions` 保持 `202 + PrivacyRequestWithJob`；隐私字段只使用 `Acme` / 保留 example 域名邮箱 / `+1-555-0100`..`+1-555-0199` 占位；id 用 UUIDv7 字面量且不出现 `tmp_`
- [x] 1.3 落地 `scripts/lint/validate_fixtures.py`（或等价 Go 实现）：schema 校验对应 `openapi.yaml` operation 的 requestBody 与 2xx/4xx/5xx response 分支；强制 6 个 AI schema 含非空 provenance；隐私 allowlist + 黑名单扫描；UUIDv7 / `tmp_` id 扫描；当前 34 operation 全覆盖；接入 `make validate-fixtures`
- [x] 1.4 Phase 1 自检：`make validate-fixtures` exit 0；删除任一 AI schema 的 `provenance` / 改 privacy export 为 202 / 写入真实邮箱 / 写入 `tmp_` id → fail 且错误指向 operationId，revert 后恢复

## Phase 2: prototype-baseline scenario 同步工具

- [x] 2.1 落地 `openapi/fixtures/PROTOTYPE_MAPPING.md`：把 `ui-design/src/data.jsx` 的 mock 数据节映射到 operationId（一对多 / 多对一显式标注）
- [x] 2.2 落地 `scripts/codegen/sync_fixtures_from_prototype.{py,ts}`：按 mapping 把数据写入每个 fixture 的 `scenarios.prototype-baseline` 节；schema 不通过 fail-fast；接入 `make sync-fixtures-from-prototype`；幂等（再跑 `git diff --exit-code` 不变）
- [x] 2.3 更新 `openapi/fixtures/README.md`：scenario 命名规则（`default` 必填、`prototype-baseline` 来自 ui 原型、其它 `<purpose>-<variant>`）+ consumer 选择 scenario 的契约（默认 fallback `default`）
- [x] 2.4 Phase 2 自检：`make sync-fixtures-from-prototype` 幂等；当前 P0 闭环关键 6 个 endpoint 的 `prototype-baseline` 节非空；`make validate-fixtures` 同时通过 `default` 与 `prototype-baseline`

## Phase 3: Mock parity 接口预演（E1 handoff）

- [x] 3.1 落地 fixtures → OpenAPI named examples 投影工具：读取 `openapi/openapi.yaml` + `openapi/fixtures/`，输出 `openapi/.generated/openapi-with-fixtures.yaml` 或临时等价产物；当前 34 个 default example 全覆盖；生成 example body 与 fixture body 字节级一致；重复运行幂等
- [x] 3.2 在 `openapi/README.md` / `openapi/fixtures/README.md` 写入 Prism 启动方式（`prism mock openapi/.generated/openapi-with-fixtures.yaml -p 4010`）+ 固定 5 个 operation（`getMe` / `listTargetJobs` / `getPracticeSession` / `getFeedbackReport` / `requestPrivacyExport`）用 curl `Prefer: example=default` 验证返回 body 与 fixture 字节级一致；不落正式 mock server 入口（归 E1）
- [x] 3.3 在 `openapi/fixtures/README.md` 明确 frontend `msw` 与 backend `mock-server` / Prism 必须共享 `openapi/fixtures/`，前端禁止 hardcode mock；该约束在 E1 / D1 后续 plan 落实，本 plan 只声明真理源位置
- [x] 3.4 工作日志记录：spec C-9 中「fixture 唯一真理源」与「default scenario → OpenAPI example → Prism response 字节级一致」由本 plan 关闭；「真实 msw / 后端 mock-server 同字节」由 E1 / D1 在 W2 闭合

## Phase 4: Verification + handoff

- [x] 4.1 spec C-6 / C-7 / C-9 partial / C-11 自检：`make validate-fixtures` exit 0；删除 fixture / 临时改 request 或 response schema / 临时去 provenance / 临时使用真实邮箱 / 临时使用 `tmp_` id → 各 fail；examples 投影工具通过；Prism 跑 `POST /privacy/exports` 返回 `501 + error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`；固定 5 个 operation Prism + curl 字节级一致；当前 6 个 P0 关键 endpoint `prototype-baseline` 非空且 schema-valid
- [x] 4.2 文档与 INDEX 同步：仅本 plan 切 completed；001 必须已 completed，003 保持 active 并由自身 Phase 4 关闭 B2 freeze handoff；`openapi/fixtures/README.md` 与 `openapi/README.md` Header 完整；`/sync-doc-index --check` 通过
- [x] 4.3 E1 handoff：工作日志声明 E1 mock-contract-suite 在 W2 直接消费 `openapi/fixtures/` 与 `openapi/openapi.yaml`，不重建 fixture 真理源；本 plan 不修改 E1 spec / plan

## Phase 5: v1.8 fixture remediation

- [x] 5.1 新增 `openapi/fixtures/auth/deleteMe.json` default fixture：request 带 `Idempotency-Key`，response `202 + PrivacyRequestWithJob`，`job.jobType="privacy_delete"`，语义与 `requestPrivacyDelete` 保持一致
- [x] 5.2 更新 `make validate-fixtures`、fixtures → examples 投影工具与 README 中 operation count；Phase 5 曾提升到 v1.8 的 37 operation，Phase 6 后当前为 34 operation；缺 `deleteMe` fixture 或 example 必须 fail
- [x] 5.3 `Debrief` / `DebriefWithJob` default fixture 不包含 P1 感谢信草稿或完整跟进建议 required 字段；如果 schema 保留这些字段，fixture 中体现 optional / hidden 口径，不阻塞 P0
- [x] 5.4 复跑 `make validate-fixtures` 与 examples 投影；Phase 5 曾确认 37 operation coverage，Phase 6 后当前确认 34 operation coverage、privacy redaction、provenance 与 fixture example parity 通过

## Phase 6: product-scope v1.2 fixture remediation

- [x] 6.1 Red: 调整 `scripts/lint/validate_fixtures.py` 到 34 operation 覆盖后，运行 `make validate-fixtures` 失败，证明旧 Mistakes / Growth fixture 集合被 gate 拦住
  - 2026-05-03: `make validate-fixtures` exit 2，明确报出 `Growth/getGrowthOverview`、`Mistakes/listMistakes`、`Mistakes/retestMistake` 不在 spec §3.1.1 freeze / openapi.yaml 中；同时暴露待修复旧字段 `openMistakeCount`、`drill/review` next action、旧 practice enum 与报告题目回顾字段缺失。
- [x] 6.2 Green: 删除 Mistakes / Growth fixtures，修订 Reports / TargetJobs fixtures 与 `PROTOTYPE_MAPPING.md`，字段改为题目回顾 / 本轮复练语义
  - 2026-05-03: 删除 `openapi/fixtures/Growth/getGrowthOverview.json`、`openapi/fixtures/Mistakes/listMistakes.json`、`openapi/fixtures/Mistakes/retestMistake.json`；更新 default fixture 生成器与 prototype sync，将 target job `openQuestionIssueCount`、report `reviewStatus` / `includedInRetryPlan` / `retryFocusTurnIds`、next action `retry_current_round` / `review_evidence` 作为当前语义。
- [x] 6.3 Verify: `make validate-fixtures` 与 fixture example render 测试通过；`openapi/fixtures/` 搜索确认无独立 Mistakes / Growth operation 或旧单题 Drill 文案
  - 2026-05-03: `make validate-fixtures` exit 0；`make render-openapi-fixture-examples` exit 0；`python3 -m unittest scripts/codegen/render_openapi_fixture_examples_test.py scripts/lint/validate_fixtures_test.py scripts/codegen/sync_fixtures_from_prototype_test.py` 27 tests OK；`rg` 未命中 `Mistakes|Growth|getGrowthOverview|listMistakes|retestMistake|MistakeStatus|openMistakeCount|writtenToMistakeBook|mistakeIds|single_drill|counter_questions|warmup|fix_mistake|单题深钻|反问专练|热身|"drill"|"review"|sprint|core_interview`。
