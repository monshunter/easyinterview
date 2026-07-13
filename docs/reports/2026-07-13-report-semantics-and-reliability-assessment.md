# Report Semantics and Reliability 交付复盘报告

> **日期**: 2026-07-13
> **审查人**: Codex

## 1 结论

本次报告模块按用户确认的最终口径完成验收：确定性的格式、枚举、长度、锚点与跨字段限制必须 100% 正确；模型生成内容不承诺 100% 语义正确，以五类固定代表场景达到约 80% 的合理置信度，并保留更严格诊断场景的真实失败。

- 最终 prompt 保留完整 JSON 示例输出；示例前增加配对 synthetic candidate input，后接 anti-copy/current-context regeneration 规则。没有以删除示例换取表面安全。
- P0.100 最终运行 `e2e-p0-100-20260713T101214Z-59381`：9/9 已生成最终输出全部通过机械格式与限制，语义 judge 8/9（88.9%），五类固定代表场景 4/5（80%）。
- 同一运行在第 9 次 `injection_resistant` 样本因 `unsupported_item`、`path=$.summary` fail-close；剩余两次和独立盲审未执行。因此严格的 11/11 P0.100 结果是 `FAIL`，不得表述为当前 PASS。
- P0.099 最终运行 `e2e-p0-099-20260713T095144Z-12381` 为 `PASS`：中文 `needs_practice`、英文 `well_prepared`、真实 `generating` 三种状态各有 desktop/mobile，共六张 full-page 截图；DB/API canonical content、报告/会话/冻结上下文 digest 与人工无 OCR 视觉审计完成绑定。
- 最终接受的中文、英文 ready 报告人工核对均未发现 unsupported fact、irrelevant advice 或 causal mismatch。另有一份早期英文输出虚构 “production incident”，已作为语义负样本计入诊断并废弃，没有进入验收报告。

## 2 多角度真实性与建议可靠性

| 角度 | 当前保障 | 结果 |
|------|----------|------|
| 输入事实 | 完成会话时冻结 JD、简历、轮次、计划与有序消息；表现判断只能引用候选人 user message seqNo | 结构闭环 |
| 描述真实性 | highlight/issue 必须有候选人消息锚点；summary 每个事实子句必须映射到输出证据；未回答的 assistant 追问不得变成候选人弱点 | runtime 机械校验 + judge/人工语义核对 |
| 准备度一致性 | 四档 preparedness 与 dimension/evidence/issue/action/focus 使用封闭跨字段规则；禁止隐藏数值换档 | 机械规则 100% |
| 建议可靠性 | retry 只能把已引用缺失行为转为重答动作；review 只能复核已引用证据；next 受 readiness 与 hasNextRound 限制；不得引入未引用的新机制、阈值、工具、框架或例子 | 当前真实样本 8/9 通过 |
| 语言与长度 | schema 200 code-point 仅作 malformed-output fuse；英文最多 24 whitespace words，中文最多 64 Unicode code points；targeted repair 内部目标 18/52 | 9/9 通过，UI 完整换行 |
| 注入与示例隔离 | 不可信上下文不能改 policy/schema；完整示例与 synthetic input 配对，并明确只能学习结构、不得复用事实 | 静态 gate 通过；严格 live 仍保留一条 summary 负样本 |

语义内容是概率性输出。当前结论是“达到约 80% 的合理置信度”，不是“保证零幻觉”。任何机械无效输出都无法成为 ready；语义偏差则通过真实样本、judge 与人工抽查暴露和计数，不在 runtime 增加脆弱文本启发式或第二个在线 judge。

## 3 用户动作级重试与恢复

- 一次用户 `GenerateReport` 动作拥有 1 次 initial + 最多 3 次 retry；只在 retryable provider/protocol failure 或输出无效时继续。
- 同一动作等待序列为 `10s/20s/40s`；第 4 次仍失败即结束本次动作。动作返回时内存计数销毁，用户重新操作从 0 开始。
- `feedback_reports` 不持久化产品 retry counter；`async_jobs.attempts/max_attempts` 只表示基础设施 lease/执行代次，不消耗用户动作额度。
- lease takeover 仍通过 job ID + claimed attempts fencing，防止 stale worker 写 report/outbox/audit/job 终态。
- P0.058 v3 证明 initial+3、精确等待、动作返回销毁、第二次动作重置、async/product attempt 分离与 non-retryable zero retry。

## 4 UI 与用户体验验收

- ready 页面直接消费后端持久化的报告与冻结 context，不从 URL、浏览器存储或当前可变 TargetJob/Resume 推导业务事实。
- required empty `retryFocusDimensionCodes` 固定编码为 `[]`，不再因 Go nil slice 输出 `null` 而阻断合法 dashboard（[BUG-0165](../bugs/BUG-0165.md)）。
- desktop 与 390px mobile 均展示准备度、维度、证据与行动；action label 完整可见，无省略、裁剪或横向溢出。
- generating 页面只展示真实生成中状态，不提前显示 ready 内容或成功暗示；桌面与 mobile 无重叠和横向溢出。
- 六张当前截图位于 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/screenshots/`，最终交付回复直接展示。

## 5 一致性、完备性与奥卡姆审计

- **一致性**：prompt/schema/hash/migration/resolved prompt/active DB 同源；OpenAPI、backend mapper/validator、frontend exact validator、scenario 与 owner docs 使用同一枚举、200/24/64 和 focus/action 合同。
- **完备性**：正常 ready、中文/英文、generating、invalid output、provider failure、动作级重试、第二动作重置、stale lease、empty array wire、desktop/mobile 均有可执行证据。
- **用户友好**：低频一次性报告使用 2C 短退避；行动文案受 24/64 可读上限而非只靠 200 fuse；失败不持久化为用户终身额度。
- **奥卡姆原则**：LLM 负责用户可见业务语义；后端只做可确定校验与有界恢复。不新增隐藏分数、runtime 文本规则、持久化产品 retry counter、在线第二 judge 或历史兼容层。

## 6 Bug 与验证证据

- [BUG-0164](../bugs/BUG-0164.md)：P0.058 verifier 把 Go subtest 计为根测试，已用精确根行计数修复。
- [BUG-0165](../bugs/BUG-0165.md)：合法空 focus 被 API mapper 编码为 `null`，已固定为 `[]` 并覆盖四状态 projection。
- [BUG-0166](../bugs/BUG-0166.md)：无配对输入的完整示例诱导事实复用；已保留完整示例并增加 paired input、anti-copy、hash/migration/lint gate。

已通过的主要门禁：backend 全量测试、`go vet`、相关 race；frontend 112 files / 795 tests 与 production build；prompt lint 24/24、8 files clean；offline eval 与 Promptfoo 28/28；OpenAPI 37 operations / 37 fixtures / 0 diff findings；migration lint 与 v19 up/down/up；P0.058；P0.099 当前 exact-six PASS。

严格 P0.100 当前为 FAIL，不列入“通过”清单。它作为高于当前约 80% 产品验收口径的稳定性诊断保留，后续只有在用户单独要求提高语义稳定性时再继续调优；本次不扩散新的报告特性。

## 7 收尾结论

本次实现满足最终确认范围：格式和限制正确，语义内容达到约 80% 的合理置信度，示例输出保留，前后端 UI 兼容，重试是用户动作会话级且使用短退避，真实六图闭环。下一步只需 review/merge 当前分支，不应在本目标内继续扩展功能。
