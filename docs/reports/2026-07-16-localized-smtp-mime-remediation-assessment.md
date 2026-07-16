# Localized SMTP MIME Remediation 交付复盘报告

> **日期**: 2026-07-16
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 `backend-auth/001-email-code-session-bootstrap` 中文验证码邮件的 RFC 2047 Subject 与 quoted-printable text/plain/text/html 编码，并原地修正 `ai-provider-and-model-routing/001-aiclient-and-profile-bootstrap` checklist 3.1 的旧 SDK 口径。
- `TestSMTPDeliveryWriterEncodesLocalizedMessageAsStandardsCompliantMIME` 先在 raw UTF-8 Subject 上 RED，随后以标准 `net/mail`、`mime/multipart` 和 quoted-printable reader 验证中文主题、两种正文与验证码无损解码；全部 `backend/internal/auth` 测试通过。
- 根 `make test` 通过：Python 567 tests / 4481 subtests、Go 全包、frontend 126 files / 1004 tests；`make build`、A3 provider terminology lint、两个 owner context、`make docs-check` 与 `git diff --check` 均通过。
- `BDD.AUTH.EMAIL.002` 保持 domain behavior test 证据层，未新增或误标真实 E2E。

## 2 会话中的主要阻点/痛点

- 既有 SMTP transport gate 覆盖 TLS、认证、错误与隐私，但只对英文 raw message 做字符串断言，没有用标准 MIME reader 覆盖中文 locale；因此 `charset=UTF-8` 声明掩盖了 Subject 和 body transfer encoding 缺失。
- A3 Phase 15 已把实现和 plan/spec 切到 adapter-private `openai-go/v3`，Phase 3 checklist 仍保留“禁止 vendor SDK”的历史已勾选描述，造成 completed owner 内部自相矛盾。

## 3 根因归类

- SMTP 问题属于 `spec-plan` 覆盖缺口：本地化是用户可见 alternate path，但原 gate 验证的是字符串存在，不是邮件协议可解析性。
- A3 问题属于 `spec-plan` 修订遗漏：后续 Phase 改变了早期实现约束，却未反向校对全部历史 checked items。
- 本次执行没有环境阻塞；不需要调整 `AGENTS.md`、README 或运行环境。

## 4 对流程资产的改进建议

- `spec-plan`：邮件类 owner 的 locale gate 应以标准 parser 解码后的用户可见结果为准，不只检查 raw bytes；本次已在 backend-auth Phase 13 和 `BDD.AUTH.EMAIL.002` 固化，无需额外 sibling plan。
- `plan-code-review`：审查 completed plan 的后续替换阶段时，可增加“反向扫描早期 checked item 是否仍描述被替代实现”的语义检查；优先级 medium。
- `spec-plan`：若后续再替换 adapter/SDK/transport，应在同一 revision 中搜索 `standard library`、`vendor SDK`、旧 transport 名称等历史约束；优先级 medium。

## 5 建议优先级与后续动作

- 最高优先级：合并前保留当前 localized MIME parser 回归和 A3 SDK import-boundary lint，避免这两个合同再次分离。
- 后续可在下一次 `/plan-code-review` 自身治理迭代中评估 completed-plan historical item sweep；本次修复不需要继续扩大产品或运行时范围。
