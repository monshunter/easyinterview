# Local Dev Stack Postgres 18 Volume Guard 交付复盘报告

> **日期**: 2026-05-22
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付范围：`/plan-code-review local-dev-stack/001-bootstrap repo --fix` 的 L2 runtime remediation，修复 Postgres 18 官方镜像与 `easyinterview-pg-data` 命名卷挂载布局不匹配导致的 `make dev-up` unhealthy，并为旧卷增加只读 preflight。
- 代码与契约修复：`deploy/dev-stack/docker-compose.yaml` 将 Postgres 卷挂到 `/var/lib/postgresql`；`deploy/dev-stack/scripts/check-postgres-volume-layout.sh` 在 `make dev-up` 前检测旧根目录 `PG_VERSION`、旧 `/data/PG_VERSION` 与半初始化 `/18` 布局；local-dev-stack spec / plan / checklist / README / history / BUG-0091 已同步。
- 成功证据：compose config 通过且只显示 `target: /var/lib/postgresql`；guard 临时目录断言覆盖空卷、旧根目录、旧 `/data`、半初始化 `/18`、有效 `/18/docker`；本机旧卷下 `make dev-up` 明确失败并输出 `Existing local data was preserved`。
- 干净 runtime 证据：使用临时 Postgres 卷名冷启动，Postgres / Redis / MinIO 均 healthy；Postgres `select 1`、Redis set/get/del、MinIO bucket probe 均通过；临时容器与临时卷已清理。
- 静态与测试证据：`validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`、`sh -n` 通过；`make test` 通过。`make lint` 失败在既有 practice voice retired-route lint，与本次 local-dev-stack 修改无关。

## 2 会话中的主要阻点/痛点

- Postgres 18 镜像升级后的 volume layout 变化没有被 plan gate 捕获。
  - **证据**：compose 原本仍挂 `/var/lib/postgresql/data`，官方镜像直接报旧数据库布局错误；修正挂载点后，既有旧卷又暴露根目录 `PG_VERSION` 与半初始化 `/18` 的权限失败。
  - **影响**：先前 registry blocker 解除后仍无法通过 `make dev-up`，需要二次 runtime 诊断。
- 旧卷处理边界没有写成可执行 preflight。
  - **证据**：没有 guard 时用户只看到 `container easyinterview-postgres-dev is unhealthy`；根因只能通过容器日志和卷内容反查。
  - **影响**：容易误导执行者直接删卷或把问题归因到 Docker 健康检查，而不是明确要求用户确认本地数据后 reset。
- L2 review 的静态 gate 不足以证明 dev stack 可启动。
  - **证据**：`docker compose config --quiet` 可以通过，但无法发现 Postgres 18 entrypoint 对 PGDATA 和旧卷的 runtime 保护。
  - **影响**：completed checklist 如果只引用结构 gate，容易 false-green。

## 3 根因归类

- Postgres 18 运行契约未进 owner spec / plan gate。
  - **类别**：spec-plan
- `make dev-up` 缺少持久化依赖旧状态 preflight。
  - **类别**：README / spec-plan
- plan-code-review 对 dev infra 的审查需要把镜像 entrypoint、volume 内容和干净冷启动作为 artifact-level truth。
  - **类别**：skill

## 4 对流程资产的改进建议

- 在 `plan-code-review` 的 dev-infra 审查路径中补一句：涉及 Docker 镜像 major 升级或命名卷时，必须读取当前镜像 entrypoint / env / default UID / volume layout，并至少执行一次干净卷 runtime gate。
  - **落点**：skill
  - **优先级**：high
- 在 local-dev-stack 后续修订 gate 中保留“干净卷通过 + 旧卷安全阻断”双证据，而不是只要求 `make dev-up` 绿灯。
  - **落点**：spec-plan
  - **优先级**：high
- 在 `deploy/dev-stack/README.md` 故障排查中继续保留“确认数据后再 reset”的表达，避免未来维护者把旧卷处理改成自动清空。
  - **落点**：README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一步最高价值：更新 `.agent-skills/plan-code-review/SKILL.md`，把 Docker image major upgrade / named volume / entrypoint contract 纳入 L2 review 检查清单。
- 可随下一次 local-dev-stack 迭代处理：把临时卷冷启动验证封装成 repo-tracked smoke target，减少手写 compose override 的验证成本。
- 可延后：若后续频繁遇到 dependency volume migration，考虑把 `dev-reset` 拆出更细粒度的 per-service reset，但当前不要自动删除用户卷。
