import { describe, expect, it, vi } from "vitest";

import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getPracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/getPracticeSession.json";
import getReportConversationFixture from "../../../openapi/fixtures/Reports/getReportConversation.json";
import listTargetJobsFixture from "../../../openapi/fixtures/TargetJobs/listTargetJobs.json";
import { EasyInterviewClient } from "./generated/client";
import type { PaginatedTargetJob, PracticeSession, RuntimeConfig, UserContext } from "./generated/types";
import { createFixtureBackedFetch, createFixtureRegistry } from "./mockTransport";

describe("fixture-backed generated client transport", () => {
	it("returns typed responses from OpenAPI fixtures", async () => {
		const client = new EasyInterviewClient({
			fetch: createFixtureBackedFetch(
				createFixtureRegistry([
					getRuntimeConfigFixture,
					getMeFixture,
					listTargetJobsFixture,
					getPracticeSessionFixture,
				]),
			),
		});

		const runtimeConfig: RuntimeConfig = await client.getRuntimeConfig();
		const me: UserContext = await client.getMe();
		const targetJobs: PaginatedTargetJob = await client.listTargetJobs();
		const session: PracticeSession = await client.getPracticeSession(
			"01918fa0-0000-7000-8000-000000005000",
		);

		expect(runtimeConfig.appVersion).toBe("1.0.0+dev.0428");
		expect(me.emailMasked).toBe("ali***@example.com");
		expect(targetJobs.items[0]?.title).toBe("Senior Frontend Engineer");
		expect(session.messages.map((message) => message.role)).toEqual([
			"assistant",
			"user",
			"assistant",
		]);
	});

	it("selects named scenarios from Prefer and rejects unknown scenarios", async () => {
		const client = new EasyInterviewClient({
			fetch: createFixtureBackedFetch(createFixtureRegistry([getMeFixture])),
		});

		const authenticated = await client.getMe({
			headers: { Prefer: "example=authenticated" },
		});

		expect(authenticated.displayName).toBe("Alice Example");
		await expect(
			client.getMe({ headers: { Prefer: "example=does-not-exist" } }),
		).rejects.toThrow("unknown fixture scenario does-not-exist for operationId: getMe");
	});

	it("returns every report conversation fixture scenario without a legacy list fallback", async () => {
		const fetch = createFixtureBackedFetch(
			createFixtureRegistry([getReportConversationFixture]),
		);
		const reportId = "01918fa0-0070-7000-8000-000000000070";
		const scenarios = [
			"default",
			"queued",
			"generating",
			"failed",
			"empty-messages",
			"markdown-gfm",
			"cross-user-not-found",
			"report-not-found",
			"invalid-report-identity",
			"invalid-message-role",
			"invalid-message-sequence",
			"invalid-report-session-binding",
		] as const;

		for (const scenario of scenarios) {
			const response = await fetch(
				`http://fixture.local/api/v1/reports/${reportId}/conversation`,
				{ headers: { Prefer: `example=${scenario}` } },
			);
			const expected = getReportConversationFixture.scenarios[scenario].response;

			expect(response.status).toBe(expected.status);
			expect(await response.json()).toEqual(expected.body);
		}

		await expect(
			fetch("http://fixture.local/api/v1/practice/sessions"),
		).rejects.toThrow("no fixture route matched GET http://fixture.local/api/v1/practice/sessions");
	});

	it("honors fixture X-Mock-Delay-Ms before resolving responses", async () => {
		vi.useFakeTimers();
		try {
			const fetch = createFixtureBackedFetch(
				createFixtureRegistry([
					{
						...getMeFixture,
						scenarios: {
							default: {
								response: {
									...getMeFixture.scenarios.default.response,
									headers: {
										"X-Mock-Delay-Ms": "25",
									},
								},
							},
						},
					},
				]),
			);
			const promise = fetch("http://fixture.local/api/v1/me");
			const settled = vi.fn();
			void promise.then(settled);

			await vi.advanceTimersByTimeAsync(24);
			expect(settled).not.toHaveBeenCalled();

			await vi.advanceTimersByTimeAsync(1);
			const response = await promise;
			expect(response.status).toBe(200);
			expect(settled).toHaveBeenCalledTimes(1);
		} finally {
			vi.useRealTimers();
		}
	});

	it("rejects delayed fixture responses when the request is aborted", async () => {
		vi.useFakeTimers();
		try {
			const fetch = createFixtureBackedFetch(
				createFixtureRegistry([
					{
						...getMeFixture,
						scenarios: {
							default: {
								response: {
									...getMeFixture.scenarios.default.response,
									headers: {
										"X-Mock-Delay-Ms": "25",
									},
								},
							},
						},
					},
				]),
			);
			const controller = new AbortController();
			const promise = fetch("http://fixture.local/api/v1/me", {
				signal: controller.signal,
			});
			controller.abort();
			await expect(promise).rejects.toMatchObject({ name: "AbortError" });
			await vi.advanceTimersByTimeAsync(25);
		} finally {
			vi.useRealTimers();
		}
	});
});
