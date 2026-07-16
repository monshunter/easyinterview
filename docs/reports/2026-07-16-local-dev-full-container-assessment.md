# Local Dev Full Container 交付复盘报告

> **日期**: 2026-07-16
> **审查人**: Codex

**关联计划**: [local-dev-stack/001-bootstrap](../spec/local-dev-stack/plans/001-bootstrap/plan.md)

## 1 复盘范围与成功证据

- 本次交付在同一 dev-stack Compose 中新增显式 `full-container` profile、backend/frontend 镜像、migration 前置与根 `dev-container-*` 生命周期入口，默认 frontend/backend host port 为 10800/10801，默认 `dev-up` host-run 语义保持不变。
- 静态与代码 gate 全部通过：full scenario contract 9 passed；根 `make test` 覆盖 Python 561 tests / 4481 subtests、backend Go 全包与 frontend 126 files / 1004 tests；`make build`、`make docs-check`、context validator、INDEX check、Compose config 与 `git diff --check` 通过。
- 真实部署 gate 通过：migrations 正常退出，doctor 为 6/6 OK，10800 frontend 与 10801 runtime-config 均返回 200。
- Chrome 使用 E2E.P0.101 标准 synthetic 邮箱完成邮箱验证码、资料补全、简历 AI 解析、JD/面试计划、真实 AI 问答和报告生成；未使用 route mock/interception，console warning/error 为 0，四张脱敏截图保存在 `.test-output/full-container-chrome/`。

## 2 会话中的主要阻点/痛点

### 2.1 现有 automated E2E 入口绑定 host-run 端口

- **证据**：E2E.P0.101 的 setup/trigger 固定使用 frontend 5173 与 backend 8080，不能直接消费本次 10800/10801 全容器部署；本次只能按计划使用 Chrome 手工闭环。
- **影响**：同一业务流在两种受支持的本地部署形态之间无法复用自动化 runner，容易让执行者误把 host-run PASS 当作 full-container 证据。

### 2.2 验收 synthetic 邮箱命名出现一次口径返工

- **证据**：首次 Chrome 验收使用了 `full-container-<timestamp>@example.test`；用户指出后，执行者废弃该轮并按 E2E.P0.101 的 `auth-email-code-e2e-p0-101-<run-id>@example.test` 重新完成全链路。
- **影响**：虽然两者都使用保留测试域且不影响产品行为，但截图与数据无法直接对应现有场景命名约定，造成一次完整登录流程返工。

### 2.3 frontend 容器首次安装了根 workspace 依赖

- **证据**：初版 Dockerfile 触发根 workspace 依赖安装，随后改为 frontend owner 的 isolated、frozen install，并补齐 frontend build 实际引用的 OpenAPI/shared fixtures 后镜像构建通过。
- **影响**：首次构建下载和缓存范围过大，延长反馈时间；未造成最终产物或仓库依赖变更。

## 3 根因归类

- E2E endpoint 固定值属于 `README/spec-plan` 与场景 runner 参数化合同缺口：框架已声明 full-container 是同一真实环境的可选形态，但 P0.101 runner 尚未支持 endpoint override。
- synthetic 邮箱返工属于 `spec-plan` 验收数据约定不够显式；本次已把标准格式写入 Phase 12 checklist，因此后续无需新增治理层规则。
- frontend install 返工属于实现阶段的 Docker build-context 细节，现有 focused contract 与真实 image build 已覆盖最终结果，归类为 `no repo change needed`。

## 4 对流程资产的改进建议

- 为 E2E.P0.101 的 frontend origin / API base 增加受控 env override，并保留 5173/8080 为默认值；full-container 验收传入 10800/同源或 10801。**落点**：`test/scenarios/e2e/p0-101-auth-email-code-profile-setup` README/scripts；**优先级**：high。
- 若更多场景需要支持两种本地部署形态，把 endpoint 解析收口到 `_shared` helper，禁止各场景重新硬编码端口。**落点**：`test/scenarios/_shared` 与 `test/scenarios/README.md`；**优先级**：medium。
- 保持 Phase 12 checklist 中的标准 synthetic 邮箱与脱敏要求，不再上升为 AGENTS.md 全局规则。**落点**：当前 `spec-plan`；**优先级**：已完成。

## 5 建议优先级与后续动作

1. 下一轮优先由 `/change-intake` 原地修订 E2E.P0.101 owner，使同一 automated 场景可通过 endpoint override 验收 full-container，而不是复制 sibling 场景。
2. 当第二个场景出现同类需求时，再抽取共享 endpoint resolver；当前不提前泛化。
3. Dockerfile 维持 isolated frontend install 与实际 image build gate，不新增重复的 workspace 依赖测试层。
