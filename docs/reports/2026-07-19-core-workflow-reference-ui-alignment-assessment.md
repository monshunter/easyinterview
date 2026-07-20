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
- [App Shell Visual System](../spec/frontend-shell/plans/002-app-shell-visual-system/plan.md)

## 1 复盘范围与成功证据

- 本次交付按三张参考图重构上传简历、面试进行和面试报告三页的导航下方起点、desktop 内容面、响应式网格、卡片层级、按钮、SVG icon、字体与间距，同时保持既有 API、路由、消息、简历注册和报告事实语义。
- 用户追加要求已闭环：AI/用户角色改为方形标识；说明胶囊带 sparkle icon；Transcript 成为唯一滚动区；Composer 整体固定在会话卡底部；说明胶囊固定贴在输入框上方 8px。
- 三张追加参考图也已闭环并按后续反馈收敛：简历预览采用约 `1512px` 内容面，删除 `1310px` 共享背景板，让 PDF 页面栈/约 `1150px` Markdown 纸张直接落在画布；报告列表采用 `1372px` 插画 Header、事实摘要卡和编号时间线；面试记录采用同宽三列 Context Strip，并让 AI/“我”共用消息卡片与头像轮廓。
- 最后五张参考图也已闭环：面试规划详情采用 `1250px` Header 右侧动作与四层卡面；Settings 采用 `1372px` 插画 Header 和三张横向功能卡；登录、验证码、退出共享 `1450px` 双栏 Auth Shell、原则卡和主操作卡，验证码页不伪造倒计时或成功状态。
- 四张异步等待态参考图也已闭环：Practice、Resume、Report Generating、JD Parse 复用同一 shell-owned transition canvas 和四种语义 SVG；共享 TopBar 始终可见，Resume 全宽、Report 白卡为 1090px、JD 使用 1–4 编号步骤轴，所有 percent/内部 provider/伪阶段均被拒绝。
- TDD 证据：Resume create 4 files / 21 tests；Practice 最终 4 files / 56 tests；报告页追加 2 files / 23 tests 的目标构图 RED/GREEN，并由根级 frontend 131 files / 1055 tests 完整覆盖。报告页 RED 明确拒绝缺失 Detail icon 与 1432px 目标构图，随后转 GREEN。
- 根回归：Python 615 tests / 4615 subtests、Go 全包、frontend 133 files / 1066 tests 全部通过；三页 owner 32 files / 242 tests、规划/Auth/Settings shared visual/detail 95 tests、frontend typecheck 与 production build 通过。
- Chrome 证据：Resume 使用正式 real-mode frontend 完成 1916×821 / 390×844 containment；Practice 的 Transcript 在 desktop/mobile 实际滚动前后，input 坐标与 helper/input 8px gap 不变。报告页使用当前真实 backend ready report 完成 desktop 1920×964 与 exact mobile 390×844 full-page 验收：desktop 主体 1432px、四列 Context、两列 709px 内容卡、四个 46px Detail icon 和首屏完整 Overall 均闭合，两档 document overflow 均为 0。
- Chrome 追加证据：规划详情、Settings、登录、验证码、退出在 `1916×821` 与 `390×844` 均按目标层级渲染且无横向溢出；真实中英文切换、退出回 Home 与 protected Settings pendingAction 回跳通过。auth probe 中间 loading/error 瞬态未被捕获，因此相关旧 Phase 15.3 继续如实保持未完成。
- Chrome 等待态证据：在真实 frontend/backend 上触发多次 Practice 启动、Resume/JD 解析和 Report 生成；1920px viewport 下 Resume scene `x=0/width=1920`，Report card `x=415/width=1090`，JD 读取到 1–4 marker/state，所有流程都真实 handoff 到 ready 页面，browser error/warning=0。最终 focused 9 files / 89 tests、production build/redeploy、环境 4/4 与根 `make test` 615 / 4615 通过。
- 2026-07-20 Settings 插画补充闭环：把旧山形人物稀疏线稿重画为资料窗口、头像信息、柱状图、锁、盾牌对勾与星芒组成的 7 层 code-native SVG；focused Settings visual/behavior 26/26、typecheck、production build、frontend redeploy、环境 4/4 和根 `make test`（Python 615 / 4615 subtests、Go 全包、frontend 134 files / 1082 tests）通过。Chrome 在真实 `1264×964` Settings 页面测得插画 `360×200`、Header/Appearance 同边界、`documentWidth=viewportWidth`，Ocean/Plum/Custom 主题预览和刷新恢复均正常。
- 2026-07-20 TopBar 补充闭环：语言入口改为带可见底板和展开旋转的 code-native SVG chevron；设置入口从 authenticated runtime `displayName` 派生首个 Unicode 字符，拉丁字母大写、空名称显示 `?`。Focused 43/43、typecheck/build、frontend redeploy、环境 4/4 与根 `make test`（Python 615 / 4615 subtests、Go 全包、frontend 134 files / 1087 tests）通过；真实 Chrome 验证 `星期无 → 星`、设置直达、无横向溢出及 0 warning/error。
- 2026-07-20 Home JD textarea 补充闭环：保留 `width: 100%`，默认 `min-height` 从 `106px` 增至 `212px`，长内容随最新 `scrollHeight` 自动增高并在删减后回缩。RED 先暴露静态尺寸与缺少重算，GREEN 后 Home 9 files / 67 tests、typecheck/build、frontend redeploy、环境 4/4 和根 `make test`（Python 615 / 4615 subtests、Go 全包、frontend 134 files / 1088 tests）通过；真实 Chrome 验证空/短内容 `212px`、36 行 JD `993px` 且 `clientHeight=scrollHeight`、无横向溢出及 0 warning/error。
- 2026-07-20 操作按钮圆角补充闭环：建立 `--ei-radius-control: 8px`，以 28 个 CSS selector + 10 个内联恢复 action 的显式 inventory 统一 TopBar/Auth/Home/Workspace/Parse/Practice/Reports/Report/Generating/Resume/Settings 有框操作，同时保留 circular/pill、无边框 link/back 与非按钮 surface。Focused 6 files / 62 tests、根 `make test`（Python 615 / 4615 subtests、Go 全包、frontend 全量）、typecheck/build、frontend redeploy、环境 4/4 通过；真实 Chrome 验证 Settings desktop/mobile、弹窗与跨页样本 computed `8px`、无横向溢出和 0 warning/error，证据不声明 E2E。
- 2026-07-20 报告返回文案补充闭环：为 trusted Reports destination 恢复“返回面试报告 / Back to interview reports”，Workspace fallback 与其它页面继续使用 shared Back；normal/error component、locale/source 和导航矩阵回归均通过。真实 Chrome 捕获短暂 Generating 文案，但页面约三秒后自动 handoff，未把未可靠执行的 live click 冒充 PASS。
- 2026-07-20 简历背景板补充闭环：删除空 `className` 与 `ei-resume-detail-preview-card` desktop/mobile/后代 CSS，把正向旧样式测试改为 selector 缺席 gate；Chrome 首轮发现透明 `article` 因失去宽度 owner 把短内容纸张缩到约 497px，follow-up RED/GREEN 以 `width: 100%` 恢复内部纸张 1150px，同时保持透明、无边框/圆角/阴影/padding。最终 focused 14/14、Resume owner 118/118、typecheck/build、根 `make test` 615 / 4615、redeploy 与 browser warning/error=0 通过。
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

