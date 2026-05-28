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

		const voiceTurn = await client.createPracticeVoiceTurn(
			"01918fa0-0000-7000-8000-000000005000",
			body,
			{ idempotencyKey: "01918fa0-0000-7000-8000-00000000f001" },
		);

		expect(voiceTurn).toMatchObject({
			voiceTurnId: "01918fa0-0000-7000-8000-00000000f201",
			providerMetaSummary: {
				sttProfile: "practice.voice.stt.default",
				chatProfile: "practice.followup.default",
				ttsProfile: "practice.voice.tts.default",
			},
		});
		expect(voiceTurn.ttsChunks[0]?.audioRef).toMatch(
			/^data:audio\/[a-z0-9.+-]+;base64,/i,
		);

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

	it("serves confirmResumeStructuredMaster through the generated fixture client", async () => {
		const client = createDevMockClient();

		const response = await client.confirmResumeStructuredMaster(
			"01918fa0-0000-7000-8000-000000001000",
			{
				displayName: "Structured master",
				language: "zh-CN",
				structuredProfile: {
					headline: "Senior frontend engineer",
					summary:
						"Owns complex product surfaces and turns interview evidence into concise resume proof.",
					skills: ["React", "TypeScript", "Design systems"],
					sections: [],
					provenance: {
						promptVersion: "resume_profile.v1",
						rubricVersion: "not_applicable",
						modelId: "fixture-model:resume-version-profile",
						language: "zh-CN",
						featureFlag: "resume-workshop-additive",
						dataSourceVersion: "resume_asset.v1",
					},
				},
			},
			{ idempotencyKey: "idem-confirm-structured-master-2026-05-17" },
		);

		expect(response).toMatchObject({
			versionType: "structured_master",
			parentVersionId: null,
			seedStrategy: null,
			structuredProfile: {
				provenance: { promptVersion: "resume_profile.v1" },
			},
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

	it("advances the created debrief job in default dev mock mode", async () => {
		const client = createDevMockClient();

		const created = await client.createDebrief(
			{
				targetJobId: "01918fa0-0000-7000-8000-000000002000",
				roundType: "technical",
				language: "zh-CN",
				questions: [
					{
						questionText: "你如何处理跨团队设计系统迁移？",
						myAnswerSummary: "我用分阶段 rollout 和 adoption 指标推动迁移。",
					},
				],
			},
			{ headers: { "Idempotency-Key": "idem-dev-debrief-mock-flow" } },
		);

		await expect(client.getJob(created.job.id)).resolves.toMatchObject({
			id: created.job.id,
			jobType: "debrief_generate",
			resourceType: "debrief",
			resourceId: created.debriefId,
			status: "succeeded",
		});
	});

	it("uses debrief-derived practice fixtures for default dev mock replay", async () => {
		const client = createDevMockClient();

		const plan = await client.createPracticePlan(
			{
				targetJobId: "01918fa0-0000-7000-8000-000000002000",
				goal: "debrief",
				mode: "assisted",
				interviewerPersona: "hiring_manager",
				difficulty: "standard",
				language: "zh-CN",
				questionBudget: 6,
				timeBudgetMinutes: 30,
				resumeAssetId: "01918fa0-0000-7000-8000-000000001000",
				sourceDebriefId: "01918fa0-0000-7000-8000-00000000a000",
				focusCompetencyCodes: [],
			},
			{ idempotencyKey: "idem-dev-debrief-plan" },
		);

		expect(plan).toMatchObject({
			id: "01918fa0-0000-7000-8000-000000004700",
			goal: "debrief",
			status: "ready",
		});

		await expect(
			client.startPracticeSession(
				{ planId: plan.id, hintsEnabled: false },
				{ idempotencyKey: "idem-dev-debrief-session" },
			),
		).resolves.toMatchObject({
			id: "01918fa0-0000-7000-8000-000000005700",
			planId: plan.id,
			status: "running",
			currentTurn: {
				questionIntent: "debrief.source_question",
			},
		});
	});
});
