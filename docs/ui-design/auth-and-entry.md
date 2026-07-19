# 认证与默认入口

> **版本**: 1.33
> **状态**: completed
> **更新日期**: 2026-07-20

## 1 文档目的

本文档定义当前默认入口、认证触发点、pendingAction 接续和账号设置边界。未登录时显示登录入口；已登录时不再显示账号 chip 或用户菜单，只显示一个由当前用户名首字符构成的圆形设置按钮。

## 2 已确认决策

1. App 默认进入 `home`。
2. 登录方式只有邮箱验证码。
3. 登录是操作级拦截，不是默认前置页。
4. 登录成功后若 `profileCompletionRequired=true`，进入 `auth_profile_setup`；完成账号资料补全后接续 pendingAction。
5. `auth_profile_setup` 是账号资料补全，不是用户画像。
6. 已登录 TopBar 账号区只显示一个带可访问名称、由当前 `displayName` 首字符构成的圆形设置按钮，直接进入 `settings`；它不是图片头像；退出登录迁入设置页。
7. `用户画像` 不是 TopBar 或设置页入口。
8. `复盘` 不是业务入口或登录触发点。
9. `/reports?targetJobId=...` 是受保护的规划上下文页面；鉴权接续只保留合法 `targetJobId`，但它不进入 TopBar 一级导航。
10. `/workspace?targetJobId=...` 是受保护的只读规划详情；`/workspace` 是列表。`/parse?targetJobId=...` 只允许接续已创建的新导入 queued/processing 命令，ready 后 replace 到 Workspace 详情。
11. 受保护 route 等待或无法确认登录状态时继续使用统一 auth gate；其 eyebrow、标题和说明全部跟随当前语言，不得在中文模式显示硬编码英文。

## 3 默认入口

```text
打开 App
└─ home
   ├─ 在唯一文本框粘贴 JD
   ├─ 最近模拟面试
   └─ 创建简历入口
```

## 4 TopBar 账号区

```text
未登录
└─ 登录

已登录
└─ 圆形用户名首字符设置按钮 -> settings
```

设置入口必须使用圆形用户 initial mark 与标准 button，具备 `aria-label="设置 / Settings"`、清晰 focus ring 和至少 40×40px 点击区域。initial mark 直接消费应用启动时已取得的 authenticated runtime `displayName`：先 trim，再取首个 Unicode 字符，拉丁字母按 locale 大写；名称为空时显示 `?`，不得回退为固定品牌字母。它不表示图片头像；desktop 与 mobile 都不得出现完整姓名、账号 dropdown 或退出动作。

TopBar 语言按钮必须在当前语言标签右侧展示 code-native SVG 向下 chevron。箭头使用独立的至少 `20×20px` 低对比度底板、`secondary` 以上前景色与清晰描边；菜单展开时旋转为向上，不能再使用难以辨认的 `9px` 文本符号。

## 5 认证页面

```text
auth_login
  -> auth_verify
     -> profileCompletionRequired?
        ├─ true -> auth_profile_setup -> continue pendingAction
        └─ false -> continue pendingAction

auth_logout
  -> home
```

范围外 `auth_register` 和 `auth_reset` 归一到 `auth_login`。

### 5.1 目标构图与视觉合同

- desktop 使用约 `1450px` 的居中宽幅双栏：左侧是当前认证阶段的 eyebrow、标题、说明与统一原则卡，右侧是唯一主操作卡；不能回退为窄居中表单或多张并列小卡。
- 登录页右侧展示邮箱输入、主 CTA 与辅助说明；验证码页保持同一 shell，展示已发送邮箱状态、六位验证码输入、主 CTA、重发入口和真实可用的帮助说明；退出页展示风险提示、全宽确认 CTA 与居中的返回入口。
- 登录与验证码页可使用仓库内 SVG/CSS 绘制的信封、验证码、盾牌和浅色几何装饰；装饰不承载业务事实、不影响阅读顺序，并在窄屏折叠或弱化。
- mobile 改为单列，先读标题和原则，再进入主操作卡；输入、CTA 与辅助入口保持完整宽度且页面不得产生横向溢出。
- 验证码倒计时、重发成功和登录状态必须来自真实运行状态；没有 producer 时不得为了贴图伪造计时器、成功提示或可点击能力。

## 6 Pending Action

可接续的 pendingAction 只覆盖当前核心业务动作：