### 2.9 插画存在性断言再次掩盖内容结构漂移

- **证据**：Settings Phase 18 已有 `.ei-settings-header-art` 存在性测试且整体 Chrome 构图通过，但 SVG 实际只包含山形折线、人物和圆形对勾；用户再次指出与目标资料窗口、图表、锁和盾牌构图差异巨大。新增 layer contract 后，旧实现按预期 RED。
- **影响**：只验证插画容器、宽高或颜色会让任意图形满足合同，无法防止“页面有插画但画错内容”的假绿；本轮需要再次原地收紧 UI design、spec、plan 与 BDD。

### 2.10 品牌标识被误用为账号身份

- **证据**：TopBar 左侧品牌已经固定展示 `E`，右侧设置入口仍重复固定 `E`；同时语言入口只使用 9px 文本 `▾`。用户明确要求右侧入口展示用户名首字符并增强箭头辨识度，旧 spec 还把固定 `E` 写成正式合同。
- **影响**：实现与文档同时一致却仍不符合当前产品语义；若不先修订 owner spec，代码层只能持续保护错误设计。

### 2.11 自然语言尺寸要求需要先翻译为轴向与 overflow 合同

- **证据**：用户要求 JD 输入区“默认宽度加倍并自动适配”，但正式页面的 textarea 已经横向 `width: 100%`；真实 bbox 为 textarea/外框 `1346/1348px`，继续横向翻倍会突破 `1916px` viewport。截图所指的实际体验缺口是 `106px` 纵向可视容量和长内容内部滚动。
- **影响**：如果机械修改 CSS width，会制造横向溢出且仍不能完整展示多行内容；本轮先把要求转译为 `min-height + scrollHeight + overflow` 可测合同，再以 RED/GREEN 与 Chrome 几何闭环。

