# UI Design Report CTA Direct Session 交付复盘报告

> **日期**: 2026-05-02
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖报告页后续动作修正：`复练当前轮` 和 `进入下一轮` 都直接进入对应面试 session，不再回到面试前准备页；`返回面试前确认` 继续作为独立返回动作保留。

已通过的验证：

- `node --test ui-design/ui-design-contract.test.mjs`，14 项通过。
- `npx --yes esbuild ui-design/src/*.jsx --outdir=/tmp/easyinterview-ui-check --format=iife`，所有 JSX 入口解析通过。
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`，Header / INDEX zero drift。
- `python3 scripts/lint/check_md_links.py docs/ui-design`，通过。
- `git diff --check`，通过。
- `bash -n ui-design/run.sh`，退出码为 0，仅有本机 locale warning。
- bundled Playwright + 本机 Chrome 冒烟验证：报告页两个 CTA 分别进入 `PracticeScreen`，未回到 `当前面试规划`。
- 已建立回归记录 [BUG-0005](../bugs/BUG-0005.md)。

## 2 会话中的主要阻点/痛点

1. 报告 CTA 的产品语义和实现路由不一致。
   - **证据**：用户指出复练当前轮和进入下一轮都应直接进入面试；代码中两条路径此前都走 `workspace`。
   - **影响**：用户从报告继续练习时被迫回准备页二次确认，报告闭环被打断。

2. 文档中的旧口径放大了实现误判。
   - **证据**：`docs/ui-design` 仍出现 `Mock Interview Plan(same round)`、`创建下一轮面试规划`、`准备下一轮` 等描述。
   - **影响**：代码、测试和文档都需要同步改口径，不能只改按钮跳转。

3. 浏览器冒烟脚本需要注意静态原型的 hash 路由特性。
   - **证据**：首次脚本复用同一个 hash，React 内部路由没有重新挂载报告页；第二次脚本把静态资源 404 日志误判为失败。
   - **影响**：验证本身多跑了两次，但最终确认了两个独立 CTA 的真实页面行为。

## 3 根因归类

1. 报告页后续动作没有把 `return to setup` 和 `start interview session` 分成互斥行为。
   - **类别**：spec-plan

2. 契约测试此前只覆盖“路径文案区分”，没有覆盖目标 route、pending action route 和 sessionId 后缀语义。
   - **类别**：no repo change needed

3. 静态 UI 冒烟脚本缺少针对 hash-only React route 的稳定写法。
   - **类别**：no repo change needed

## 4 对流程资产的改进建议

1. 对报告、复盘、练习这类闭环 CTA，契约测试必须断言目标 route 和 action payload，而不只断言按钮文案。
   - **落点**：no repo change needed
   - **优先级**：high

2. UI 文档中涉及“下一步动作”的位置应明确区分返回、复练、进入下一轮三类动作。
   - **落点**：spec-plan
   - **优先级**：medium

3. 后续 Playwright 冒烟脚本对 hash route 使用独立页面或不同 hash 参数，避免复用同 URL 造成的路由状态残留。
   - **落点**：no repo change needed
   - **优先级**：low

## 5 建议优先级与后续动作

最高优先级是继续把关键 CTA 的“目标 route + payload 语义”纳入契约测试。本次已覆盖报告后续动作，后续如果扩展复盘或错题本入口，也应先写同类断言。

中等优先级是维护 `docs/ui-design` 的动作边界词汇，避免 `准备下一轮`、`进入下一轮`、`复练当前轮` 在不同文档中漂移。