| 动作 | 目标 route | 说明 |
|------|------------|------|
| 打开当前面试规划 | `workspace` | 从 ready Home/Workspace 卡片、Reports Back 或 Practice terminal recovery 发起；只携带 `targetJobId` |
| 查看新导入解析进度 | `parse` | 只接续已创建 TargetJob 的 queued/processing 命令；只携带 `targetJobId`，ready 后 replace 到 Workspace 详情 |
| 立即面试 | `practice` | 从 Workspace 详情或列表 quick-start 发起 |
| 打开当前规划报告 | `reports` | 从规划详情内容区入口或受保护深链发起；只携带 `targetJobId` |
| 复练当前轮 | `practice` | 从 report header 发起 |
| 进入下一轮 | `practice` | 从 report header 发起 |
| 保存简历 | `resume_versions` | 创建或编辑简历 |
| 打开设置 | `settings` | 已登录用户通过圆形用户名首字符设置按钮或受保护深链进入账号和隐私设置；route 不携带用户字段 |

不得创建 pendingAction 到 `debrief`、`debrief_full` 或 `profile`。

## 7 范围外入口归一

| 输入 | 处理 |
|------|------|
| `debrief` / `debrief_full` | 归一到 `home` |
| `profile` | 归一到 `home` |
| `auth_register` | 归一到 `auth_login` |
| `auth_reset` | 归一到 `auth_login` |

## 8 后续实现输入

1. TopBar 已登录账号区只能有从 runtime `displayName` 派生的圆形用户名首字符设置按钮；不得包含图片头像、完整姓名 chip、dropdown、退出按钮或 `用户画像`。语言按钮使用清晰 SVG chevron 表达展开状态。
2. 未登录保护路由不得把 `debrief` 或 `profile` 当业务目标。
3. 正式前端、URL fallback 和 scenario 都必须覆盖范围外入口负向。
4. 设置页为无 tab 单页，承载账号级主题偏好、真实账号字段、退出登录和当前可执行的隐私状态/动作；语言与暗色继续由 TopBar 控制，字体不再可配置。
5. `import_jd` pendingAction 只接续粘贴文本入口；路由只携带 opaque pending id 与业务 ID，不携带 JD 原文或导入类型。
6. `reports` pendingAction 只允许 `targetJobId`；`section`、report/status/round、原文和其他业务状态必须剔除。登录恢复后仍由受保护 API 校验规划归属。
7. `workspace` / `parse` pendingAction 只允许合法 `targetJobId`；不得接续 `planId`、`resumeId`、analysis status、动画步数或 auto-start 等业务事实。恢复后分别由受保护 API 判定只读详情或命令进度，ready Parse 必须 replace 到 Workspace detail。

## 9 设置页单页合同

```text
Settings
├─ Appearance
│  ├─ 一级 Theme: Ocean / Plum / Custom accent（始终可见）
│  ├─ 二级 Custom accent: hue + saturation（仅选择 Custom 后在一级下方展示）
│  └─ Save: 单次 updateMe，成功响应直接刷新 runtime
├─ Account
│  ├─ Display name: /me.displayName（只读）
│  ├─ Login email: /me.email（完整账号邮箱，只读）
│  └─ Sign out -> auth_logout
└─ Privacy & data
   ├─ Export data: 暂不可用（不可触发伪成功）
   └─ Delete account -> destructive confirmation
```