### 2.12 页面私有圆角值让同类 action 持续分叉

- **证据**：Settings 退出/注销为 `2px`，Parse/Practice/Report/Generating 恢复动作和其他页面 action 又分别使用 `6px`、`7px` 或 surface small token；第一次尝试批量替换时还会同时命中 input、card icon 等非操作 surface，必须逐项反查并恢复。
- **影响**：缺少语义 token 与 consumer inventory 时，单页修复会继续制造新的局部半径；若改用全局 `button` selector，则会反向破坏 circular initial、pill toggle 和无边框 link/back。

### 2.13 共享等待组件仍可保留单消费者视觉分叉

- **证据**：报告生成已接入 `AsyncTransitionScene`，但 shared component 仍保留仅 Generating 使用的 `card` prop/class，desktop/mobile CSS 又为该分支增加白色 surface、边框、阴影与局部毛玻璃；用户按修订稿明确要求内容直接落在氛围画布。
- **影响**：仅证明“使用共享组件”仍会让旧视觉分支获得错误完成感；必须把无卡片要求转为 source/DOM/CSS 负向 gate，并在短暂真实 pending 窗口读取 computed style。

### 2.14 返回文案与导航目标被错误绑定

- **证据**：14 个正式二级/三级页面分别消费 20 个目标特定 action key，产生“返回首页/报告/简历工坊/面试规划”等标签；统一 `common.back` 后，65 files / 495 tests 证明原 target、push/replace、trusted-context 与 fail-closed 语义无需改变。
- **影响**：如果每个业务 owner 同时拥有动作标签与目标，文案会在 route 行为全部正确时持续漂移；共享 Shell 应只拥有统一标签，页面 owner 继续拥有目标和安全边界。

### 2.15 desktop 可接受的 Composer overlay 在 mobile 形成新退化

- **证据**：首轮修订把 send 绝对定位到新 input surface 右下角，并用 `126px` 右内边距避免文本覆盖；desktop Chrome containment 通过，但 `390×844` 截图显示正文被压成窄列，同时 textarea 全局 focus ring 与 surface `:focus-within` 形成双框。
- **影响**：如果只在实现后做单一 desktop 截图，视觉归属修复会制造新的 mobile 可读性问题。本轮依据同一 Chrome 反馈重开 GREEN，把 action 改为 surface 内非叠加底部区域，并补单一 focus owner gate。

### 2.16 文案一致性被误解为“任何页面都不得有例外”

- **证据**：二三级页面统一为 shared Back 后，Generating 的 trusted Reports destination 仍安全正确，但用户明确指出该页需要“返回面试报告”而不是过度抽象的“返回”；Workspace fallback 又不能跟随变成长文案。
- **影响**：如果把统一性实现成无条件单 key，会在 route 行为测试全绿时损失用户依赖的关键去向语义；修复必须把例外绑定 owner 已有的可信 destination，而不能恢复跨页面文案分叉。

### 2.17 删除视觉 wrapper 可能同时删除布局 owner

