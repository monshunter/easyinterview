# Parse Waiting Back Action Removal 交付复盘报告

> **日期**: 2026-07-21
> **审查人**: Codex

**关联计划**:

- [Home JD Import and Parse](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md)
- [Resume Listing, Routing and Detail Readonly](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)

## 1 复盘范围与成功证据

本次交付移除 JD 与简历 queued/processing 解析等待场景中的内联“返回 / Back”动作，同时保留失败态恢复、轮询、ready replace、路由和 shared transition component 的可选 action 能力。

- TDD RED 证明旧实现仅在两条新等待态 action negative 上失败；随后 6 个 focused 文件、49 个测试通过。
- 根 `make test` 通过：backend `626 passed, 4628 subtests passed`，frontend `137` 个测试文件、`1126` 个测试通过。
- frontend typecheck 与 production build 通过；build 仅保留既有 chunk-size warning。
- `make dev-container-up` 重建并启动真实容器化 frontend/backend；`make dev-container-doctor` 证明 6/6 服务为 `OK`。
- 真实 Chrome 在 `1916x821` 与 `390x844` 下分别捕获简历与 JD processing 场景：`mainButtons=0`、`actionWraps=0`；移动端 `scrollWidth=390`，console issue 为 0。
- 两个 owner context、Header/INDEX、Markdown 链接、Spec contract ID、core-loop pruning 与 `git diff --check` 均通过。

## 2 会话中的主要阻点/痛点

### 2.1 初始运行拓扑判断错误

- **证据**：首次前端/后端重部署走了 host-run `env-redeploy` 路径；用户明确补充“当前项目是全 Docker 容器部署”后，改用 `make dev-container-up`，并由 6/6 container doctor 结果重新建立验收环境。
- **影响**：增加了一次无效的宿主机邮件路由排查和额外重部署成本，也延迟了 Chrome 验收。

### 2.2 Chrome 扩展的本地文件访问权限不可用

- **证据**：合成 Markdown 简历的 file chooser `setFiles` 被浏览器拒绝；随后改走产品正式“粘贴内容”入口，同样真实创建 queued/processing 简历并完成等待态验收。
- **影响**：无法用 Chrome 自动覆盖本地文件上传入口，但不影响本次等待态 UI 合同；正式 backend、持久化、AI parse 与轮询链路仍被粘贴路径驱动。

## 3 根因归类

- 运行拓扑误判属于**无需仓库改动**：`test/scenarios/README.md`、`deploy/dev-stack/README.md` 与 `/scenario-env` 已明确要求用户指定全容器时使用 `make dev-container-up`；问题是执行时沿用了旧的 host-run 假设。
- Chrome 文件上传失败属于**无需仓库改动**：这是本机扩展权限，不是产品代码、测试设计或场景环境缺陷；同一产品 owner 已有可用的粘贴入口作为当前验收路径。
- 两个 completed owner 原地恢复为 active、共同承接同一等待态 action 不变量，说明当前 **spec/plan** 原地修订规则足以覆盖本次跨页面小修订。

## 4 对流程资产的改进建议

- 不新增治理规则。后续本地验收在执行第一条环境命令前，应把用户指定的运行形态写成当前会话约束，并直接映射到既有 `/scenario-env` 入口。
  - **落点**：无需仓库改动；执行纪律
  - **优先级**：high
- 保留本次已固化的 owner negative tests 与 Chrome BDD 证据，不把 shared `AsyncTransitionScene.action` 改成全局禁用。
  - **落点**：spec/plan（已完成）
  - **优先级**：high
- 如未来必须验收本地文件上传，可由用户在 Chrome 扩展中开启 file URL access；不应为绕过本机权限增加产品测试后门或 mock 路径。
  - **落点**：无需仓库改动
  - **优先级**：low

## 5 建议优先级与后续动作

1. 请求用户确认两个 owner plan 可以恢复 `completed` lifecycle；确认后同步 Header/INDEX 并复跑文档门禁。
2. 后续全容器 UI 验收直接从 `make dev-container-up` 开始，避免再次进入 host-run 邮件/端口诊断。
3. Chrome 本地文件权限仅在后续确有上传入口验收需求时处理，本次无需追加代码或场景资产。
