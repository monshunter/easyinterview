# Config

应用配置、feature flag 与 AI profile 的根容器。

Owner specs: [A4 secrets-and-config](../docs/spec/secrets-and-config/spec.md) owns config schema, env dictionary, feature flags, and runtime config; [A3 ai-gateway-and-model-routing](../docs/spec/ai-gateway-and-model-routing/spec.md) consumes `config/ai-profiles/`; [F3 prompt-rubric-registry](../docs/spec/prompt-rubric-registry/spec.md) owns prompt/rubric registry paths that later live under this root.
