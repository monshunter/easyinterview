import { afterEach, describe, expect, it, vi } from "vitest";

import { createAppClient } from "./clientFactory";
import type {
	FeedbackReport,
	Job,
	PaginatedFeedbackReport,
	PaginatedPracticeSession,
	PaginatedResume,
	PracticePlan,
	PracticeSession,
	PracticeVoiceTurnResult,
	ReportWithJob,
	Resume,
	ResumeTailorRun,
	ResumeTailorRunWithJob,
	ResumeWithJob,
	SessionEventResult,
} from "./generated/types";

const API_BASE_URL = "http://api.test/api/v1";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-00000000f001";
const PRACTICE_PLAN_ID = "01918fa0-0000-7000-8000-00000000f101";
const PRACTICE_SESSION_ID = "01918fa0-0000-7000-8000-00000000f102";
const PRACTICE_TURN_ID = "01918fa0-0000-7000-8000-00000000f103";
const REPORT_ID = "01918fa0-0000-7000-8000-00000000f201";
const RESUME_ID = "01918fa0-0000-7000-8000-00000000f301";
const DUPLICATE_RESUME_ID = "01918fa0-0000-7000-8000-00000000f302";
const TAILOR_RUN_ID = "01918fa0-0000-7000-8000-00000000f303";
const REPORT_JOB_ID = "01918fa0-0000-7000-8000-00000000f901";
const RAW_RESUME_TEXT = "Led confidential platform migration with private metrics.";
const PROVENANCE = {
	dataSourceVersion: "frontend-owner-real-backend-2026-05-23",
	featureFlag: "frontend.owner.real_backend",
	language: "en",
	modelId: "model-profile:frontend-owner-live",
	promptVersion: "frontend-owner.v1",
	rubricVersion: "frontend-owner-rubric.v1",
};

function buildJob(overrides: Partial<Job> = {}): Job {
	return {
		id: REPORT_JOB_ID,
		jobType: "report_generate",
		resourceId: REPORT_ID,
		resourceType: "feedback_report",
		status: "queued",
		createdAt: "2026-05-23T01:00:00Z",
		updatedAt: "2026-05-23T01:00:00Z",
		...overrides,
	};
}

function buildPracticePlan(overrides: Partial<PracticePlan> = {}): PracticePlan {
	return {
		id: PRACTICE_PLAN_ID,
		difficulty: "standard",
		goal: "baseline",
		interviewerPersona: "technical_manager",
		language: "en",
		mode: "assisted",
		questionBudget: 4,
		resumeId: RESUME_ID,
		status: "ready",
		targetJobId: TARGET_JOB_ID,
		timeBudgetMinutes: 30,
		createdAt: "2026-05-23T01:00:00Z",
		...overrides,
	};
}

function buildPracticeSession(overrides: Partial<PracticeSession> = {}): PracticeSession {
	return {
		id: PRACTICE_SESSION_ID,
		planId: PRACTICE_PLAN_ID,
		targetJobId: TARGET_JOB_ID,
		status: "waiting_user_input",
		hintsEnabled: true,
		language: "en",
		turnCount: 1,
		currentTurn: {
			id: PRACTICE_TURN_ID,
			questionText: "Tell me about a complex platform migration.",
			status: "asked",
			turnIndex: 1,
		},
		createdAt: "2026-05-23T01:00:00Z",
		updatedAt: "2026-05-23T01:02:00Z",
		...overrides,
	};
}

function buildFeedbackReport(overrides: Partial<FeedbackReport> = {}): FeedbackReport {
	return {
		id: REPORT_ID,
		sessionId: PRACTICE_SESSION_ID,
		targetJobId: TARGET_JOB_ID,
		status: "ready",
		preparednessLevel: "basically_ready",
		provenance: PROVENANCE,
		questionAssessments: [
			{
				turnId: PRACTICE_TURN_ID,
				questionIntent: "system_design",
				reviewStatus: "queued_for_retry",
				includedInRetryPlan: true,
				dimensionResults: { communication: "meets_bar" },
			},
		],
		retryFocusTurnIds: [PRACTICE_TURN_ID],
		createdAt: "2026-05-23T01:05:00Z",
		updatedAt: "2026-05-23T01:06:00Z",
		...overrides,
	};
}

