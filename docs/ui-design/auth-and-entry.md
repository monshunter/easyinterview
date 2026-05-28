# EasyInterview UI 认证与默认入口

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-05-28

## 1 文档目的

本文档定义当前静态 UI 中默认入口、操作级登录拦截和完整认证页面流。认证能力存在，但不作为进入产品的前置门槛。

## 2 已确认决策

1. App 默认进入首页。
2. 不再展示独立未登录欢迎页。
3. 用户未登录时，仍然可以看到首页并开始输入 JD。
4. 当用户执行需要身份、保存、同步、导出或敏感数据处理的动作时，产品目标是再触发登录；当前静态稿已把轻量 `requestAuth(pendingAction)` 接入 `立即面试`、`复练当前轮` 和 `进入下一轮`。
5. 当前静态稿中邮箱验证成功后先检查 `/me.profileCompletionRequired`；需要补全资料时进入 `auth_profile_setup`，完成后才恢复 `pendingAction`；没有待恢复动作时回到 `Home`。
6. 未登录顶部用户区只显示 `登录`，登录入口同时承担首次账号创建。
7. 已登录用户菜单显示 `用户画像`、`设置与隐私`、`退出登录`。
8. 认证页面包括单入口登录、邮箱验证、首次资料补全、重置登录和退出登录；旧 `auth_register` 不再是 live route。
9. 顶栏主题色、暗色模式和语言下拉对未登录用户也可见，不触发认证，不改变 `signedIn`。
10. Home 可未登录访问，但 Recent mock interviews 属于账号历史数据；未登录时不展示该模块，也不触发读取历史面试 / target job 的后端请求。
11. 面试相关业务 route（岗位推荐、JD 解析、工作台、简历、模拟面试、报告、复盘、用户画像、设置与隐私）在正式 runtime 中必须先确认登录；未登录直开或点击这些入口时统一进入 `auth_login(pendingAction)`。

## 3 默认入口

```text
打开 App
  -> Home
     ├─ 粘贴 JD
     ├─ 上传 JD / URL
     ├─ 查看最近模拟面试（仅已登录）
     ├─ 创建第一份简历
     ├─ 打开岗位推荐
     └─ 打开复盘
```

默认入口不检查 `signedIn` 后再决定是否进入 `welcome`。登录状态只影响哪些动作能继续、用户区展示什么入口。

## 4 顶部用户区与全局控制

```text
全局显示控制:
├─ 主题色 -> 暖陶 / 苔林 / 深海 / 梅子 / 自定义
├─ 暗色模式
└─ 语言下拉

未登录:
└─ 用户区
   └─ 登录 -> auth_login

已登录:
└─ 用户菜单
   ├─ 用户画像 -> profile
   ├─ 设置与隐私 -> settings
   └─ 退出登录 -> auth_logout
```

全局显示控制始终属于 TopBar，不属于认证流程。`signedIn` 只影响用户区展示 `登录` 还是头像菜单。

`用户画像` 与 `设置与隐私` 不应混用：画像是系统理解用户的结构化资料，设置是账号、登录安全和隐私控制。

## 5 认证页面流

```text
Auth
├─ auth_login
│  ├─ 邮箱
│  ├─ 发送 6 位验证码
│  ├─ 首次使用该邮箱会在验证后创建账号
│  ├─ 忘记密码 -> auth_reset
│  └─ 继续 -> auth_verify
├─ auth_verify
│  ├─ 输入 6 位邮箱验证码
│  ├─ 验证码 5 分钟内有效
│  ├─ 重新发送登录验证码
│  └─ 验证并继续；若 profileCompletionRequired=true -> auth_profile_setup
├─ auth_profile_setup
│  ├─ 显示姓名
│  ├─ 同意条款
│  └─ 完成资料后恢复 pendingAction 或回 Home
├─ auth_reset
│  ├─ 输入账号邮箱
│  ├─ 发送重置说明
│  └─ 返回登录
└─ auth_logout
   ├─ 确认退出
   ├─ 清除本机登录态
   ├─ 重新登录
   └─ 返回首页
```

退出登录只清除本机 session，不删除用户数据。

## 6 认证拦截模型

当前静态 UI 已实现独立认证页面、顶部用户区入口和轻量业务级 `requestAuth(pendingAction)`。运行时行为是：

```text
TopBar 登录
  -> auth_login
  -> auth_verify
  -> success
  -> profileCompletionRequired?
     ├─ true -> auth_profile_setup -> resume pendingAction 或 Home
     └─ false -> resume pendingAction 或 Home

UserMenu 退出登录
  -> auth_logout
  -> 清除本机登录态
  -> 重新登录 或 返回首页

业务动作:
  -> requestAuth(pendingAction)
     ├─ signedIn=true -> 执行原动作
     └─ signedIn=false -> auth_login(pendingAction)
        └─ success -> resume pendingAction.route(pendingAction.params)
```

后续接入保存、上传、查看更多历史数据等真实业务动作时，应继续沿用同一模型：

