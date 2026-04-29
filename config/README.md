# Config

应用配置、feature flag 与 AI profile 的根容器。A1 只锁定 `config/` 顶层容器，不在本目录写入 secret 示例值。

Owner specs: [A4 secrets-and-config](../docs/spec/secrets-and-config/spec.md) owns `config/config.yaml`, `config/{dev,staging,prod}.yaml`, `config/feature-flags.yaml`, `.env.example` dictionary, and runtime config schema; [A3 ai-gateway-and-model-routing](../docs/spec/ai-gateway-and-model-routing/spec.md) consumes `config/ai-profiles/`; [F3 prompt-rubric-registry](../docs/spec/prompt-rubric-registry/spec.md) owns prompt/rubric registry paths that later live under this root.
