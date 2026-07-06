# Frontend Home Resume Pre-bind 交付复盘报告

> **日期**: 2026-07-06
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 Home 新建模拟面试规划入口：删除冗余 hero sub，把主 CTA 从“解析并确认面试”改为“立即面试”，并要求用户在 Home 选择已有 ready 简历后才允许 paste / upload / URL import；成功进入 `parse` 时携带真实 `resumeId`。
- 同步范围覆盖 `ui-design/` 静态原型、`docs/ui-design/`、owner spec/plan/checklist/BDD、正式 frontend、Parse route resume 继承、P0.014/P0.015/P0.016 场景脚本和 BUG-0131。
- 成功证据：`HomeResumeSelection.test.tsx` 3 tests PASS，`ParseResumeBinding.test.tsx` 5 tests PASS，Home import/auth/recent/screen/i18n focused suites PASS。
- BDD 成功证据：`test/scenarios/e2e/p0-014-home-default-render/scripts/trigger.sh` + `verify.sh` PASS；`p0-015-jd-import-and-parse/scripts/trigger.sh` + `verify.sh` PASS；`p0-016-parse-confirm-to-workspace/scripts/trigger.sh` + `verify.sh` PASS。
- 收尾证据：`node ui-design/ui-design-contract.test.mjs` 28 tests PASS；`playwright test tests/pixel-parity/home.spec.ts` desktop/mobile 8 tests PASS；`sync-doc-index --check`、`make docs-check`、`git diff --check` PASS。

## 2 会话中的主要阻点/痛点

- Home 与 Parse 的职责边界需要重新落地到多个资产。
  - **证据**：用户指出 Home 不是单纯解析页入口，而是“模拟面试”新建规划快捷方式；修复必须同时改 UI 真理源、正式前端、Parse 继承、owner plan、BDD 和 scenario verify。
  - **影响**：如果只改按钮文案，会留下“立即面试但未绑定简历”的语义缺口。

- 新增 `listResumes` 后，既有 Home 单测的全局 fixture scenario 会误伤简历列表请求。
  - **证据**：`HomeResumeSelection.test.tsx` 首次 green 前失败于 `unknown fixture scenario manual-text-primary for operationId: listResumes`。
  - **影响**：测试需要区分“JD import fixture scenario”和“简历列表 fixture response”，否则会把测试工具限制误判为业务失败。

- 场景 verify 的旧文案负向搜索初次扫到了测试文件里的负向断言。
  - **证据**：P0.014 verify 首跑失败于 `HomeResumeSelection.test.tsx` 中的 `queryByText("解析并确认面试")`。
  - **影响**：负向源码搜索必须限定实现源码或排除测试文件；否则会惩罚必要的 regression assertion。

## 3 根因归类

- Home/Parse 职责边界漂移。
  - **类别**：spec-plan
  - **说明**：旧 plan 已在 Phase 7 修复 Parse 绑定简历，但 Home 入口仍按“先 import 再 parse 选择简历”的旧心智运行。

- Fixture scenario 选择粒度过粗。
  - **类别**：no repo change needed
  - **说明**：本次已在 Home focused tests 中显式 mock `listResumes`，让 import scenario 继续服务 `importTargetJob`，没有必要改全局 mock transport。

- 负向搜索范围过宽。
  - **类别**：spec-plan
  - **说明**：已在 P0.014/P0.015 `verify.sh` 中使用 `--exclude='*.test.ts' --exclude='*.test.tsx'`，保留测试负向断言，同时拒绝实现源码残留。

## 4 对流程资产的改进建议

- 将“CTA 语义升级必须核对业务前置实体”保留在当前 owner plan Phase 8 gate 中。
  - **落点**：spec-plan
  - **优先级**：high

- 后续若 Home 再新增依赖其他 API 的选择控件，focused tests 应默认隔离各 operation 的 fixture response，而不是复用全局 scenario。
  - **落点**：README
  - **优先级**：medium

- 场景 verify 中的旧口径负向搜索应默认排除测试文件，除非目标就是检查测试断言本身。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮围绕 Home/Parse/Workspace 的入口语义变更时，先在 owner plan 中写明“入口 CTA、必需业务实体、route params、BDD 负向 marker”四件套，再实施。
- 中优先级：如果同类 fixture scenario 误伤再次出现，再考虑在 mock transport helper 层支持 per-operation scenario；本次不需要额外抽象。
- 可延后：把“负向搜索排除测试文件”的约定提炼到场景 README 或共享 verify helper；当前 P0.014/P0.015 已局部修复。
