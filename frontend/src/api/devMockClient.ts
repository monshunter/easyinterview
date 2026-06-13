import deleteMeFixture from "../../../openapi/fixtures/Auth/deleteMe.json";
import completeMyProfileFixture from "../../../openapi/fixtures/Auth/completeMyProfile.json";
import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import logoutFixture from "../../../openapi/fixtures/Auth/logout.json";
import startAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/startAuthEmailChallenge.json";
import verifyAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/verifyAuthEmailChallenge.json";
import createDebriefFixture from "../../../openapi/fixtures/Debriefs/createDebrief.json";
import getDebriefFixture from "../../../openapi/fixtures/Debriefs/getDebrief.json";
import suggestDebriefQuestionsFixture from "../../../openapi/fixtures/Debriefs/suggestDebriefQuestions.json";
import getJobFixture from "../../../openapi/fixtures/Jobs/getJob.json";
import createPracticePlanFixture from "../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import getPracticePlanFixture from "../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import appendSessionEventFixture from "../../../openapi/fixtures/PracticeSessions/appendSessionEvent.json";
import completePracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/completePracticeSession.json";
import createPracticeVoiceTurnFixture from "../../../openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json";
import getPracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/getPracticeSession.json";
import listPracticeSessionsFixture from "../../../openapi/fixtures/PracticeSessions/listPracticeSessions.json";
import startPracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/startPracticeSession.json";
import getPrivacyRequestFixture from "../../../openapi/fixtures/Privacy/getPrivacyRequest.json";
import requestPrivacyDeleteFixture from "../../../openapi/fixtures/Privacy/requestPrivacyDelete.json";
import requestPrivacyExportFixture from "../../../openapi/fixtures/Privacy/requestPrivacyExport.json";
import createExperienceCardFixture from "../../../openapi/fixtures/Profile/createExperienceCard.json";
import getMyProfileFixture from "../../../openapi/fixtures/Profile/getMyProfile.json";
import listExperienceCardsFixture from "../../../openapi/fixtures/Profile/listExperienceCards.json";
import updateExperienceCardFixture from "../../../openapi/fixtures/Profile/updateExperienceCard.json";
import updateMyProfileFixture from "../../../openapi/fixtures/Profile/updateMyProfile.json";
import getFeedbackReportFixture from "../../../openapi/fixtures/Reports/getFeedbackReport.json";
import listTargetJobReportsFixture from "../../../openapi/fixtures/Reports/listTargetJobReports.json";
import getResumeTailorRunFixture from "../../../openapi/fixtures/ResumeTailor/getResumeTailorRun.json";
import requestResumeTailorFixture from "../../../openapi/fixtures/ResumeTailor/requestResumeTailor.json";
import acceptResumeTailorSuggestionFixture from "../../../openapi/fixtures/Resumes/acceptResumeTailorSuggestion.json";
import archiveResumeAssetFixture from "../../../openapi/fixtures/Resumes/archiveResumeAsset.json";
import branchResumeVersionFixture from "../../../openapi/fixtures/Resumes/branchResumeVersion.json";
import confirmResumeStructuredMasterFixture from "../../../openapi/fixtures/Resumes/confirmResumeStructuredMaster.json";
import exportResumeVersionFixture from "../../../openapi/fixtures/Resumes/exportResumeVersion.json";
import getResumeFixture from "../../../openapi/fixtures/Resumes/getResume.json";
import getResumeVersionFixture from "../../../openapi/fixtures/Resumes/getResumeVersion.json";
import listResumesFixture from "../../../openapi/fixtures/Resumes/listResumes.json";
import listResumeVersionsFixture from "../../../openapi/fixtures/Resumes/listResumeVersions.json";
import rejectResumeTailorSuggestionFixture from "../../../openapi/fixtures/Resumes/rejectResumeTailorSuggestion.json";
import registerResumeFixture from "../../../openapi/fixtures/Resumes/registerResume.json";
import updateResumeVersionFixture from "../../../openapi/fixtures/Resumes/updateResumeVersion.json";
import getTargetJobFixture from "../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import importTargetJobFixture from "../../../openapi/fixtures/TargetJobs/importTargetJob.json";
import listTargetJobsFixture from "../../../openapi/fixtures/TargetJobs/listTargetJobs.json";
import updateTargetJobFixture from "../../../openapi/fixtures/TargetJobs/updateTargetJob.json";
import createUploadPresignFixture from "../../../openapi/fixtures/Uploads/createUploadPresign.json";
import {
	ALL_OPERATION_IDS,
	EasyInterviewClient,
	type OperationId,
} from "./generated/client";
import type { UserContext } from "./generated/types";
import {
	createFixtureBackedFetch,
	createFixtureRegistry,
	type FixtureBackedFetchOptions,
	type FixtureRegistry,
	type OperationFixture,
} from "./mockTransport";
import { JOB_TYPE_DEBRIEF_GENERATE } from "../lib/jobs/jobs";

