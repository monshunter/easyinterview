# Auth Mail-Link E2E Scenario 交付复盘报告

> **日期**: 2026-05-27
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付把 BUG-0112 的 ad-hoc Playwright smoke 固化为标准 E2E 场景 `E2E.P0.101`：登录和注册分别提交 synthetic 邮箱、读取 Mailpit、打开 frontend `/auth/verify` callback、确认 `/me=200`、URL token 已清理且 console/page/http failure 为 0。
- 新增资产包括 `test/scenarios/e2e/p0-101-auth-mail-link-login-register/` 全套 README / data / setup / trigger / verify / cleanup，`frontend/tests/e2e/auth-mail-link.spec.ts`，以及 `frontend/playwright.auth-mail-link.config.ts`。
- 通过验证：
  - `test/scenarios/env-verify.sh`
  - `test/scenarios/e2e/p0-101-auth-mail-link-login-register/scripts/setup.sh && .../trigger.sh && .../verify.sh; .../cleanup.sh`
  - `pnpm --filter @easyinterview/frontend exec playwright test --config=playwright.auth-mail-link.config.ts --list auth-mail-link.spec.ts`
  - `pnpm --filter @easyinterview/frontend typecheck`
  - `python3 -m pytest scripts/lint/scenario_env_contract_test.py -q`
  - `make docs-check`
  - `git diff --check`

## 2 会话中的主要阻点/痛点

- 默认 Playwright config 误指向 pixel parity。
  - **证据**：首次触发 `playwright test tests/e2e/auth-mail-link.spec.ts` 时启动了 `serve-pixel-parity`，并返回 `No tests found`。
  - **影响**：脚本不是自包含验收入口，必须补充专用 config 后才能稳定发现目标 E2E spec。

- Shell wrapper 的失败收尾需要避免 zsh 保留变量。
  - **证据**：一次性命令使用 `status=$?` 在 zsh 下触发 `read-only variable: status`。
  - **影响**：虽然不是 repo 脚本缺陷，但说明手工组合命令应使用 `rc` 等中性变量，或直接按四段脚本逐条执行。

## 3 根因归类

- Playwright config 漂移属于 `README/skill` 类问题：frontend 已有默认 `playwright.config.ts` 服务 pixel parity，新增 E2E 场景如果不显式指定 config，就会被错误 testDir / webServer 接管。
- zsh 变量问题属于 `no repo change needed`：仓库脚本本身没有使用 `status` 变量，失败来自临时 shell glue。

## 4 对流程资产的改进建议

- 在 `/scenario-create` 或 `test/scenarios/e2e/README.md` 增加一条 Playwright 场景创建规则：当目标 package 已存在多个 Playwright config 或默认 config 不服务当前 suite 时，必须创建/指定场景专用 config，并在 `trigger.sh` / `verify.sh` 中写入 config marker。
  - **落点**：skill 或 README
  - **优先级**：medium

- 在场景运行示例中避免使用 `status` 作为 shell 临时变量，统一使用 `rc`。
  - **落点**：README 示例或无需仓库改动
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一轮最值得做的是把“Playwright 场景必须显式 config”的经验沉淀到 `/scenario-create`，防止未来新增 frontend E2E 时再次误用 pixel parity 默认 config。
- `E2E.P0.101` 当前已经是可重复自动化入口；后续若将 auth 邮件链路纳入 suite 批量执行，可直接按 `/scenario-run -i E2E.P0.101` 的四段脚本协议调用。
