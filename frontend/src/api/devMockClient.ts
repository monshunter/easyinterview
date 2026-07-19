import deleteMeFixture from "../../../openapi/fixtures/Auth/deleteMe.json";
import updateMeFixture from "../../../openapi/fixtures/Auth/updateMe.json";
import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import logoutFixture from "../../../openapi/fixtures/Auth/logout.json";
import startAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/startAuthEmailChallenge.json";
import verifyAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/verifyAuthEmailChallenge.json";
import getJobFixture from "../../../openapi/fixtures/Jobs/getJob.json";
import createPracticePlanFixture from "../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import getPracticePlanFixture from "../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import sendPracticeMessageFixture from "../../../openapi/fixtures/PracticeSessions/sendPracticeMessage.json";
import completePracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/completePracticeSession.json";
import createPracticeVoiceTurnFixture from "../../../openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json";
import getPracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/getPracticeSession.json";
import startPracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/startPracticeSession.json";
import getPrivacyRequestFixture from "../../../openapi/fixtures/Privacy/getPrivacyRequest.json";
import requestPrivacyDeleteFixture from "../../../openapi/fixtures/Privacy/requestPrivacyDelete.json";
import requestPrivacyExportFixture from "../../../openapi/fixtures/Privacy/requestPrivacyExport.json";
import getFeedbackReportFixture from "../../../openapi/fixtures/Reports/getFeedbackReport.json";
import getReportConversationFixture from "../../../openapi/fixtures/Reports/getReportConversation.json";
import listTargetJobReportsFixture from "../../../openapi/fixtures/Reports/listTargetJobReports.json";
import regenerateFeedbackReportFixture from "../../../openapi/fixtures/Reports/regenerateFeedbackReport.json";
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
	updateMeFixture,
	getMeFixture,
	createPracticePlanFixture,
	getPracticePlanFixture,
	startPracticeSessionFixture,
	getPracticeSessionFixture,
	completePracticeSessionFixture,
	sendPracticeMessageFixture,
	createPracticeVoiceTurnFixture,
	requestPrivacyDeleteFixture,
	requestPrivacyExportFixture,
	getPrivacyRequestFixture,
	getFeedbackReportFixture,
	getReportConversationFixture,
	regenerateFeedbackReportFixture,
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
		email: "new.user@example.com",
		lastEmail: "",
		displayPreferences: { theme: "ocean", customAccent: null } as UserContext["displayPreferences"],
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
				authState.email = body.email;
			}
		}
		if (response.ok && request.method === "GET" && request.path === "/auth/email/verify") {
			authState.signedIn = true;
		}
		if (response.ok && request.method === "POST" && request.path === "/auth/logout") {
			authState.signedIn = false;
		}
		if (response.ok && request.method === "DELETE" && request.path === "/me") {
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
		email: string;
		lastEmail: string;
		displayPreferences: UserContext["displayPreferences"];
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
		const acceptedTerms = body?.acceptedTerms;
		const hasDisplayPreferences = body != null && Object.prototype.hasOwnProperty.call(body, "displayPreferences");
		const displayPreferences = isAccountDisplayPreferences(body?.displayPreferences)
			? body.displayPreferences
			: null;
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
		const hasProfileFields = body?.displayName !== undefined || body?.acceptedTerms !== undefined;
		if (
			(!hasProfileFields && !displayPreferences) ||
			(hasProfileFields && (!displayName || acceptedTerms !== true)) ||
			(hasDisplayPreferences && !displayPreferences)
		) {
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
		if (hasProfileFields) {
			state.displayName = displayName;
			state.profileComplete = true;
		}
		if (displayPreferences) state.displayPreferences = displayPreferences;
		return jsonResponse(200, buildMockUserContext(state));
	}
	return null;
}

function buildMockUserContext(state: {
	profileComplete: boolean;
	displayName: string;
	email: string;
	displayPreferences: UserContext["displayPreferences"];
}): UserContext {
	return {
		id: "01918fa0-0000-7000-8000-000000000101",
		email: state.email,
		displayName: state.profileComplete ? state.displayName || "Alice Example" : "",
		profileCompletionRequired: !state.profileComplete,
		displayPreferences: state.displayPreferences,
	};
}

function isAccountDisplayPreferences(value: unknown): value is UserContext["displayPreferences"] {
	if (!hasExactKeys(value, ["theme", "customAccent"]) || (value.theme !== "ocean" && value.theme !== "plum")) return false;
	if (value.customAccent === null) return true;
	return hasExactKeys(value.customAccent, ["h", "c"]) &&
		typeof value.customAccent.h === "number" && Number.isFinite(value.customAccent.h) && value.customAccent.h >= 0 && value.customAccent.h < 360 &&
		typeof value.customAccent.c === "number" && Number.isFinite(value.customAccent.c) && value.customAccent.c >= 0 && value.customAccent.c <= 0.28;
}

function jsonResponse(status: number, body: unknown): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { "Content-Type": "application/json" },
	});
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

function hasExactKeys<K extends string>(value: unknown, expected: readonly K[]): value is Record<K, unknown> {
	if (!isObject(value) || Array.isArray(value)) return false;
	const keys = Object.keys(value);
	return keys.length === expected.length && expected.every((key) => keys.includes(key));
}

function stripApiBase(path: string): string {
	return path.startsWith("/api/v1/") ? path.slice("/api/v1".length) : path;
}
