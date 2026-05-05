# Provider Registry and Capability Profiles Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-05

**关联计划**: [plan](./plan.md)

## Phase 1: Provider Registry schema 与 loader

- [x] 1.1 定义 `config/ai-providers.yaml` schema：`name` / `protocol` / `base_url_env` / `api_key_env` / `capabilities[]` / `version`；`stub` 可不声明 secret env ref，网络出站 provider 必须声明
- [x] 1.2 落地 registry loader + A4 SecretSource 解析，覆盖 provider name 唯一、protocol 合法、capability 非空、按 protocol 校验 secret env ref、被选中真实 provider 非 test fail-fast，且 `stub` provider 不需要伪造 secret
- [x] 1.3 落地 registry/profile snapshot 热加载语义：≤30s 生效、进行中调用使用旧快照、reload 失败不污染当前快照
- [x] 1.4 补齐 registry negative fixtures：重复 provider、未知 protocol、capability 拼写错误、网络出站 provider 缺 env ref、provider ref 不存在、capability mismatch、被选中真实 provider secret 缺失、fallback 超 2 跳，并补 `stub` 无 secret 正向 fixture

## Phase 2: Capability-scoped Model Profile schema

- [x] 2.1 将 profile schema 从 `task_type` / 全局 provider 口径迁移到 `capability` / `provider_ref` / `status`，为 `disabled` / `unsupported` 强制校验 `unsupported_reason`，不保留旧 schema key fallback
- [x] 2.2 补齐 F3 12 个 baseline default profile fixture 与 spec §4.5 非 F3 placeholder profile fixture，并为 P1/P2/002+ profile 使用 `status=disabled` / `status=unsupported` + `unsupported_reason` 表达不可执行状态
- [x] 2.3 建立 Product/UI capability coverage 检查，确保 spec §4.5 每个默认 profile 都是具体 profile name，且与 F3 feature_key 字典和 profile catalog 同步
- [x] 2.4 同步 `backend/internal/ai/aiclient/README.md`、`config/README.md` 与 fixture 注释

## Phase 3: AIClient routing, fallback, and fail-closed behavior

- [ ] 3.1 AIClient 按 profile `capability` + `provider_ref` + `status` 路由；disabled / unsupported profile 或 unsupported capability 返回 B1-owned `AI_UNSUPPORTED_CAPABILITY` 或同义 approved `AI_*` code 并记录 meta/log，不降级到 chat 或 stub
- [ ] 3.2 实现 profile central fallback chain，最多 2 跳，业务代码不得自行 retry-with-different-model
- [ ] 3.3 更新 observability / privacy tests，覆盖 capability meta、fallback metric/log、DB/audit metadata 无明文
- [ ] 3.4 重构 openai_compatible adapter 的 base URL / API key 来源为 provider ref secret，并保留 `/v1` 归一化测试

## Phase 4: A4 / B1 / F3 integration

- [ ] 4.1 A4 env/config 字典扩展为 `AI_PROVIDER_REGISTRY_PATH` + `AI_MODEL_PROFILE_PATH` + provider-specific secret env refs，并同步 `.env.example`、bindings、validator、redaction 与 `make lint-config`
- [ ] 4.2 B1 shared vocabulary 新增或迁移 AI capability enum、provider registry field names、profile field names、meta field names 与 provider/profile routing `AI_*` 错误码，codegen parity tests 通过
- [ ] 4.3 F3 + Product/UI profile coverage lint 覆盖 12 个 baseline feature_key 的默认 `model_profile_name` 与 spec §4.5 默认 profile
- [ ] 4.4 同步 ADR-Q6、A3 history、A4/F3 spec、engineering-roadmap A3 职责描述与 docs/spec INDEX

## Phase 5: Verification and handoff

- [ ] 5.1 Focused tests 通过：registry loader、profile schema、AIClient routing/fallback、openai_compatible adapter、observability/privacy、A4 config、B1 vocabulary、F3 + Product/UI profile coverage
- [ ] 5.2 Global gates 通过：`make lint-config`、`make codegen-check`、`make docs-check`、`make lint`、`make test`、`make build`
- [ ] 5.3 Active-scope negative search 通过：不含旧 schema key，不把 AI provider 描述为独立 provider-proxy 业务语义或单一全局 endpoint 当前目标架构
- [ ] 5.4 将 plan/checklist Header 切到 `completed`，同步 INDEX 与工作日志，并给 002 / C14 / practice / report / resume / debrief / F3 eval owner 留出 handoff
