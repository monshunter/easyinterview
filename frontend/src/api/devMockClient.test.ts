import { describe, expect, it } from "vitest";

import { ALL_OPERATION_IDS } from "./generated/client";
import {
	createDevMockClient,
	getDevMockFixtureOperationIds,
} from "./devMockClient";

describe("frontend dev fixture-backed mock client", () => {
	it("covers every generated operationId with a fixture", () => {
		expect(getDevMockFixtureOperationIds().sort()).toEqual(
			[...ALL_OPERATION_IDS].sort(),
		);
	});

	it("returns fixture-backed data for dev preview pages", async () => {
		const client = createDevMockClient();

		const runtime = await client.getRuntimeConfig();
		const me = await client.getMe({
			headers: { Prefer: "example=authenticated" },
		});
		const targets = await client.listTargetJobs();

		expect(runtime.appVersion).toBe("1.0.0+dev.0428");
		expect(me.displayName).toBe("Alice Example");
		expect(targets.items[0]?.title).toBe("Senior Frontend Engineer");
	});

	it("serves createPracticeVoiceTurn through the generated fixture client", async () => {
		const client = createDevMockClient();
		const body = {
			clientVoiceTurnId: "01918fa0-0000-7000-8000-00000000f101",
			turnId: "01918fa0-0000-7000-8000-000000006000",
			audio: {
				contentBase64: "T2dnUw==",
				contentType: "audio/webm" as const,
				durationMs: 4320,
			},
			language: "zh-CN",
			practiceMode: "assisted" as const,
		};

		await expect(
			client.createPracticeVoiceTurn(
				"01918fa0-0000-7000-8000-000000005000",
				body,
				{ idempotencyKey: "01918fa0-0000-7000-8000-00000000f001" },
			),
		).resolves.toMatchObject({
			voiceTurnId: "01918fa0-0000-7000-8000-00000000f201",
			providerMetaSummary: {
				sttProfile: "practice.voice.stt.default",
				chatProfile: "practice.followup.default",
				ttsProfile: "practice.voice.tts.default",
			},
		});

		await expect(
			client.createPracticeVoiceTurn(
				"01918fa0-0000-7000-8000-000000005000",
				body,
				{
					idempotencyKey: "01918fa0-0000-7000-8000-00000000f002",
					headers: { Prefer: "example=does-not-exist" },
				},
			),
		).rejects.toThrow(
			"unknown fixture scenario does-not-exist for operationId: createPracticeVoiceTurn",
		);
	});

	it("parses declared non-OK export fallback responses", async () => {
		const client = createDevMockClient();

		await expect(
			client.exportResumeVersion("0195f2d0-0002-7000-8000-000000000201"),
		).resolves.toMatchObject({
			error: { code: "RESUME_EXPORT_NOT_AVAILABLE" },
		});
		await expect(client.requestPrivacyExport()).resolves.toMatchObject({
			error: { code: "PRIVACY_EXPORT_NOT_AVAILABLE" },
		});
	});

	it("returns the async branch response shape for ai_select", async () => {
		const client = createDevMockClient();

		const response = await client.branchResumeVersion(
			{
				parentVersionId: "0195f2d0-0002-7000-8000-000000000201",
				targetJobId: "01918fa0-0030-7a00-8a00-000000000030",
				seedStrategy: "ai_select",
			},
			{ headers: { Prefer: "example=ai-select-202-with-job" } },
		);

		expect(response).toMatchObject({
			resumeVersionId: "0195f2d0-0002-7000-8000-000000000204",
			job: { jobType: "resume_tailor" },
		});
	});

	it("models dev auth session state across /me, verify, and logout", async () => {
		const client = createDevMockClient();

		await expect(client.getMe()).rejects.toThrow(/HTTP 401/);

		await client.verifyAuthEmailChallenge({
			query: { token: "654321" },
		});
		await expect(client.getMe()).resolves.toMatchObject({
			displayName: "Alice Example",
			emailMasked: "ali***@example.com",
		});

		await client.logout();
		await expect(client.getMe()).rejects.toThrow(/HTTP 401/);
	});
});
