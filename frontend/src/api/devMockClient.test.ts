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
			emailMasked: "new***r@example.com",
			profileCompletionRequired: true,
		});

		await client.completeMyProfile({
			displayName: "Alice Example",
			acceptedTerms: true,
		});
		await expect(client.getMe()).resolves.toMatchObject({
			displayName: "Alice Example",
			emailMasked: "new***r@example.com",
			profileCompletionRequired: false,
		});

		await client.logout();
		await expect(client.getMe()).rejects.toThrow(/HTTP 401/);
	});

});
