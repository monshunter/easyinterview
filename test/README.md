# test

跨服务测试根目录。`test/scenarios/` 是当前 BDD / E2E 场景契约入口，包含单一本地 Kind 目标环境约定、场景索引和当前 Ready 场景脚本。缺少部署资产的 Ready 场景可通过 repo-tracked Go / Vitest / Playwright 脚本验证同一行为契约。

Current owner: [engineering-roadmap S3](../docs/spec/engineering-roadmap/spec.md#64-s3--true-integration-and-release-gate)。
