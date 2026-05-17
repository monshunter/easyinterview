# 001 Debrief Screen and Handoff BDD Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-17

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## E2E.P0.065 — Debrief Default Render + 3 Picker Modal

- [ ] 065.A 创建 scenario 目录 `test/scenarios/e2e/p0-065-debrief-default-render-and-pickers/`
- [ ] 065.B 编写 fixtures：用户 A；fixtures listTargetJobs/listPracticeSessions/listResumes/listResumeVersions(get by resumeAssetId)/getTargetJob/getResumeVersion/getPracticeSession 数据
- [ ] 065.C 编写 setup.sh：登录用户 A + 加载 fixture transport
- [ ] 065.D 编写 trigger.sh：Playwright script 触发 nav + 三次 picker 交互
- [ ] 065.E 编写 verify.sh：`debrief_full` normalize 到 `debrief` + DOM 锚点存在 + control type 一致 + suggestDebriefQuestions 自动调用一次 + TopBar debrief 高亮 + 正式 route catalog 不含 `debrief_full` + privacy/legacy 反查
- [ ] 065.F 编写 cleanup.sh：清空 InterviewContext + localStorage + 登出
- [ ] 065.G 确认 `scripts/setup.sh` / `scripts/trigger.sh` / `scripts/verify.sh` / `scripts/cleanup.sh` 四段脚本均可独立执行；`trigger.sh` 保留 Playwright exit code；verify.sh 含前端 runner pass marker + 旧口径 grep
- [ ] 065.H 编写 scenario README 描述 isolation / setup / cleanup 协议；登记到 `test/scenarios/e2e/INDEX.md`
- [ ] 065.I 在场景目录内按 `setup.sh -> trigger.sh -> verify.sh -> cleanup.sh` 执行通过；记录证据
- [ ] 065.J BDD-Gate 通过：plan checklist 8.8 勾选

## E2E.P0.066 — Text Mode AI Suggestions + Entries + createDebrief Submit

- [ ] 066.A 创建 scenario 目录 `test/scenarios/e2e/p0-066-debrief-text-suggestions-and-submit/`
- [ ] 066.B 编写 fixtures：suggestDebriefQuestions=default 6 items + createDebrief=default 202 + suggestDebriefQuestions=fail variant
- [ ] 066.C 编写 setup.sh：登录 + 通过 P0.065 完成三选状态
- [ ] 066.D 编写 trigger.sh：Playwright script 触发 4 个 CTA + 重新生成失败 + 切换 voice/text + Submit
- [ ] 066.E 编写 verify.sh：entries 3 行（source ai_confirmed/ai_edited/manual）+ AI failure inline error + voice UI shell + Submit createDebrief call + IK UUID + 202 响应 + InterviewContext.debriefId/debriefJobId 设置 + 不覆盖 jobId + setStep(1) + polling 启动 + privacy 反查
- [ ] 066.F 编写 cleanup.sh
- [ ] 066.G 确认四段脚本可独立执行
- [ ] 066.H 登记到 INDEX
- [ ] 066.I 执行 scenario 通过
- [ ] 066.J BDD-Gate 通过：plan checklist 8.9 勾选

## E2E.P0.067 — Polling Happy + Analysis Render

- [ ] 067.A 创建 scenario 目录 `test/scenarios/e2e/p0-067-debrief-polling-happy-and-analysis/`
- [ ] 067.B 编写 fixtures：getJob 4 状态序列 + getDebrief=default
- [ ] 067.C 编写 setup.sh：经过 P0.066 完成 submit + step 1 启动
- [ ] 067.D 编写 trigger.sh：等待 polling 完成 + 切到 step 1 + 展开 provenance
- [ ] 067.E 编写 verify.sh：getJob 调用 4 次 + getDebrief 调用 1 次 + risk_items 渲染 3 项 + 维度卡 3 张 + provenance 6 字段 + 不渲染 P1 fields + Step 1 testid 命中
- [ ] 067.F 编写 cleanup.sh
- [ ] 067.G 确认四段脚本可独立执行
- [ ] 067.H 登记到 INDEX
- [ ] 067.I 执行 scenario 通过
- [ ] 067.J BDD-Gate 通过：plan checklist 8.10 勾选

## E2E.P0.068 — Failure States + Cross-Owner Handoff

- [ ] 068.A 创建 scenario 目录 `test/scenarios/e2e/p0-068-debrief-failure-and-handoff/`
- [ ] 068.B 编写 fixtures：4 类失败 + 1 类成功 handoff variants
- [ ] 068.C 编写 setup.sh：5 个 sub-scenarios chained
- [ ] 068.D 编写 trigger.sh：(1) MissingContext (2) Failure (3) Timeout (4) 编辑+重试 (5) handoff
- [ ] 068.E 编写 verify.sh：3 种失败 state DOM + 文案 + CTA 行为 + nav practice 时 payload 含 `practiceGoal='debrief'` 且完整 + spy `createPracticePlan/startPracticeSession` 在 debrief 模块内 0 调用
- [ ] 068.F 编写 cleanup.sh
- [ ] 068.G 确认四段脚本可独立执行
- [ ] 068.H 登记到 INDEX
- [ ] 068.I 执行 scenario 通过
- [ ] 068.J BDD-Gate 通过：plan checklist 8.11 勾选

## E2E.P0.069 — Pixel Parity + i18n + Privacy + Legacy Negative

- [ ] 069.A 创建 scenario 目录 `test/scenarios/e2e/p0-069-debrief-pixel-parity-and-legacy-negative/`
- [ ] 069.B 编写 fixtures：ui-design 静态原型 desktop+mobile 截图 + frontend live screen 截图工具
- [ ] 069.C 编写 setup.sh：完整 happy flow + marker 注入 + 准备截图对比基线
- [ ] 069.D 编写 trigger.sh：8 sub-scenarios（desktop+mobile×light+dark+customAccent + zh+en + privacy spy + legacy grep）
- [ ] 069.E 编写 verify.sh：pixel diff < 0.5% × 6 个 viewport+主题组合；i18n zh/en 文案显示；marker 0 命中；legacy terms 0 命中；createPracticePlan/startPracticeSession 在 debrief 模块内 0 命中
- [ ] 069.F 编写 cleanup.sh
- [ ] 069.G 确认四段脚本可独立执行
- [ ] 069.H 登记到 INDEX
- [ ] 069.I 执行 scenario 通过
- [ ] 069.J BDD-Gate 通过：plan checklist 8.12 勾选

## 收口

- [ ] 9.A 所有 5 个 scenario `Ready` 状态登记到 `test/scenarios/e2e/INDEX.md`
- [ ] 9.B 所有 5 个 scenario 一次性顺序执行通过：`for s in p0-065-debrief-default-render-and-pickers p0-066-debrief-text-suggestions-and-submit p0-067-debrief-polling-happy-and-analysis p0-068-debrief-failure-and-handoff p0-069-debrief-pixel-parity-and-legacy-negative; do (cd test/scenarios/e2e/$s && bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh); rc=$?; (cd test/scenarios/e2e/$s && bash scripts/cleanup.sh); [ $rc -eq 0 ] || break; done`
- [ ] 9.C 全部 scenario 证据 `*.evidence.log` 已记录
