# Local Dev Stack Loopback Backend 交付复盘报告

> **日期**: 2026-06-15
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付范围：修复本地测试环境中首次登录用户打开简历模块时出现 500 的问题；实际 owner 为 `local-dev-stack/001-bootstrap` Phase 9，修复 host-run backend redeploy 继承通配 `APP_LISTEN_ADDR=:8080` 的监听问题。
- 成功证据：
  - Red gate：`python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k redeploy_script_documents_host_run_artifact_boundary` 在实现前失败，缺少 `backend_listen_addr()`。
  - Static green gate：`python3 -m pytest scripts/lint/scenario_env_contract_test.py -q` 通过，12 tests passed。
  - Shell gate：`bash -n test/scenarios/_shared/scripts/local-dev-runtime.sh` 与 `bash -n test/scenarios/env-redeploy.sh` 通过；仅出现既有 locale warning。
  - Live redeploy gate：无关 `172.18.0.6:8080` bridge listener 仍存在时，`test/scenarios/env-redeploy.sh backend` 成功，输出 `APP_LISTEN_ADDR=127.0.0.1:8080`；backend log 记录 `addr=127.0.0.1:8080`。
  - API regression gate：新 synthetic 首次登录用户在 profile setup 前后 `GET /api/v1/resumes` 均返回 200 empty list。
  - Browser gate：`agent-browser` 打开 `/resume-versions` 显示简历空态；同一浏览器上下文 fetch `/api/v1/resumes` 返回 `status=200`、`itemCount=0`、`errorCode=null`。

## 2 会话中的主要阻点/痛点

- 症状指向简历模块，但根因在本地环境 redeploy。
  - **证据**：真实 API 首次复现 `/api/v1/resumes` 为 500；迁移状态、`resumes` 表结构、直接 SQL empty-list 查询均正常；`env-redeploy.sh backend` 失败于 `listen tcp :8080: bind: address already in use`。
  - **影响**：需要从 materials/frontend 方向切换到 local-dev-stack owner，额外做 DB 与 runtime 两条线排除。

- Phase 8 redeploy gate 没有约束 effective listen address。
  - **证据**：既有 plan 只要求 build + restart + endpoint/log/PID 可接管；没有要求通配 `APP_LISTEN_ADDR` 在 host-run 场景中收敛为 loopback。
  - **影响**：历史 gate 能在无 bridge listener 的机器上通过，但遇到 interface-level 冲突时 false-green。

- 浏览器验收需要 fallback 到 `agent-browser`。
  - **证据**：内置 in-app browser 返回 `Browser is not available: iab`；随后使用 `agent-browser` 完成页面空态和同浏览器 fetch 验证。
  - **影响**：验证路径多一步，但未影响交付结果；当前不需要仓库代码改动。

## 3 根因归类

- `spec-plan`：`local-dev-stack` spec/plan 缺少 host-run backend loopback 监听不变量，导致 redeploy contract 未覆盖 bridge listener regression。
- `README`：`deploy/dev-stack/README.md` 原先只写 `APP_LISTEN_ADDR=:8080` 和 backend API loopback base，没有说明 scenario redeploy 会规范化通配监听地址。
- `no repo change needed`：浏览器 surface fallback 是当前工具可用性问题，本次已通过现有 `agent-browser` 完成等价验收。

## 4 对流程资产的改进建议

- 在后续修改 `test/scenarios/_shared/scripts/local-dev-runtime.sh` 或 `env-redeploy.sh` 时，保持 Phase 9 的 static contract：必须同时检查 effective env、detached process、PID/log、loopback endpoint。
  - **落点**：spec-plan
  - **优先级**：high

- 若后续反复出现 host-run env/redeploy 类问题，可把 BUG-0127 的检查模式补入 `docs/bugs/PATTERNS.md`：通配监听地址、interface-level 端口冲突、不得误杀非 owner listener。
  - **落点**：Bug pattern library
  - **优先级**：medium

- `scenario-redeploy` skill 的说明可在后续专门维护时补充：host-run backend/frontend redeploy 不应假设 `:PORT` 安全，且遇到端口冲突应先判断 owner pidfile / loopback listener。
  - **落点**：skill
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：保留并执行 `local-dev-stack/001-bootstrap` Phase 9 的 loopback redeploy gate，避免未来环境脚本修改重新引入通配监听。
- 次优先级：在下一次处理 scenario env / redeploy 相关问题时，把 BUG-0127 抽象进 `PATTERNS.md` 或 `scenario-redeploy` skill；当前 session 已在 owner spec/plan/README 固化核心规则，不需要再扩大本次 diff。
