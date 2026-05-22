# 001 Debrief Screen and Handoff BDD Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-05-23

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## 2026-05-17 L2 Close-out Evidence

- P0.065-P0.069 were re-run sequentially with `setup.sh -> trigger.sh -> verify.sh -> cleanup.sh`; all five scenarios exited 0.
- P0.065-P0.068 are Vitest-backed scenario runners with `verify.sh` marker checks and `scripts/lint/frontend_debrief_legacy.py` negative gates.
- P0.069 additionally runs `pnpm --filter @easyinterview/frontend build` and `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/debrief.spec.ts`; Playwright result was 11 passed / 1 skipped.
- Scenario evidence is written under `.test-output/e2e/<scenario>/trigger.log` and consumed by each scenario `verify.sh`.

## E2E.P0.065 — Debrief Default Render + 3 Picker Modal

- [x] 065.A 创建 scenario 目录 `test/scenarios/e2e/p0-065-debrief-default-render-and-pickers/`
- [x] 065.B 编写 fixtures：用户 A；fixtures listTargetJobs/listPracticeSessions/listResumes/listResumeVersions(get by resumeAssetId)/getTargetJob/getResumeVersion/getPracticeSession 数据
- [x] 065.C 编写 setup.sh：登录用户 A + 加载 fixture transport
- [x] 065.D 编写 trigger.sh：Vitest-backed runner 覆盖 DebriefScreen / Header / ContextStrip / Stepper / route normalize 测试
- [x] 065.E 编写 verify.sh：`debrief_full` normalize 到 `debrief` + DOM 锚点存在 + control type 一致 + TopBar debrief 高亮 + 正式 route catalog 不含 `debrief_full` + legacy 反查
- [x] 065.F 编写 cleanup.sh：清空 InterviewContext + localStorage + 登出
- [x] 065.G 确认 `scripts/setup.sh` / `scripts/trigger.sh` / `scripts/verify.sh` / `scripts/cleanup.sh` 四段脚本均可独立执行；`trigger.sh` 保留 runner exit code；verify.sh 含前端 runner pass marker + 旧口径 grep
- [x] 065.H 编写 scenario README 描述 isolation / setup / cleanup 协议；登记到 `test/scenarios/e2e/INDEX.md`
- [x] 065.I 在场景目录内按 `setup.sh -> trigger.sh -> verify.sh -> cleanup.sh` 执行通过；记录证据
- [x] 065.J BDD-Gate 通过：plan checklist 8.8 勾选

## E2E.P0.066 — Text Mode AI Suggestions + Entries + createDebrief Submit

- [x] 066.A 创建 scenario 目录 `test/scenarios/e2e/p0-066-debrief-text-suggestions-and-submit/`
- [x] 066.B 编写 fixtures：suggestDebriefQuestions=default 6 items + createDebrief=default 202 + suggestDebriefQuestions=fail variant
- [x] 066.C 编写 setup.sh：登录 + 通过 P0.065 完成三选状态
- [x] 066.D 编写 trigger.sh：Vitest-backed runner 覆盖 debrief 模块、InterviewContext reducer、pendingAction 与 privacy boundary
- [x] 066.E 编写 verify.sh：entries 3 行（source ai_confirmed/ai_edited/manual）且 `myAnswerSummary` 非空 + AI failure/empty 时 manual CTA 可用 + voice UI shell + Submit createDebrief call + IK UUID + 202 响应 + InterviewContext.debriefId/debriefJobId 设置 + 不覆盖 jobId + setStep(1) + polling 启动 + privacy 反查
- [x] 066.F 编写 cleanup.sh
- [x] 066.G 确认四段脚本可独立执行
- [x] 066.H 登记到 INDEX
- [x] 066.I 执行 scenario 通过
- [x] 066.J BDD-Gate 通过：plan checklist 8.9 勾选

## E2E.P0.067 — Polling Happy + Analysis Render

