# Full Funnel Real Provider Manual UAT Checklist

> 配合 [`README.md`](./README.md) 执行；填写副本建议放在 `.test-output/manual-uat/full-funnel/checklist-YYYYMMDD.md`。
> Owner: `e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel`
> Scenario: `E2E.P0.100`

走查日期：________　走查人：________　UI 语言：☐ 中文 ☐ English

## 0 环境与账号

| # | 检查项 | 期望结果 | 结果 |
|---|--------|----------|------|
| 0.1 | `make dev-up` + `make dev-doctor` | summary down==0 && degraded==0；Postgres / Redis / MinIO / Mailpit 均 OK | ☐ Pass ☐ Fail |
| 0.2 | `make migrate-up` | 退出 0，schema 最新 | ☐ Pass ☐ Fail |
| 0.3 | backend `APP_ENV=dev go run ./backend/cmd/api` | 监听 :8080；无 auth/AI secret 缺失错误 | ☐ Pass ☐ Fail |
| 0.4 | frontend `VITE_EI_API_MODE=real` | 请求命中 `http://127.0.0.1:8080/api/v1`，无 fixture `Prefer` header | ☐ Pass ☐ Fail |
| 0.5 | Mailpit magic-link | `manual-uat-full-funnel@example.test` 收到本地 Mailpit 邮件，点击或复制 token 验证 | ☐ Pass ☐ Fail |
| 0.6 | magic-link 验证后刷新 | TopBar 已登录态，显示 UAT 账号；未直接写 sessions 表 | ☐ Pass ☐ Fail |

## 1 Home -> Parse

| # | 检查项 | 期望结果 | 结果 |
|---|--------|----------|------|
| 1.1 | Home 渲染 | 五入口 TopBar 与 Home 主入口正常 | ☐ Pass ☐ Fail |
| 1.2 | 粘贴 `materials/jd-backend-engineer.<lang>.md` | `importTargetJob` 真实请求成功 | ☐ Pass ☐ Fail |
| 1.3 | Parse loading | 真实 runner / polling 推进，不是 mock 即时返回 | ☐ Pass ☐ Fail |
| 1.4 | Parse ready | 结构化 JD 字段可见，符合 `materials/expected-observations.md` 的角色/资历/技能观察点 | ☐ Pass ☐ Fail |

## 2 Workspace -> Practice

| # | 检查项 | 期望结果 | 结果 |
|---|--------|----------|------|
| 2.1 | Confirm 进入 Workspace | route/query 携带 target / resume / plan 上下文 | ☐ Pass ☐ Fail |
| 2.2 | 点击立即面试 | 真实 `createPracticePlan` + `startPracticeSession` 成功 | ☐ Pass ☐ Fail |
| 2.3 | 首题呈现 | 首题由真实 AI provider 生成，问题与 JD/简历相关 | ☐ Pass ☐ Fail |
| 2.4 | 提交作答样例 | 使用 `answer-sample-backend-engineer.<lang>.md` 后，follow-up 正常出现 | ☐ Pass ☐ Fail |

## 3 Generating -> Report -> Next Round

| # | 检查项 | 期望结果 | 结果 |
|---|--------|----------|------|
| 3.1 | 完成面试 | `completePracticeSession` 返回 report/job handoff | ☐ Pass ☐ Fail |
| 3.2 | Generating 轮询 | 真实 `report_generate` runner 推进到 ready | ☐ Pass ☐ Fail |
| 3.3 | Report 呈现 | ReportDashboard 与 detail tabs 正常，内容符合 `materials/expected-observations.md` 的 report 观察点 | ☐ Pass ☐ Fail |
| 3.4 | 进入下一轮 | `next_round` 派生 plan/session 与首轮不同，符合 next-round 观察点 | ☐ Pass ☐ Fail |

## 4 真实 AI 与隐私证据

| # | 检查项 | 期望结果 | 结果 |
|---|--------|----------|------|
| 4.1 | provider evidence | evidence 只记录 provider/profile/model/latency/task-run count 摘要 | ☐ Pass ☐ Fail |
| 4.2 | no stub/mock | 无 `APP_ENV=test`、stub provider、fixture transport、`Prefer: example=` 完成证据 | ☐ Pass ☐ Fail |
| 4.3 | URL/storage/console 隐私 | 不出现 JD 原文、答案全文、报告 prose、session cookie value | ☐ Pass ☐ Fail |
| 4.4 | legacy-negative | 无 welcome/growth/mistakes/drill/followup/experiences/star/onboarding/独立 voice/`mode=debrief` 物化 | ☐ Pass ☐ Fail |

## 5 判定

- ☐ **通过**：0-4 全部 Pass，且证据路径已记录。
- ☐ **阻断**：记录失败项编号、截图/日志路径和推断 owner，按 `/change-intake` 路由。

证据路径：

```text
.test-output/manual-uat/full-funnel/
```
