# Fixture-backed Mock Runtime Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-05

**关联计划**: [plan](./plan.md)

## Phase 1: Fixture registry 与 coverage

- [x] 1.1 建立 operationId -> fixture registry；验证: 新增 focused unit test 覆盖至少 Auth、TargetJobs、PracticeSessions、Reports、Resumes、Debriefs 的 fixture lookup，未知 operationId 返回明确错误
- [x] 1.2 补齐 fixture coverage lint；验证: negative fixture 或临时缺失 fixture 测试先 Red 后 Green，并实际运行 `make validate-fixtures` 与 `make lint-openapi`，确认 `openapi/openapi.yaml` operationId 与 fixture registry 无缺失/多余

## Phase 2: Frontend mock transport

- [x] 2.1 接入 generated client mock transport；验证: frontend focused test 断言 generated client 在 mock mode 下返回 typed fixture response，至少覆盖 `getRuntimeConfig`、`getMe`、`listTargetJobs`、`getPracticeSession`
- [x] 2.2 拦截 prototype data 运行时依赖；验证: lint/test 断言 `frontend/src` 不 import `ui-design/src/data.jsx`，且 mock response 不含 prototype-only display fields

## Phase 3: Backend mock harness

- [x] 3.1 提供后端 mock handler / router；验证: backend focused test 通过 HTTP request 命中 mock handler，返回与 fixture registry 相同的 JSON shape 和 status code
- [x] 3.2 建立错误态和 seed profile；验证: tests 覆盖未登录、已登录、缺 session、缺简历、报告生成中、隐私删除请求至少六种 named fixture scenarios，每种 scenario 可重置；请求未知 scenario 必须 fail loudly，不能静默回落到 `default`

## Phase 4: Drift gates and handoff

- [x] 4.1 接入本地质量门禁；验证: 本地 lint/codegen gate 执行 fixture coverage、prototype import boundary 和 scoped retired token negative search，失败时输出 owner spec 指引；收口实际运行 `make validate-fixtures`、`make lint-openapi`、`make codegen-check`、`make docs-check`
- [x] 4.2 Handoff 给 frontend-shell；验证: `frontend-shell/001-app-shell-auth-settings` context references 可指向本 spec，mock runtime README 或 package docs 说明可消费入口、seed profile 和阻塞条件
- [x] 4.3 L2 remediation: `lint-mock-contract` 必须执行 operation registry metadata test，Go generated route table 注释不得保留旧 36-row 口径，前端 mock transport 必须有 named scenario / unknown scenario 回归测试；验证: focused Red-Green 后运行 registry unittest、Go codegen test、frontend mockTransport test、`make lint-mock-contract`、`make codegen-check`、`make docs-check`
  <!-- verified: 2026-05-05 method=tdd-red-green focused=makefile-dry-run,go-codegen-openapi,frontend-mockTransport,fixture-registry gates=lint-mock-contract,docs-check,codegen-check-temp-index -->