- [x] 067.A 创建 scenario 目录 `test/scenarios/e2e/p0-067-debrief-polling-happy-and-analysis/`
- [x] 067.B 编写 fixtures：getJob 4 状态序列 + getDebrief=default
- [x] 067.C 编写 setup.sh：经过 P0.066 完成 submit + step 1 启动
- [x] 067.D 编写 trigger.sh：等待 polling 完成 + 切到 step 1 + 展开 provenance
- [x] 067.E 编写 verify.sh：getJob 调用 4 次 + getDebrief 调用 1 次 + risk_items 渲染 3 项 + 维度卡 3 张 + provenance 6 字段 + 不渲染 P1 fields + Step 1 testid 命中
- [x] 067.E2 BUG-0070 runtime route gate：真实 `GET /api/v1/jobs/{jobId}` route、Jobs handler/store owner scope 与 focused Go tests 通过；证据: `go test ./internal/jobs ./internal/api/jobs ./internal/store/jobs ./cmd/api -count=1`
- [x] 067.F 编写 cleanup.sh
- [x] 067.G 确认四段脚本可独立执行
- [x] 067.H 登记到 INDEX
- [x] 067.I 执行 scenario 通过
- [x] 067.J BDD-Gate 通过：plan checklist 8.10 勾选

## E2E.P0.068 — Failure States + Cross-Owner Handoff

- [x] 068.A 创建 scenario 目录 `test/scenarios/e2e/p0-068-debrief-failure-and-handoff/`
- [x] 068.B 编写 fixtures：4 类失败 + 1 类成功 handoff variants
- [x] 068.C 编写 setup.sh：5 个 sub-scenarios chained
- [x] 068.D 编写 trigger.sh：(1) MissingContext (2) Failure (3) Timeout (4) 编辑+重试 (5) handoff
- [x] 068.E 编写 verify.sh：3 种失败 state DOM + 文案 + CTA 行为 + Step 2 调用 `createPracticePlan(goal='debrief', sourceDebriefId)` + `startPracticeSession` 创建 fresh session + nav practice 时 payload 含 `practiceGoal='debrief'` / `planId` / fresh `sessionId` 且完整
- [x] 068.F 编写 cleanup.sh
- [x] 068.G 确认四段脚本可独立执行
- [x] 068.H 登记到 INDEX
- [x] 068.I 执行 scenario 通过
- [x] 068.J BDD-Gate 通过：plan checklist 8.11 勾选

## E2E.P0.069 — Pixel Parity + i18n + Privacy + Legacy Negative

- [x] 069.A 创建 scenario 目录 `test/scenarios/e2e/p0-069-debrief-pixel-parity-and-legacy-negative/`
- [x] 069.B 编写 fixtures / tooling：generated API fixture mocks + frontend live screen screenshot smoke 工具
- [x] 069.C 编写 setup.sh：清理 scenario output 并准备 P0.069 trigger log
- [x] 069.D 编写 trigger.sh：Vitest i18n/privacy/devMock gates + frontend build + Playwright desktop/mobile debrief parity spec + legacy grep
- [x] 069.E 编写 verify.sh：assert Vitest marker、Playwright marker、debrief spec 文件名、passed counts、legacy lint OK、scenario-tree legacy grep clean，并拒绝 no tests / failed
- [x] 069.F 编写 cleanup.sh
- [x] 069.G 确认四段脚本可独立执行
- [x] 069.H 登记到 INDEX
- [x] 069.I 执行 scenario 通过
- [x] 069.J BDD-Gate 通过：plan checklist 8.12 勾选

## 收口

- [x] 9.A 所有 5 个 scenario `Ready` 状态登记到 `test/scenarios/e2e/INDEX.md`
- [x] 9.B 所有 5 个 scenario 一次性顺序执行通过：`for s in p0-065-debrief-default-render-and-pickers p0-066-debrief-text-suggestions-and-submit p0-067-debrief-polling-happy-and-analysis p0-068-debrief-failure-and-handoff p0-069-debrief-pixel-parity-and-legacy-negative; do (cd test/scenarios/e2e/$s && bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh); rc=$?; (cd test/scenarios/e2e/$s && bash scripts/cleanup.sh); [ $rc -eq 0 ] || break; done`
- [x] 9.C 全部 scenario 证据 `.test-output/e2e/<scenario>/trigger.log` 已记录并由对应 `verify.sh` 消费
- [x] 9.D 2026-05-23 real-backend gate：P0.065-P0.069 trigger scripts now run `frontendOwners.realApiMode.test.ts` before fixture-backed debrief UI subcases, and verify scripts reject missing real-mode marker / default backend base URL / test-file marker; focused real-mode Vitest PASS.
