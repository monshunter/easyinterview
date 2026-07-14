# Secrets and Config Runtime Content Limits BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

复用现有端到端场景证明方案 A 的内容大小缺省不会再把正常材料误判为“材料与对话过长”，同时在真实边界外给出明确、可恢复的错误。场景只验证用户可感知合同；typed config、provider response body、profile fallback 与 zero-reference 由主 checklist 的 TDD/contract gates 承接。

## 2 场景映射

| BDD-Gate | 场景 | 主路径 | 边界/失败路径 | owner handoff |
|----------|------|--------|---------------|---------------|
| BDD-13.1 | `E2E.P0.010` | 低于 96KiB 的真实 JD 可保存并进入解析 | 96KiB+1 拒绝，刷新后既有 TargetJob 不被污染 | backend-targetjob + frontend-home |
| BDD-13.2 | `E2E.P0.046` | 单条低于 32KiB 且会话累计低于 256KiB 可继续对话 | 单条或累计 limit+1 拒绝，已持久化消息保持一致 | backend/frontend Practice |
| BDD-13.3 | `E2E.P0.081` | 10MiB 内上传与 384KiB 内粘贴/提取可完成 Resume intake | upload/paste limit+1 拒绝且不创建半成品资产 | backend-upload/resume + frontend-resume-workshop |
| BDD-13.4 | `E2E.P0.056` | 62,397-byte 真实失败样本及 896KiB 内 framed input 可进入 report provider | 896KiB+1 返回报告不可用 receipt，provider 不被调用且可回到面试报告 | backend-review + A3 |

## 3 执行约束

- 场景数据按 UTF-8 bytes 构造，必须同时覆盖 limit 与 limit+1；不得用字符数近似 bytes。
- 前端公开边界从 `runtime-config.contentLimits` 读取；场景不得把默认值复制成另一套产品真理源，fixture 只用于构造边界输入。
- Report 场景验证 provider call/no-call 与 receipt 恢复路径，不要求浏览器展示内部容量数值。
- 复用现有场景 ID，不新增 sibling 场景目录；若现有脚本缺断言，原地扩展对应场景资产并同步 INDEX。

## 4 完成标准

- 四个 BDD-Gate 均有当前运行证据。
- 正常材料、边界值、越界值与恢复路径均被覆盖。
- API、持久化与前端提示在刷新/重试后保持一致，无静默截断或半成品业务数据。