function buildResume(overrides: Partial<Resume> = {}): Resume {
	return {
		id: RESUME_ID,
		title: "Frontend Platform Resume",
		displayName: "Frontend Platform Resume",
		language: "en",
		parseStatus: "ready",
		sourceType: "paste",
		status: "active",
		originalText: null,
		parsedSummary: null,
		structuredProfile: { headline: "Frontend platform lead" },
		createdAt: "2026-05-23T01:00:00Z",
		updatedAt: "2026-05-23T01:01:00Z",
		...overrides,
	};
}

function buildTailorRun(overrides: Partial<ResumeTailorRun> = {}): ResumeTailorRun {
	return {
		id: TAILOR_RUN_ID,
		resumeId: RESUME_ID,
		targetJobId: TARGET_JOB_ID,
		status: "ready",
		matchSummary: { strengths: ["Platform scope"], gaps: ["Backend depth"] },
		suggestions: [
			{
				originalBullet: "Owned frontend systems.",
				suggestedBullet: "Led frontend systems tied to platform outcomes.",
				reason: "Connects scope to evidence.",
			},
		],
		provenance: PROVENANCE,
		createdAt: "2026-05-23T01:00:00Z",
		updatedAt: "2026-05-23T01:03:00Z",
		...overrides,
	};
}

function jsonResponse(body: unknown, status = 200): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { "Content-Type": "application/json" },
	});
}

