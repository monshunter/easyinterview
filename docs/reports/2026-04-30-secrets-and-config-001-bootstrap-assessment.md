# secrets-and-config/001-bootstrap 交付复盘报告

> **日期**: 2026-04-30
> **审查人**: Claude

## 1 复盘范围与成功证据

- **范围**：secrets-and-config A4 第一份实施 plan `001-bootstrap`，覆盖三层
  config loader、`SecretSource` / `FeatureFlagClient` 抽象、`config/*.yaml`
  与 `.env.example` 字典、`os.Getenv` 边界 lint、pre-commit secret 拦截、
  `runtime-config` builder + stub handler + 前端 fetcher，关联
  [plan v1.1](../spec/secrets-and-config/plans/001-bootstrap/plan.md)
  与 [checklist v1.1](../spec/secrets-and-config/plans/001-bootstrap/checklist.md)。
  32/32 checklist 项全部勾选；plan/checklist Header 已切到 `completed`，
  plans/INDEX 同步。
- **成功证据**：
  - `cd backend && go test ./internal/platform/... -count=1`：14 个 config
    + 11 个 featureflag + 3 个 secrets 测试全部通过（含三路径 RedactedString
    断言、koanf 四层合并、file 热加载与 invalid YAML 保留快照、PostHog
    mock /decide 命中 + 5xx last-known-good + 无缓存 degraded、`POSTHOG_SELF_HOSTED=false`
    在 staging/prod 启动 fail-fast、async.queueWeights 缺失 fail-fast）。
  - `cd frontend && npx vitest run src/lib/runtime-config/`：4/4 通过
    （fetch + 缓存命中 + HTTP 错误重试 + forceRefresh）。
  - `make lint-config`：`lint-getenv-boundary` / `lint-env-dict` /
    `lint-secrets-pattern` 三层全绿；env_dict 三方求差集 24/24 keys 对齐。
  - C-1：`APP_LISTEN_ADDR=:9090 go run ./backend/cmd/api -dump-config` 输出
    `app.listenAddr=:9090`、`log.level=debug`，证明 D-1 四层合并。
  - C-2：`APP_ENV=prod ./bin/api`（无 secret）退出码 1 且 stderr 列出 5 项
    `missing required secret: ...` 行。
  - C-7..C-9 / C-12：构造 `backend/internal/auth/violator.go` 越界 / 删除
    `.env.example` 中 `AI_GATEWAY_BASE_URL` / 临时生成 `AKIA*` / `sk-*` /
    `xox*` 三正则假数据，三类故意失败 case 均被 lint / hook 拦截，验证
    后立即 revert，未污染主分支。
  - C-6 partial：`curl http://localhost:8090/api/v1/runtime-config` 返回
    5 项 `public: true` flag、`ai_fallback_model_enabled` 已过滤、
    `analyticsEnabled=false`、无 `postHogPublicKey`、无任何 secret；完整
    OpenAPI verification 待 [B2](../spec/openapi-v1-contract/spec.md) 与
    [D1 frontend-shell](../spec/engineering-roadmap/spec.md#54-layer-d--frontend7-份p04--p12--p21)
    后续 plan 接入。
  - `python3 .claude/skills/sync-doc-index/scripts/sync-doc-index.py --check`：
    `All documents are in sync. Zero drift detected.`

## 2 会话中的主要阻点/痛点

- **A4 lint 红线遇到 pre-existing 越界点 `os.Getenv("APP_ENV")` in
  `internal/ai/aiclient/providers/stub/stub.go:62`**
  - **证据**：首次跑 `go run scripts/lint/getenv_boundary.go -root backend`
    输出 1 处 `internal/ai` 违规，4 处 `cmd/codegen` `flag.*` 违规。
  - **影响**：本来只是落地新 lint 规则，结果带出 stub provider 的清理：
    必须移除 `os.Getenv("APP_ENV")` fallback、改用 `WithAppEnv` 显式注入、
    更新 5 处 stub_test.go 调用点；同时还要决定是否把 `flag.*` 一并禁掉。
    最终选择只覆盖 `os.Getenv` 家族（与 spec C-7 实际验收一致），把
    `flag.*` 留给 CLI 二进制自然使用，避免误伤 codegen 工具链。
- **`yaml.Unmarshal(":::not yaml:::", &fileSchema)` 不报错导致首版
  FileFlagProvider 在收到非法 YAML 时把 last-known-good 快照清空**
  - **证据**：`TestFileProviderInvalidYAMLKeepsLastSnapshot` 首跑断言失败
    （`practice_hint_enabled` 被擦掉）。
  - **影响**：定位 + 修复 `loadOnce` 增加 `flags:` 顶层 key 探测，并补
    `yaml.Unmarshal(raw, &probe map[string]any)` 前置探针；约 10 分钟。
- **prod 部分在 `validate` 检查 prod 必填字段时顺带要求 async.queueWeights
  全字段为正，原 fixture 没写齐导致 `TestValidateProdAllSecretsPasses`
  误报失败**
  - **证据**：第一次跑 validator_test.go 失败信息：
    `async.queueWeights must declare positive critical/default/low values`。
  - **影响**：修正测试 fixture 而不是放宽 validator；保持 spec C-12 必填
    语义不漂移。
- **`backend/cmd/api` 与 `backend/cmd/worker` 之前不存在，但 lint
  allowlist 与 spec §4.1 已经将它们写为唯一允许 `os.Getenv` 的 entry
  point**
  - **证据**：spec §4.1 / plan §3 phase 4.1 文本均假定 `cmd/{api,worker}`
    存在；但仓库当前只有 `cmd/codegen`。
  - **影响**：必须在 phase 5 / phase 6 之间补 `cmd/api/main.go` 与
    `cmd/worker/main.go` 才能跑 C-1 dump-config 与 C-2 fail-fast 的活
    smoke。skeleton 二进制总计 ~150 行 Go，但需要把所有 EnvBindings /
    SecretBindings 与 spec §3.1.1 24 项一一对应回填，容易漏字段。

## 3 根因归类

- **stub provider 越界 `os.Getenv` 是历史遗留代码**：A3 ai-gateway 在
  `001-bootstrap` 时就用 `os.Getenv("APP_ENV")` 简化 stub 工厂，但当时
  A4 的 `internal/platform/config` 还不存在，所以没有自然的注入路径。
  - **类别**：spec-plan（A3 / A4 spec 之前没显式声明 stub 必须通过
    `WithAppEnv` 注入；A4 落地后立刻产生跨 spec 调整）。
- **`yaml.Unmarshal` 对错误顶层结构不报错** 是 `gopkg.in/yaml.v3`
  的预期行为，不是仓库 bug。
  - **类别**：no repo change needed（实现细节，已用顶层 key 探针解决；
    不需要文档化）。
- **validator 测试 fixture 与新增 C-12 检查没有对齐** 是会话内一次性
  执行偏差，不是流程缺陷。
  - **类别**：no repo change needed。
- **cmd/api 与 cmd/worker 缺位** 反映 plan 的 phase 安排假设了 entry
  point 已存在，但 W2 之前并未真正落地。
  - **类别**：spec-plan（建议 plan §3 在 phase 5 显式列出「补 cmd/api +
    cmd/worker skeleton」步骤，避免下游 plan 再次踩到同样 gap）。

## 4 对流程资产的改进建议

- **建议在 spec-centric plan 模板里加入「pre-existing violator sweep」
  小节**，提示在落地新 lint 红线前先扫描既有越界点，给出修复方案与
  scope 影响（本次会话临时决策只 ban `os.Getenv` 家族、不 ban `flag.*`，
  应在 plan 内显式记录依据）。
  - **落点**：spec-plan（[secrets-and-config plan §3 phase 4.1](../spec/secrets-and-config/plans/001-bootstrap/plan.md#41-golangci-lint-自定义规则拒绝-osgetenv-越界)）
    或 `/plan-review` 自检项。
  - **优先级**：medium。
- **`backend/cmd/api` + `backend/cmd/worker` 的 EnvBindings /
  SecretBindings 可以 codegen 自 spec §3.1.1 表**，避免新增 env key 时
  仍需人工同步 main.go map。
  - **落点**：spec-plan（A4 future plan）+ skill（潜在 codegen 任务由
    [B1 shared-conventions-codified](../spec/shared-conventions-codified/spec.md)
    consumer 共享）。
  - **优先级**：medium（W2 内可用手工同步，不阻塞）。
- **`/tdd` skill 在 `--phase-commit` 与 user-overridden「不要切换分支」
  之间应有显式策略**：本次按 user 指令跳过了 phase-boundary
  `/work-journal --auto` + base-branch merge，但 skill workflow 本身要
  求 phase-commit 必须执行。建议在 `/tdd` SKILL.md 增加一节描述「user
  禁止 branch checkout 时如何降级 phase-commit（仅做 work-journal
  本地提交而不 merge）」。
  - **落点**：skill（[/tdd SKILL.md](../../.agent-skills/tdd/SKILL.md)）。
  - **优先级**：medium。
- **B2 / D1 跨 plan handoff token 应固化在 plan §6.3** —— 当前 plan
  写了「在工作日志记录 token」，但实际工作日志由 user 决定何时提交。
  建议 plan §6.3 增加「在本 plan handoff 段落写入跨 plan 引用 anchor
  (e.g., `<a id="a4-runtime-config-handoff">`)」，B2 / D1 后续 plan 直接
  cross-link，无需依赖工作日志条目。
  - **落点**：spec-plan（plan §6.3 / spec §7）。
  - **优先级**：low（不阻塞本次交付，下游 plan 可在引用时回头补 anchor）。

## 5 建议优先级与后续动作

- **High（none）**：本次交付未发现 high 优先级阻断项。
- **Medium**：
  1. plan §3 phase 4.1 增补 pre-existing violator sweep 步骤。
  2. cmd/api / cmd/worker EnvBindings 增加 codegen 路径（A4 future 或 B1
     shared-conventions consumer）。
  3. `/tdd` SKILL.md 增加「user 禁 branch checkout 时的 phase-commit 降级
     策略」。
- **Low**：plan §6.3 / §7 在 handoff 段落埋入 cross-link anchor，方便 B2 /
  D1 后续 plan 直接引用。
- **可延后**：当前未触发 `/bug-report`。stub provider 的 os.Getenv 行为
  是设计取舍，不属于 bug，无需新建 BUG 记录；如未来发现因 stub provider
  在 prod 误启用导致事故，再回头评估。
