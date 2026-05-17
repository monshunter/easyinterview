import deleteMeFixture from "../../../openapi/fixtures/Auth/deleteMe.json";
import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import logoutFixture from "../../../openapi/fixtures/Auth/logout.json";
import startAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/startAuthEmailChallenge.json";
import verifyAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/verifyAuthEmailChallenge.json";
import createDebriefFixture from "../../../openapi/fixtures/Debriefs/createDebrief.json";
import getDebriefFixture from "../../../openapi/fixtures/Debriefs/getDebrief.json";
import suggestDebriefQuestionsFixture from "../../../openapi/fixtures/Debriefs/suggestDebriefQuestions.json";
import getJobFixture from "../../../openapi/fixtures/Jobs/getJob.json";
import addToWatchlistFixture from "../../../openapi/fixtures/JobMatch/addToWatchlist.json";
import createSavedSearchFixture from "../../../openapi/fixtures/JobMatch/createSavedSearch.json";
import getAgentScanStatusFixture from "../../../openapi/fixtures/JobMatch/getAgentScanStatus.json";
import getJobMatchProfileFixture from "../../../openapi/fixtures/JobMatch/getJobMatchProfile.json";
import getJobRecommendationFixture from "../../../openapi/fixtures/JobMatch/getJobRecommendation.json";
import getMarketSignalsFixture from "../../../openapi/fixtures/JobMatch/getMarketSignals.json";
import listJobRecommendationsFixture from "../../../openapi/fixtures/JobMatch/listJobRecommendations.json";
import listSavedSearchesFixture from "../../../openapi/fixtures/JobMatch/listSavedSearches.json";
import listWatchlistFixture from "../../../openapi/fixtures/JobMatch/listWatchlist.json";
import markJobNotRelevantFixture from "../../../openapi/fixtures/JobMatch/markJobNotRelevant.json";
import removeFromWatchlistFixture from "../../../openapi/fixtures/JobMatch/removeFromWatchlist.json";
import searchJobsFixture from "../../../openapi/fixtures/JobMatch/searchJobs.json";
import createPracticePlanFixture from "../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import getPracticePlanFixture from "../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import appendSessionEventFixture from "../../../openapi/fixtures/PracticeSessions/appendSessionEvent.json";
import completePracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/completePracticeSession.json";
import createPracticeVoiceTurnFixture from "../../../openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json";
import getPracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/getPracticeSession.json";
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
	createDebriefFixture,
	getDebriefFixture,
	suggestDebriefQuestionsFixture,
	getJobFixture,
	getJobMatchProfileFixture,
	getAgentScanStatusFixture,
	listJobRecommendationsFixture,
	getJobRecommendationFixture,
	markJobNotRelevantFixture,
	listWatchlistFixture,
	addToWatchlistFixture,
	removeFromWatchlistFixture,
	searchJobsFixture,
	listSavedSearchesFixture,
	createSavedSearchFixture,
	getMarketSignalsFixture,
	deleteMeFixture,
	getMeFixture,
	createPracticePlanFixture,
	getPracticePlanFixture,
	startPracticeSessionFixture,
	getPracticeSessionFixture,
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
	let signedIn = false;
	const fetch: typeof globalThis.fetch = async (input, init) => {
		const request = readRequest(input, init);
		const nextInit = withStatefulAuthScenario(init, request, signedIn);
		const response = await fixtureFetch(input, nextInit);
		if (response.ok && request.method === "GET" && request.path === "/auth/email/verify") {
			signedIn = true;
		}
		if (response.ok && request.method === "POST" && request.path === "/auth/logout") {
			signedIn = false;
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
): { method: string; path: string; headers: Headers } {
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
	};
}

function withStatefulAuthScenario(
	init: RequestInit | undefined,
	request: { method: string; path: string; headers: Headers },
	signedIn: boolean,
): RequestInit | undefined {
	if (
		request.method !== "GET" ||
		request.path !== "/me" ||
		request.headers.has("Prefer")
	) {
		return init;
	}
	const headers = new Headers(init?.headers);
	headers.set("Prefer", `example=${signedIn ? "authenticated" : "unauthenticated"}`);
	return { ...init, headers };
}

function stripApiBase(path: string): string {
	return path.startsWith("/api/v1/") ? path.slice("/api/v1".length) : path;
}
