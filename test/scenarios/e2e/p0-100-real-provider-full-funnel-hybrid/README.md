# E2E.P0.100 Real Provider Full Funnel Hybrid

> **Status**: active
> **更新日期**: 2026-05-27
> **Owner plan**: [`e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel`](../../../../docs/spec/e2e-scenarios-p0/plans/002-manual-uat-real-provider-full-funnel/plan.md)
> **BDD scenario**: `E2E.P0.100`

本场景用于统一管理真实 provider 全漏斗联调：AI Agent 先执行环境与材料 preflight，人工或浏览器 Agent 再补齐真实浏览器、真实 backend/frontend、真实 AI provider 的脱敏证据。流程覆盖 Home -> 导入 JD -> Parse -> Workspace -> Practice -> Generating -> Report -> 进入下一轮。

## 1 当前执行状态

本目录已经有 owner plan，且本地邮箱入口已由 `deploy/dev-stack` 的 Mailpit 承接。因此：

- AI Agent 可自动执行：共享环境 setup/verify、场景材料结构检查、`deploy/dev-stack/.env` preflight、mock/stub 禁用检查、result artifact 写入。
- 需要真实本地上下文：真实 `AI_PROVIDER_API_KEY`、host-run backend/frontend 进程、Mailpit email-code 浏览器操作和脱敏证据记录。
- 缺少真实本地上下文时，场景结果是 `MANUAL_REQUIRED`，不是 `PASS`，也不是框架 `ERROR`。

## 2 严格边界

真实 provider UAT 必须同时满足：

- backend: `APP_ENV=dev` 的 `go run ./backend/cmd/api` 真进程。
- frontend: `VITE_EI_API_MODE=real` 且 `VITE_EI_API_BASE_URL` 指向 backend。
- storage: `make dev-up` 提供的真实 Postgres / Redis / MinIO / Mailpit。
- AI: `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 指向真实 OpenAI-compatible provider，当前默认 DeepSeek。
- raw debug: `AI_DEBUG_PRINT_RAW_OUTPUT=true` 必须来自 `deploy/dev-stack/.env`，用于本地捕获真实 provider 输出格式；raw 内容只保留在本机 backend stderr / `.test-output/` 调试日志，不写入验收报告。
- account: 使用 synthetic 邮箱 `manual-uat-full-funnel@example.test` 触发真实 email-code flow，并从 Mailpit `http://127.0.0.1:8025` 读取 6 位 code。

`test/scenarios` 目录只承接 runbook、材料、shell/Python 辅助和检查脚本；不得新增 `backend/cmd` / Go helper，也不得通过直接写 `sessions` 表绕过被测 auth flow。

以下都不是本场景的完成证据：

- `APP_ENV=test`
- `EI_E2E_P0_099_SERVER=1`
- deterministic / fixture AI client
- frontend fixture-backed mock transport
- `Prefer: example=<scenario>`
- 只运行 `E2E.P0.098` / `E2E.P0.099`

## 3 前置工具

| 工具 | 用途 |
|------|------|
| Docker + Docker Compose v2 | 启动 Postgres / Redis / MinIO / Mailpit |
| Go | 启动 backend |
| Python 3 | 仅用于可选材料检查或后续 Python 辅助脚本；当前登录不需要直接 DB helper |
| Node + pnpm | 构建/启动 frontend |
| Chrome 或同级现代浏览器 | 人工走查和读取 Mailpit email code |
| 真实 AI provider key | `AI_PROVIDER_API_KEY`，不得提交 |

## 4 环境变量

本地真实联调只使用一个 env 文件：`deploy/dev-stack/.env`。它是 `make dev-up` / `scenario-env-setup`、host-run backend、frontend real mode 和真实 AI provider 的共同配置源，不为本场景维护独立 `.env`。

首次运行时可以从 dev-stack 模板生成并只在本地填写：

```bash
cp deploy/dev-stack/.env.example deploy/dev-stack/.env
$EDITOR deploy/dev-stack/.env
```

必须填写的真实值：

- `SESSION_COOKIE_SECRET`
- `AUTH_CHALLENGE_TOKEN_PEPPER`
- `AI_PROVIDER_API_KEY`

`deploy/dev-stack/.env.example` 已给出本地邮箱默认值：`EMAIL_PROVIDER=mailpit`、`EMAIL_SMTP_HOST=127.0.0.1`、`EMAIL_SMTP_PORT=1025`、`EMAIL_VERIFY_BASE_URL=http://127.0.0.1:5173/auth/verify`。不要填写真实个人邮箱账号或外部 SMTP 凭证。当前邮件正文为 code-only，`EMAIL_VERIFY_BASE_URL` 仅用于本地 frontend origin / dev CORS 推导。若人工 UAT 使用 `vite preview --port 4174` 作为唯一前端入口，本地 `.env` 的 `EMAIL_VERIFY_BASE_URL` 应同步改为 `http://127.0.0.1:4174/auth/verify`。

`AI_PROVIDER_BASE_URL` 默认是 `https://api.deepseek.com`；如使用其他 OpenAI-compatible endpoint，必须同步确认 `config/ai-providers.yaml` / `config/ai-profiles.yaml` 支持。