function createRealApiFetchSpy() {
	return vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
		const url = new URL(String(input));
		const method = init?.method ?? "GET";
		const route = url.pathname.replace("/api/v1", "");

		if (route === "/practice/plans" && method === "POST") {
			return jsonResponse(buildPracticePlan(), 201);
		}
		if (route === `/practice/plans/${PRACTICE_PLAN_ID}` && method === "GET") {
			return jsonResponse(buildPracticePlan());
		}
		if (route === "/practice/sessions" && method === "GET") {
			const body: PaginatedPracticeSession = {
				items: [buildPracticeSession()],
				pageInfo: { hasMore: false, nextCursor: null, pageSize: 10 },
			};
			return jsonResponse(body);
		}
		if (route === "/practice/sessions" && method === "POST") {
			return jsonResponse(buildPracticeSession(), 201);
		}
		if (route === `/practice/sessions/${PRACTICE_SESSION_ID}` && method === "GET") {
			return jsonResponse(buildPracticeSession());
		}
		if (route === `/practice/sessions/${PRACTICE_SESSION_ID}/events` && method === "POST") {
			const body: SessionEventResult = {
				acknowledged: true,
				session: buildPracticeSession({ status: "running" }),
				assistantAction: {
					type: "ask_follow_up",
					sessionStatus: "waiting_user_input",
					turnId: PRACTICE_TURN_ID,
					questionText: "What did you defer?",
					provenance: PROVENANCE,
				},
			};
			return jsonResponse(body);
		}
		if (route === `/practice/sessions/${PRACTICE_SESSION_ID}/voice-turns` && method === "POST") {
			const body: PracticeVoiceTurnResult = {
				voiceTurnId: "01918fa0-0000-7000-8000-00000000f104",
				userTranscriptFinal: "I sequenced the migration.",
				assistantTextDraft: "What was the risk control?",
				session: buildPracticeSession({ status: "waiting_user_input" }),
				ttsChunks: [],
				ttsError: null,
				providerMetaSummary: {
					sttProvider: "test-stt",
					sttProfile: "stt-live",
					chatProvider: "test-chat",
					chatProfile: "chat-live",
					ttsProvider: "test-tts",
					ttsProfile: "tts-live",
				},
			};
			return jsonResponse(body);
		}
		if (route === `/practice/sessions/${PRACTICE_SESSION_ID}/complete` && method === "POST") {
			const body: ReportWithJob = {
				reportId: REPORT_ID,
				job: buildJob(),
			};
			return jsonResponse(body, 202);
		}
		if (route === `/reports/${REPORT_ID}` && method === "GET") {
			return jsonResponse(buildFeedbackReport());
		}
		if (route === `/targets/${TARGET_JOB_ID}/reports` && method === "GET") {
			const body: PaginatedFeedbackReport = {
				items: [buildFeedbackReport()],
				pageInfo: { hasMore: false, nextCursor: null, pageSize: 5 },
			};
			return jsonResponse(body);
		}
		if (route === "/resumes" && method === "GET") {
			const body: PaginatedResume = {
				items: [buildResume()],
				pageInfo: { hasMore: false, nextCursor: null, pageSize: 10 },
			};
			return jsonResponse(body);
		}
		if (route === "/resumes" && method === "POST") {
			const body: ResumeWithJob = {
				resumeId: RESUME_ID,
				job: buildJob({
					id: "01918fa0-0000-7000-8000-00000000f305",
					jobType: "resume_parse",
					resourceType: "resume_asset",
					resourceId: RESUME_ID,
				}),
			};
			return jsonResponse(body, 202);
		}
		if (route === `/resumes/${RESUME_ID}` && method === "GET") {
			return jsonResponse(buildResume());
		}
		if (route === `/resumes/${RESUME_ID}` && method === "PATCH") {
			return jsonResponse(buildResume({ displayName: "Frontend Platform v3" }));
		}
		if (route === `/resumes/${RESUME_ID}/archive` && method === "POST") {
			return jsonResponse(buildResume({ status: "archived" }));
		}
		if (route === `/resumes/${RESUME_ID}/duplicate` && method === "POST") {
			return jsonResponse(
				buildResume({ id: DUPLICATE_RESUME_ID, displayName: "Frontend Platform Resume (copy)" }),
				201,
			);
		}
		if (route === `/resumes/${RESUME_ID}/exports` && method === "POST") {
			return jsonResponse({ error: { code: "RESUME_EXPORT_NOT_AVAILABLE", message: "Export unavailable" } }, 501);
		}
		if (route === "/resume/tailor" && method === "POST") {
			const body: ResumeTailorRunWithJob = {
				tailorRunId: TAILOR_RUN_ID,
				job: buildJob({
					id: "01918fa0-0000-7000-8000-00000000f306",
					jobType: "resume_tailor",
					resourceType: "resume_tailor_run",
					resourceId: TAILOR_RUN_ID,
				}),
			};
			return jsonResponse(body, 202);
		}
		if (route === `/resume/tailor-runs/${TAILOR_RUN_ID}` && method === "GET") {
			return jsonResponse(buildTailorRun());
		}
		if (route === `/jobs/${REPORT_JOB_ID}` && method === "GET") {
			return jsonResponse(buildJob({
				status: "succeeded",
			}));
		}
		return jsonResponse({ error: `unexpected ${method} ${route}` }, 500);
	});
}

