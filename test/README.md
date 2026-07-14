# test

跨服务测试根目录。`test/scenarios/` 是当前 BDD / E2E 场景契约入口，包含共享环境约定、场景索引和当前 Ready 场景脚本。E2E 只通过真实 HTTP API，或通过浏览器访问真实 frontend 且让业务请求落到真实 backend；外部依赖按需由 `make dev-up` 提供。

`go test`、Vitest、pytest、lint、build、source contract 与 package smoke 属于代码层 gate，不得嵌入 E2E 脚本或充当场景证据。开发中可运行 focused test 获取反馈，阶段收口时由仓库根 `make test` 统一执行前后端全量单测。

Current owner: [engineering-roadmap S3](../docs/spec/engineering-roadmap/spec.md#64-s3--true-integration-and-release-gate)。
