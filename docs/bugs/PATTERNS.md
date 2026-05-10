# Bug 模式库

> 从历史 Bug 中归纳的通用问题模式。Agent 在诊断新问题时应先查阅本文件。

<!-- 模式模板：

## 模式 N：[模式名称]

- **相关 Bug**：BUG-XXXX, BUG-YYYY
- **典型症状**：...
- **检查清单**：
  1. ...
  2. ...

-->

## 模式 1：Cleanup commit 漏同步 test consumer

- **相关 Bug**：BUG-0023
- **典型症状**：truth source（`shared/*.yaml` / spec / generated 主体）已干净，但单包构建失败，报 `undefined: <RemovedConstant>`；`make codegen-check` 全绿但 `make test` 在某包失败；cleanup commit message 自述同步 owner 文档与 generated artifacts，未提及 test consumer。
- **检查清单**：
  1. 删除 enum value / JobType / capability / feature flag 等 cross-cutting 标识符前后，对仓库做一次 reverse-grep（`*.go` / `*.ts` / `*.tsx` / `*.yaml` / `*.json` / `*.tmpl` / `*.sql`），把命中点全部过一遍而不仅看 truth source。
  2. cleanup 类提交除 `make codegen-check` 之外，必须额外运行触达包的 `go test ./...` 与对应前端 `pnpm --filter ... test`，避免 generated 主体清理但 test 没跟上。
  3. 把 `internal-only` / `forbidden` / `allowed` 等测试断言数组当作契约 anchor — 一旦上游 enum 集合调整，必须把数组中的对应元素同步删除或追加，禁止保留悬空引用。
  4. 落地新 lint 时优先考虑跨 yaml/code 的双向 enum 一致性（如 `lint-jobs-consumers`）：truth source 删除后，凡引用该常量的测试断言数组必须能被静态识别。

## 模式 2：入口 Skill 在分支门禁前修改默认父分支

- **相关 Bug**：BUG-0033
- **典型症状**：用户报 bug / 回归后，`/change-intake` 或其它入口 skill 先修改 spec / plan / checklist / docs，再进入 `/implement`；随后 `git status` 显示未提交改动落在 `main` 等默认父分支上；下游 `/implement` 的 branch resolution 已经无法防止前置文档改动污染父分支。
- **检查清单**：
  1. 任何可能写文件的入口 skill 在首次 `apply_patch`、formatter、codegen、doc creation、bug/report/journal 写入前，先运行 `git status --short --branch`。
  2. 若当前在默认父分支且工作区干净，先 fast-forward-only 更新父分支，再创建 feature branch；不得在父分支上做 spec / plan / checklist 原地修订。
  3. 若已经在默认父分支产生当前会话改动，先确认父分支与远端同步，再 `git switch -c <feature-branch>` 保留改动并报告恢复动作。
  4. 若 dirty 内容来源不明或可能属于用户，停止并询问用户；禁止擅自 `stash`、`reset`、`checkout` 或把不明改动提交进当前任务。
