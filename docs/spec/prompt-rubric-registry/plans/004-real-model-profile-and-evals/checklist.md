# F3 Real Model Profile and Evals Checklist

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-24

**关联计划**: [plan](./plan.md)

## Phase 0: 当前契约预读与现状快照

- [x] 0.1 前后端契约预读（development §2 + config/backend README + A3/F1 spec），记录 operation matrix N/A 与替代 gate
- [x] 0.2 Judge 与 profile 现状快照（judge.default unsupported / NotImplementedJudge / types.go 单 Score / catalog_test unsupported / AIClient Complete chat-only / bootstrap judge_compatible unsupported / judge-placeholder provider / 13 chat active+deepseek / ai_profile_coverage 现状）
- [x] 0.3 eval 维度映射锁定：从 `config/rubrics/*/v0.1.0.yaml` 提取真实维度名，映射 spec §3.2 质量指标（含异常高分率/离群评分→`score_outlier`），禁止新造同义维度
- [x] 0.4 Promptfoo single-source 与执行模式不变量锁定（经 registry 解析、不复制 prompt；fixture 默认 / EVAL_LIVE opt-in / live 不进 make test）

## Phase 1: Judge 接口演进为逐维度 []Score

- [x] 1.1 重写 `types_test.go` `TestJudgeSignature` 期望 `[]Score` 返回（先红）— 验证：`go test ./backend/internal/ai/registry -run TestJudgeSignature`
- [x] 1.2 `types.go` Judge 接口返回 `([]Score, Reasoning, error)`，Score 语义改为逐 rubric dimension（按 spec D-9 v2.8）— 验证：`go build ./backend/...`
- [x] 1.3 同步 `judge.go` `NotImplementedJudge`、`judge_test.go`、`types_test.go` stubJudge 与所有 caller — 验证：`go vet ./backend/internal/ai/registry`
- [x] 1.4 全包绿灯 — 验证：`go test ./backend/internal/ai/registry -count=1`（含既有 resolver/loader/cache 测试）

## Phase 2: judge capability dispatch + 真实 LLMJudge 实现

- [x] 2.1 A3 judge dispatch / adapter 红灯：`Complete` 继续拒绝 judge profile，`CompleteJudge`/等价窄接口只接受 `CapabilityJudge`，`CompleteJudge` 调 chat profile fail-close，bootstrap 不再把 `judge_compatible` 判为未实现 — 验证：`go test ./backend/internal/ai/aiclient/... -run 'Test.*Judge|Test.*judge' -count=1`
- [x] 2.2 实现 judge capability dispatch 与 `providers/judge_compatible` adapter；延续 secret fail-fast、fallback、observability 与 privacy red-line — 验证：`go test ./backend/internal/ai/aiclient/... -run 'Test.*Judge|Test.*judge' -count=1`
- [x] 2.3 新增 `judge_llm_test.go`：录制 fixture judge 响应断言逐维度 `[]Score`（len==dimensions、维度名匹配、Value∈[0,1]、Reasoning.Summary 非空、EvidenceQuotes 允许为空数组而不报错）（先红）— 验证：`go test ./backend/internal/ai/registry -run TestLLMJudge`
- [x] 2.4 实现 `LLMJudge`（注入 RegistryClient+judge model client，经 judge.default 调用，逐维度产出 []Score+Reasoning；business prompt 经 registry 解析，judge 评分指令从 `config/evals/` 读取且不 hardcode）— 验证：`go test ./backend/internal/ai/registry -run TestLLMJudge` + `make lint-prompts-hardcode`
- [x] 2.5 fail-close 负向单测（judge profile 不可用 / 输出不可解析 / 被评估 output schema invalid / 维度数量与 rubric 不匹配 → error，不静默补零；schema 校验复用 A3 同一子集语义）— 验证：`go test ./backend/internal/ai/registry -run TestLLMJudgeFailClose`
- [x] 2.6 `var _ Judge = (*LLMJudge)(nil)` 编译期断言；保留 NotImplementedJudge 安全默认；记录构造注入方式 — 验证：`go build ./backend/...`

## Phase 3: judge.default 激活 + coverage 门禁扩展