const DEV_MOCK_FIXTURES = [
	startAuthEmailChallengeFixture,
	verifyAuthEmailChallengeFixture,
	logoutFixture,
	createDebriefFixture,
	getDebriefFixture,
	suggestDebriefQuestionsFixture,
	getJobFixture,
	deleteMeFixture,
	completeMyProfileFixture,
	getMeFixture,
	createPracticePlanFixture,
	getPracticePlanFixture,
	startPracticeSessionFixture,
	getPracticeSessionFixture,
	listPracticeSessionsFixture,
	completePracticeSessionFixture,
	appendSessionEventFixture,
	createPracticeVoiceTurnFixture,
	requestPrivacyDeleteFixture,
	requestPrivacyExportFixture,
	getPrivacyRequestFixture,
	getMyProfileFixture,
	updateMyProfileFixture,
	listExperienceCardsFixture,
	createExperienceCardFixture,
	updateExperienceCardFixture,
	getFeedbackReportFixture,
	requestResumeTailorFixture,
	getResumeTailorRunFixture,
	acceptResumeTailorSuggestionFixture,
	archiveResumeAssetFixture,
	branchResumeVersionFixture,
	confirmResumeStructuredMasterFixture,
	exportResumeVersionFixture,
	getResumeVersionFixture,
	listResumesFixture,
	listResumeVersionsFixture,
	rejectResumeTailorSuggestionFixture,
	updateResumeVersionFixture,
	registerResumeFixture,
	getResumeFixture,
	getRuntimeConfigFixture,
	listTargetJobsFixture,
	importTargetJobFixture,
	getTargetJobFixture,
	updateTargetJobFixture,
	listTargetJobReportsFixture,
	createUploadPresignFixture,
] as readonly OperationFixture[];

export function getDevMockFixtureOperationIds(): OperationId[] {
	return DEV_MOCK_FIXTURES.map((fixture) => fixture.operationId as OperationId);
}

export function createDevMockFixtureRegistry(): FixtureRegistry {
	const registry = createFixtureRegistry(DEV_MOCK_FIXTURES);
	const missing = ALL_OPERATION_IDS.filter((operationId) => !registry[operationId]);
	if (missing.length > 0) {
		throw new Error(`missing dev mock fixtures for operationId(s): ${missing.join(", ")}`);
	}
	return registry;
}

