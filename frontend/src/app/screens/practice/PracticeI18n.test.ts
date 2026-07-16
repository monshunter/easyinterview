import { describe, expect, it } from "vitest";

import { translate } from "../../i18n/messages";

describe("Practice recovery i18n", () => {
  it("keeps thinking, retry, and unresolved guidance localized in zh/en", () => {
    expect(translate("zh", "practice.transcript.thinking")).toBe("面试官正在思考");
    expect(translate("en", "practice.transcript.thinking")).toBe("The interviewer is thinking");
    expect(translate("zh", "practice.input.thinkingPlaceholder")).toBe("面试官正在思考…");
    expect(translate("en", "practice.input.thinkingPlaceholder")).toBe("The interviewer is thinking…");
    expect(translate("zh", "practice.message.retry")).toBe("重试这条消息");
    expect(translate("en", "practice.message.retry")).toBe("Retry message");
    expect(translate("zh", "practice.errors.completionFailed")).toBe("面试无法结束，报告也没有生成。请稍后再试。");
    expect(translate("en", "practice.errors.completionFailed")).toBe("We couldn't end the interview or create its report. Try again in a moment.");
    expect(translate("zh", "practice.errors.sessionLoadFailed")).toBe("面试内容加载失败，请稍后再试。");
    expect(translate("en", "practice.errors.sessionLoadFailed")).toBe("We couldn't load this interview. Try again in a moment.");
    expect(translate("zh", "practice.finishDisabled.unresolvedReply")).not.toBe(translate("en", "practice.finishDisabled.unresolvedReply"));
    expect(translate("zh", "practice.terminal.title")).toBe("面试官的回复中断了");
    expect(translate("en", "practice.terminal.title")).toBe("The interviewer's reply was interrupted");
    expect(translate("zh", "practice.terminal.description")).toBe("请返回当前面试规划，重新开始一场模拟面试。");
    expect(translate("en", "practice.terminal.description")).toBe("Return to this interview plan and start a new mock interview.");
    expect(translate("zh", "practice.terminal.backToPlan")).toBe("返回当前面试规划");
    expect(translate("en", "practice.terminal.backToPlan")).toBe("Return to this interview plan");
  });
});
