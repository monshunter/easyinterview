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
