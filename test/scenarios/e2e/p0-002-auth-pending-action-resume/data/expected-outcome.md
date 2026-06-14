# Expected Outcome

- 点击 `立即面试` 后立即进入 `auth_login`；TopBar
  `topbar-user-area` `data-signed-in=false`。
- 输入邮箱并触发 `startAuthEmailChallenge` 后跳转 `auth_verify`，输入
  6 位登录 code 并触发 `verifyAuthEmailChallenge` 成功。
- 最终 `route-practice` 渲染，且 `data-route-params` 同时包含 `planId` /
  `targetJobId` / `jdId` / `resumeId` / `roundId` 五个原始值。
- 整个过程 vitest 报告 `1 passed`，且 trigger.log 不出现旧入口
  (`topbar-nav-mistakes` / `topbar-nav-growth` / `topbar-nav-drill` /
  `topbar-nav-voice` / `topbar-nav-welcome`) 或 prototype data 引用
  (`ui-design/src/data`)。
