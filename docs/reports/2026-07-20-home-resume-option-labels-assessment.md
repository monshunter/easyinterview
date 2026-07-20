# Home 简历选项名称精简交付复盘报告

> **日期**: 2026-07-20
> **审查人**: Codex

**关联计划**: [001 Home + JD Import + Parse](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md)

## 1 复盘范围与成功证据

- 本次原地重开 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` Phase 30，将 Home 简历下拉业务选项从“名称 + 语言 + 来源 + 日期”收敛为仅显示 `displayName || title`；selectable predicate、更新时间排序、option `value`、选择状态与 import request 保持不变。
- TDD RED 精确捕获旧文本 `Zhang San - Backend Engineer · zh-CN · upload · 2026-07-19`；GREEN 后 `HomeResumeSelection.test.tsx` 7/7，通过 displayName 优先、title fallback 与四类元信息负向断言。
- `BDD.HOME.RESUME.OPTION.008` 的 selection/import owner behavior tests 16/16 通过；根 `make test` 通过 Python 615 tests / 4615 subtests、Go 全包与 frontend 全量测试。
- frontend typecheck、production build、owner context、Header/INDEX、docs links、pruning、scoped source search 与 `git diff --check` 全部通过；post-pass `/plan-review` 无实质性发现。

## 2 会话中的主要阻点/痛点

### 2.1 Matcher 未直接命中已有 Home owner

- **证据**：`match_change_context.py` 对截图文字与“只显示简历名字”的组合查询返回 `confidence=none`；随后通过 `HomeScreen.tsx` 的 `resumeMeta`、UI 文档和最近工作日志定位到原 plan。
- **影响**：增加了一次 live-search 与 owner 反查，但现有低置信度兜底流程正确阻止了新建 sibling plan 或误选 backend-resume owner。

### 2.2 连续交付分支未通过 plan-name 分支探测

- **证据**：当前 `feat/frontend-list-pages-ui-alignment-0719` 最近两个 Home commit 均属于同一 owner plan，但 `detect_session_branch.py` 只按 plan-name stem 判断，返回 `matchesSessionBranch=false`；分支没有 upstream 或同名远端。
- **影响**：需要额外检查分支远端状态与最近 owner commit，才能确认这是同一 plan 的连续恢复执行，而不是在不相关 dirty 分支上实施。

### 2.3 首次 focused Vitest 命令实际执行了全量套件

- **证据**：`pnpm --filter @easyinterview/frontend test -- <file>` 被 package script 解析为 `vitest run -- <file>`，实际运行 136 个文件；后续改用 `pnpm exec vitest run <file>` 才得到真正 focused 的 1-file run。
- **影响**：RED 反馈延迟，但没有造成错误结论；全量失败仍只有新增断言一项。

## 3 根因归类

- Matcher 的词面召回不足属于 `skill` 边界，但当前 live-search 兜底已按设计工作；本次不足以证明需要立即调整 ranking 规则。
- 连续交付分支只按名称判断属于 `skill` 可改进点：真实 owner 证据还包括近期 plan/journal commit 与 clean/unshared branch 状态。
- focused Vitest 参数透传差异属于 `README / spec-plan` 命令精度问题；测试入口应直接命名 runner，而不是依赖 package script 的参数透传语义。

## 4 对流程资产的改进建议

- 在 `.agent-skills/implement` 的 branch resolution 中增加“同一 owner 连续交付”只读证据规则：仅当当前分支未共享、近期 commit 明确关联同一 plan、dirty 文件也在 owner 范围内时，允许作为 resume；否则仍停止。**落点**：skill；**优先级**：medium。
- 后续 Home plan 的 focused 命令统一写为 `cd frontend && pnpm exec vitest run <files...>`；如该误用再次出现，再把精确命令补进 `frontend/README.md`。**落点**：spec-plan，必要时 README；**优先级**：low。
- Matcher 暂不为单次简短文案修订扩充通用关键词；继续以现有 live-search 兜底和原 plan 原地修订规则承接。**落点**：无需仓库改动；**优先级**：low。

## 5 建议优先级与后续动作

- 下一轮最值得处理的是 implement 分支恢复判定：补充受约束的 owner-history 证据，可减少同一长生命周期 UI 分支上的误阻塞，同时不放宽不明 dirty worktree 防线。
- focused Vitest 命令可在下一次修改同一 Home plan 时顺带统一；在再次出现前不单独建立治理任务。
- 本次行为已经闭环，不需要新增 Bug、API、fixture、backend、persistence、route、E2E ID 或浏览器验收。