`AI_DEBUG_PRINT_RAW_OUTPUT` 在本地测试与本地真实联调中默认是 `true`。如果本地 `.env` 缺失或被改成其它值，本场景必须保持 `MANUAL_REQUIRED`，不能用无法检查 raw output 的真实 provider run 冒充闭环。

## 5 AI Agent 入口

从仓库根目录运行标准场景脚本：

```bash
bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/setup.sh
bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/trigger.sh
bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/verify.sh
bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/cleanup.sh
```

Agent 阶段会把结果写入：

```text
.test-output/e2e/p0-100-real-provider-full-funnel-hybrid/result.json
```

如果 `deploy/dev-stack/.env` 缺真实 provider / auth / frontend real-mode 配置，或浏览器证据尚未准备好，`result.json` 会标记 `MANUAL_REQUIRED`。补齐 `.test-output/e2e/p0-100-real-provider-full-funnel-hybrid/evidence.md` 后可重跑 `trigger.sh` / `verify.sh`。

`setup.sh` 会清理上一轮 `evidence.md` 并在 `setup.env` 写入本轮 `RUN_ID`。人工或浏览器 Agent 补证时，`evidence.md` 必须包含同一个 `run_id`；`trigger.sh` 只会在 evidence 属于本轮且通过脱敏红线扫描后写入 `PASS`。

## 6 启动真实联调环境

### 6.1 外部依赖与 migration

```bash
bash test/scenarios/env-setup.sh --with-migrations
bash test/scenarios/env-verify.sh

export DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable'
make migrate-up
```

### 6.2 后端真实进程

```bash
set -a
. deploy/dev-stack/.env
set +a

go run ./backend/cmd/api
```

期望：

- 后端监听 `:8080`。
- 缺 `SESSION_COOKIE_SECRET` / `AUTH_CHALLENGE_TOKEN_PEPPER` 时 fail-fast。
- 缺真实 `AI_PROVIDER_API_KEY` 时 AIClient-enabled runtime fail-fast 或后续 AI 调用失败；不得自动回退到 stub。

### 6.3 Mailpit Email-Code 登录

确认 Mailpit 已随 `make dev-up` 启动：

```bash
open http://127.0.0.1:8025
```

在前端登录页或任一操作级 auth gate 中输入 `manual-uat-full-funnel@example.test` 并提交。随后在 Mailpit Web UI 打开最新邮件，把邮件中的 6 位 code 输入前端验证页。前端会调用 `verifyAuthEmailChallenge`，浏览器收到 `ei_session` 后回到目标页面，TopBar 应显示已登录用户。

不要把 raw code 写入 tracked 文件或日志。

### 6.4 前端真实模式

推荐 build + preview：

```bash
cd frontend
set -a
. ../deploy/dev-stack/.env
set +a
pnpm build
pnpm exec vite preview --host 127.0.0.1 --port 4174
```

也可在开发热更新时使用：

```bash
cd frontend
set -a
. ../deploy/dev-stack/.env
set +a
pnpm --filter @easyinterview/frontend dev
```

## 7 登录态确认

1. 打开 `http://127.0.0.1:4174`。
2. 触发登录，输入 `manual-uat-full-funnel@example.test`。
3. 打开 Mailpit `http://127.0.0.1:8025`，从最新邮件读取 6 位 code 并在前端验证页提交。
4. 前端刷新登录态，TopBar 应显示已登录用户 `manual-uat-full-funnel@example.test`。

## 8 走查流程

材料见 [`data/`](./data/)：

1. Home：粘贴 `jd-backend-engineer.<lang>.md`。
2. Parse：等待真实 `target_import` runner 与真实 AI parse 完成，确认结构化结果。
3. Workspace：确认绑定 ready resume / target job，点击立即面试。
4. Practice：使用 `answer-sample-backend-engineer.<lang>.md` 作答，推进至少一轮 follow-up。
5. Complete：完成 session，进入 Generating。
6. Report：等待真实 `report_generate` runner 和真实 AI report 完成。
7. Next round：点击进入下一轮，确认派生 practice plan / session。

## 9 真实 AI 调用证据

记录到 `.test-output/e2e/p0-100-real-provider-full-funnel-hybrid/evidence.md`，只写脱敏摘要：

- run_id（来自同一输出目录下 `setup.env` 的 `RUN_ID`）
- provider ref（例如 `deepseek`）
- model profile（例如 `target.import.default`、`practice.first_question.default`、`report.generate.default`）
- model id
- latency / token count（如可见）
- `ai_task_runs` 行数或 backend log 中的脱敏 task-run marker
- raw debug 开关状态与 raw log 路径（不复制 raw 内容）

禁止记录：

- `AI_PROVIDER_API_KEY`
- prompt 明文
- provider response 明文
- JD 原文、答案全文、报告 prose
- session cookie value

## 10 清理

默认 cleanup 走真实产品隐私删除路径，只作用于当前 UAT 邮箱对应的用户，不清空整个 dev DB。命令见 [`data/account.md`](./data/account.md#cleanup)。

如果用户要求“保留现场”，保留 dev DB 行和 `.test-output/e2e/p0-100-real-provider-full-funnel-hybrid/` 脱敏证据；退出浏览器登录态并不要导出 session cookie。Mailpit 中的邮件只属于本地 dev 环境，可通过 Mailpit UI 清空。