- [ ] 3.1 改 `catalog_test.go`（judge.default 期望 active）+ `ai_profile_coverage` 测试期望（judge.default active + provider_ref/model 非 placeholder + 13 chat 非 placeholder + 负向 placeholder 拒绝）（先红）— 验证：`go test ./backend/internal/ai/aiclient/profile -run TestCatalog` + coverage lint 测试
- [ ] 3.2 `config/ai-providers.yaml` 新增/替换非 placeholder `judge_compatible` provider ref；`config/ai-profiles.yaml` judge.default `unsupported→active` + 指向该 provider + 移除 unsupported_reason；不改任何 chat 业务 profile status — 验证：`git diff config/ai-providers.yaml config/ai-profiles.yaml`（judge provider/profile 范围内）
- [ ] 3.3 扩展 `scripts/lint/ai_profile_coverage.py` placeholder 黑名单断言（provider_ref + model）和 judge provider protocol/capability 匹配断言，并接线顶层 `make lint` — 验证：`python3 scripts/lint/ai_profile_coverage.py --repo-root .`
- [ ] 3.4 coverage 绿灯 + 负向断言（placeholder 回填到 judge/chat profile 或 provider → exit 非 0）— 验证：`make lint-ai-profile-coverage` + `go test ./backend/internal/ai/aiclient/profile -count=1`

## Phase 4: ≥50 题离线评估集 + Promptfoo runner

- [ ] 4.1 落地 `config/evals/<feature_key>/` 用例（覆盖 §3.1.1 chat feature_key，总量 ≥50，含 1 个 en→multi fallback 用例）— 验证：eval count≥50 断言
- [ ] 4.2 新增 repo-owned pinned Promptfoo dependency 与 runner 落点（根 devDependency 或 workspace package；同步 lockfile；禁止未固定 `pnpm dlx`/全局安装）— 验证：`pnpm install --lockfile-only` drift clean + runner version 可由仓库脚本输出
- [ ] 4.3 Promptfoo registry-driven 配置（custom provider 经 RegistryClient 解析 + AIClient；LLMJudge 作 grader；不复制 prompt 正文）— 验证：registry-single-source drift check
- [ ] 4.4 录制 fixture 默认 + `EVAL_LIVE=1` opt-in；EVAL_LIVE 未设不打网络 — 验证：`make eval-offline`（fixture）+ no-network 断言
- [ ] 4.5 新增 `make eval-offline`（.PHONY+help，不纳入 make test）+ count≥50 断言 + single-source drift gate — 验证：`make eval-offline` + drift gate exit 行为
- [ ] 4.6 `make lint-prompts-hardcode` 仍 green（未复制第二份 prompt）— 验证：`make lint-prompts-hardcode`

## Phase 5: 验证、生命周期与收口

- [ ] 5.1 focused + 聚合验证（registry/aiclient/profile go tests、lint-ai-profile-coverage、eval-offline、make lint、make test、make build）— 验证：`make lint` / `make test` / `make build` + 上述命令全绿
- [ ] 5.2 active-scope zero-reference grep 门禁（完整旧编号、短写旧编号、stale spec version 在 active specs、F3 plans index、README 与代码/配置/脚本 truth source = 0，排除本 plan 自身 gate 定义目录与 completed plan/history/reports/work-journal/bugs 历史资产）— 验证：`rg -n '002-real-model|003-grayscale|F3 后续 002|prompt-rubric-registry/spec\.md.*v2\.[0-8]' docs/spec/*/spec.md docs/spec/prompt-rubric-registry/plans/INDEX.md config backend scripts --glob '!docs/spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/**'`
- [ ] 5.3 profile status 范围负向断言（仅 judge.default 改 status，13 chat 未翻动；judge provider ref 非 placeholder）— 验证：`git diff config/ai-profiles.yaml config/ai-providers.yaml`
- [ ] 5.4 docs/lifecycle 收口（validate_context、sync-doc-index --check、docs-check、git diff --check；004 置 completed；INDEX/work-journal/retrospective；缺陷建档）— 验证：`make docs-check` + `python3 .agent-skills/implement/shared/scripts/validate_context.py ...`