export function createDevMockClient(
	options: FixtureBackedFetchOptions = {},
): EasyInterviewClient {
	const registry = createDevMockFixtureRegistry();
	const fixtureFetch = createFixtureBackedFetch(registry, options);
	const authState = {
		signedIn: false,
		profileComplete: false,
		displayName: "",
		emailMasked: "new***r@example.com",
		lastEmail: "",
	};
	const debriefJobIds = new Set<string>();
	const debriefPracticePlanIds = new Set<string>();
	const fetch: typeof globalThis.fetch = async (input, init) => {
		const request = readRequest(input, init);
		const authResponse = respondToStatefulAuthRequest(request, authState);
		if (authResponse) return authResponse;
		const scenarioInit = withStatefulDebriefScenario(
			init,
			request,
			debriefJobIds,
			debriefPracticePlanIds,
		);
		const response = await fixtureFetch(input, scenarioInit);
		if (response.ok && request.method === "POST" && request.path === "/auth/email/start") {
			const body = parseJsonObject(request.bodyText);
			if (typeof body?.email === "string") {
				authState.lastEmail = body.email;
				authState.emailMasked = maskEmail(body.email);
			}
		}
		if (response.ok && request.method === "GET" && request.path === "/auth/email/verify") {
			authState.signedIn = true;
		}
		if (response.ok && request.method === "POST" && request.path === "/auth/logout") {
			authState.signedIn = false;
		}
		if (response.ok && request.method === "POST" && request.path === "/debriefs") {
			await rememberDebriefJob(response, debriefJobIds);
		}
		if (response.ok && request.method === "POST" && request.path === "/practice/plans") {
			await rememberDebriefPracticePlan(response, debriefPracticePlanIds);
		}
		return response;
	};
	return new EasyInterviewClient({
		fetch,
	});
}

function readRequest(
	input: RequestInfo | URL,
	init?: RequestInit,
): { method: string; path: string; headers: Headers; bodyText?: string } {
	const method =
		init?.method ??
		(input instanceof Request ? input.method : "GET");
	const rawUrl =
		typeof input === "string"
			? input
			: input instanceof URL
				? input.href
				: input.url;
	const headers = new Headers(input instanceof Request ? input.headers : undefined);
	new Headers(init?.headers).forEach((value, key) => headers.set(key, value));
	return {
		method: method.toUpperCase(),
		path: stripApiBase(new URL(rawUrl, "http://fixture.local").pathname),
		headers,
		bodyText: typeof init?.body === "string" ? init.body : undefined,
	};
}

function respondToStatefulAuthRequest(
	request: { method: string; path: string; headers: Headers; bodyText?: string },
	state: {
		signedIn: boolean;
		profileComplete: boolean;
		displayName: string;
		emailMasked: string;
		lastEmail: string;
	},
): Response | null {
	if (request.headers.has("Prefer")) return null;
	if (request.method === "GET" && request.path === "/me") {
		if (!state.signedIn) {
			return jsonResponse(401, {
				error: {
					code: "AUTH_UNAUTHORIZED",
					message: "Session cookie is missing or expired.",
					requestId: "req_dev_mock_auth_unauthenticated",
					retryable: false,
					details: { seedProfile: "unauthenticated" },
				},
			});
		}
		return jsonResponse(200, buildMockUserContext(state));
	}
	if (request.method === "PATCH" && request.path === "/me") {
		const body = parseJsonObject(request.bodyText);
		const displayName = typeof body?.displayName === "string" ? body.displayName.trim() : "";
		const acceptedTerms = body?.acceptedTerms === true;
		if (!state.signedIn) {
			return jsonResponse(401, {
				error: {
					code: "AUTH_UNAUTHORIZED",
					message: "Session cookie is missing or expired.",
					requestId: "req_dev_mock_profile_unauthenticated",
					retryable: false,
					details: {},
				},
			});
		}
		if (!displayName || !acceptedTerms) {
			return jsonResponse(400, {
				error: {
					code: "VALIDATION_FAILED",
					message: "displayName and acceptedTerms are required.",
					requestId: "req_dev_mock_profile_validation",
					retryable: false,
					details: {},
				},
			});
		}
		state.displayName = displayName;
		state.profileComplete = true;
		return jsonResponse(200, buildMockUserContext(state));
	}
	return null;
}

