# TargetJob Archive Import Terminal 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：修复 [BUG-0151](../bugs/BUG-0151.md)，让已归档或已不可见的 TargetJob 被 `target_import` worker 读取时以 non-retryable `TARGET_JOB_NOT_FOUND` 终结，不再进入 parse failure cleanup 或制造 retryable `TARGET_IMPORT_FAILED`。
- Owner 文档范围：原地同步 `backend-targetjob/001-targetjob-import-and-parse-bootstrap` spec / plan / checklist，补齐归档与 queued/retrying import job 的交叉 gate。
- 成功证据：
  - `go test ./backend/internal/targetjob -run TestParseExecutor_MissingTargetIsTerminalWithoutFailureCleanup -count=1` 修复前 RED，修复后 PASS。
  - `go test ./backend/internal/targetjob -run 'TestParseExecutor|TestSQLStore_ArchiveTargetJob|TestService_ArchiveTargetJob|TestHandler_ArchiveTargetJob|TestSQLStore_CompleteParseFailure' -count=1` PASS。
  - `go test ./backend/internal/targetjob -count=1` PASS。

## 2 会话中的主要阻点/痛点

- 归档实现和 parse failure 实现各自有测试，但缺少跨生命周期交叉场景。
  - **证据**：`archiveTargetJob` focused tests 覆盖 handler/service/store、read-side soft-delete 和 idempotency；`CompleteParseFailure` tests 覆盖失败删除；code review 仍发现 queued import job 在归档后会进入 retryable cleanup failure。
  - **影响**：需要新增 `TestParseExecutor_MissingTargetIsTerminalWithoutFailureCleanup`，并把该 gate 固化到 Phase 12。

- 失败清理路径把 cleanup 的 `ErrTargetJobNotFound` 转成 retryable `TARGET_IMPORT_FAILED`。
  - **证据**：红测输出 `{ErrorCode:TARGET_IMPORT_FAILED Retryable:false}`，并且原代码会调用 `p.fail()`，而 `p.fail()` 在 cleanup error 时固定返回 retryable `TARGET_IMPORT_FAILED`。
  - **影响**：如果不在入口识别已归档/缺失目标，async job 可能重试到 dead，给后台诊断带来噪声。

## 3 根因归类

- Phase 12 归档 gate 只覆盖 API/DB/read-side，未覆盖异步 runner 仍持有旧 resource id 的情况。
  - **类别**：spec-plan。

- 入口错误分类没有区分“目标已不可见的终态”与“系统读取失败”。
  - **类别**：spec-plan。

## 4 对流程资产的改进建议

- 对任何 soft-delete/archive 类 user action，在 owner checklist 中增加“pending async job after deletion/archive”交叉测试。
  - **落点**：spec-plan。
  - **优先级**：high。

- 后续 L2 code review 对 async job owner 增加一项反查：resource read 返回 not-found 时，handler 是否会终结、取消或重试。
  - **落点**：skill。
  - **优先级**：medium。

## 5 建议优先级与后续动作

- 最高优先级：保留本次新增的 Phase 12.5 gate，后续归档/删除类变更都以该测试作为交叉回归模板。
- 次优先级：若后续引入更多 user-owned async resource，可把 `ErrTargetJobNotFound` 终态处理模式抽成 handler-level helper，减少每个 executor 自行分类错误的风险。
