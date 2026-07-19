# 核心工作流参考图对齐交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

**关联 Bug**: [BUG-0192](../bugs/BUG-0192.md)

**关联计划**:

- [Resume Create Flow](../spec/frontend-resume-workshop/plans/002-create-flow/plan.md)
- [Resume Listing and Readonly Detail](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)
- [Practice Text Event Loop](../spec/frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md)
- [Report Screen and Generating Handoff](../spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/plan.md)
- [Home + JD Import + Parse](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md)
- [App Shell, Auth Gate, and Settings Entrypoints](../spec/frontend-shell/plans/001-app-shell-auth-settings/plan.md)

## 1 复盘范围与成功证据

- 本次交付按三张参考图重构上传简历、面试进行和面试报告三页的导航下方起点、desktop 内容面、响应式网格、卡片层级、按钮、SVG icon、字体与间距，同时保持既有 API、路由、消息、简历注册和报告事实语义。
- 用户追加要求已闭环：AI/用户角色改为方形标识；说明胶囊带 sparkle icon；Transcript 成为唯一滚动区；Composer 整体固定在会话卡底部；说明胶囊固定贴在输入框上方 8px。
- 三张追加参考图也已闭环：简历预览采用 `1512/1310/1150px` Header/背景板/纸张构图；报告列表采用 `1372px` 插画 Header、事实摘要卡和编号时间线；面试记录采用同宽三列 Context Strip，并让 AI/“我”共用消息卡片与头像轮廓。
- 最后五张参考图也已闭环：面试规划详情采用 `1250px` Header 右侧动作与四层卡面；Settings 采用 `1372px` 插画 Header 和三张横向功能卡；登录、验证码、退出共享 `1450px` 双栏 Auth Shell、原则卡和主操作卡，验证码页不伪造倒计时或成功状态。
- 四张异步等待态参考图也已闭环：Practice、Resume、Report Generating、JD Parse 复用同一 shell-owned transition canvas 和四种语义 SVG；共享 TopBar 始终可见，Resume 全宽、Report 白卡为 1090px、JD 使用 1–4 编号步骤轴，所有 percent/内部 provider/伪阶段均被拒绝。
- TDD 证据：Resume create 4 files / 21 tests；Practice 最终 4 files / 56 tests；报告页追加 2 files / 23 tests 的目标构图 RED/GREEN，并由根级 frontend 131 files / 1055 tests 完整覆盖。报告页 RED 明确拒绝缺失 Detail icon 与 1432px 目标构图，随后转 GREEN。
- 根回归：Python 615 tests / 4615 subtests、Go 全包、frontend 133 files / 1066 tests 全部通过；三页 owner 32 files / 242 tests、规划/Auth/Settings shared visual/detail 95 tests、frontend typecheck 与 production build 通过。
- Chrome 证据：Resume 使用正式 real-mode frontend 完成 1916×821 / 390×844 containment；Practice 的 Transcript 在 desktop/mobile 实际滚动前后，input 坐标与 helper/input 8px gap 不变。报告页使用当前真实 backend ready report 完成 desktop 1920×964 与 exact mobile 390×844 full-page 验收：desktop 主体 1432px、四列 Context、两列 709px 内容卡、四个 46px Detail icon 和首屏完整 Overall 均闭合，两档 document overflow 均为 0。
- Chrome 追加证据：规划详情、Settings、登录、验证码、退出在 `1916×821` 与 `390×844` 均按目标层级渲染且无横向溢出；真实中英文切换、退出回 Home 与 protected Settings pendingAction 回跳通过。auth probe 中间 loading/error 瞬态未被捕获，因此相关旧 Phase 15.3 继续如实保持未完成。
- Chrome 等待态证据：在真实 frontend/backend 上触发多次 Practice 启动、Resume/JD 解析和 Report 生成；1920px viewport 下 Resume scene `x=0/width=1920`，Report card `x=415/width=1090`，JD 读取到 1–4 marker/state，所有流程都真实 handoff 到 ready 页面，browser error/warning=0。最终 focused 9 files / 89 tests、production build/redeploy、环境 4/4 与根 `make test` 615 / 4615 通过。
- 文档证据：Home/Parse owner 恢复 `completed`；Shell 视觉 Phase 完成但因独立 Phase 15.3 继续保持 `active`；两个 context、Header/INDEX、docs links 与 `git diff --check` 通过。

## 2 会话中的主要阻点/痛点

### 2.1 参考图一致不等于滚动行为一致