- **证据**：移除 `ei-resume-detail-preview-card` 后，第一轮 focused 测试只证明旧 CSS 缺席、透明外层成立；真实 Chrome 发现直接 flex child 按短内容 shrink-to-fit，`article` 与 Markdown 页都约 497px，未达到仍有效的 1150px 纸张合同。补透明 `width: 100%` 后，外层/纸张恢复为 1443/1150px。
- **影响**：只做 class/CSS 清理会在消除突兀背景的同时制造新的内容缩窄；必须把 presentation surface 和 structural sizing 分成两条可执行合同，并以短内容样本验收。

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
- Settings Header 视觉合同只有“有账号/安全插画”的抽象描述和容器存在性测试，没有目标对象集合、前后景和透明层级的可执行定义。
  - **类别**：spec/plan
- TopBar 把品牌字母和字体三角当作稳定 UI 合同，没有定义 authenticated runtime identity 派生、Unicode/空值边界和 SVG 开合状态。
  - **类别**：spec/plan
- Home textarea 只拥有静态 `106px` 高度断言，没有定义空/短/长/删减四态的轴向、回缩与 overflow 不变量。
  - **类别**：spec/plan
- 视觉系统缺少有框 action 专属语义 token 和跨页面 consumer/exception 清单，页面私有半径值可以在各自测试全绿时持续漂移。
  - **类别**：spec/plan
- shared transition 把统一画布与单消费者卡片分支放在同一组件，没有 caller inventory 和“无旧分支”负向 gate。
  - **类别**：spec/plan
- 返回动作标签由各页面 locale key 重复持有，缺少 Shell 统一文案与业务 target 分离的 ownership 合同。
  - **类别**：spec/plan
- Composer 合同只规定“右下角”，没有定义 input surface ownership、overlay/flow、窄屏文本宽度和 focus ring owner。
  - **类别**：spec/plan
- 返回文案治理把 shared default 错当成绝对无例外，没有规定例外必须绑定既有 trusted destination、同时覆盖 normal/error/fallback。
  - **类别**：spec/plan
- 简历背景板同时承担 presentation 与 flex sizing；旧测试只锁定 class 样式或 selector 缺席，没有在短内容下断言直接子项和 renderer page 的实际宽度。
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
- 装饰性 SVG 也必须把参考稿中的对象集合、前景/背景关系、主题色透明层与 `aria-hidden` 写成 layer/CSS contract；“插画节点存在”只能作为最弱 smoke，不能作为视觉完成 gate。本轮已在 Shell Phase 21 与 Settings UI design 原地固化。
  - **落点**：相关 UI design / spec / plan / visual test
  - **优先级**：high
- 共享 TopBar 的品牌标识与账号身份必须分开建模：设置入口只消费现有 authenticated runtime 的最小身份投影，并覆盖中文、拉丁、前导空白和空值；语言开合提示使用 code-native SVG、可见底板和 transform 状态。本轮已在 Shell Phase 22 原地固化。
  - **落点**：Shell UI design / spec / plan / TopBar tests
  - **优先级**：high
- 对 textarea / editor 的“尺寸加倍、自动适配”需求，UI owner 必须显式拆成 `width` containment、`min-height`、内容驱动 height、缩短回收、overflow 与 mobile no-overflow 六个断言；本轮已在 Home Phase 28 和 `BDD.HOME.JD.TEXTAREA.006` 原地固化。
  - **落点**：相关 UI design / spec / plan / component test / Chrome bbox gate
  - **优先级**：high
- 跨页面有框 action 应共享一个语义 control radius；source contract 必须显式枚举 consumer，并把 circular、pill、borderless 与非按钮 surface 固化为负向例外，禁止用全局 `button` selector 掩盖分类缺失。本轮已在 Shell Visual Phase 23 原地固化。
  - **落点**：Shell UI design / spec / visual plan / token source contract
  - **优先级**：high
- 共享异步场景的视觉 variant 必须维护 caller inventory；修订稿删除某一构图时，source test 同时拒绝旧 prop/class/testid 和 desktop/mobile CSS，真实 Chrome 在 pending 窗口补 computed-style 证据。本轮已在 Report Phase 20 原地固化。
  - **落点**：Report/Shell transition spec / plan / source contract / Chrome gate
  - **优先级**：high
