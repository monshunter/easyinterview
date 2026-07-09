import deleteMeFixture from "../../../openapi/fixtures/Auth/deleteMe.json";
import completeMyProfileFixture from "../../../openapi/fixtures/Auth/completeMyProfile.json";
import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import logoutFixture from "../../../openapi/fixtures/Auth/logout.json";
import startAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/startAuthEmailChallenge.json";
import verifyAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/verifyAuthEmailChallenge.json";
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
import getFeedbackReportFixture from "../../../openapi/fixtures/Reports/getFeedbackReport.json";
import listTargetJobReportsFixture from "../../../openapi/fixtures/Reports/listTargetJobReports.json";
import getResumeTailorRunFixture from "../../../openapi/fixtures/ResumeTailor/getResumeTailorRun.json";
import requestResumeTailorFixture from "../../../openapi/fixtures/ResumeTailor/requestResumeTailor.json";
import archiveResumeFixture from "../../../openapi/fixtures/Resumes/archiveResume.json";
import duplicateResumeFixture from "../../../openapi/fixtures/Resumes/duplicateResume.json";
import exportResumeFixture from "../../../openapi/fixtures/Resumes/exportResume.json";
import getResumeFixture from "../../../openapi/fixtures/Resumes/getResume.json";
import getResumeSourceFixture from "../../../openapi/fixtures/Resumes/getResumeSource.json";
import listResumesFixture from "../../../openapi/fixtures/Resumes/listResumes.json";
import registerResumeFixture from "../../../openapi/fixtures/Resumes/registerResume.json";
import updateResumeFixture from "../../../openapi/fixtures/Resumes/updateResume.json";
import getTargetJobFixture from "../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import archiveTargetJobFixture from "../../../openapi/fixtures/TargetJobs/archiveTargetJob.json";
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

const DEV_MOCK_FIXTURES = [
	startAuthEmailChallengeFixture,
	verifyAuthEmailChallengeFixture,
	logoutFixture,
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
	getFeedbackReportFixture,
	requestResumeTailorFixture,
	getResumeTailorRunFixture,
	listResumesFixture,
	registerResumeFixture,
	getResumeFixture,
	getResumeSourceFixture,
	updateResumeFixture,
	duplicateResumeFixture,
	archiveResumeFixture,
	exportResumeFixture,
	getRuntimeConfigFixture,
	listTargetJobsFixture,
	importTargetJobFixture,
	getTargetJobFixture,
	updateTargetJobFixture,
	archiveTargetJobFixture,
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
	const fetch: typeof globalThis.fetch = async (input, init) => {
		const request = readRequest(input, init);
		const authResponse = respondToStatefulAuthRequest(request, authState);
		if (authResponse) return authResponse;
		const response = await fixtureFetch(input, init);
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

function parseJsonObject(bodyText: string | undefined): Record<string, unknown> | null {
	if (!bodyText) return null;
	try {
		const parsed: unknown = JSON.parse(bodyText);
		return isObject(parsed) ? parsed : null;
	} catch {
		return null;
	}
}

function isObject(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}

function stripApiBase(path: string): string {
	return path.startsWith("/api/v1/") ? path.slice("/api/v1".length) : path;
}
