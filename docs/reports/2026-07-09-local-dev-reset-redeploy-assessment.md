# Local Dev Reset Redeploy 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

**关联计划**: [local-dev-stack/001-bootstrap](../spec/local-dev-stack/plans/001-bootstrap/plan.md)

## 1 复盘范围与成功证据

- 交付范围：在 `local-dev-stack/001-bootstrap` 原计划内新增 Phase 10，提供根 `Makefile` 一键 `scenario-env-reset-redeploy` target，固定执行清数据、迁移、重编译重启 backend/frontend、最终 verify 的调试链路。
- 已同步资产：`Makefile`、`scripts/lint/scenario_env_contract_test.py`、`deploy/dev-stack/README.md`、`test/scenarios/README.md`、`test/scenarios/e2e/README.md`、local-dev-stack spec / history / context / plan / checklist / INDEX。
- 通过证据：
  - `python3 -m pytest scripts/lint/scenario_env_contract_test.py -q`，13 passed。
  - `make scenario-env-reset-redeploy ARGS=--dry-run`，按 reset -> setup/migrations -> redeploy backend/frontend -> verify 顺序预览，无实际清数据。
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`，zero drift。
  - `make docs-check`，Header / INDEX / Markdown link checks 通过。
  - `git diff --check`，通过。

## 2 会话中的主要阻点/痛点

- “普通重启”和“清数据重部署”在日常调试中容易被混用。
  - **证据**：现有 `scenario-env-redeploy TARGET=all` 只负责 build + restart host-run backend/frontend，不会清理命名卷；本次用户需要的是清数据、重跑迁移、重编译、重部署的组合入口。
  - **影响**：若只复用 redeploy target，后续调试仍可能带着旧数据库/缓存/对象存储状态复现，导致误判。
- 组合命令需要保持脚本复用和顺序可测。
  - **证据**：本次新增 lint contract 验证 Makefile target 使用 `SCENARIO_ENV_*` 变量，并用 dry-run 输出断言 reset、setup/migrations、redeploy backend/frontend、verify 的顺序。
  - **影响**：避免把清理、迁移或重启逻辑复制进 Makefile，后续 env scripts 调整时不会出现两套实现漂移。

## 3 根因归类

- 原有生命周期入口覆盖 setup/status/verify/cleanup/redeploy，但缺少一个面向“清数据后重新编译部署”的组合调试入口。
  - **类别**：spec-plan
- 文档只分别说明 cleanup 和 redeploy，未把危险的 `--with-volumes` reset 与普通 restart 明确并排区分。
  - **类别**：README

## 4 对流程资产的改进建议

- 后续新增环境组合命令时，继续先加 `scripts/lint/scenario_env_contract_test.py` 的 dry-run order contract，再接 Makefile target。
  - **落点**：spec-plan
  - **优先级**：high
- 调试前优先用 `make scenario-env-reset-redeploy ARGS=--dry-run` 预览将要执行的 reset 范围，确认无误后再去掉 dry-run。
  - **落点**：README
  - **优先级**：medium

## 5 建议优先级与后续动作

- high：下一轮需要复现全链路问题时，先执行 `make scenario-env-reset-redeploy`，再运行对应 `test/scenarios/e2e/*/scripts/trigger.sh` / `verify.sh`，保证场景从干净数据状态开始。
- medium：如果后续继续频繁清数据调试，可补充一个 `scenario-env-reset-redeploy TARGET=backend|frontend|all` 的拆分能力；当前用户请求的“一键全量调试入口”已经闭环。