function callSummary(fetchSpy: ReturnType<typeof createRealApiFetchSpy>) {
	return fetchSpy.mock.calls.map(([input, init]) => {
		const url = new URL(String(input));
		return {
			method: init?.method ?? "GET",
			path: `${url.pathname}${url.search}`,
			credentials: init?.credentials,
			headers: new Headers(init?.headers ?? {}),
			body: init?.body,
		};
	});
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe("frontend owner real API mode", () => {
	it("routes practice, report, resume, job, and privacy owners through the real generated client", async () => {
		if (import.meta.env.VITE_EI_API_MODE !== undefined) {
			expect(import.meta.env.VITE_EI_API_MODE).toBe("real");
		}

		const fetchSpy = createRealApiFetchSpy();
		vi.stubGlobal("fetch", fetchSpy);
		const client = createAppClient({
			DEV: true,
			VITE_EI_API_MODE: "real",
			VITE_EI_API_BASE_URL: `${API_BASE_URL}/`,
		});

		expect(client.baseUrl).toBe(API_BASE_URL);

		const createdPlan = await client.createPracticePlan(
			{
				difficulty: "standard",
				goal: "baseline",
				interviewerPersona: "technical_manager",
				language: "en",
				mode: "assisted",
				questionBudget: 4,
				resumeId: RESUME_ID,
				targetJobId: TARGET_JOB_ID,
				timeBudgetMinutes: 30,
			},
			{ idempotencyKey: "ik_practice_plan" },
		);
		const plan = await client.getPracticePlan(PRACTICE_PLAN_ID);
		const sessions = await client.listPracticeSessions({ query: { targetJobId: TARGET_JOB_ID, status: "completed", pageSize: 10 } });
		const session = await client.startPracticeSession(
			{ planId: PRACTICE_PLAN_ID, hintsEnabled: true },
			{ idempotencyKey: "ik_practice_start" },
		);
		const refreshedSession = await client.getPracticeSession(PRACTICE_SESSION_ID);
		const eventResult = await client.appendSessionEvent(PRACTICE_SESSION_ID, {
			clientEventId: "client-event-001",
			kind: "answer_submitted",
			occurredAt: "2026-05-23T01:02:00Z",
			payload: { answerSummary: "Sequenced the migration." },
		});
		const voiceTurn = await client.createPracticeVoiceTurn(
			PRACTICE_SESSION_ID,
			{
				turnId: PRACTICE_TURN_ID,
				clientVoiceTurnId: "client-voice-001",
				language: "en",
				practiceMode: "assisted",
				manualTranscriptFallback: "I sequenced the migration.",
				audio: {
					contentType: "audio/webm",
					contentBase64: "UklGRg==",
					durationMs: 1200,
					byteLength: 8,
				},
			},
			{ idempotencyKey: "ik_voice_turn" },
		);
		const completed = await client.completePracticeSession(
			PRACTICE_SESSION_ID,
			{ clientCompletedAt: "2026-05-23T01:04:00Z" },
			{ idempotencyKey: "ik_practice_complete" },
		);
		const report = await client.getFeedbackReport(REPORT_ID);
		const reportList = await client.listTargetJobReports(TARGET_JOB_ID, { query: { pageSize: 5 } });

		const resumeList = await client.listResumes({ query: { pageSize: 10 } });
		const registered = await client.registerResume(
			{ title: "Frontend Platform Resume", language: "en", sourceType: "paste", rawText: RAW_RESUME_TEXT },
			{ idempotencyKey: "ik_register_resume" },
		);
		const resume = await client.getResume(RESUME_ID);
		const updated = await client.updateResume(
			RESUME_ID,
			{ displayName: "Frontend Platform v3", structuredProfile: { headline: "Updated lead" } },
			{ idempotencyKey: "ik_update_resume" },
		);
		const duplicated = await client.duplicateResume(
			RESUME_ID,
			{ displayName: "Frontend Platform Resume (copy)" },
			{ idempotencyKey: "ik_duplicate_resume" },
		);
		const archived = await client.archiveResume(RESUME_ID, { idempotencyKey: "ik_archive_resume" });
		const exportUnavailable = await client.exportResume(RESUME_ID, { idempotencyKey: "ik_export_resume" });
		const tailor = await client.requestResumeTailor(
			{ mode: "bullet_suggestions", resumeId: RESUME_ID, targetJobId: TARGET_JOB_ID },
			{ idempotencyKey: "ik_resume_tailor" },
		);
		const tailorRun = await client.getResumeTailorRun(TAILOR_RUN_ID);
		const reportJob = await client.getJob(REPORT_JOB_ID);

		expect(createdPlan.id).toBe(PRACTICE_PLAN_ID);
		expect(plan.targetJobId).toBe(TARGET_JOB_ID);
		expect(sessions.items[0]?.id).toBe(PRACTICE_SESSION_ID);
		expect(session.currentTurn?.id).toBe(PRACTICE_TURN_ID);
		expect(refreshedSession.status).toBe("waiting_user_input");
		expect(eventResult.assistantAction.provenance).toEqual(PROVENANCE);
		expect(voiceTurn.providerMetaSummary.chatProfile).toBe("chat-live");
		expect(completed.reportId).toBe(REPORT_ID);
		expect(report.provenance).toEqual(PROVENANCE);
		expect(reportList.items[0]?.retryFocusTurnIds).toEqual([PRACTICE_TURN_ID]);
		expect(resumeList.items[0]?.id).toBe(RESUME_ID);
		expect(registered.job.jobType).toBe("resume_parse");
		expect(registered.resumeId).toBe(RESUME_ID);
		expect(resume.originalText).toBeNull();
		expect(updated.displayName).toBe("Frontend Platform v3");
		expect(duplicated.id).toBe(DUPLICATE_RESUME_ID);
		expect(archived.status).toBe("archived");
		expect(exportUnavailable.error.code).toBe("RESUME_EXPORT_NOT_AVAILABLE");
		expect(tailor.tailorRunId).toBe(TAILOR_RUN_ID);
		expect(tailorRun.resumeId).toBe(RESUME_ID);
		expect(tailorRun.provenance).toEqual(PROVENANCE);
		expect(reportJob.status).toBe("succeeded");

		const summary = callSummary(fetchSpy);
		expect(summary.map(({ method, path }) => `${method} ${path}`)).toEqual([
			"POST /api/v1/practice/plans",
			`GET /api/v1/practice/plans/${PRACTICE_PLAN_ID}`,
			`GET /api/v1/practice/sessions?targetJobId=${TARGET_JOB_ID}&status=completed&pageSize=10`,
			"POST /api/v1/practice/sessions",
			`GET /api/v1/practice/sessions/${PRACTICE_SESSION_ID}`,
			`POST /api/v1/practice/sessions/${PRACTICE_SESSION_ID}/events`,
			`POST /api/v1/practice/sessions/${PRACTICE_SESSION_ID}/voice-turns`,
			`POST /api/v1/practice/sessions/${PRACTICE_SESSION_ID}/complete`,
			`GET /api/v1/reports/${REPORT_ID}`,
			`GET /api/v1/targets/${TARGET_JOB_ID}/reports?pageSize=5`,
			"GET /api/v1/resumes?pageSize=10",
			"POST /api/v1/resumes",
			`GET /api/v1/resumes/${RESUME_ID}`,
			`PATCH /api/v1/resumes/${RESUME_ID}`,
			`POST /api/v1/resumes/${RESUME_ID}/duplicate`,
			`POST /api/v1/resumes/${RESUME_ID}/archive`,
			`POST /api/v1/resumes/${RESUME_ID}/exports`,
			"POST /api/v1/resume/tailor",
			`GET /api/v1/resume/tailor-runs/${TAILOR_RUN_ID}`,
			`GET /api/v1/jobs/${REPORT_JOB_ID}`,
		]);

		for (const call of summary) {
			expect(call.credentials).toBe("include");
			expect(call.headers.get("Prefer")).toBeNull();
			expect(String(call.path)).not.toContain(RAW_RESUME_TEXT);
		}

		expect(summary[0]?.headers.get("Idempotency-Key")).toBe("ik_practice_plan");
		expect(summary[3]?.headers.get("Idempotency-Key")).toBe("ik_practice_start");
		expect(summary[5]?.headers.get("Idempotency-Key")).toBeNull();
		expect(summary[6]?.headers.get("Idempotency-Key")).toBe("ik_voice_turn");
		expect(summary[7]?.headers.get("Idempotency-Key")).toBe("ik_practice_complete");
		expect(summary[11]?.headers.get("Idempotency-Key")).toBe("ik_register_resume");
		expect(summary[12]?.headers.get("Idempotency-Key")).toBeNull();
		expect(summary[13]?.headers.get("Idempotency-Key")).toBe("ik_update_resume");
		expect(summary[14]?.headers.get("Idempotency-Key")).toBe("ik_duplicate_resume");
		expect(summary[15]?.headers.get("Idempotency-Key")).toBe("ik_archive_resume");
		expect(summary[16]?.headers.get("Idempotency-Key")).toBe("ik_export_resume");
		expect(summary[17]?.headers.get("Idempotency-Key")).toBe("ik_resume_tailor");
		expect(summary[19]?.headers.get("Idempotency-Key")).toBeNull();
		expect(JSON.parse(String(summary[11]?.body))).toMatchObject({
			sourceType: "paste",
			rawText: RAW_RESUME_TEXT,
		});
	});
});