- 返回动作应采用“shared default + owner-scoped trusted exception + page-owned target”合同：默认 locale key 定义“返回 / Back”，确需更具体文案时只能绑定 owner 已有的可信 destination，同时覆盖 normal/error/fallback 并保留 route、history 与 fail-closed tests。本轮已在 Shell Phase 25 与 Report Phase 21 原地固化唯一例外。
  - **落点**：Shell UI architecture / spec / plan / locale source contract
  - **优先级**：high
- Composer/editor 的附属 CTA 必须同时定义“所属边界 + 是否与内容叠加 + exact mobile 内容宽度 + focus owner”；第一轮真实 Chrome 若暴露新几何退化，应回写同一 phase 后再收口。本轮已在 Practice Phase 15 原地固化非叠加 action area。
  - **落点**：Practice UI design / spec / plan / visual tests / Chrome bbox gate
  - **优先级**：high
- 删除 visual wrapper/class 时，owner gate 必须同时覆盖旧 selector 零残留、renderer 正向存在、透明外层和短内容 geometry；如果 wrapper 曾承担尺寸，应用无 presentation 的结构规则接管，不得以恢复背景板修复 shrink-to-fit。
  - **落点**：对应 UI design / spec / plan / source parity / Chrome bbox gate
  - **优先级**：high

## 5 建议优先级与后续动作

- 下一轮最高价值动作：把“参考图标框 → 组件所有权/几何关系 → source/component RED → real Chrome bbox”的闭环补入 UI plan 模板或前端验证规范，避免再次出现宽度和圆角已变、页面结构仍未改造的假绿。
- 同一优先级的后续项：把“滚动内容 + 固定操作区”的结构不变量补入聊天类 UI owner，覆盖短/长内容与滚动前后 bbox。
- 第二优先级：单独修复 secrets scanner 的现有治理文本误报，使根 `make lint` 恢复可直接作为聚合门禁使用。
- 第二优先级：为 Shell Phase 15.3 增加可控的 auth probe 请求冻结方式，捕获中文 loading/error 中间 DOM 后再恢复 owner `completed`。
- 下一轮同样应保留本次新固化的 transition caller-bbox gate；新增异步页面时必须先接入 shared scene，再在真实 pending 状态验收嵌入后的 viewport geometry。
- 新增或重画装饰插画时，先列出必须出现与明确禁止的图层，再写 component/CSS RED，最后用当前主题下的真实 Chrome 截图和 bbox 验收；不得再次以单个 SVG selector 存在作为完成依据。
- 共享导航下一轮继续优先审计“品牌、导航、账号、语言”四类语义是否被同一视觉符号混用；身份入口必须来自当前 runtime，语言入口必须有独立开合状态，且不得以新增账号请求换取显示信息。
- 后续涉及 textarea、Markdown editor 或会话输入区尺寸调整时，先复用本轮的六项尺寸合同，不再只断言单个 `min-height`；Chrome 必须同时记录 `clientHeight/scrollHeight` 和 `documentWidth/viewportWidth`。
- 后续新增有框 action 时，优先消费 `--ei-radius-control` 并更新 inventory test；只有 circular、pill 或明确无边框 action 才能进入例外清单，避免页面私有 `border-radius` 再次回流。
- 后续修改共享 transition variant 时，先更新 caller inventory；没有剩余消费者的 prop/class/CSS 必须同步删除，并在真实 pending 窗口保存 desktop/mobile computed-style 与截图证据。
- 后续新增二级/三级页面时，返回控件默认消费 `common.back`；只有用户确认的关键去向才允许 owner-scoped key，并必须由 trusted destination 分支、normal/error/fallback tests 与 source inventory 共同约束。
- 后续修改 Composer/editor 内的发送、保存或格式化动作时，优先使用同一 surface 内的非叠加 action area；必须同时跑 normal desktop 与 exact `390×844`，测量 action containment、内容宽度、overlap 和单一 focus ring。
- 后续删除卡片、背景或 wrapper class 时，先列出它当前承担的 presentation 与 geometry 职责；Chrome 必须使用短内容样本同时测量透明外层和内部页面宽度，避免“旧 selector 已清零”掩盖 shrink-to-fit。
- 可延后：为 fixture visual acceptance 增加更多长内容 fixture；当前通过压缩 desktop/mobile viewport 已真实产生 Transcript overflow 并验证了固定 Composer，不构成本次交付 blocker。