```text
UserAction
  -> requiresAuth(action)
     ├─ false -> continue
     └─ true
        -> AuthGate(pendingAction)
           ├─ success + profileCompletionRequired=false -> resume pendingAction
           ├─ success + profileCompletionRequired=true -> auth_profile_setup -> resume pendingAction
           ├─ cancel -> return source screen
           └─ error -> stay AuthGate with retry
```

`AuthGate` 可以是弹层或认证页，但它不是默认落地页，也不出现在顶部导航里。

## 7 动作级登录触发点

| 动作 | 是否触发登录 | 原因 |
|------|--------------|------|
| 打开首页 | 否 | 用户应直接开始 |
| 粘贴 JD 文本 | 否 | 只是输入草稿 |
| 打开岗位推荐列表 | 是 | 会读取/生成与用户简历、目标岗位和偏好相关的推荐 |
| 上传 JD 文件 | 是 | 文件上传到服务端并绑定用户上下文 |
| URL 导入 JD | 是 | 服务端抓取并保存目标岗位上下文 |
| 解析并保存岗位 / 面试规划 | 是 | 会创建用户上下文 |
| 打开已有模拟面试规划 | 是 | 读取用户历史数据 |
| 上传简历 | 是 | 涉及敏感个人资料 |
| 粘贴简历 | 是 | 涉及敏感个人资料 |
| 轻量简历问答 | 是 | 会保存个人经历信息 |
| 创建岗位定制简历版本 | 是 | 会基于用户简历和目标岗位生成新的个人资料版本 |
| 采纳 / 编辑简历改写建议 | 是 | 会修改当前简历版本内容 |
| 导出简历 PDF / 复制纯文本 | 是 | 涉及用户个人资料导出 |
| 更换当前规划绑定简历 | 是 | 会修改用户面试规划上下文 |
| 立即面试 | 是 | 会生成个人面试记录 |
| 结束并生成报告 | 是 | 会写入报告和证据 |
| 查看历史报告 | 是 | 会读取用户历史数据 |
| 复练当前轮 | 是 | 会创建新的面试 session |
| 进入下一轮 | 是 | 会直接创建并进入下一轮面试 session |
| 开始复盘 | 是 | 会保存真实面试问题、回答和反馈 |
| 开始复盘面试 | 是 | 会基于复盘创建新的面试 session |
| 打开用户画像 | 是 | 读取系统沉淀的个人结构化资料 |
| 打开设置与隐私 | 是 | 账户与隐私设置依赖身份 |
| 导出数据 | 是 | 涉及用户数据 |
| 删除数据 | 是 | 高风险账户操作 |
| 切换主题 / 暗色 / 语言 | 否 | 只改变本地显示偏好 |
| 切换字体预设 | 否 | 界面偏好，进入设置页本身仍需要身份 |

## 8 Pending Action

`Pending Action` 是动作级登录拦截的恢复对象。当前静态 UI 会在登录、邮箱验证和首次资料补全页展示待恢复动作。登录成功后必须先处理 `profileCompletionRequired`，确认资料补全完成后才恢复目标 route 和 params。

```text
pendingAction
├─ type
├─ label
├─ route
└─ params
```

### 8.1 成功路径

```text
Auth success
  -> if profileCompletionRequired
       -> auth_profile_setup
       -> complete profile
  -> read pendingAction.route
  -> restore pendingAction.params
  -> navigate target route
```

### 8.2 取消路径

```text
Auth cancel
  -> return sourceRoute(sourceParams)
  -> restore draftState
  -> show non-blocking status
```

取消登录不应清空用户已经输入的 JD、简历草稿或面试设置。

## 9 页面框架影响

### 9.1 删除的页面职责

```text
Welcome
├─ 产品欢迎
├─ 登录前门
└─ Start with a JD
```

以上职责不再由独立页面承担。

### 9.2 新职责归属

```text
Home
├─ 产品默认入口
├─ JD 输入
├─ 最近模拟面试（仅已登录）
└─ 简历创建入口

Auth Pages
├─ 登录
├─ 邮箱验证
├─ 首次资料补全
├─ 重置登录
└─ 退出登录
```

## 10 后续实现输入

1. `hideTopBar` 不应依赖 `welcome` 作为默认未登录态。
2. `signedIn` 不应决定是否能渲染 Home。
3. 登录成功必须先检查 `/me.profileCompletionRequired`；资料补全完成后才恢复 `pendingAction`，没有待恢复动作时才回 `Home`。
4. 认证 UI 需要明确“登录后继续刚才动作”的文案。
5. 接入真实业务保存 / 上传 / 查看历史数据时，所有需要登录的按钮必须走统一 `requestAuth(pendingAction)`，不能在各页面散落自定义跳转。
6. 退出登录页面必须说明账号数据不会被删除。
7. 主题色、暗色和语言下拉不得被绑定到认证状态；登录前后应保持同一套显示控制。
8. 旧 `auth_register`、注册按钮、注册页文案和 displayName-before-verify 入口不得作为 live UI 继续出现。
9. 未登录 Home 不展示 Recent mock interviews，也不调用历史面试 / target job API；任何受保护业务 route 在 auth loading 或 unauthenticated 状态下都不能先挂载业务 screen 再依赖后端 401 兜底。