- **证据**：第一轮将说明胶囊放到 Transcript 最后一项，标准截图视觉接近参考图，但用户指出聊天增长后它会随记录移动；进一步要求输入框本身也必须始终贴底。
- **影响**：仅验收最终静态截图会漏掉内容增长后的相对定位，导致一次明确返工。

### 2.2 真实环境健康与完整视觉状态不是同一类证据

- **证据**：正式 real-mode frontend/backend 健康，但当前数据库没有可复用的 active session / ready report；Practice / Report 的完整视觉状态只能使用 repository fixture 验收。
- **影响**：若不显式区分，容易把 fixture 视觉验收错误描述成真实业务 E2E PASS。

### 2.3 通用 secrets scanner 被既有 Skill 文本阻断

- **证据**：`make lint` 唯一失败来自未改动的 `.agent-skills/change-intake/SKILL.md:120`，`generic-api-key` 规则把普通英文示例短语误判为密钥；其余 lint owner、Go lint 和 frontend lint 全部通过。
- **影响**：完整 lint 聚合无法给出绿色结果，需要额外拆分运行并解释现有误报。

### 2.4 报告页第一轮视觉 gate 允许“旧结构换皮”假绿

- **证据**：Phase 16 只锁定 1336px 宽度、背景、圆角和部分语义图标；用户再次指出报告页完全没有按目标稿改造，并用标框明确 Header CTA、单体 Context Strip 和贯穿各卡片的左侧语义图标轨。源码反查确认旧的四卡 Context 和无 Detail icon 结构仍在。
- **影响**：历史 focused/root/fixture 结果全部为绿，但没有回答目标图的组件结构问题，造成一次完整返工和用户二次纠正。

### 2.5 面试记录角色样式被错误设计为非对称层级

- **证据**：初次追加改造让 AI 消息直接显示在主体背景中，只给“我”增加独立 surface，并让“我”的头像轮廓与 AI 不同；用户明确要求两种消息与头像轮廓保持一致。
- **影响**：共同组件没有形成共同视觉合同，导致一次可避免的细节返工；若只看页面整体截图，仍可能漏掉角色间的 computed-style 差异。

### 2.6 历史行为绿灯掩盖 page composition 漂移

- **证据**：规划详情、Settings 与 Auth 的现有业务测试持续通过，但页面仍分别停留在 `1200px` inline detail、`980px` 纵向 Settings 和 `1160px` 窄 Auth Shell；只有新增 source/visual RED 才拒绝这些旧结构。
- **影响**：若只复跑已有行为回归，会把“动作仍能点击”误当作“页面已按批准稿改造”；不同页面族需要同时拥有业务 gate 和 page-scoped composition gate。

### 2.7 瞬态 auth gate 不能用最终登录页替代验收

- **证据**：Chrome 已验证登录、验证码、退出、Settings、真实语言切换和 pendingAction 回跳，但 5ms 尝试未捕获 auth probe loading/error 中间状态。
- **影响**：若为了恢复 owner `completed` 而用最终登录页代替瞬态 gate，会形成虚假证据；本轮选择完成视觉 Phase，同时保留 Phase 15.3 与 BDD 条目未勾选。

### 2.8 首轮共享等待态仍需要真实 Chrome 做结构校正

- **证据**：共享组件与 124 个 focused assertions 首轮全绿，但真实 Chrome 对照参考图仍发现 Resume 被详情容器限制在 1512px、Report 白卡只有 920px、JD current step 是整行 bordered card 且 marker 无编号。
- **影响**：如果在组件测试后直接收口，会再次留下“共享了代码但没有对齐参考稿”的假完成；本轮追加 RED 固定全宽 Resume、1090px Report card 和 JD marker 1–4，再重新部署验收。

## 3 根因归类

- 固定 Composer / 滚动 Transcript 最初没有成为可执行 DOM/CSS/BBox 不变量，只描述了视觉形态。
  - **类别**：spec/plan
- Chrome 验收入口缺少对“real-mode health”“fixture visual state”“real API/UI E2E”三类证据的统一命名，容易产生边界混淆。
  - **类别**：README
- secrets scanner 对治理 Skill 的普通示例文本产生稳定误报，属于 lint owner 与 Skill 文本的契约漂移，不是本次 UI 实现缺陷。
  - **类别**：skill
- Chrome finalize 参数的两次格式尝试没有影响产品结果，属于一次性工具调用失误。
  - **类别**：无需仓库改动
- 报告页视觉合同把 max-width/rounded/overflow 当成主要完成证据，没有将参考图标框转换为共享 surface、divider、icon rail 和首屏 Overall 的结构性 RED。
  - **类别**：spec/plan
- 面试记录 visual gate 只验证用户消息 surface，没有对 assistant/user 的边框、圆角、padding、阴影和头像几何做成对称断言。
  - **类别**：spec/plan
