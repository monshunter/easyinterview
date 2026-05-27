# 002 Real Provider Hybrid Full Funnel

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-05-27

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 P0 完整漏斗从 `001-full-funnel-happy-journey` 的自动化 stub-AI 证明，推进到由场景框架统一管理的真实 provider hybrid 联调用例：

- 保留原 owner plan，不创建 sibling，把 `E2E.P0.100` 统一登记到 `test/scenarios/e2e/INDEX.md`。
- 删除独立 `test/scenarios/manual-uat` companion 入口，避免场景发现、执行、证据和 cleanup 工作流割裂。
- 让 AI Agent 成为第一执行者：先执行共享环境 preflight、四段脚本、材料/配置/隐私检查和统一 result artifact。
- 人工或浏览器 Agent 作为第二执行者，在同一场景输出目录补齐真实凭证、真实浏览器、真实 AI provider 的脱敏证据。
- 提供可执行的真实本地联调启动路径：Docker Compose 只提供 Postgres / Redis / MinIO / Mailpit 外部依赖，backend / frontend 使用宿主机真实进程，并统一从 `deploy/dev-stack/.env` 读取 `APP_ENV=dev`、frontend real mode、auth secrets 和真实 AI provider env。
- 提供完整 hybrid 场景材料：Mailpit 本地 magic-link 登录、JD、简历、作答样例、验收 checklist、`deploy/dev-stack/.env` 配置说明、证据归档路径与清理说明。
- 明确禁止用 `APP_ENV=test`、deterministic stub AI、fixture-backed frontend mock transport、`Prefer: example=<scenario>` 或 P0.099 test server 冒充真实 AI 联调。
- 将真实 AI LLM 连接作为验收边界：当前开发主力 provider ref 为 `deepseek` / `judge-deepseek`，真实调用通过 `AI_PROVIDER_BASE_URL` + `AI_PROVIDER_API_KEY` 指向 OpenAI-compatible provider；无真实 key 时本计划不得标记完成。

## 2 背景

`001-full-funnel-happy-journey` 已完成 `E2E.P0.098` / `E2E.P0.099`，证明 P0 happy 主干在真后端、真 PostgreSQL、真实 runner、但 deterministic stub AI 下端到端贯通。该交付对自动化回归是正确的，但它不覆盖人工验收前必须具备的材料完整性，也不证明真实 LLM provider 能在本地联调中被 backend runtime 使用。

本计划 v1.3 曾把 `E2E.P0.100` 落在 `test/scenarios/manual-uat/` companion 目录，并在 `e2e/INDEX.md` 之外维护 runbook。用户反馈指出这导致执行体验割裂：执行者首先应是 AI Agent，然后才是人工；所有用例都应支持 Agent 自动执行与人工接手，且运行前必须准备并验证对应环境。

当前代码事实：

- `docs/development.md` §2 规定本地集成为 `make dev-up` 外部依赖 + host-run backend/frontend + repo-tracked runner，不默认 Kind / K8s / Helm。
- `backend/cmd/api` 在 `APP_ENV=dev` 下会通过 A3 `bootstrap.NewClient` 构建真实 AIClient；`APP_ENV=test` 才允许 stub。
- `cmd/api` 启动 auth runtime 需要 `SESSION_COOKIE_SECRET` 与 `AUTH_CHALLENGE_TOKEN_PEPPER`；真实本地联调必须用 `deploy/dev-stack/.env` 承接 auth、AI、frontend real-mode 和共享依赖配置，不得再让单个场景维护独立 env。
- passwordless local dev 登录已由 `local-dev-stack/001` Mailpit revision 承接：`make dev-up` 启动 Mailpit，`EMAIL_PROVIDER=mailpit` 的真实 `go run ./backend/cmd/api` 通过 SMTP writer 投递 magic-link。