- `SettingsScreen` 复用应用启动或认证恢复时已取得的 authenticated user，不为页面挂载或 route 切换重复调用 `getMe`；loading/error/unauthenticated 仍由统一 route guard 处理。
- Appearance 的 Ocean / Plum / Custom accent 是始终可见的一级主题选择器。desktop 下一级选择器与“保存主题”共享固定主操作行，选项在左、保存动作在右；只有选择 Custom accent 时才在该行下方挂载色相与饱和度二级编辑器，展开或收起不得让保存动作上下跳动。二级编辑器必须进入正常文档流，不得覆盖、替换或遮挡主操作行。色相轨道使用完整光谱渐变，彩度轨道从当前色相的低彩中性色渐变到高彩色，使调节方向无需数值也可理解；thumb、键盘和 focus 仍沿用可访问 range 语义。选择 Ocean 或 Plum 后立即隐藏二级编辑器并退出自定义色，用户在任何状态下都能回退到预定义主题。
- 主题草稿在设置页本地预览；拖动 hue/chroma 不发网络请求。点击保存只发一次 `updateMe`，成功响应返回完整用户上下文并直接更新 runtime/display context，不追加 `getMe`。失败保留草稿和错误供重试，不覆盖最近一次服务端确认值；离开未保存页面恢复确认值。重新登录或其他平台启动后由首个 `getMe` 恢复同一账号主题。
- 不渲染 tab rail、手机号、界面语言行、时区、登录与安全、字体预设、产品信息、数据留存开关、数据概览、删除单次会话或删除所有练习数据等没有当前数据源/operation 的静态条目。
- 数据导出沿用 P0 `501 PRIVACY_EXPORT_NOT_AVAILABLE` 契约，默认显示禁用态和可读的“暂不可用”原因；不得展示为可触发按钮，也不得把静态文案或未发请求状态显示成已受理。
- 删除账号点击后打开二次确认；确认前无副作用。对话框使用 destructive title/description、初始焦点、焦点约束、Escape/取消和关闭后焦点归还；pending 时禁止关闭和重复提交。失败保留对话框、显示可恢复错误并允许重试；`401` 进入统一认证重探测，不继续展示可重试删除；成功收到 `202` 后调用现有 `refreshAuth()` 重探测 `/me`（预期 401），提交 unauthenticated 状态并 replace Home。一次确认生命周期内的网络重试复用同一 idempotency key，不新增第二套清 session 方法；重探测若遇网络/服务错误，诚实保留 auth error，不伪装成已确认退出。
- 退出登录继续进入既有 `auth_logout` 确认页，不在设置页复制第二套 logout mutation 或确认文案。
- desktop 设置页使用约 `1372px` 内容列：顶部左侧为 eyebrow、标题与说明，右侧为低对比度账号/安全插画；下方依次是外观、账号、隐私与数据三张横向功能卡。Header 插画使用仓库内 code-native SVG，主体必须是带顶部栏、头像与资料行的半透明账号窗口，并在窗口右下组合真实柱状图；左下前景是带锁孔的圆角锁卡，右下前景是带对勾的盾牌，两侧以小型四角星平衡留白。各层使用当前主题 accent 的透明度层级与柔和阴影，不得退化为山形折线、单一人物轮廓和独立圆形对勾的稀疏线稿；整组仅作装饰，保持 `aria-hidden`，不承载账号状态或隐私事实。
- 每张设置卡均使用统一的左侧图标轨道与内容区。外观卡在内容区内以固定主操作行共同承载一级主题选项和保存动作，并把条件 Custom 编辑器放在下方；账号卡保留真实账号字段与退出，隐私卡保留不可用导出说明和删除账号状态机；不得因视觉重排复制 mutation、增加静态伪功能或改变 owner。
- 窄屏将标题、插画、信息与操作按原 DOM 顺序叠放；所有按钮、radio、slider 和 destructive dialog 继续满足键盘、focus、状态和错误恢复合同。

## 10 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-20 | 1.33 | 固定 Settings 主题一级选项与保存动作的同一主行归属，自定义编辑器只在下方展开，避免切换 Custom 时保存按钮跳动。 |
| 2026-07-20 | 1.32 | 将 TopBar 设置入口从固定 `E` 改为 authenticated runtime 用户名首字符，并把语言菜单的弱文本箭头改为带状态反馈的清晰 SVG chevron；不新增账号请求。 |
| 2026-07-20 | 1.31 | 收紧 Settings Header 安全插画结构：以半透明资料窗口、头像信息、柱状图、前景锁、盾牌对勾和星芒取代稀疏人物线稿，并保持主题色与装饰语义。 |
| 2026-07-19 | 1.30 | 明确 Appearance 的 Ocean / Plum / Custom 一级选择器始终可见；Custom 色相/饱和度仅在选择后于一级下方按正常文档流展开，并以完整色相光谱和当前色相的彩度渐变表达调节方向。 |
| 2026-07-19 | 1.29 | 按五张参考稿锁定登录、验证码、退出的宽幅双栏认证构图，以及设置页 Header 插画与三张横向功能卡；不新增伪倒计时或业务操作。 |
| 2026-07-19 | 1.28 | 将已登录单一设置入口从齿轮视觉改为圆形 E initial mark；保持直达 Settings、无用户头像数据、无账号 dropdown。 |
| 2026-07-19 | 1.27 | 设置页更名为“设置”，新增账号级外观主题保存；锁定 bootstrap 单次读取、route 零重复读取和单次保存更新合同。 |
| 2026-07-16 | 1.26 | 明确 auth loading/error route gate 属于本地化 shell UI，中文模式不得出现英文硬编码。 |
| 2026-07-15 | 1.25 | 采用设置简化方案 A：已登录 TopBar 仅保留设置齿轮，设置页改为无 tab 的真实账号/隐私单页，并明确退出、导出不可用与账号删除状态机。 |
