import { describe, expect, it } from "vitest";

import type { FeedbackReport } from "./generated/types";

describe("Reports generated contract", () => {
	it("FeedbackReport exposes failed-state errorCode", () => {
		const report: FeedbackReport = {
			id: "01918fa0-0000-7000-8000-00000000a001",
			sessionId: "01918fa0-0000-7000-8000-00000000a002",
			targetJobId: "01918fa0-0000-7000-8000-00000000a003",
			status: "failed",
			errorCode: "AI_PROVIDER_TIMEOUT",
			createdAt: "2026-05-15T00:00:00Z",
			updatedAt: "2026-05-15T00:00:00Z",
		};

		expect(report.errorCode).toBe("AI_PROVIDER_TIMEOUT");
	});
});
