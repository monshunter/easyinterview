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
    expect(translate("zh", "practice.errors.completionFailed")).toBe("暂时无法结束面试并生成报告，请稍后重试。");
    expect(translate("en", "practice.errors.completionFailed")).toBe("We could not finish the interview and start the report. Please try again.");
    expect(translate("zh", "practice.errors.sessionLoadFailed")).toBe("面试会话暂时无法读取，请稍后重试。");
    expect(translate("en", "practice.errors.sessionLoadFailed")).toBe("The interview session is temporarily unavailable. Please try again.");
    expect(translate("zh", "practice.finishDisabled.unresolvedReply")).not.toBe(translate("en", "practice.finishDisabled.unresolvedReply"));
    expect(translate("zh", "practice.terminal.title")).toBe("本次回复未能完成。");
    expect(translate("en", "practice.terminal.title")).toBe("This reply could not be completed.");
    expect(translate("zh", "practice.terminal.description")).toBe("请返回当前面试规划，准备好后重新开始一场面试。");
    expect(translate("en", "practice.terminal.description")).toBe("Return to this interview plan, then start a new session when you are ready.");
    expect(translate("zh", "practice.terminal.backToPlan")).toBe("返回当前面试规划");
    expect(translate("en", "practice.terminal.backToPlan")).toBe("Return to this interview plan");
  });
});
