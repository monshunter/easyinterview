# Provider Registry and Capability Profiles Checklist

> **版本**: 1.15
> **状态**: active
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Provider Registry schema 与 loader

- [x] 1.1 定义 `config/ai-providers.yaml` schema：`name` / `protocol` / `base_url_env` / `api_key_env` / `capabilities[]` / `version`；`stub` 可不声明 secret env ref，网络出站 provider 必须声明
- [x] 1.2 落地 registry loader + A4 SecretSource 解析，覆盖 provider name 唯一、protocol 合法、capability 非空、按 protocol 校验 secret env ref、被选中真实 provider 非 test fail-fast，且 `stub` provider 不需要伪造 secret
- [x] 1.3 落地 registry/profile snapshot 热加载语义：≤30s 生效、进行中调用使用原快照、reload 失败不污染当前快照
- [x] 1.4 补齐 registry negative fixtures：重复 provider、未知 protocol、capability 拼写错误、网络出站 provider 缺 env ref、provider ref 不存在、capability mismatch、被选中真实 provider secret 缺失、fallback 超 2 跳，并补 `stub` 无 secret 正向 fixture
- [x] 1.5 L2 remediation: 生产 bootstrap 实际读取 registry/profile path、调用 `ResolveSelectedProviders`、按 provider ref materialize adapter，并用 focused tests 覆盖非 test selected provider secret fail-fast
- [x] 1.6 L2 remediation: profile hot reload 失败保持原快照且通过 `OnWarn` 输出可观测 warning，并补 focused test

## Phase 2: Capability-scoped Model Profile schema

- [x] 2.1 将 profile schema 从 `task_type` / 全局 provider 口径迁移到 `capability` / `provider_ref` / `status`，为 `disabled` / `unsupported` 强制校验 `unsupported_reason`，loader 只接受 current schema keys 并拒绝 out-of-scope key
- [x] 2.2 当前 F3 6 个 baseline feature_key 的 default profile 引用与 spec §4.5 非 F3 fail-closed profile fixture 均存在，并为 P1/P2/002+ profile 使用 `status=disabled` / `status=unsupported` + `unsupported_reason` 表达不可执行状态
- [x] 2.3 建立 Product/UI capability coverage 检查，确保 spec §4.5 每个默认 profile 都是具体 profile name，且与 F3 feature_key 字典和 profile catalog 同步
- [x] 2.4 同步 `backend/internal/ai/aiclient/README.md`、`config/README.md` 与 fixture 注释
- [x] 2.5 将 per-profile YAML directory active truth source 收敛为单一 `config/ai-profiles.yaml` catalog，并用 profile loader / tracked catalog / coverage lint focused tests 验证 catalog 文件路径、重复 profile、缺失 profile 与 out-of-scope 目录引用被拦截

## Phase 3: AIClient routing, fallback, and fail-closed behavior

- [x] 3.1 AIClient 按 profile `capability` + `provider_ref` + `status` 路由；disabled / unsupported profile 或 unsupported capability 返回 B1-owned `AI_UNSUPPORTED_CAPABILITY` 或同义 approved `AI_*` code 并记录 meta/log，不降级到 chat 或 stub
- [x] 3.2 实现 profile central fallback chain，最多 2 跳，业务代码不得自行 retry-with-different-model
- [x] 3.3 更新 observability / privacy tests，覆盖 capability meta、fallback metric/log、DB/audit metadata 无明文
- [x] 3.4 重构 openai_compatible adapter 的 base URL / API key 来源为 provider ref secret，并保留 `/v1` 归一化测试

## Phase 4: A4 / B1 / F3 integration

- [x] 4.1 A4 env/config 字典扩展为 `AI_PROVIDER_REGISTRY_PATH` + `AI_MODEL_PROFILE_PATH` + provider-specific secret env refs，并同步 `.env.example`、bindings、validator、redaction 与 `make lint-config`
- [x] 4.2 B1 shared vocabulary 新增或迁移 AI capability enum、provider registry field names、profile field names、meta field names 与 provider/profile routing `AI_*` 错误码，codegen parity tests 通过
- [x] 4.3 F3 + Product/UI profile coverage lint 覆盖当前 6 个 baseline feature_key 的默认 `model_profile_name` 与 spec §4.5 默认 profile
- [x] 4.4 同步 ADR-Q6、A3 history、A4/F3 spec、engineering-roadmap A3 职责描述与 docs/spec INDEX
- [x] 4.5 L2 remediation: `make lint-ai-profile-coverage` 拒绝 repo-tracked active profile 指向 `stub` protocol provider，并将当前 active default profiles 切到 non-stub provider ref

## Phase 5: Verification and handoff

