import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

import { ALL_OPERATION_IDS } from "./generated/client";
import {
	createDevMockClient,
	createDevMockFixtureRegistry,
} from "./devMockClient";

describe("frontend dev fixture-backed mock client", () => {
	it("does not expose a test-only fixture operation observer", () => {
		const observer = ["getDevMockFixture", "OperationIds"].join("");
		expect(readFileSync(resolve(__dirname, "devMockClient.ts"), "utf8")).not.toContain(
			observer,
		);
	});

	it("covers every generated operationId with a fixture", () => {
		expect(Object.keys(createDevMockFixtureRegistry()).sort()).toEqual(
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

	it("serves the report-scoped conversation fixture through the generated operation", async () => {
		const client = createDevMockClient();

		await expect(
			client.getReportConversation("01918fa0-0070-7000-8000-000000000070"),
		).resolves.toMatchObject({
			reportId: "01918fa0-0070-7000-8000-000000000070",
			reportStatus: "ready",
			messages: [
				{ sequence: 1, role: "user" },
				{ sequence: 2, role: "assistant" },
			],
		});
	});

	it("serves failed-report regeneration through the generated fixture operation", async () => {
		const client = createDevMockClient();
		const reportId = "01918fa0-0079-7000-8000-000000000079";

		await expect(
			client.regenerateFeedbackReport(reportId, {
				idempotencyKey: "idem-regenerate-report-2026-07-16-default",
			}),
		).resolves.toMatchObject({
			reportId,
			job: {
				jobType: "report_generate",
				status: "queued",
				resourceType: "feedback_report",
				resourceId: reportId,
			},
		});
	});

	it("keeps voice mode fail-closed in the generated fixture client", async () => {
		const client = createDevMockClient();
		const body = {
			clientVoiceTurnId: "01918fa0-0000-7000-8000-00000000f101",
			audio: {
				contentBase64: "T2dnUw==",
				contentType: "audio/webm" as const,
				durationMs: 4320,
			},
			language: "zh-CN",
		};

		await expect(client.createPracticeVoiceTurn(
			"01918fa0-0000-7000-8000-000000005000",
			body,
			{ idempotencyKey: "01918fa0-0000-7000-8000-00000000f001" },
		)).rejects.toThrow(/HTTP 422.*AI_UNSUPPORTED_CAPABILITY/);

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
			client.exportResume("0195f2d0-0002-7000-8000-000000000201"),
		).resolves.toMatchObject({
			error: { code: "RESUME_EXPORT_NOT_AVAILABLE" },
		});
		await expect(client.requestPrivacyExport()).resolves.toMatchObject({
			error: { code: "PRIVACY_EXPORT_NOT_AVAILABLE" },
		});
	});

	it("models dev auth session state across /me, verify, and logout", async () => {
		const client = createDevMockClient();

		await expect(client.getMe()).rejects.toThrow(/HTTP 401/);

		await client.verifyAuthEmailChallenge({
			query: { token: "654321" },
		});
		await expect(client.getMe()).resolves.toMatchObject({
			displayName: "",
			email: "new.user@example.com",
			profileCompletionRequired: true,
		});

		await client.updateMe({
			displayName: "Alice Example",
			acceptedTerms: true,
		});
		await expect(client.getMe()).resolves.toMatchObject({
			displayName: "Alice Example",
			email: "new.user@example.com",
			profileCompletionRequired: false,
		});

		await client.logout();
		await expect(client.getMe()).rejects.toThrow(/HTTP 401/);
	});

	it("models account deletion as an authenticated session termination", async () => {
		const client = createDevMockClient();

		await client.verifyAuthEmailChallenge({
			query: { token: "654321" },
		});
		await client.updateMe({
			displayName: "Alice Example",
			acceptedTerms: true,
		});

		await client.deleteMe({
			idempotencyKey: "01918fa0-0000-7000-8000-00000000d001",
		});

		await expect(client.getMe()).rejects.toThrow(/HTTP 401/);
	});

	it("rejects a combined update when displayPreferences is present but invalid", async () => {
		const client = createDevMockClient();

		await client.verifyAuthEmailChallenge({
			query: { token: "654321" },
		});
		await expect(client.updateMe({
			displayName: "Alice Example",
			acceptedTerms: true,
			displayPreferences: { theme: "forest", customAccent: null },
		} as never)).rejects.toThrow(/HTTP 400/);

		await expect(client.getMe()).resolves.toMatchObject({
			displayName: "",
			profileCompletionRequired: true,
			displayPreferences: { theme: "ocean", customAccent: null },
		});
	});

});
