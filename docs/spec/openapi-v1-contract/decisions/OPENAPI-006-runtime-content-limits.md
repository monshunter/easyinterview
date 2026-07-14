# OPENAPI-006 · Runtime content limits

> **ID**: OPENAPI-006
> **状态**: accepted
> **日期**: 2026-07-14
> **版本**: 1.0

## 1 背景

Resume 上传/粘贴、TargetJob JD 文本与 Practice 消息需要在浏览器发请求前按后端当前配置进行 UTF-8 byte 预检。历史前端分别持有 2MiB 或字符数常量，无法跟随后端配置覆盖，造成前后端边界漂移。项目尚未上线；用户于 2026-07-14 明确批准方案 A 及修订默认值。

## 2 决策

- `RuntimeConfig` 新增 required `contentLimits`。
- `ContentLimits` 是 `additionalProperties=false` 的 required object，精确包含五个 positive int64：`resumeUploadBytes`、`resumePasteTextBytes`、`targetJobRawTextBytes`、`practiceMessageBytes`、`practiceSessionTextBytes`。
- fixture 默认值精确为 `10485760/393216/98304/32768/262144`。
- report framed input、HTTP request body、AI provider response body、profile context/output token 等内部限制不得出现在 public schema。
- 业务 request wire、37 operations / 10 tags、`GET /runtime-config` method/path/operationId/200 保持不变；后端 domain validation 仍是权威边界。
- OpenAPI source、fixture、Go/TS generated types、backend runtime builder 与 Resume/Home/Practice consumers 同批迁移，不保留可选字段或 `any` fallback。

## 3 迁移与回滚

- 003 Phase 10 必须在 baseline 不变时 exact-match old baseline → proposed OpenAPI finding set，并保存 audit artifact。
- lint、fixture、codegen、backend、frontend 与相关 BDD gate 全部通过后才允许原地 re-freeze v1.0.0 baseline。
- 任一 consumer 未迁移时整体回滚本 correction；不得把 `contentLimits` 改成 optional 来维持旧客户端形态。

## 4 审计

| 项 | 内容 |
|----|------|
| 提议人 | product owner |
| Review | user explicitly approved Scheme A and revised defaults on 2026-07-14 |
| 实施分支 | `fix/runtime-size-limits-0714` |
| expected finding oracle | `OPENAPI-006-runtime-content-limits.expected-findings.json` |
| baseline | `openapi/baseline/openapi-v1.0.0.yaml`；consumer gates 全绿后才允许 re-freeze |
| history | `2026-07-14 | 1.60 | OPENAPI-006 Runtime content limits` |

## 5 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 1.0 | 接受 required closed 五字段 RuntimeConfig public content-limit projection。 |
