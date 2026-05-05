# shared

跨语言真理源落点：枚举 / 错误码 / 异步 Job 状态 / API 包装结构 / ID 与时间约定。`conventions.yaml` 是 Go 与 TS 共享类型的唯一镜像，由 `make codegen-conventions` 渲染到 `backend/internal/shared/` 与 `frontend/src/lib/`。

Owner subspec: [shared-conventions-codified](../docs/spec/shared-conventions-codified/spec.md)。当前实现以本目录 YAML、生成物和 active spec 为准。
