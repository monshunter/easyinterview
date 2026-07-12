# E2E.P0.022 Practice Plan Baseline Create And Read

> **场景 ID**: E2E.P0.022
> **执行方式**: automated
> **隔离级别**: real PostgreSQL integration + focused generated handler tests
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

已登录用户拥有 `targetJobId`、`resumeId`、结构化 2..5 轮 TargetJob 与 baseline practice plan 输入；相邻轮次可具有相同时长，sequence 可不连续。场景使用真实 PostgreSQL 验证 normalized round identity，不依赖真实 provider。

## 2 When

执行 generated handler/domain/store create/read focused gates，并在真实 PostgreSQL 创建 baseline plan、完成第一轮、创建等时长下一轮和 sequence 跳号的紧邻 successor；同时读取 session start 的真实 round name/type/focus。

## 3 Then

新 plan 与 API response 成对保存 `roundId/roundSequence`；客户端不提交 sequence；等时长轮次不会误复用；sequence `1,2,4` 的下一轮是 `4`；session prompt context 使用真实 round name/type/focus，不用 persona 冒充；错误路径不插入 plan。

## 4 执行

```bash
./test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/scripts/setup.sh
./test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/scripts/trigger.sh
./test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/scripts/verify.sh
./test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/scripts/cleanup.sh
```

## 5 污染控制

integration test 使用固定测试用户并在前后删除该用户及关联 audit 证据，不清理其他业务数据。`cleanup.sh` 只清理 setup marker，保留 trigger log 作为 BDD evidence。
