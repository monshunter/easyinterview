# test

跨服务测试根目录。`test/scenarios/` 是当前 BDD / E2E 场景契约入口，包含本地 runner 约定、场景索引和当前 Ready 场景脚本。Ready 场景默认通过 repo-tracked Go / Vitest / Playwright / browser smoke 脚本验证同一行为契约；外部依赖按需由 `make dev-up` 提供。

Current owner: [engineering-roadmap S3](../docs/spec/engineering-roadmap/spec.md#64-s3--true-integration-and-release-gate)。
