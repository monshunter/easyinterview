# 001 Debrief Record and Analysis BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-16

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## E2E.P0.060 — Debrief Create + Worker Generation Happy

- [ ] 060.A 创建 scenario 目录 `test/scenarios/e2e/p0-060-debrief-create-and-generate/`
- [ ] 060.B 编写 fixtures：`fixtures/users.json`（用户 A）、`fixtures/target_jobs.json`（target T ready）、`fixtures/createDebriefRequest.json`（合法 request with 3 questions）
- [ ] 060.C 编写 `setup.sh`：插入 user/target_job/practice_plan 等前置；初始化 IK pepper
- [ ] 060.D 编写 `trigger.sh`：调用 `POST /debriefs` + 等待 drainer 处理；保留真实 exit code（与 backend-practice/003 wrapper 规范一致）
- [ ] 060.E 编写 `verify.sh`：assert 202 + DebriefWithJob 形态；DB debriefs draft→completed；async_jobs queued→succeeded；outbox 两行 debrief.created + debrief.completed；ai_task_runs success row；含 PASS + ok + privacy + legacy grep 反查
- [ ] 060.F 编写 `cleanup.sh`：删除 user A → cascade 所有相关行
- [ ] 060.G 编写 `run.sh` wrapper：串联 setup / trigger / verify / cleanup；调用真实 backend (`cd backend && go test ./test/scenarios/p0_060 -count=1` 或 HTTP fixture flow)
- [ ] 060.H 编写 scenario README 描述 isolation / setup / cleanup 协议；登记到 `test/scenarios/e2e/INDEX.md`
- [ ] 060.I 执行 `bash test/scenarios/e2e/p0-060-debrief-create-and-generate/run.sh` 通过；记录证据（`*.evidence.log`）
- [ ] 060.J BDD-Gate 通过：plan checklist 6.6 勾选

## E2E.P0.061 — Debrief Get Draft/Completed + Cross-User Isolation

- [ ] 061.A 创建 scenario 目录 `test/scenarios/e2e/p0-061-debrief-get-and-cross-user/`
- [ ] 061.B 编写 fixtures：用户 A / 用户 B 各一份；debriefs draft + completed 两种
- [ ] 061.C 编写 setup.sh：插入两个 user + 两个 debriefs（A 拥有 draft + completed；B 拥有 0）
- [ ] 061.D 编写 trigger.sh：4 个 GET 请求（A→draft / A→completed / B→draft / A→NOT_EXIST）
- [ ] 061.E 编写 verify.sh：assert HTTP 200 placeholder / HTTP 200 full / HTTP 404 DEBRIEF_NOT_FOUND × 2；provenance 仅 6 wire 字段
- [ ] 061.F 编写 cleanup.sh：删除 user A + B + cascade
- [ ] 061.G 编写 run.sh wrapper
- [ ] 061.H 登记到 INDEX
- [ ] 061.I 执行 scenario 通过；记录证据
- [ ] 061.J BDD-Gate 通过：plan checklist 6.7 勾选

## E2E.P0.062 — Worker AI Failure Graceful + Retry + Permanent Fail

- [ ] 062.A 创建 scenario 目录 `test/scenarios/e2e/p0-062-debrief-worker-failure-and-retry/`
- [ ] 062.B 编写 fixtures：用户 A；createDebrief request；A3 mock 配置 5 次 timeout
- [ ] 062.C 编写 setup.sh：插入 user + target_job + 触发 createDebrief
- [ ] 062.D 编写 trigger.sh：循环触发 drainer 5 次 lease + handle
- [ ] 062.E 编写 verify.sh：assert attempts 1→5；前 4 次 status='queued'；第 5 次 status='failed' (permanent)；debriefs.status 始终保持 'draft'；ai_task_runs 5 行 failed；outbox debrief.completed 0 行
- [ ] 062.F 编写 cleanup.sh
- [ ] 062.G 编写 run.sh wrapper
- [ ] 062.H 登记到 INDEX
- [ ] 062.I 执行 scenario 通过
- [ ] 062.J BDD-Gate 通过：plan checklist 6.8 勾选

## E2E.P0.063 — suggestDebriefQuestions Sync + AI Failure

- [ ] 063.A 创建 scenario 目录 `test/scenarios/e2e/p0-063-suggest-debrief-questions/`
- [ ] 063.B 编写 fixtures：用户 A + target_job + 可选 session + 可选 resume_version；A3 mock 3 种响应（valid / timeout / invalid JSON）
- [ ] 063.C 编写 setup.sh：插入 user / target_job / session / resume_version
- [ ] 063.D 编写 trigger.sh：3 次 POST `/debriefs/question-suggestions`
- [ ] 063.E 编写 verify.sh：assert (1) 200 + 6 suggestions；(2) 502 AI_PROVIDER_TIMEOUT；(3) 502 AI_OUTPUT_INVALID；ai_task_runs 3 行（success / timeout / invalid）；audit 3 行；不应有 debriefs 行写入
- [ ] 063.F 编写 cleanup.sh
- [ ] 063.G 编写 run.sh wrapper
- [ ] 063.H 登记到 INDEX
- [ ] 063.I 执行 scenario 通过
- [ ] 063.J BDD-Gate 通过：plan checklist 6.9 勾选

## E2E.P0.064 — Debrief Privacy + Legacy Negative

- [ ] 064.A 创建 scenario 目录 `test/scenarios/e2e/p0-064-debrief-privacy-and-legacy-negative/`
- [ ] 064.B 编写 fixtures：用户 A 完整 createDebrief request with marker string `__SECRET_RAW_TEXT__` 注入 notes 与 questionText 字段
- [ ] 064.C 编写 setup.sh：完整执行 P0.060 happy flow + 注入 marker
- [ ] 064.D 编写 trigger.sh：触发 verify scan（不调用 API）
- [ ] 064.E 编写 verify.sh：marker 在 outbox / audit / metric 0 命中；retired 标识符 `mistakes_count` / `generatedMistakeCount` / `experience_library` / `drill_builder` / `growth_center` / `star_editor` / `debrief_voice` 在源码 / 契约 / scenario runtime 0 命中；ai_task_runs 行字段完整
- [ ] 064.F 编写 cleanup.sh
- [ ] 064.G 编写 run.sh wrapper
- [ ] 064.H 登记到 INDEX
- [ ] 064.I 执行 scenario 通过；记录证据
- [ ] 064.J BDD-Gate 通过：plan checklist 6.10 勾选

## 4 收口

- [ ] 9.A 所有 5 个 scenario `Ready` 状态登记到 `test/scenarios/e2e/INDEX.md`
- [ ] 9.B 所有 5 个 scenario 一次性顺序执行通过：`for s in p0-060-debrief-create-and-generate p0-061-debrief-get-and-cross-user p0-062-debrief-worker-failure-and-retry p0-063-suggest-debrief-questions p0-064-debrief-privacy-and-legacy-negative; do bash test/scenarios/e2e/$s/run.sh || break; done`
- [ ] 9.C 全部 scenario 证据 `*.evidence.log` 已记录
