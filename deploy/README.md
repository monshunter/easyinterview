# deploy

部署资产根目录。当前已落地的是 `deploy/dev-stack/` Docker Compose 外部依赖栈；Helm / Kind / K8s 不是当前 P0 本地测试或部署默认前提，后续 staging / 生产部署清单由 release workstream 按需原地设计。

Current owner: release workstream 按 [engineering-roadmap S3](../docs/spec/engineering-roadmap/spec.md#64-s3--true-integration-and-release-gate) on-demand 创建；本地开发栈由 [local-dev-stack](../docs/spec/local-dev-stack/spec.md) 约束。