- [x] 5.1 Focused tests 通过：registry loader、profile schema、AIClient routing/fallback、openai_compatible adapter、observability/privacy、A4 config、B1 vocabulary、F3 + Product/UI profile coverage
- [x] 5.2 Global gates 通过：`make lint-config`、`make codegen-check`、`make docs-check`、`make lint`、`make test`、`make build`
- [x] 5.3 Active-scope negative search 通过：不含 out-of-scope schema key，不把 AI provider 描述为独立 provider-proxy 业务语义或单一全局 endpoint 当前目标架构
- [x] 5.4 将 plan/checklist Header 切到 `completed`，同步 INDEX 与工作日志，并给 002 / C14 / practice / report / resume / F3 eval owner 留出 handoff
- [x] 5.5 L2 remediation verification: focused tests、profile coverage lint、config lint、context validation、negative search 与必要全局 gate 通过，plan/checklist Header 与 active spec 003 状态投影均为 `completed`
- [x] 5.6 Catalog consolidation verification: focused Go profile tests、profile coverage pytest、`make lint-ai-profile-coverage`、`make lint-config`、context validation、`make docs-check`、active-scope out-of-scope profile-directory 负向搜索与必要全局 gate 通过，completed 状态保持一致
- [x] 5.7 L2 remediation: 修复 dev-stack / product owner matrix out-of-scope profile directory 漂移，并补强 deploy/profile semantic drift gate；验证 `make lint-ai-profile-coverage`、`make lint-config`、context validation、active-scope out-of-scope profile-directory 负向搜索与必要 focused tests 通过

## Phase 6: DeepSeek baseline and retrieval cleanup

- [x] 6.1 删除当前 active scope 的向量化 / 重排 capability、profile、provider protocol、AIClient 方法、OpenAI-compatible wire、stub、job type、migration 表/索引与 dev-stack 依赖。验证: focused Go tests、B1/B3 codegen drift、migration lint、active-scope negative search
- [x] 6.2 将 repo-tracked AI provider 开发主力收敛为 `deepseek`，chat profile 只使用 `deepseek-v4-flash` / `deepseek-v4-pro`；STT / realtime 在本阶段继续 fail-closed，judge 仅由后续 Phase 9 独立激活。验证: profile catalog tests、`make lint-ai-profile-coverage`、out-of-scope 模型别名负向搜索
- [x] 6.3 同步 A3 / B1 / B3 / B4 / F3 active spec、README、lint、fixtures 与 generated artifacts，使文档、配置、代码和基础设施契约一致。验证: `make docs-check`、context validation、`make lint-config`
- [x] 6.4 完成全局验证并确认 Header 状态：focused tests、codegen idempotency、`make lint-ai-profile-coverage`、`make lint-config`、`make docs-check`、active-scope negative search 通过，plan/checklist 均保持 `completed`

## Phase 7: Provider startup error contract cleanup

- [x] 7.1 删除仅由测试自证的 `providerregistry.SharedErrorCode` 映射层；保留 `ErrProviderConfigInvalid` / `ErrProviderSecretMissing` 的 `errors.Is` 合同，并通过 focused tests、staticcheck、production deadcode、symbol inventory、AI/config lints、owner contexts 与 docs/diff/pruning gates 验证
  <!-- verified: 2026-07-10 method=shared-provider-error-mapper-removal evidence="Production deadcode RED listed SharedErrorCode. Removed the mapper and two self-only assertions while preserving both errors.Is startup sentinels. Focused/full AI tests, staticcheck, deadcode/symbol inventory, AI/config lints and owner contexts PASS." -->

## Phase 8: report generation profile budget

- [x] 8.1 RED: profile loader/coverage tests require `report.generate.default` context_window_tokens=1000000, max_tokens=6144, timeout_ms=60000, rate_limit.tpm=60000, version=1.2.0 and unchanged DeepSeek Pro route; missing/non-positive/`<=max_tokens` capacity, 4096, budget-without-version-bump and unrelated profile mutations fail.
  <!-- verified: 2026-07-12 method=tdd-red evidence="Profile tests failed to compile without ContextWindowTokens; coverage lint incorrectly accepted missing context_window_tokens; B1 vocabulary lacked the generated field constant." -->
- [x] 8.2 GREEN: update the single `config/ai-profiles.yaml` catalog with context_window_tokens=1000000 and max_tokens 4096→6144/version 1.1.0→1.2.0 only; run focused loader tests, `make lint-ai-profile-coverage` and config lint. TPM remains a separately parsed throughput hint.
  <!-- verified: 2026-07-12 method=tdd-green evidence="Profile package, 13 coverage tests, lint-ai-profile-coverage, lint-conventions and full lint-config PASS; tracked report profile changed only context_window_tokens/max_tokens/version." -->
- [x] 8.3 BOUNDARY-GATE: consume backend-review `REPORT_BOUNDARY_FIXTURES_READY`, exact 48,000/+1-byte framed inputs and current-schema worst-case zh/en outputs; offline prove `48,000 + 2,048 framing reserve + 6,144 < 1,000,000 context window`. TPM arithmetic is forbidden as capacity proof; A3 does not reimplement +1-byte business failure behavior.
  <!-- verified: 2026-07-12 method=offline-boundary evidence="Review marker REPORT_BOUNDARY_FIXTURES_READY PASS; A3 verified manifest, exact bytes/SHA-256/current output shape and 48000+2048+6144=56192<1000000. TPM was checked only as unchanged throughput." -->
