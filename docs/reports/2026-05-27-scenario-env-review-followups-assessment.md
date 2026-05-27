# Scenario Env Review Follow-ups 交付复盘报告

> **日期**: 2026-05-27
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 review 指出的四个 follow-up：frontend redeploy build 未加载 dev-stack Vite env、scenario setup migration 未消费 dev-stack `DATABASE_URL` / `POSTGRES_*`、P0.100 hybrid PASS 可能被陈旧或未脱敏 evidence false-green、P0.100 README Owner plan 相对链接错误。
- 关联 Bug：[BUG-0111](../bugs/BUG-0111.md)。
- 成功证据：
  - `python3 -m pytest scripts/lint/scenario_env_contract_test.py -q`：11 passed。
  - `bash -n test/scenarios/env-redeploy.sh && bash -n test/scenarios/env-setup.sh && bash -n test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/{setup,trigger,verify}.sh`：通过。
  - `make scenario-env-setup ARGS='--with-migrations --dry-run'`：migration 分支显示 source `deploy/dev-stack/.env` 并按 `POSTGRES_*` 派生 `DATABASE_URL`。
  - `make scenario-env-redeploy TARGET=frontend ARGS='--dry-run'` / `TARGET=all`：frontend build 分支显示 source `deploy/dev-stack/.env` 并要求 `VITE_EI_API_MODE` / `VITE_EI_API_BASE_URL`。
  - `test/scenarios/env-redeploy.sh frontend`：真实 frontend build 通过；构建产物中命中 `http://127.0.0.1:8080/api/v1`。
  - 行为探针确认旧 `run_id` evidence 不会 PASS，包含 `prompt body` 的 PASS evidence 会被 `verify.sh` 拒绝。
  - `python3 scripts/lint/check_md_links.py test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid --check-fragments`、`make docs-check`、`git diff --check`：均通过。

## 2 会话中的主要阻点/痛点

- single-env-source contract 没有覆盖共享 env 脚本。
  - **证据**：review 指出 `env-redeploy.sh` 和 `env-setup.sh` 分别遗漏 frontend build-time Vite env 与 migration `DATABASE_URL` 来源。
  - **影响**：非默认端口、账号或 real-mode frontend preview 会连接错误目标，且问题只在本地真实联调时出现。
- Hybrid PASS 证据门禁过弱。
  - **证据**：旧 `trigger.sh` 只检查 provider/profile/model/task-run 四个 marker；`setup.sh` 不清理 evidence；`verify.sh` 只扫描 trigger log。
  - **影响**：标准 `setup -> trigger -> verify` 可能被旧 evidence false-green，且未脱敏 prompt/response/API key 内容不会阻止 PASS。
- Scenario README link check 没有覆盖全部场景目录。
  - **证据**：P0.100 README Owner plan 链接少一层 `../`；目标目录 link check 能发现并验证修复，但全 `test/scenarios` 扫描还暴露多个既有无关坏链。
  - **影响**：owner plan 追溯失效，后续 review/fix 容易错过真实 owner 文档。

## 3 根因归类

- BUG-0110 的 contract test 只验证“唯一 env 文件名”和 hybrid 目录结构，没有验证所有消费入口实际 source 该文件。
  - **类别**：lint contract / spec-plan
- Hybrid `PASS` 缺少 run-scoped evidence invariant 和 evidence redline invariant。
  - **类别**：README / lint contract
- `docs-check` 只覆盖 `docs/`，而 scenario README 链接需要目标目录级 gate。
  - **类别**：README / lint contract

## 4 对流程资产的改进建议

- 将 `scenario_env_contract_test.py` 保持为 P0.100 和共享 env 脚本的 regression gate，后续新增 env 消费入口时同步加入 source `.env` / real-mode / migration 断言。
  - **落点**：lint/test contract
  - **优先级**：high
- 后续 hybrid 场景统一要求 `RUN_ID` 或等价 run-scoped marker；`PASS` 前必须扫描 evidence 文件，而不是只扫描 runner log。
  - **落点**：test/scenarios README / scenario-create 模板 / lint contract
  - **优先级**：high
- 单独排期清理 `test/scenarios` 目录下既有坏相对链接，并把场景 README link check 纳入合适的 scenario contract gate。
  - **落点**：test/scenarios README / lint contract
  - **优先级**：medium

## 5 建议优先级与后续动作

- 推荐下一步：用 `/plan-code-review e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel scenario --fix` 做一次 owner plan 级 follow-up sweep，把 `RUN_ID` / evidence redline / env consumer gate 固化回 plan/checklist/BDD，而不只停在脚本和 contract test。
- 备选下一步：单独开一个 `test/scenarios` 链接清理任务，修复全目录 link check 已暴露的既有无关坏链。
