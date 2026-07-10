# Change Intake Owner Matching 交付复盘报告

> **日期**: 2026-07-10
> **审查人**: Codex

**关联计划**: [Core Loop Module Pruning](../spec/product-scope/plans/001-core-loop-module-pruning/plan.md)
**关联 Bug**: [BUG-0155](../bugs/BUG-0155.md)

## 1 复盘范围与成功证据

- 6.231 收敛 `change-intake` matcher 的重复词计分、action 词形、status tie-break 和精确 scenario Owner 证据，并补齐 Create Flow discovery vocabulary。
- Focused TDD 先得到旧 9 项通过、新 5 项失败；补充 stopword-only edge RED 后，最终 matcher/Skill contracts 15/15 通过。
- Python compile、Skill quick validation、Product/Create Flow context validation 通过。
- 两条真实 repo query 分别 high-confidence 命中 Create Flow 002（66 对 47）和 P0.037 README Owner（109，首条 reason 为 `scenarioOwner`）。
- `git diff --check` 与 pruning surface 通过，`real_residuals=0`。

## 2 会话中的主要阻点/痛点

- 同类误路由先后在 P0.037 和 Resume Create Flow 两次交付中出现，单次人工纠偏没有形成脚本修复。
  - **证据**：两份 2026-07-10 retrospective 都把 matcher owner precision 列为 high priority。
  - **影响**：high confidence 结果仍需人工反查，且存在重开错误 completed plan 的风险。
- 初始降噪实现遗漏了 stopword-only exact hit。
  - **证据**：自审新增 `and of` 回归时得到 1 fail / 9 pass；token 集虽为空，substring exact floor 仍给出 6 分。
  - **影响**：若未补该 edge，长 context 名中的连接词仍可能绕过停用词过滤。
- 精确 action 匹配既需要算法去重，也依赖 owner discovery metadata。
  - **证据**：只读实验去重后错误分差从 56 降到 5，但只有补入 Save-and-open / waiting-detail vocabulary 后 Create Flow 才稳定领先。
  - **影响**：仅调权重会把错误从确定性偏差变成脆弱近似，无法形成可维护 owner 证据。

## 3 根因归类

- Matcher 将 metadata value 数量当成独立证据，根因属于 `skill`；本批已改为字段内 unique-token 计分。
- Exact scenario ID 没有连接 README Owner，根因属于 `skill`；本批已增加确定性 owner edge。
- Create Flow action vocabulary 不完整，根因属于 `spec-plan` context discovery；本批已原地补齐 metadata，不改变完成态行为合同。
- Stopword exact edge 是本次实现自审发现的一次性缺口，已在同一 TDD 循环修复，无需额外流程资产。

## 4 对流程资产的改进建议

- 保留 hermetic matcher fixture + real repo query 双层门禁；新增 owner 证据类型时两层都必须覆盖。
  - **落点**：skill
  - **优先级**：high
- Scenario 数量显著增长前不增加缓存或索引；当前只在查询含精确 P0 ID 时扫描 README，实测完整查询约 0.1 秒。
  - **落点**：无需仓库改动
  - **优先级**：low
- 后续 context owner 修订出现新的用户 action 文案时，同步更新 discovery keywords，避免依赖通用 route/API 词补偿。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮技术债扫描回到脚本/测试资产，优先寻找可由现有 contract 证明的重复解析、无引用 helper 或固定数量脆弱断言。
- Matcher 当前不需要全文索引、向量模型或常驻缓存；只有新的实证误路由出现时再扩展 owner evidence 类型。