function buildMockUserContext(state: {
	profileComplete: boolean;
	displayName: string;
	emailMasked: string;
}): UserContext {
	return {
		id: "01918fa0-0000-7000-8000-000000000101",
		emailMasked: state.emailMasked,
		displayName: state.profileComplete ? state.displayName || "Alice Example" : "",
		uiLanguage: "zh-CN",
		preferredPracticeLanguage: "zh-CN",
		profileCompletionRequired: !state.profileComplete,
	};
}

function jsonResponse(status: number, body: unknown): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { "Content-Type": "application/json" },
	});
}

function maskEmail(email: string): string {
	const [local, domain] = email.split("@");
	if (!local || !domain) return "new***r@example.com";
	const prefix = local.slice(0, Math.min(3, local.length));
	return `${prefix}***@${domain}`;
}

function withStatefulDebriefScenario(
	init: RequestInit | undefined,
	request: {
		method: string;
		path: string;
		headers: Headers;
		bodyText?: string;
	},
	debriefJobIds: ReadonlySet<string>,
	debriefPracticePlanIds: ReadonlySet<string>,
): RequestInit | undefined {
	if (request.headers.has("Prefer")) return init;
	if (request.method === "GET") {
		const jobId = readPathId(request.path, "/jobs/");
		if (jobId && debriefJobIds.has(jobId)) {
			return withPreferScenario(init, request, "debrief-succeeded");
		}
	}
	if (
		request.method === "POST" &&
		request.path === "/practice/plans" &&
		isDebriefPracticePlanRequest(request.bodyText)
	) {
		return withPreferScenario(init, request, "debrief-derived");
	}
	if (request.method === "POST" && request.path === "/practice/sessions") {
		const planId = readPracticeSessionPlanId(request.bodyText);
		if (planId && debriefPracticePlanIds.has(planId)) {
			return withPreferScenario(init, request, "debrief-derived-first-question");
		}
	}
	return init;
}

function withPreferScenario(
	init: RequestInit | undefined,
	request: { headers: Headers },
	scenario: string,
): RequestInit {
	const headers = new Headers(request.headers);
	headers.set("Prefer", `example=${scenario}`);
	return { ...init, headers };
}

async function rememberDebriefJob(
	response: Response,
	debriefJobIds: Set<string>,
): Promise<void> {
	const body = await readJsonObject(response);
	const job = isObject(body?.job) ? body.job : null;
	if (
		typeof job?.id === "string" &&
		job.jobType === JOB_TYPE_DEBRIEF_GENERATE &&
		job.resourceType === "debrief"
	) {
		debriefJobIds.add(job.id);
	}
}

async function rememberDebriefPracticePlan(
	response: Response,
	debriefPracticePlanIds: Set<string>,
): Promise<void> {
	const body = await readJsonObject(response);
	if (typeof body?.id === "string" && body.goal === "debrief") {
		debriefPracticePlanIds.add(body.id);
	}
}

async function readJsonObject(response: Response): Promise<Record<string, unknown> | null> {
	try {
		const body: unknown = await response.clone().json();
		return isObject(body) ? body : null;
	} catch {
		return null;
	}
}

function isDebriefPracticePlanRequest(bodyText: string | undefined): boolean {
	const body = parseJsonObject(bodyText);
	return body?.goal === "debrief";
}

function readPracticeSessionPlanId(bodyText: string | undefined): string | null {
	const body = parseJsonObject(bodyText);
	return typeof body?.planId === "string" ? body.planId : null;
}

function parseJsonObject(bodyText: string | undefined): Record<string, unknown> | null {
	if (!bodyText) return null;
	try {
		const parsed: unknown = JSON.parse(bodyText);
		return isObject(parsed) ? parsed : null;
	} catch {
		return null;
	}
}

function readPathId(path: string, prefix: string): string | null {
	if (!path.startsWith(prefix)) return null;
	const rest = path.slice(prefix.length);
	return rest && !rest.includes("/") ? decodeURIComponent(rest) : null;
}

function isObject(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}

function stripApiBase(path: string): string {
	return path.startsWith("/api/v1/") ? path.slice("/api/v1".length) : path;
}
