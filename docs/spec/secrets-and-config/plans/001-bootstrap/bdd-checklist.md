# Secrets and Config Runtime Content Limits BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

- [ ] BDD-13.1 / `E2E.P0.010`: TargetJob 96KiB 内主路径与 96KiB+1 拒绝/不污染路径通过。
- [ ] BDD-13.2 / `E2E.P0.046`: Practice 单条 32KiB、会话累计 256KiB 的 limit/limit+1、持久化一致性与继续对话路径通过。
- [ ] BDD-13.3 / `E2E.P0.081`: Resume upload 10MiB、paste/extracted 384KiB 的 limit/limit+1 与无半成品路径通过。
- [ ] BDD-13.4 / `E2E.P0.056`: Report 62,397-byte 样本、896KiB limit、limit+1、provider call/no-call 与 receipt 恢复路径通过。
- [ ] BDD-13.5: 四场景均记录当前环境、命令、时间与证据路径；历史 PASS 不作为完成依据。
