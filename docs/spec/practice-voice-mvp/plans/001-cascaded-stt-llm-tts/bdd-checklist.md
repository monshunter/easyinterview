# Cascaded STT LLM TTS Voice MVP BDD Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-17

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.007 完整级联语音面试 turn

- [ ] 创建场景目录 `test/scenarios/e2e/p0-007-cascaded-voice-turn/`
- [ ] 准备测试数据：登录用户、practice session、active fixture STT/chat/TTS profiles、当前题目、可播放 TTS chunk metadata
- [ ] 实现 setup / trigger / verify / cleanup；trigger 写入真实 runner 日志，verify 断言 runner marker、目标测试路径、pass marker、transcript、assistant text、TTS playback state、voice turn event、profile meta 摘要和下一题入口
- [ ] 更新 `test/scenarios/e2e/INDEX.md` 并执行场景，记录验证证据

## E2E.P0.008 插话只提交已播放上下文

- [ ] 创建场景目录 `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/`
- [ ] 准备测试数据：multi-chunk TTS response、played chunk event、barge-in event、下一轮用户输入
- [ ] 实现 setup / trigger / verify / cleanup；trigger 写入真实 runner 日志，verify 断言 runner marker、目标测试路径、pass marker、未播放 draft 不在 committed context / prompt 中，interruption note 存在
- [ ] 更新 `test/scenarios/e2e/INDEX.md` 并执行场景，记录验证证据

## E2E.P0.009 Provider failure fallback

- [ ] 创建场景目录 `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/`
- [ ] 准备测试数据：STT secret missing、TTS provider error、unsupported realtime profile 三类 fixture
- [ ] 实现 setup / trigger / verify / cleanup；trigger 写入真实 runner 日志，verify 断言 runner marker、目标测试路径、pass marker、fail-fast / 文本 fallback / 不走 stub 或 realtime / privacy grep 无明文
- [ ] 更新 `test/scenarios/e2e/INDEX.md` 并执行场景，记录验证证据