因此本计划必须明确：真实 UAT 不能复用 001 的 test server；账号入口必须走真实 passwordless flow + Mailpit 本地邮箱，不得通过 direct DB session bootstrap、新增 `backend/cmd` / Go helper 或真实外部邮箱账号完成。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior + tooling + docs + contract`。本计划覆盖用户可感知的 hybrid UAT 工作流、Mailpit 本地邮箱账号入口、真实 provider runtime 验证、Agent-first 场景脚本与人工验收材料。
- **TDD 策略**: 本计划 v1.5 涉及场景脚本、lint contract 与 skill 指令逻辑，必须通过 `/tdd` 红绿重构执行。Red gate 为 `python3 -m pytest scripts/lint/scenario_env_contract_test.py -q`，覆盖 `E2E.P0.100` e2e 注册、旧 `manual-uat` 入口删除、`scenario-run` 环境 preflight、`MANUAL_REQUIRED` 语义，以及禁止场景专属 env、统一读取 `deploy/dev-stack/.env` 的配置契约。
- **BDD 策略**: Feature plan requires BDD。本计划新增人工可接手的端到端业务流，BDD 场景 `E2E.P0.100` 记录在 [bdd-plan.md](./bdd-plan.md)，主 checklist 使用 `BDD-Gate:` 引用。该场景为 `hybrid`：AI Agent 自动验证环境/材料/配置/隐私和 result artifact，人工或浏览器 Agent 记录真实 UI/AI 结果。
- **替代验证 gate**:
  - Mailpit / SMTP writer / config focused tests（由 local-dev-stack/backend-auth/A4 owner 承接）。
  - `test ! -d backend/cmd/devsession && test ! -d backend/internal/devsession`，确认不把 hybrid UAT 依赖放进正式 backend cmd / internal 包。
  - `go build ./backend/cmd/api`，确认被测真实 backend 入口未被破坏。
  - `make docs-check`、`sync-doc-index --check`、`validate_context.py`。
  - `rg` 负向 gate：active 场景文档不得要求 `APP_ENV=test`、`EI_E2E_P0_099_SERVER`、fixture-backed mock transport、`Prefer: example=` 或 deterministic stub AI 作为真实 UAT 完成条件。
  - secret redline：tracked materials 中不得出现真实 `AI_PROVIDER_API_KEY`、真实 session cookie、真实个人邮箱、真实手机号或可还原 token。

### 3.1 Operation Matrix

本计划不新增公开 OpenAPI operation。真实 UAT 消费 001 已锁定的完整漏斗 operation，并使用既有 auth operation + local-dev-stack Mailpit 作为本地账号入口。

| # | operation / tool | fixture | frontend consumer | backend handler / tool | persistence | AI dependency | scenario coverage |
|---|------------------|---------|-------------------|------------------------|-------------|---------------|-------------------|
| 0 | `startAuthEmailChallenge` + Mailpit magic link + `verifyAuthEmailChallenge` | 不使用 fixture；synthetic `.example.test` email | AuthLogin/AuthVerify 或操作级 auth gate | real backend-auth handler + `email_dispatch` handler + Mailpit SMTP writer | `auth_challenges` + `async_jobs` + `users` + `sessions` | none | `E2E.P0.100` preflight |
| 1 | `registerResume` | 不使用 fixture；人工粘贴材料或由真实 UI 创建 | ResumeVersions / Home handoff | real resume handler + `resume_parse` runner | `resume_assets` + `async_jobs` | `resume.parse.default` via real provider | `E2E.P0.100` |
| 2 | `importTargetJob` | 不使用 fixture；人工粘贴 JD | HomeScreen | real targetjob handler + `target_import` runner | `target_jobs` + `jobs` | `target.import.default` via real provider | `E2E.P0.100` |
| 3 | `getTargetJob` | N/A | ParseScreen polling | real handler | `target_jobs` | none | `E2E.P0.100` |
| 4 | `createPracticePlan` | N/A | Workspace / Report CTA | real practice handler | `practice_plans` | none | `E2E.P0.100` |
| 5 | `startPracticeSession` | N/A | Workspace CTA | real practice handler | `practice_sessions` + `session_events` | `practice.first_question.default` via real provider | `E2E.P0.100` |
| 6 | `appendSessionEvent` | N/A | PracticeScreen | real practice handler | `session_events` | `practice.followup.default` / `practice.turn_observe.default` via real provider | `E2E.P0.100` |
| 7 | `completePracticeSession` | N/A | PracticeScreen | real practice handler + report runner | `feedback_reports` + `jobs` + outbox | `report.generate.default` via real provider | `E2E.P0.100` |
| 8 | `getFeedbackReport` / `getJob` | N/A | Generating / ReportDashboard | real reports/jobs handlers | `feedback_reports` / `jobs` | none | `E2E.P0.100` |

### 3.2 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-27 | 1.5 | 修正 hybrid UAT env 边界：`deploy/dev-stack/.env` 是真实本地联调唯一 env 来源，删除 `p0-100` 场景专属 `dev-real.env` 模板，脚本只读取 dev-stack env。 |
| 2026-05-27 | 1.4 | 将 `E2E.P0.100` 从独立 `manual-uat` companion 迁移为标准 `e2e` hybrid 场景；新增 AI Agent first-run preflight、`MANUAL_REQUIRED` 结果语义和统一输出目录。 |

## 4 实施步骤

### Phase 0: 设计归位与现状清理

#### 0.1 owner 计划落地

创建本 002 plan / checklist / BDD / context，并修订 `e2e-scenarios-p0` spec / history / plans INDEX，明确真实 provider hybrid UAT 是 001 之后的新验收层，不修改 001 的 stub-AI 边界。

#### 0.2 未提交 manual-uat 目录归属

审查现有 `test/scenarios/manual-uat/` 文件，把它们改为本计划的产物；如果内容仍描述 Tier A stub AI 为可交付真实联调，必须重写。

### Phase 1: Mailpit 账号入口与边界收口

#### 1.1 owner 能力确认

确认并引用 owner gate：

- `make dev-up` 启动 Mailpit，并且 `make dev-doctor` 报 `mailpit-dev` OK。
- `EMAIL_PROVIDER=mailpit` 的 `cmd/api` 使用 SMTP writer 投递 magic link。
- A4 env/config 字典包含 Mailpit SMTP keys。

#### 1.2 runbook 登录路径

更新 `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/README.md` 与 `data/account.md`：人工输入 `manual-uat-full-funnel@example.test`，在 Mailpit `http://127.0.0.1:8025` 打开 magic-link 邮件并完成 `verifyAuthEmailChallenge`；不得保存 magic token 或 cookie value。

