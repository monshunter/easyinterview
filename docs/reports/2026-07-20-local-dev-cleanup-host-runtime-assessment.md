# Local Dev Cleanup Host Runtime 交付复盘报告

> **日期**: 2026-07-20
> **审查人**: Codex

**关联计划**: [local-dev-stack/001-bootstrap](../spec/local-dev-stack/plans/001-bootstrap/plan.md)
**关联 Bug**: [BUG-0196](../bugs/BUG-0196.md)

## 1 复盘范围与成功证据

- 交付范围：原地重开 `local-dev-stack/001-bootstrap` Phase 18，修复标准 scenario cleanup 只停止 Compose 依赖、遗漏 repo-managed host-run backend/frontend 的生命周期缺口。
- 实现结果：cleanup 默认与 `--with-volumes` 两条路径都先通过共享 ownership helper 停止 repo-owned host runtime，再执行 `dev-down` 或显式 reset；命名卷与无关进程安全边界不变。
- 成功证据：28 个 scenario environment contracts、shell syntax、dry-run 双路径、真实 setup/redeploy/cleanup、context、Header/INDEX、docs links、Spec ID、diff gate均通过；根 `make test` 最终通过 Python 624 tests / 4628 subtests、Go 全包与 frontend 137 files / 1126 tests。

## 2 会话中的主要阻点/痛点

- cleanup 的名称与当前实际 owner 范围不一致。
  - **证据**：依赖容器停止后，`10900/10901` 与两个 repo pidfile 仍存在；需要额外调用内部 `_stop_host_runtimes` 才达到用户理解的“本地测试环境已清理”。
  - **影响**：后续调试可能误连旧 frontend/backend，且标准 cleanup 给出成功退出码却留下运行态。
- 既有 contract 只冻结了 `make dev-down` / `dev-reset` 委派，没有冻结跨拓扑 clean state。
  - **证据**：历史 live gate 记录容器与 network 清除、卷保留，但没有 listener/pidfile 断言。
  - **影响**：host-run 在后续 revision 成为正式环境组成后，cleanup consumer 漂移未被检测。
- 首次根 `make test` 出现无关 frontend timing failure。
  - **证据**：`HomeRecentMocks` 在全量并发下短暂看到 loading；focused 14/14 和完整重跑 1126/1126 均通过。
  - **影响**：增加一次回归轮次，但没有证据支持把它并入本修复。

## 3 根因归类

- 环境拓扑从“Compose 依赖”演进为“Compose 依赖 + host-run app”时，plan/runbook 没有要求反查所有 lifecycle consumer。
  - **类别**：spec/plan
- `/scenario-env` cleanup workflow 没有明确 repo-owned process 与 manual/unrelated process 的安全边界。
  - **类别**：skill
- runbook 对 cleanup 只写“默认保留卷”，缺少进程与容器的完整 clean-state 定义。
  - **类别**：README
- 单次 frontend timing failure 已由当前证据排除为非本分支稳定回归。
  - **类别**：no repo change needed

## 4 对流程资产的改进建议

- 保留本次新增的 contract：每次修改 environment lifecycle，都同时检查 default/reset dry-run 顺序、repo-owned termination、unowned preservation 与 docs/Skill 投影。
  - **落点**：spec/plan
- cleanup 验收继续采用“listener + pidfile + Compose state + volumes”四类事实，避免只用 `dev-doctor DOWN` 证明清理完成。
  - **落点**：README
- 不为一次性通过重跑的无关 frontend timing failure扩展本 plan；若同一测试再次在独立任务中复现，再由 frontend owner 走问题入口。
  - **落点**：no repo change needed

## 5 建议优先级与后续动作

- high：后续 local-dev-stack lifecycle revision 必须反查 `setup/status/verify/cleanup/redeploy` 全部入口及其真实进程、容器、pidfile、listener、volume owner。
- medium：若 `HomeRecentMocks` 的 loading-to-empty 时序在未来完整回归中再次失败，单独交由 frontend-home owner 复现和修复，不与环境 cleanup 混合提交。
