# Backend TargetJob BDD Checklist

> **版本**: 1.12
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.010 Text JD import 走完异步解析并可列表 / 详情 / 更新

- [x] 创建场景目录 `test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/`，并在 `test/scenarios/e2e/INDEX.md` 添加 `E2E.P0.010` 行（关联需求：backend-targetjob C-1/C-3/C-6/C-7/C-12）
- [x] 历史主路径证据已存在；Phase 18 必须把 payload 改为合法 `{rawText,targetLanguage,resumeId}`，并保留 cookie、IK、stub profile、outbox 与 DB 清理边界。
- [x] 历史 setup/trigger/verify/cleanup 已覆盖 import→runner→list/get/patch 与 IK replay；Phase 18 在同一场景上替换 wire，并新增 DB 仅写 `raw_jd_text`、无来源 row/file ref/refresh job 的断言。
- [x] 历史 outbox 证据已存在；Phase 18 必须进一步断言 payload 不含 `raw_jd_text` 或来源字段。
- [x] 执行并通过场景验证，记录 `.test-output/runs/.../E2E.P0.010/result.json` 证据
  <!-- verified: 2026-05-08 method=cmd-api-http run=targetjob-http-20260508 validBddEvidence=true -->
- [x] BUG-0146 C-16 historical regression: valid real-provider pasted JD without company name parses ready, uses fallback company display, renders authenticated `/parse` without `JD 解析失败`, and records `ai_task_runs` jd_parse evidence
  <!-- verified: 2026-07-09 method=local-real-provider-browser targetJobId=019f44a1-b43e-754f-ba0b-3cd9ed11ce1f evidence=analysisStatus=ready companyName=未提供 title=AI应用技术负责人 browserRoute=route-parse aiTaskRun=success -->

## E2E.P0.012 Parse 失败 retryable / non-retryable

- [x] 创建场景目录 `test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/`，并在 `test/scenarios/e2e/INDEX.md` 添加 `E2E.P0.012` 行（关联需求：backend-targetjob C-4/C-5/C-10）
- [x] 准备测试数据：可注入失败的 stub provider（`AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID`）、F3 unsupported / disabled profile、paste-only request、cookie jar 与 `Idempotency-Key`。
- [x] 实现 setup / trigger / verify / cleanup：start auth → 注入 `AI_PROVIDER_TIMEOUT` → import → runner kernel drain → 验证 `target.analysis.failed.retryable=true`、`GET /targets/{id}` 返回 404 且 `GET /targets` 不含失败 job；切换 stub 到 `AI_OUTPUT_INVALID` → 重新 import → 验证 `retryable=false` 与同样的不可见资产语义；切换 F3 `target.import.parse` 为 disabled / unsupported → 验证 import 启动 / runner kernel 阶段 fail-closed；切换 A3 缺 secret / config invalid → 验证 `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID`
- [x] 断言 error envelope / log / metric label / audit / outbox payload 不含 prompt body、response body、provider secret、`Authorization:` 或 raw JD；失败 TargetJob 不作为可继续规划资产，用户重试必须重新 import。
- [x] 执行并通过场景验证，记录 `.test-output/runs/.../E2E.P0.012/result.json` 证据
  <!-- verified: 2026-05-08 method=cmd-api-http run=targetjob-http-20260508 validBddEvidence=true -->
- [x] Revision 2026-07-09 trigger covers parse-failure admission: failed imports emit `target.analysis.failed` but `GET /targets/{id}` is 404 and `GET /targets` excludes the failed job.
- [x] Revision 2026-07-09 verify covers failed TargetJob deletion / no dirty interview-list admission / no prompt-response or raw JD leakage.

## E2E.P0.018 Workspace 删除图标持久归档

- [x] `test/scenarios/e2e/p0-018-workspace-default-render/` trigger/verify 覆盖 generated `archiveTargetJob` frontend 调用路径、右上角删除按钮定位和事件隔离；local real-backend browser smoke 覆盖 delete refresh 语义（关联需求：backend-targetjob C-7a/C-8）
- [x] 准备测试数据：已登录用户、ready TargetJob、workspace real API mode、generated `Idempotency-Key`、archive 后 DB readback
- [x] 实现并验证 smoke：打开 workspace → 点击右上角删除图标 → DB readback `status='archived'` + `deleted_at is not null` → 刷新 workspace → 列表不含该 target
- [x] 重复归档 `TARGET_INVALID_STATE_TRANSITION` conflict、跨用户归档 `TARGET_JOB_NOT_FOUND` 由 focused backend tests 覆盖；删除不触发卡片主体导航由 focused frontend tests 覆盖
- [x] 执行并通过场景/浏览器验证，记录 `E2E.P0.018 PASS`、`.test-output/e2e/workspace-archive-real-browser/workspace-card-before-delete.png`、`.test-output/e2e/workspace-archive-real-browser/workspace-after-delete.png`、`archive-db-state.txt` 和 `workspace-after-refresh-text.txt`

## E2E.P0.098 Practice round progress persistence

- [x] Seed a multi-round TargetJob plus exact round-identified plans/sessions and execute real completion transactions.<!-- verified: 2026-07-12 method=P0.098 real-postgres -->
- [x] Verify Get/List both return first→next-existing→final projections for a `1,2,4` ladder; wrong-resume/duplicate completion contributes zero/once respectively, and a newer old-round retry plan cannot replace the current-round plan.<!-- verified: 2026-07-12 method=P0.098 markers="wrong-resume-completion-ignored,non-contiguous-round-1-2-4,get-list-first-next-final-parity" -->
- [x] Verify report queued/ready/failed and TargetJob lifecycle status variants leave identical progress.<!-- verified: 2026-07-12 method=P0.098 marker=target-report-status-independent=PASS -->
- [x] Verify final `status=completed`, `currentRound=null`, `currentPracticePlanId=null`, and static frontend scope tests contain no business-progress storage/URL fact source.<!-- verified: 2026-07-12 method=P0.098-real-postgres+frontend-scope-test; not-live-browser -->

## Phase 18 paste-only scenario revision

- [x] RED: `E2E.P0.010` / `E2E.P0.012` 当前仍接受旧 source-shaped payload 或缺少 paste-only persistence negative assertion，先记录失败证据。
- [x] GREEN: P0.010 使用 `{rawText,targetLanguage,resumeId}`，证明 queued→ready、IK replay、`raw_jd_text` 唯一事实源、无来源字段/row/file ref/refresh job。
- [x] GREEN: P0.012 使用同一 wire，证明 AI retryable/non-retryable 失败、失败资产不可见、可重新粘贴，且日志/事件不泄漏原文。
- [x] 删除 `E2E.P0.011` / `E2E.P0.013` 场景实体与 INDEX 行，并通过 suite inventory zero-reference gate。
- [x] 执行 P0.010/P0.012 setup→trigger→verify→cleanup，记录当前运行证据。
  <!-- verified: 2026-07-13 method=cmd-api-http evidence="P0.010 result verifiedAt=2026-07-13T17:23:42Z and P0.012 result verifiedAt=2026-07-13T17:23:43Z both record status=passed, validBddEvidence=true; retired P0.011/P0.013 entities and active INDEX rows are absent." -->