- 规划/Auth/Settings 只有历史行为 gate，缺少拒绝旧 max-width、Header/action 关系和卡片层级的 source/component contract。
  - **类别**：spec/plan
- Chrome 对最终稳定页面的验收无法证明短暂 auth probe gate；需要可控请求冻结或专门的 inter-request DOM gate。
  - **类别**：test/browser verification
- 共享组件抽取只保证实现一致，不能自动保证参考稿几何；缺少首轮真实 pending-state bbox 反馈时，父容器约束和内部步骤结构仍会逃逸。
  - **类别**：spec/plan

## 4 对流程资产的改进建议

- 对所有聊天类页面，owner spec/plan 必须显式写出“唯一滚动区、固定区、DOM ownership、短/长内容、滚动前后 bbox 不变量”，并把这几项同时落到 source/component 与 Chrome gate。
  - **落点**：相关 UI spec / plan template
  - **优先级**：high
- 在场景或前端验证说明中统一定义三档证据标签：`real-mode health`、`formal frontend fixture visual acceptance`、`real API/UI E2E`，结果中必须逐档声明，不允许相互替代。
  - **落点**：`test/scenarios/README.md` 或 frontend verification README
  - **优先级**：medium
- 由 secrets-config / skills owner 单独处理 `.agent-skills/change-intake/SKILL.md:120` 的稳定误报：优先重写示例短语或增加精确、最小 allowlist，并保留真实 generic-key negative fixture。
  - **落点**：change-intake Skill + secrets lint owner
  - **优先级**：medium
- UI 参考图修订必须先列出可观察的组件所有权与几何关系，再以 source/component RED 拒绝旧结构；宽度、背景、圆角和无溢出只能是辅助 gate，不能单独作为完成条件。本轮已在原 report spec/plan Phase 17 固化该规则。
  - **落点**：相关 UI spec / plan
  - **优先级**：high
- 聊天记录的角色 visual contract 必须同时断言“共享基类 surface”和“仅身份 token 不同”：两类卡片的 border/radius/padding/shadow 相同，两类头像的 width/height/radius 相同。本轮已在 report owner Phase 18 原地固化。
  - **落点**：ReportConversation owner spec / plan / source contract test
  - **优先级**：high
- 页面参考图对齐的 RED 必须覆盖内容列宽度、Header/action ownership、卡片层级、图标轨和 mobile containment；已有业务测试继续负责请求与状态机，两者不可互相替代。本轮已在 Home/Parse Phase 26 与 Shell Phase 18 原地固化。
  - **落点**：对应 UI spec / plan / visual tests
  - **优先级**：high
- auth probe 等短暂 UI 状态应通过可控请求冻结测试 inter-request DOM，再由 Chrome 只验证可稳定到达的真实页面；如果真实 Chrome 无法稳定捕获，必须保留未完成而不是用最终态代替。
  - **落点**：Shell Phase 15.3 / auth route gate verification
  - **优先级**：medium
- 共享异步组件的视觉 gate 必须同时覆盖“组件自身尺寸”与“嵌入各 caller 后的实际 viewport bbox”；至少用一轮真实 pending-state Chrome 检查全宽/居中尺寸、TopBar 与当前步骤，避免父容器把共享构图再次收窄。
  - **落点**：Shell transition plan + 各业务 owner BDD gate
  - **优先级**：high

## 5 建议优先级与后续动作

- 下一轮最高价值动作：把“参考图标框 → 组件所有权/几何关系 → source/component RED → real Chrome bbox”的闭环补入 UI plan 模板或前端验证规范，避免再次出现宽度和圆角已变、页面结构仍未改造的假绿。
- 同一优先级的后续项：把“滚动内容 + 固定操作区”的结构不变量补入聊天类 UI owner，覆盖短/长内容与滚动前后 bbox。
- 第二优先级：单独修复 secrets scanner 的现有治理文本误报，使根 `make lint` 恢复可直接作为聚合门禁使用。
- 第二优先级：为 Shell Phase 15.3 增加可控的 auth probe 请求冻结方式，捕获中文 loading/error 中间 DOM 后再恢复 owner `completed`。
- 下一轮同样应保留本次新固化的 transition caller-bbox gate；新增异步页面时必须先接入 shared scene，再在真实 pending 状态验收嵌入后的 viewport geometry。
- 可延后：为 fixture visual acceptance 增加更多长内容 fixture；当前通过压缩 desktop/mobile viewport 已真实产生 Transcript overflow 并验证了固定 Composer，不构成本次交付 blocker。
