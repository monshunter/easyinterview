# 001 Debrief Record and Analysis BDD Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-06-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## E2E.P0.060 — Debrief Create + Worker Generation Happy

- [x] 060.A 创建 scenario 目录 `test/scenarios/e2e/p0-060-debrief-create-worker-happy/`
- [x] 060.B 编写 fixtures：以 `data/seed-input.md` 和 unit/integration fixture-backed tests 固化用户 A、target T、合法 request
- [x] 060.C 编写 `setup.sh`：初始化 scenario output 与 setup marker
- [x] 060.D 编写 `trigger.sh`：执行 createDebrief + worker happy path Go tests；保留真实 exit code
- [x] 060.E 编写 `verify.sh`：assert runner marker、目标测试名、PASS、legacy lint
- [x] 060.F 编写 `cleanup.sh`：幂等删除 setup marker
- [x] 060.G 确认 `scripts/setup.sh` / `scripts/trigger.sh` / `scripts/verify.sh` / `scripts/cleanup.sh` 四段脚本均可独立执行
- [x] 060.H 编写 scenario README 描述 isolation / setup / cleanup 协议；登记到 `test/scenarios/e2e/INDEX.md`
- [x] 060.I 在场景目录内按 `setup.sh -> trigger.sh -> verify.sh -> cleanup.sh` 执行通过；证据位于 `.test-output/e2e/p0-060-debrief-create-worker-happy/trigger.log`
- [x] 060.J BDD-Gate 通过：plan checklist 6.6 勾选

## E2E.P0.061 — Debrief Get Draft/Completed + Cross-User Isolation

- [x] 061.A 创建 scenario 目录 `test/scenarios/e2e/p0-061-debrief-get-isolation/`
- [x] 061.B 编写 fixtures：以 `data/seed-input.md` 和 store/handler tests 固化 draft + completed + cross-user/not-found
- [x] 061.C 编写 setup.sh：初始化 scenario output 与 setup marker
- [x] 061.D 编写 trigger.sh：执行 GetDebrief store/service/handler tests
- [x] 061.E 编写 verify.sh：assert draft/completed/cross-user/not-found tests 与 PASS marker
- [x] 061.F 编写 cleanup.sh：幂等删除 setup marker
- [x] 061.G 确认四段脚本可独立执行
- [x] 061.H 登记到 INDEX
- [x] 061.I 执行 scenario 通过；证据位于 `.test-output/e2e/p0-061-debrief-get-isolation/trigger.log`
- [x] 061.J BDD-Gate 通过：plan checklist 6.7 勾选

## E2E.P0.062 — Worker AI Failure Graceful + Retry + Permanent Fail

- [x] 062.A 创建 scenario 目录 `test/scenarios/e2e/p0-062-debrief-worker-retry-failure/`
- [x] 062.B 编写 fixtures：以 `data/seed-input.md` 和 handler/finalizer tests 固化 F3/A3/parse failure 与 max-attempt case
- [x] 062.C 编写 setup.sh：初始化 scenario output 与 setup marker
- [x] 062.D 编写 trigger.sh：执行 worker failure 与 retry policy Go tests
- [x] 062.E 编写 verify.sh：assert failure/retry/permanent tests 与 PASS marker
- [x] 062.F 编写 cleanup.sh
- [x] 062.G 确认四段脚本可独立执行
- [x] 062.H 登记到 INDEX
- [x] 062.I 执行 scenario 通过
- [x] 062.J BDD-Gate 通过：plan checklist 6.8 勾选

## E2E.P0.063 — suggestDebriefQuestions Sync + AI Failure

- [x] 063.A 创建 scenario 目录 `test/scenarios/e2e/p0-063-debrief-suggest-questions/`
- [x] 063.B 编写 fixtures：以 `openapi/fixtures/Debriefs/suggestDebriefQuestions.json`、`data/seed-input.md` 和 store/service/API/cmd-api tests 固化 `sessionId` + `resumeId` request、target_job + completed practice session derived summary + resume structured_profile + valid / timeout / invalid JSON
- [x] 063.C 编写 setup.sh：初始化 scenario output 与 setup marker
- [x] 063.D 编写 trigger.sh：执行 suggestDebriefQuestions store/service/API/cmd-api tests（含 sessionId/resumeId context tests）与 `make validate-fixtures`
- [x] 063.E 编写 verify.sh：assert store/service/API/cmd-api success/failure/count-boundary tests、fixture `sessionId` + `resumeId` marker、`resumeVersionId` 负向 gate 与 PASS marker
- [x] 063.F 编写 cleanup.sh
- [x] 063.G 确认四段脚本可独立执行
- [x] 063.H 登记到 INDEX
- [x] 063.I 执行 scenario 通过
- [x] 063.J BDD-Gate 通过：plan checklist 6.9 勾选

## E2E.P0.064 — Debrief Privacy + Legacy Negative

- [x] 064.A 创建 scenario 目录 `test/scenarios/e2e/p0-064-debrief-privacy-legacy/`
- [x] 064.B 编写 fixtures：`data/seed-input.md` 固化 marker string `__SECRET_RAW_TEXT__`
- [x] 064.C 编写 setup.sh：初始化 scenario output 与 setup marker
- [x] 064.D 编写 trigger.sh：触发 privacy / task-run / legacy lint / fixture validation gates
- [x] 064.E 编写 verify.sh：marker 在 outbox / audit 0 命中；retired 标识符通过 `backend_debrief_legacy.py` 反查；ai_task_runs 行字段完整
- [x] 064.F 编写 cleanup.sh
- [x] 064.G 确认四段脚本可独立执行
- [x] 064.H 登记到 INDEX
- [x] 064.I 执行 scenario 通过；证据位于 `.test-output/e2e/p0-064-debrief-privacy-legacy/trigger.log`
- [x] 064.J BDD-Gate 通过：plan checklist 6.10 勾选

## 4 收口

- [x] 9.A 所有 5 个 scenario `Ready` 状态登记到 `test/scenarios/e2e/INDEX.md`
- [x] 9.B 所有 5 个 scenario 一次性顺序执行通过：`for s in p0-060-debrief-create-worker-happy p0-061-debrief-get-isolation p0-062-debrief-worker-retry-failure p0-063-debrief-suggest-questions p0-064-debrief-privacy-legacy; do test/scenarios/e2e/$s/scripts/setup.sh && test/scenarios/e2e/$s/scripts/trigger.sh && test/scenarios/e2e/$s/scripts/verify.sh && test/scenarios/e2e/$s/scripts/cleanup.sh; done`
- [x] 9.C 全部 scenario 证据 `.test-output/e2e/*/trigger.log` 已记录
