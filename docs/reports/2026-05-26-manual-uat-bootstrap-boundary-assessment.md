# Manual UAT Mailpit Boundary 交付复盘报告

> **日期**: 2026-05-26
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel` 中账号/session bootstrap 越界进入 `backend/cmd` 的问题，并补齐本地 Mailpit，使验收账号不依赖真实外部邮箱服务或真实邮箱账号。
- 已完成：删除未提交的 `backend/cmd/devsession` / `backend/internal/devsession` 和 scenario direct-session Python helper；`local-dev-stack/001-bootstrap` 原地追加 Mailpit；backend `EMAIL_PROVIDER=mailpit` 走 SMTP `DeliveryWriter`；A4 env/config、dev-stack README、scenario README、manual UAT runbook/materials、e2e spec/plan/checklist/BDD 已同步到 synthetic 邮箱 + Mailpit magic-link。
- 成功证据：focused auth/cmd-api/config Go tests、`go build ./backend/cmd/api`、compose config、`dev-doctor.sh` shell check、`make -C deploy/dev-stack -n up`、live `make dev-up && make dev-doctor`、no-backend-cmd/no-Go/no-bootstrap-account negative gates、`make lint-config`、两个 `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`。
- 限制：`E2E.P0.100` 的真实 provider 人工 UAT 仍未执行，本报告只复盘本地账号入口与 Mailpit 边界修复，不代表 002 plan 完成。

## 2 会话中的主要阻点/痛点

- 场景验收依赖被误提升为正式 backend cmd。
  - **证据**：初始实现新增 `backend/cmd/devsession` / `backend/internal/devsession`；用户指出这会把场景测试依赖变成正式进程。
  - **影响**：删除 Go helper，并把账号入口收回真实 passwordless flow；`test/scenarios` 只保留 runbook、材料、shell/Python 编排和 negative gates。
- local mailbox 能力缺口没有先回 owner。
  - **证据**：初始核对时默认 dev-stack 只有 Postgres / Redis / MinIO；C1 `DevMailSink` 仅为进程内测试 sink，无法支撑人工 UI 收信登录。
  - **影响**：最终需要修订 `local-dev-stack/001-bootstrap`、`backend-auth`、A4 config 与 manual UAT 材料，而不是继续维护 direct session helper。
- 场景目录 shell/Python-only 规则不能等同于允许 direct session bootstrap。
  - **证据**：用户进一步明确 `test/scenarios` 不允许 Go/cmd 实施，场景通常只允许 shell + Python；同时验收账号仍必须属于被测真实 auth flow。
  - **影响**：negative gate 同时检查 `test/scenarios` 无 Go 文件、无 `backend/cmd/devsession`、无 `bootstrap_account.py`。

## 3 根因归类

- 场景工具边界没有写入当前 README / spec gate。
  - **类别**：README + spec-plan
- TDD 策略把账号 helper 描述成 Go CLI，导致实现自然落到 `backend/cmd`。
  - **类别**：spec-plan
- 当前本地邮箱能力没有在 manual UAT runbook 中先声明缺口，也没有要求先由 `local-dev-stack` / `backend-auth` 补齐 Mailpit。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 保留并执行 `test/scenarios` shell/Python-only 场景工具规则，并加入 no-Go/no-backend-cmd negative gate。
  - **落点**：`test/scenarios/README.md`
  - **优先级**：high
- 对需要账号/session 的 manual UAT plan，默认要求走真实 auth flow；本地邮箱由 `local-dev-stack` Mailpit 或明确 owner dependency 提供。
  - **落点**：spec-plan checklist 模板或 plan-review checklist
  - **优先级**：high

## 5 建议优先级与后续动作

- 下一步优先：继续执行 `e2e-scenarios-p0/002` 的真实 provider manual UAT 4.3，用 Mailpit magic-link 登录并记录脱敏证据。
- 后续可选：把 no-backend-cmd/no-Go/no-direct-session-bootstrap negative gate 抽成 `test/scenarios` 通用 lint，避免后续 manual UAT 或 BDD 材料再次绕过真实 auth flow。
