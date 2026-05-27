# Real Provider Hybrid Scenario Unification 交付复盘报告

> **日期**: 2026-05-27
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel` Phase 5 原地修订，将 `E2E.P0.100` 从框架外 `manual-uat` companion 迁入标准 `e2e` hybrid 场景。
- 代码与流程资产：`test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/`、`test/scenarios/e2e/INDEX.md`、`test/scenarios` README、`.agent-skills/scenario-run` / `scenario-env` / `scenario-redeploy`、`scripts/lint/scenario_env_contract_test.py`、`e2e-scenarios-p0` spec/plan/checklist/BDD。
- 关联 Bug：[BUG-0110](../bugs/BUG-0110.md)。
- 成功证据：
  - `bash test/scenarios/env-verify.sh`：Postgres / Redis / MinIO / Mailpit 四个 dependency 全 OK。
  - `python3 -m pytest scripts/lint/scenario_env_contract_test.py -q`：8 passed。
  - `bash -n test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/*.sh`：标准四段脚本语法通过。
  - `test ! -e test/scenarios/manual-uat && rg -n 'MANUAL_REQUIRED|hybrid|AI Agent' test/scenarios/README.md test/scenarios/e2e/README.md`：旧入口删除且 README 描述 hybrid 语义。
  - `bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/trigger.sh && bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/verify.sh && jq . .test-output/e2e/p0-100-real-provider-full-funnel-hybrid/result.json`：生成 `MANUAL_REQUIRED` result artifact。
  - `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`：均通过。

## 2 会话中的主要阻点/痛点

- 框架外 manual UAT 造成执行入口割裂。
  - **证据**：`E2E.P0.100` 资产曾位于 `test/scenarios/manual-uat/`，未出现在 `test/scenarios/e2e/INDEX.md`。
  - **影响**：AI Agent 无法通过标准场景发现、运行和交接证据，人工 runbook 与场景 runner 变成两套体验。
- Hybrid 场景缺少可表达的中间结果。
  - **证据**：`scenario-run` 原先没有把 `MANUAL_REQUIRED` 定义为合法结果，缺真实 provider env 或浏览器证据时只能在 PASS / FAIL / ERROR 之间误判。
  - **影响**：缺密钥时容易被当成失败或被人为标绿，无法清楚表达“自动 preflight 已完成，等待人工/浏览器 Agent 证据”。
- 环境 preflight 未成为运行入口的强制前置条件。
  - **证据**：用户明确指出运行用例前必须准备对应环境；旧 `scenario-run` 说明没有要求先执行 `test/scenarios/env-setup.sh` / `env-verify.sh`。
  - **影响**：场景失败时难以区分是共享依赖环境未准备、真实 provider secret 缺失，还是业务链路本身回归。

## 3 根因归类

- 真实 provider UAT 被误建模为人工 companion，而不是标准 scenario 的 hybrid 执行模式。
  - **类别**：spec-plan / README
- 场景运行 skill 缺少环境 preflight 与 `MANUAL_REQUIRED` contract。
  - **类别**：skill
- 缺少负向 contract 防止 `E2E.P0.100` 再次脱离标准 e2e INDEX 与四段脚本结构。
  - **类别**：spec-plan / lint contract

## 4 对流程资产的改进建议

- 保持 `scripts/lint/scenario_env_contract_test.py` 对 `E2E.P0.100` 的注册、脚本、旧入口删除和 `scenario-run` hybrid 语义检查，后续新增 hybrid 场景时先扩展该 contract。
  - **落点**：lint/test contract
  - **优先级**：high
- 后续新增需要人工或浏览器 Agent 介入的场景时，统一要求标准 `e2e` 目录 + `MANUAL_REQUIRED` artifact，而不是新建 companion 目录。
  - **落点**：test/scenarios README / spec-plan
  - **优先级**：high
- 若后续要把真实 provider browser flow 进一步自动化，应先为未跟踪 secret 文件、浏览器证据目录和截图脱敏规则设计独立 gate。
  - **落点**：spec-plan / README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一步优先项：用 `$scenario-run -i E2E.P0.100` 走一次 skill 入口级验证，确认真实调用路径按新 `scenario-run` 文档先执行环境 preflight，再落到 `MANUAL_REQUIRED` 或 PASS 汇总。
- 可延后项：为真实 provider browser evidence 增加更细的自动化检查，例如检查 `.test-output/e2e/p0-100-real-provider-full-funnel-hybrid/evidence.md` 的 provider/profile/model/task-run 脱敏字段。