- [x] 8.4 LIVE-TOKEN-GATE + HANDOFF: opt-in real-provider smoke records redacted AICallMeta usage for exact framed input and zh/en output-fixture token-count probes, rejects missing usage/`finish_reason=length`/over-budget results, then emits `REPORT_PROFILE_6144_PASS` with executable offline evidence; P0.100 must repeat the live gate before final acceptance. YAML grep alone cannot satisfy this marker.
  <!-- verified: 2026-07-12 method=real-provider-smoke evidence="framed-48000 input/output=7050/1710 stop reserved=15242; en fixture=1298/41 stop reserved=9490; zh-CN fixture=2732/82 stop reserved=10924. All usage present, fixture input <=6144, no length finish, all below 1M. P0.100 must repeat current-run live gate." -->

## Phase 9: context-aware judge final-content reliability

- [x] 9.1 RED: tracked catalog test requires judge.default thinking=disabled / response_format=json_object / max_tokens=6144 / timeout=60000 / tpm=60000 / version=1.2.0; adapter request test requires the exact wire. Reasoning-only fixture must fail with AI_OUTPUT_INVALID and no reasoning text leakage.
  <!-- verified: 2026-07-12 method=tdd-red evidence="catalog still had max_tokens=2048/v1.1.0/no thinking or JSON params; adapter omitted both wire fields and returned reasoning-only length response as nominal empty completion" -->
- [x] 9.2 GREEN: change only judge.default thinking/response format/max tokens/version; keep provider/model/route/fallback/timeout/TPM unchanged. Adapter emits only redacted finish/token/reasoning-presence metadata on empty final content and preserves invalid provenance.
  <!-- verified: 2026-07-12 method=tdd-green evidence="judge_compatible and profile focused tests PASS; reasoning-only error contains finish_reason=length, completion_tokens and presence=true but not private reasoning; lint-ai-profile-coverage and lint-config PASS" -->
- [x] 9.3 LIVE-JUDGE-GATE + HANDOFF: real complete+judge smoke requires stop, positive usage, non-empty JSON, thresholds/item/causal/critical pass; emit `JUDGE_FINAL_CONTENT_V120_PASS`. P0.100 must still rerun all five cases / 11 attempts.
  <!-- verified: 2026-07-12 method=real-provider-smoke evidence="completion v1.2.0 stop input/output=1928/1230 bytes=1235; judge v1.2.0 stop input/output=1318/791 bytes=2999; weighted=1, item verdicts=7, causal checks=1, zero tolerance=0, critical=true; JUDGE_FINAL_CONTENT_V120_PASS. Full P0.100 remains a separate gate." -->

## Phase 10: report generation non-thinking structured output

- [x] 10.1 RED: openai-compatible contract tests require `thinking=enabled|disabled` to map to the exact official object while object `output_schema` still drives `response_format=json_object`; adapter invalid values must fail before network, and loader/tracked catalog/coverage lint reject missing or illegal report thinking.
  <!-- verified: 2026-07-13 method=tdd-red evidence="adapter omitted thinking and sent all three invalid values; loader accepted auto/boolean/number; tracked catalog lacked report thinking; coverage exact-coordinate positive fixture failed after adding thinking" -->
- [x] 10.2 GREEN: set only `report.generate.default.default.params.thinking=disabled`; add the official openai-compatible thinking object and strict allowlist to Complete/Stream request construction; keep report response format absent from profile params and output-schema driven.
  <!-- verified: 2026-07-13 method=tdd-green evidence="focused provider/profile tests and report profile coverage cases PASS; enabled/disabled exact wire, object-schema JSON mode, invalid pre-request fail-close, tracked catalog non-thinking coordinate and missing/invalid lint cases are executable" -->
- [x] 10.3 VERIFY + HANDOFF: run provider/profile full and race tests, full profile coverage/lint, owner context/index/docs gates and `git diff --check`; P0.100 must rerun the live report matrix before final acceptance.
  <!-- verified: 2026-07-13 method=focused-race-contract-docs evidence="openai-compatible/profile full and race PASS; 13 profile-coverage tests and tracked lint PASS; smoke-tag consumer compiles without live call; owner context, zero-drift index/docs links and diff check PASS. Live P0.100 remains root-owner final acceptance." -->

## Phase 11: typed profile defaults and provider response cap

- [ ] 11.1 RED: active profile missing token fields、合法 override、显式零/负数/invalid capacity 与四 adapter response body limit/+1 测试先失败。
- [ ] 11.2 GREEN: 每个 active profile 增加与 tracked catalog 一致的 typed code default；显式非法值 fail-closed，catalog 坐标不漂移。
- [ ] 11.3 GREEN: 四个 provider adapter 统一消费 A4 `ai.maxResponseBodyBytes=4194304`，删除 adapter-local 4MiB 与无界 response read；limit 接受、limit+1 返回 typed provider error。
- [ ] 11.4 CAPACITY-GATE: 消费 917,504/917,505-byte review fixtures，证明 `917504+2048+6144=925696<1000000`；62,397-byte regression sample 调用 provider；TPM 不参与裁决。
- [ ] 11.5 VERIFY: provider/profile focused/full/race、coverage/config lint、P0.056 与 opt-in P0.100 token usage、contexts/docs/diff gates 全部通过。
