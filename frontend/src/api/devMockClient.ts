import deleteMeFixture from "../../../openapi/fixtures/Auth/deleteMe.json";
import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import logoutFixture from "../../../openapi/fixtures/Auth/logout.json";
import startAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/startAuthEmailChallenge.json";
import verifyAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/verifyAuthEmailChallenge.json";
import createDebriefFixture from "../../../openapi/fixtures/Debriefs/createDebrief.json";
import getDebriefFixture from "../../../openapi/fixtures/Debriefs/getDebrief.json";
import getJobFixture from "../../../openapi/fixtures/Jobs/getJob.json";
import createPracticePlanFixture from "../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import getPracticePlanFixture from "../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import appendSessionEventFixture from "../../../openapi/fixtures/PracticeSessions/appendSessionEvent.json";
import completePracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/completePracticeSession.json";
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
import getResumeFixture from "../../../openapi/fixtures/Resumes/getResume.json";
import registerResumeFixture from "../../../openapi/fixtures/Resumes/registerResume.json";
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
	getJobFixture,
	deleteMeFixture,
	getMeFixture,
	createPracticePlanFixture,
	getPracticePlanFixture,
	startPracticeSessionFixture,
	getPracticeSessionFixture,
	completePracticeSessionFixture,
	appendSessionEventFixture,
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
	return new EasyInterviewClient({
		fetch: createFixtureBackedFetch(createDevMockFixtureRegistry(), options),
	});
}
