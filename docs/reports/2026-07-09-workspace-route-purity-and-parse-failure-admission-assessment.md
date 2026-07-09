# Workspace Route Purity and Parse Failure Admission 交付复盘

> 日期: 2026-07-09
> 类型: #assessment #BUG-0147
> 状态: completed

## 1 范围

本次修复覆盖两个同源回归：

- JD 解析失败不应持久化为可继续规划的 TargetJob，也不能进入面试规划列表。
- `workspace` 是纯列表页，不应从 URL 参数或 stale `InterviewContext` 派生目标岗位详情或启动面试上下文。

## 2 交付内容

- 后端 TargetJob failure path 改为事务内写 `target.analysis.failed` 后删除失败 `target_jobs`，保留诊断证据但删除用户可继续规划资产。
- TargetJob get/list 读侧排除 failed 资产，list 即使传 `analysisStatus=failed` 也不返回失败记录。
- 前端 workspace 只请求 ready TargetJob 列表，并过滤 failed / non-ready / 空 title 记录。
- `workspace` 从 interview context carry routes 中移除；TopBar 或 legacy `/workspace?...` 都清理上下文并显示列表。
- URL codec 清空 workspace safe params；legacy `/workspace?...` 会 canonicalize 为 `/workspace`。
- 卡片进入 `parse` 统一规划详情；`parse` 和 report 直接 start practice，不再经 workspace `autoStartPractice`。
- P0.012 / P0.018 场景文档和 verify gate 均加入当前准入和 route purity 断言。

## 3 验证摘要

- Frontend: typecheck PASS；workspace / parse / report / App focused Vitest PASS。
- Backend: `go test ./internal/targetjob -count=1` PASS；`cmd/api` TargetJob HTTP scenario focused PASS。
- Integration: real Postgres `TestSQLStoreIntegration_CompleteParseFailureDeletesTargetAndSources` PASS。
- BDD: `E2E.P0.012` setup / trigger / verify PASS；`E2E.P0.018` setup / trigger / verify PASS。
- Runtime: local env redeploy + verify PASS。
- Browser: real email-code login 后打开带 legacy context 参数的 `/workspace`，URL canonicalize 为 `/workspace`，页面显示 plan list；未出现 `JD 解析失败`、`缺少目标岗位 ID`、workspace detail anchors、console/page/HTTP failures。
- Docs: context validation PASS；sync-doc-index zero drift；docs-check PASS；`git diff --check` PASS。

## 4 后续建议

- 下一步建议执行 `/work-journal` 并用 `fix(workspace): enforce parse failure admission and route purity (BUG-0147)` 提交，锁定本次 BUG-0147 的代码和文档证据。
- 备选路径是再跑一次全量 frontend suite，确认非本 owner 的旧路由场景没有隐藏依赖。
