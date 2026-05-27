# Scenario Env Independent Lifecycle 交付复盘报告

> **日期**: 2026-05-27
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`local-dev-stack/001-bootstrap` Phase 6 原地修订，新增共享测试环境 / 本地前后端联调环境的独立 setup/status/verify/cleanup/redeploy 生命周期。
- 代码与流程资产：`test/scenarios/env-*.sh`、根 `Makefile` `scenario-env-*` target、`.agent-skills/scenario-env`、`.agent-skills/scenario-redeploy`、`test/scenarios` / `deploy/dev-stack` README、`local-dev-stack` spec/plan/checklist。
- 成功证据：
  - `python3 -m pytest scripts/lint/scenario_env_contract_test.py .agent-skills/init-docs/scripts/test_cross_reference_consistency.py -q`：12 passed。
  - `make docs-check`：Header / INDEX zero drift，docs link checks OK。
  - `make scenario-env-* ... --dry-run`：setup/status/verify/cleanup/redeploy 均输出预期顶层 env script 命令。
  - `test/scenarios/env-setup.sh && test/scenarios/env-verify.sh && test/scenarios/env-cleanup.sh`：Postgres / Redis / MinIO / Mailpit 四个 dependency 全 OK，cleanup 停容器并保留命名卷。
  - `test/scenarios/env-redeploy.sh backend && test/scenarios/env-redeploy.sh frontend`：backend `go build ./cmd/...` 与 frontend production build 通过。
  - `git diff --check`：通过。

## 2 会话中的主要阻点/痛点

- 环境管理缺少中间层入口。
  - **证据**：`test/scenarios/README.md` 只把 env scripts 当约定路径，实际没有 `env-*.sh`；根 `Makefile` 也没有 `scenario-env-*`。
  - **影响**：开发者或 Agent 只能从具体场景 wrapper 或 `make dev-*` 侧面拼流程，无法只准备环境后暂停给人工/Agent 验证。
- Skill 声明和仓库可执行事实不一致。
  - **证据**：`scenario-redeploy` 提到 `env-redeploy.sh`，但入口不存在；`scenario-env` 未覆盖 rebuild/redeploy。
  - **影响**：skill 可以给出流程建议，但没有稳定 repo-tracked command 可执行，容易回到场景耦合或历史 Kind/Helm 口径。
- 已完成 plan 仍是正确 owner。
  - **证据**：`change-intake` 匹配 `local-dev-stack/001-bootstrap` 为 high confidence completed plan；新需求属于同一 subject 的 environment lifecycle revision。
  - **影响**：必须重开原 plan，而不是另开 sibling；否则环境入口、Makefile、skill、README 的 owner 会分散。

## 3 根因归类

- 缺少 framework-owned shared environment lifecycle。
  - **类别**：spec-plan / README
- Skill 没有把“声明能力必须有 repo-tracked entrypoint”固化为检查点。
  - **类别**：skill
- 场景 runner 与环境生命周期在文档上只做原则区分，没有对应可执行 target。
  - **类别**：README / Makefile

## 4 对流程资产的改进建议

- 保持 `scripts/lint/scenario_env_contract_test.py` 作为长期回归门禁，后续新增环境能力必须先扩展该 contract。
  - **落点**：lint/test contract
  - **优先级**：high
- 后续如果要让 skill 自动启动长驻 backend/frontend 进程，先在 `local-dev-stack` plan 中新增明确进程管理协议、pid/log 目录、secret 读取边界和 cleanup 规则。
  - **落点**：spec-plan / README
  - **优先级**：medium
- `/scenario-run` 前置条件可在后续补充为“优先接受 `test/scenarios/env-verify.sh` 证据”，让运行 skill 与环境 skill 的 handoff 更直接。
  - **落点**：skill
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一步优先项：用 `/plan-code-review local-dev-stack/001-bootstrap repo --fix` 做一次 L2 复核，重点检查新增 env scripts 是否存在 shell 边界遗漏、cleanup 是否足够保守、skill 文案是否仍有历史环境假设。
- 可延后项：为长驻 backend/frontend 本地进程管理设计专门 plan；当前 host-run 边界已经满足“构建共享环境 + 刷新 artifacts + 人工/Agent 后续验证”的目标。
