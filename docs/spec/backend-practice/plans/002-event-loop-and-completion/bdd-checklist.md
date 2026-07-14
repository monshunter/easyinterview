# Practice Event Loop and Completion BDD Checklist

> **版本**: 2.12
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 静态资产审计

- [x] `E2E.P0.098` 的 BDD 合同只描述真实登录、completion API、Home/Workspace/TargetJob progress refresh 与 TargetJob detail read。
- [x] chat、session start、opening message、quick-start 与下一轮 plan 创建未归入该 E2E。
- [x] event loop、retry 与 provider 失败的用户可观察行为明确为当前无真实 E2E owner 的合同。
- [x] 前后端代码层回归由仓库根 `make test` 独立承接，不作为 E2E 证据。

## `BDD.PRACTICE.EVENT_LOOP.001` 对话与完成

- [x] Owner behavior tests 覆盖消息顺序、same-ID retry、completion 原子性、replay 与失败零重复副作用。
- [x] 根 `make test` 已执行对应 Go tests；该结果不声明 `E2E.P0.098` PASS。

## 真实环境证据边界

本 checklist 只完成 owner 关联与静态资产审计；本轮未执行 `E2E.P0.098`，当前真实环境结果以场景 INDEX 的 `Ready` 为准，后续只由显式 `/scenario-run` 产生。