禁止把本登录辅助放入 `backend/cmd`、`backend/internal` 或任何正式 backend runtime package；禁止直接写 `sessions` 表绕过 auth flow。

#### 1.3 清理说明

cleanup 默认走真实产品隐私删除路径或本地登出；默认不清空整个 dev DB，不直接删除全量 `users`。

### Phase 2: 真实联调环境 runbook

#### 2.1 单一 env 来源

补齐 `deploy/dev-stack/.env.example` / runbook 说明，覆盖：

- `APP_ENV=dev`
- `DATABASE_URL`
- `SESSION_COOKIE_SECRET`
- `AUTH_CHALLENGE_TOKEN_PEPPER`
- `AI_PROVIDER_REGISTRY_PATH`
- `AI_MODEL_PROFILE_PATH`
- `AI_PROVIDER_BASE_URL=https://api.deepseek.com`
- `AI_PROVIDER_API_KEY=<真实 key，不提交>`
- frontend `VITE_EI_API_MODE=real`
- frontend `VITE_EI_API_BASE_URL=http://127.0.0.1:8080/api/v1`

真实本地联调只允许读取 `deploy/dev-stack/.env`。场景目录不得维护 `dev-real.env`、`env-template/dev-real.env.example` 或其它独立 env 副本。

#### 2.2 启动步骤

runbook 必须按顺序给出：

1. `make dev-up` / `make dev-doctor`。
2. `DATABASE_URL=... make migrate-up`。
3. 导出真实 env，启动 `go run ./backend/cmd/api`。
4. 在前端提交 synthetic 邮箱，并从 Mailpit 打开 magic link。
5. 启动 frontend real mode。
6. magic-link 验证后刷新前端进入已登录态。

#### 2.3 禁用 mock 路径说明

runbook 必须显式说明以下路径不是本计划完成证据：

- `APP_ENV=test`
- `EI_E2E_P0_099_SERVER=1`
- deterministic parse/report/practice stub AI
- frontend fixture-backed mock transport
- `Prefer: example=<scenario>`
- 只跑 `E2E.P0.098` / `E2E.P0.099`

### Phase 3: 人工验收材料包

#### 3.1 输入材料

补齐双语 JD / 简历 / 作答样例 / 期望观察点。材料必须是合成数据，不含真实 PII；它们是验收输入，不是 mock transport 或 fixture response。

#### 3.2 账号材料

新增账号材料说明，包含 UAT 邮箱、Mailpit URL、magic-link 验证步骤、过期/重跑规则和清理说明。禁止把 magic token 或真实 cookie value 写入 tracked 文件。

#### 3.3 验收 checklist

更新 `e2e/p0-100-real-provider-full-funnel-hybrid/checklist.md`，覆盖：

- 环境 / secret / provider 就绪。
- backend 真进程 + frontend real mode。
- 账号登录态。
- Home -> Parse -> Workspace -> Practice -> Generating -> Report -> next_round。
- AI 真实调用证据：至少检查 backend log / `ai_task_runs` / report/provider metadata 中的 provider/profile/model 摘要，不要求在 tracked 文件中保存 provider response。
- 隐私与 legacy-negative spot-check。

### Phase 4: Gate 与收口

#### 4.1 文档与结构 gate

运行 `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`，并对 `manual-uat` 做必需文件结构检查。

#### 4.2 代码 gate

运行 focused Go tests、`go build ./backend/cmd/api`、`make lint-config` 与 no-backend-cmd/no-Go scenario negative gate。

#### 4.3 人工执行证据

本计划最终完成需要一次真实 provider hybrid UAT 证据，至少包含：

