# E2E Scenarios P0 History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-24

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-24 | 1.1 | L1 plan-review 修订：校正 P0 场景实施前基线为 87 条切片场景（最高编号 `E2E.P0.097`），将 operation matrix 口径统一为 9 行（8 个主链必经 operation + `getJob` 备选轮询 / handler gate），明确 Playwright 全栈必须用 `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL` 指向真后端，并把 legacy-negative 加固为 route-aware 旧 route / 独立 voice / `mode=debrief` 反查且避免误伤合法 `createPracticePlan` / `resumeAssetId`。 | 001-full-funnel-happy-journey |
| 2026-05-24 | 1.0 | 初始创建：定义 P0 完整漏斗端到端 journey owner subject；锁定 D-1~D-7（真后端全栈 + stub AI + happy 主干 + 两种 driver + 接续编号）；派生 `001-full-funnel-happy-journey`（`E2E.P0.098` API-level + `E2E.P0.099` Playwright 全栈） | 001-full-funnel-happy-journey |
