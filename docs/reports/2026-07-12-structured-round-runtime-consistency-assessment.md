# Structured Round Runtime Consistency 交付复盘报告

> **日期**: 2026-07-12
> **审查人**: Codex

**关联计划**: [Workspace and Interview Context](../spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md)
**关联 Bug**: [BUG-0161](../bugs/BUG-0161.md)

## 1 复盘范围与成功证据

- 本次修复让 TargetJob 结构化轮次成为 PracticePlan 时长、Practice Top Bar 总时长和 Report 下一轮推进的共同真理源。
- 共享 start-practice 使用当前轮 `durationMinutes` 创建 plan，并拒绝复用预算不匹配的旧 plan；Practice 读取持久化 plan budget，失败时显示未知值而非固定时长。
- Report 只推进到 ordered rounds 的 immediate successor；末轮、未知、重复 ID、loading/failure fail closed，任一 CTA in-flight 时锁定两个入口。
- Frontend 全量 112 个测试文件、727 个测试、typecheck 与 production build 通过；UI contract 45/45、Practice/Report pixel parity 16/16 通过。
- E2E.P0.021、E2E.P0.045、E2E.P0.057 均按 setup → trigger → verify → cleanup 串行通过；owner context、文档索引、链接、diff 和旧固定值负向搜索通过。

## 2 会话中的主要阻点/痛点

- 结构化轮次此前已在 `BUG-0148` 中完成建模，但 runtime downstream consumers 仍保留三套独立默认值。
  - **证据**：shared start 固定 30 分钟、Practice Top Bar 固定 25 分钟、Report handoff 固定 ladder/default；新 RED tests 分别复现。
  - **影响**：已有“structured round 已完成”的历史结论无法证明 plan snapshot、练习显示和报告推进真正同源。
- 场景资产存在旧实现口径，首次加强 P0.057 negative gate 时发现正式 Report README 仍引用 `inferNextRoundId`。
  - **证据**：P0.057 runner 34/34 通过后，verify 的旧符号负向搜索仍失败；修订 README 后场景完整通过。
  - **影响**：只运行行为测试会留下文档级错误指引，后续维护者可能重新引入固定 ladder。
- 新增 P0.045 verifier 时，Node TAP 不输出测试文件名，且首个 source marker 与真实测试断言写法不一致。
  - **证据**：runner 全绿但 verify 两次拒绝；增加明确 runner marker并绑定真实断言锚点后通过。
  - **影响**：场景证据链需要额外一轮收敛，但没有造成实现误判。
- 外层场景编排曾使用 zsh 只读变量 `status`，最终 gate 首次使用了过期 validator 路径。
  - **证据**：P0.021 runner 24/24 后外层赋值报错；context gate 报脚本不存在，均立即按当前仓库入口重跑。
  - **影响**：属于一次性执行噪声，没有改变产品代码或测试结论。

## 3 根因归类

- Structured round consumer coverage 不完整
  - **类别**：spec/plan
  - 原计划证明结构化 round mapper 进入 Parse/Home/navigation，但没有建立 create plan、plan reuse、Practice display、Report successor 的跨消费者矩阵。
- Report README 与场景实现口径漂移
  - **类别**：README / spec-plan
  - 行为测试覆盖 direct start，却没有把固定 ladder 符号和 fallback 作为 negative contract。
- P0.045 runner evidence marker 不稳定
  - **类别**：README
  - 场景规范要求 runner marker，但 Node TAP 的默认输出无法提供文件名；脚本需要显式 marker。
- zsh 只读变量与 validator 路径错误
  - **类别**：无需仓库改动
  - 都是单次命令编排错误，当前 Skill 已声明正确路径，且没有重复成为系统性阻塞。

## 4 对流程资产的改进建议

- 对“结构化业务对象替代固定假设”的 owner plan，要求列出 producer、shared resolver、persistence snapshot、每个 UI/route/report consumer 和 failure boundary，逐项提供正向及旧 fallback 负向 gate。
  - **落点**：spec-plan（后续 `frontend-workspace-and-practice` 或同类 plan checklist）
  - **优先级**：high
- 场景 verifier 若依赖 runner 身份，应由 trigger 显式输出稳定 marker，不依赖具体 test runner 是否回显文件名；source marker 应绑定语义断言而非对象字面量格式。
  - **落点**：`test/scenarios/README.md`
  - **优先级**：medium
- 将 fixed round ladder、default next-round fallback 和 fixed practice budget 保持为 P0.045/P0.057 的 active negative gates。
  - **落点**：spec-plan / scenario assets
  - **优先级**：high

## 5 建议优先级与后续动作

1. 最高优先级是保留 P0.045/P0.057 的预算与 successor 负向 gate；任何 Practice/Report handoff 调整都必须同时运行这两条场景。
2. 下一次引入结构化业务对象时，在设计阶段先建立完整 consumer matrix，避免“schema 与首批页面已同源、后续运行时仍默认”的假完成。
3. 可在后续独立治理任务中把稳定 runner marker 规则补入 `test/scenarios/README.md`；本次场景脚本已局部采用该做法，无需阻塞当前交付。
