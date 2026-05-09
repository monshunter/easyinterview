# Frontend Home / Job Picks / Parse History

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-05-09

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-09 | 1.3 | plan `002-jd-match-recommendations` 实施前用户决策修订：D-12 由「简化为单一动态加载文案 + 若 ui-design 仍是 5 步 panel 则 STOP」改为「以 ui-design 当前 5 步 AGENT panel 形态为真理源 + 前端 i18n 5 step key 源级复刻 + 动态 JD 数字（`248`/`87`/`unique postings`）替换为不含数字的静态文案 + 保留 opacity 渐变作为唯一动画效果」；C-14 验收场景同步从「单一加载文案」改为「5 步 AGENT panel + opacity + accent label + 动态数字 0 命中负向断言」；plan 002 不修改 ui-design 静态文件，未来 ui-design owner 任何 SearchTab `searching` 形态修订由后续 plan 接力 | 002-jd-match-recommendations |
| 2026-05-09 | 1.2 | 启动 plan `002-jd-match-recommendations`：§2.1 In Scope 中 jd_match 屏从「P1 placeholder shell」升级为「plan 001 placeholder + plan 002 完整三 tab 业务」；§2.2 Out of Scope 改写为剥离真实 backend handler / agent scan / 真实联网搜索 / 候选池抓取 / market signals 计算，明确 `backend-jobs-recommendations` 为后续承接 subspec；§3 D-1 改写为「契约先行 + frontend fixture 消费」模式；§3 新增 D-8（jd_match → parse 反向数据流仅锁定出口）/ D-9（Watchlist + Saved Searches 服务端持久化）/ D-10（Agent scan 状态来源 polling）/ D-11（jd_match 隐私红线扩展）/ D-12（Search loading 单一加载文案）；§5 模块边界拆为 frontend 三 tab + 未来 backend；§6 新增 C-12（Recommended + Profile chip + AGENT 状态完整渲染）/ C-13（JobMatchCard 详情 + Save/Mark not relevant/Confirm/Open source 闭环）/ C-14（Search tab 自然语言搜索 + savedSearches + filter + 单一加载文案 + failure）/ C-15（jd_match Auth pending action）/ C-16（Watchlist + Market signals + chevron handoff）；§7 关联计划 002 状态 `保留编号` → `active starting 2026-05-09` | 002-jd-match-recommendations |
| 2026-05-08 | 1.1 | 与 plan 001 落地一并 fix-up：D-3 / D-4 / D-6 / D-7 锁定值与措辞细化；C-1～C-11 验收场景与 fixture variant / privacy 反查口径对齐 plan 001 实现证据 | 001-home-jd-import-and-parse |
| 2026-05-08 | 1.0 | 初始创建：从 `frontend-shell/spec.md` §2.1 与 `engineering-roadmap/spec.md` 预占行派生新 subspec；定义 home / parse / jd_match 三屏 P0 范围、决策 D-1～D-7、设计约束、模块边界、acceptance criteria C-1～C-11；`jd_match` 完整 Recommended/Search/Watchlist 三 tab 显式 Out of Scope，等待 backend recommendations API 与对应 OpenAPI operationId 落地后由 plan `002-jd-match-recommendations` 承接 | 001-home-jd-import-and-parse |
