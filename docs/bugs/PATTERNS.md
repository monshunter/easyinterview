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

- **相关 Bug**：BUG-0035
- **典型症状**：用户报 bug / 回归后，`/change-intake` 或其它入口 skill 先修改 spec / plan / checklist / docs，再进入 `/implement`；随后 `git status` 显示未提交改动落在 `main` 等默认父分支上；下游 `/implement` 的 branch resolution 已经无法防止前置文档改动污染父分支。
- **检查清单**：
  1. 任何可能写文件的入口 skill 在首次 `apply_patch`、formatter、codegen、doc creation、bug/report/journal 写入前，先运行 `git status --short --branch`。
  2. 若当前在默认父分支且工作区干净，先 fast-forward-only 更新父分支，再创建 feature branch；不得在父分支上做 spec / plan / checklist 原地修订。
  3. 若已经在默认父分支产生当前会话改动，先确认父分支与远端同步，再 `git switch -c <feature-branch>` 保留改动并报告恢复动作。
  4. 若 dirty 内容来源不明或可能属于用户，停止并询问用户；禁止擅自 `stash`、`reset`、`checkout` 或把不明改动提交进当前任务。

## 模式 3：Vite dev 中相对 API base URL 误打前端端口

- **相关 Bug**：BUG-0036
- **典型症状**：前端 dev server 运行在 `5173`，页面请求 `/api/v1/...` 时 Network 面板显示目标也是 `localhost:5173`；后端未启动时大量真实 API 报错，已开发页面因 bootstrap data 失败而无法查看；组件测试能过，但真实 `main.tsx` 启动路径失败。
- **检查清单**：
  1. 检查 generated client 是否默认使用相对 `/api/v1`，以及 Vite config 是否有显式 `/api` proxy；没有 proxy 时相对 URL 会落到前端 origin。
  2. 检查 `main.tsx` 是否直接 `new EasyInterviewClient()`；正式 app bootstrap 必须通过可测试 factory 选择 dev mock / real backend 模式。
  3. Vite dev 默认应能在 backend absent 时展示 fixture-backed 页面；真实 backend 模式必须显式 opt-in，并指向 backend port 或 `VITE_EI_API_BASE_URL`。
  4. Playwright smoke 不应只靠 route mock；至少有一条 dev-preview smoke 断言真实页面加载期间没有意外 `/api/v1` network request。

## 模式 4：Completed checklist 掩盖未执行的 runner gate

- **相关 Bug**：BUG-0064, BUG-0066, BUG-0067
- **典型症状**：plan/checklist 标记 `completed`，但 test checklist / BDD checklist 仍有未勾选项；scenario `verify.sh` 只检查 spec 文件存在、历史说明或宽泛 `PASS` 字样；pixel parity / scenario wrapper 被写成 deferred 或外部运行，仍被计入完成证据。
- **检查清单**：
  1. 对 completed plan 先 `rg "\\[ \\]|deferred|pending|no tests|Playwright.*待|pixel parity 待"`，把空勾选、延期口径和 no-op 风险当作 blocking drift。
  2. 直接读取每个 scenario 的 `trigger.sh` / `verify.sh`，确认 `trigger.sh` 真正调用 runner，`verify.sh` 检查 runner marker、目标测试名或 spec path、pass marker，并显式拒绝 failed / no tests。
  3. Pixel parity gate 必须证明浏览器 runner 执行过；不能只检查 Playwright spec 文件存在或在 README 中写“可手动运行”。
  4. 文档收口时把证据 artifact 名称写成当前脚本真实产物，例如 `.test-output/e2e/<scenario>/trigger.log`，避免 checklist 引用不存在的 `*.evidence.log`。