- 使用真实 `AI_PROVIDER_API_KEY` 启动 backend 的命令摘要（不记录 key）。
- backend / frontend 真实进程健康证据。
- Mailpit magic-link 登录证据摘要（不记录 token 或 cookie value）。
- checklist 勾选结果与截图/日志路径。
- AI provider 调用摘要：provider ref / model profile / model id / latency / task-run count，且不包含 prompt/response 明文。

### Phase 5: 统一场景框架与 Agent-first 执行

#### 5.1 标准场景目录迁移

将 `E2E.P0.100` 的 runbook、checklist 和材料迁移到 `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/`，补齐 `data/seed-input.md`、`data/expected-outcome.md` 和四段脚本。删除 `test/scenarios/manual-uat/` 独立入口。

#### 5.2 INDEX 与框架文档

在 `test/scenarios/e2e/INDEX.md` 登记 `E2E.P0.100` 为 `hybrid | Ready`，更新 `test/scenarios/README.md` 与 `test/scenarios/e2e/README.md`：运行前统一执行环境 preflight；hybrid 场景由 AI Agent 先运行，缺人工证据时输出 `MANUAL_REQUIRED`。

#### 5.3 `scenario-run` 技能契约

更新 `.agent-skills/scenario-run/SKILL.md`，让场景运行先执行 `test/scenarios/env-setup.sh` / `env-verify.sh`，并把 `MANUAL_REQUIRED` 作为独立结果状态汇总，而不是误判为 PASS 或 ERROR。

#### 5.4 契约测试与零残留 gate

新增/更新 `scripts/lint/scenario_env_contract_test.py`，覆盖 `E2E.P0.100` 标准场景资产、旧 `test/scenarios/manual-uat` 入口删除、`scenario-run` 环境 preflight 与 hybrid 结果语义。

### Phase 6: 单一真实 env 来源修正

#### 6.1 删除场景专属 env

删除 `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/env-template/dev-real.env.example`，并清理 README / seed-input / scripts 中的 `dev-real.env` 口径。场景配置只引用 `deploy/dev-stack/.env`。

#### 6.2 dev-stack env 承接真实联调配置

更新 `deploy/dev-stack/.env.example`，让它包含 host-run backend、auth runtime、real AI provider 和 frontend real mode 所需字段。`trigger.sh` 必须只 source `deploy/dev-stack/.env` 并检查 `APP_ENV=dev`、`EMAIL_PROVIDER=mailpit`、`VITE_EI_API_MODE=real`、auth secrets 与真实 AI key；缺失时输出 `MANUAL_REQUIRED`。

## 5 验收标准

- 本计划文档集完整，`test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/` 是 `E2E.P0.100` 唯一 active 场景入口。
- `test/scenarios/manual-uat/` 不再作为独立 companion 目录存在；active 场景框架文档不再引导用户去框架外执行真实 provider UAT。
- `scenario-run` 运行 `E2E.P0.100` 时能完成 AI Agent preflight，并在缺人工/浏览器证据时生成 `MANUAL_REQUIRED` result artifact。
- 用户可按 runbook 启动真实前端 + 真实后端 + 真实 AI provider 联调环境，不依赖 frontend mock transport 或 backend test stub。
- `deploy/dev-stack/.env` 是真实本地联调唯一 env 来源；`p0-100` 不再维护独立 `dev-real.env` 或 env template。
- UAT 材料包包含 Mailpit 本地登录说明、JD、简历、作答样例、检查清单、证据归档与清理说明。
- hybrid UAT 不新增正式 backend cmd，不直接写 session 表，且不依赖真实外部邮箱服务或真实邮箱账号。
- BDD-Gate `E2E.P0.100` 的材料结构、Agent preflight、人工执行证据和 cleanup 边界在同一场景目录下闭环。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 真实 AI key 不在仓库内，自动验证不能证明 live provider 成功 | runbook 只记录脱敏命令和 provider/task-run 摘要；真实 key 由人工本地注入，缺 key 时不得标记完成 |
| dev-only 登录路径被误用为生产账号入口 | Mailpit 只在 `deploy/dev-stack` local dependency 中启用；runbook 只允许 synthetic `.example.test` 邮箱；tracked 文件禁止保存 token/cookie，negative gate 禁止新增 `backend/cmd` helper 或直接 session bootstrap |
| 手工材料被误当成 fixture/mock 数据 | 文档明确材料是用户输入样本，不是 response fixture；frontend 必须 `VITE_EI_API_MODE=real` |
| 真实 LLM 输出不稳定 | checklist 只要求结构、provider 调用和用户可观察质量记录，不把具体文案写成硬基线 |
| 直接 DB 清理误删本地数据 | cleanup 仅针对 UAT 邮箱和关联资源；全量 `dev-reset` 必须显式人工确认 |
