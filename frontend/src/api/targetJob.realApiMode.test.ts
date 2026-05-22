import { afterEach, describe, expect, it, vi } from "vitest";

import { createAppClient } from "./clientFactory";
import type {
	Job,
	PaginatedTargetJob,
	TargetJob,
	UploadPresign,
} from "./generated/types";

const API_BASE_URL = "http://api.test/api/v1";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000000014";
const FILE_OBJECT_ID = "01918fa0-0000-7000-8000-000000000015";
const RAW_JD_TEXT = "Lead React platform and design system programs.";
const SOURCE_URL = "https://jobs.example.test/frontend-platform?secret=redacted";
const PROVENANCE = {
	dataSourceVersion: "targetjob-live-test-2026-05-22",
	featureFlag: "targetjob.real_backend",
	language: "en",
	modelId: "model-profile:target-import-live",
	promptVersion: "target.import.parse.v1",
	rubricVersion: "target-import-rubric.v1",
};

function buildTargetJob(overrides: Partial<TargetJob> = {}): TargetJob {
	return {
		id: TARGET_JOB_ID,
		status: "preparing",
		analysisStatus: "ready",
		title: "Senior Frontend Engineer",
		companyName: "Acme AI",
		locationText: "Remote",
		targetLanguage: "en",
		sourceType: "manual_text",
		sourceUrl: null,
		requirements: [
			{
				id: "01918fa0-0000-7000-8000-000000000016",
				kind: "must_have",
				label: "React platform leadership",
				evidenceLevel: "explicit",
			},
		],
		summary: {
			coreThemes: ["frontend platform"],
			interviewHypotheses: ["Expect design-system tradeoff questions."],
			provenance: PROVENANCE,
		},
		fitSummary: {
			strengths: ["Design systems"],
			gaps: ["Backend depth"],
			riskSignals: ["Distributed systems scope"],
			provenance: PROVENANCE,
		},
		latestReportId: null,
		openQuestionIssueCount: 0,
		createdAt: "2026-05-22T03:00:00Z",
		updatedAt: "2026-05-22T03:30:00Z",
		...overrides,
	};
}

function buildTargetImportJob(): Job {
	return {
		id: "01918fa0-0000-7000-8000-000000000017",
		jobType: "target_import",
		resourceId: TARGET_JOB_ID,
		resourceType: "target_job",
		status: "queued",
		createdAt: "2026-05-22T03:00:00Z",
		updatedAt: "2026-05-22T03:00:00Z",
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
		const targetJob = buildTargetJob();

		if (route === "/targets" && method === "GET") {
			const body: PaginatedTargetJob = {
				items: [targetJob],
				pageInfo: { hasMore: false, nextCursor: null, pageSize: 12 },
			};
			return jsonResponse(body);
		}
		if (route === "/uploads/presign" && method === "POST") {
			const body: UploadPresign = {
				fileObjectId: FILE_OBJECT_ID,
				uploadUrl: "https://uploads.example.test/target-job-attachment",
				method: "PUT",
				headers: { "Content-Type": "application/pdf" },
				expiresAt: "2026-05-22T04:00:00Z",
			};
			return jsonResponse(body, 201);
		}
		if (route === "/targets/import" && method === "POST") {
			return jsonResponse({ targetJobId: TARGET_JOB_ID, job: buildTargetImportJob() }, 202);
		}
		if (route === `/targets/${TARGET_JOB_ID}` && method === "GET") {
			return jsonResponse(targetJob);
		}
		if (route === `/targets/${TARGET_JOB_ID}` && method === "PATCH") {
			return jsonResponse(buildTargetJob({ locationText: "San Francisco, CA" }));
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

describe("TargetJob real API mode", () => {
	it("routes home/import/parse operations through the real generated client with auth, IK, and provenance", async () => {
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

		const targets = await client.listTargetJobs({ query: { pageSize: 12 } });
		const upload = await client.createUploadPresign(
			{
				purpose: "target_job_attachment",
				fileName: "frontend-platform.pdf",
				contentType: "application/pdf",
				byteSize: 42_000,
			},
			{ idempotencyKey: "ik_real_upload" },
		);
		const imported = await client.importTargetJob(
			{
				source: { type: "manual_text", rawText: RAW_JD_TEXT },
				targetLanguage: "en",
				titleHint: "Senior Frontend Engineer",
			},
			{ idempotencyKey: "ik_real_import" },
		);
		const detail = await client.getTargetJob(TARGET_JOB_ID);
		const updated = await client.updateTargetJob(
			TARGET_JOB_ID,
			{ locationText: "San Francisco, CA", notes: "Focus on platform scope." },
			{ idempotencyKey: "ik_real_update" },
		);

		expect(targets.items[0]?.summary?.provenance).toEqual(PROVENANCE);
		expect(upload.fileObjectId).toBe(FILE_OBJECT_ID);
		expect(imported.targetJobId).toBe(TARGET_JOB_ID);
		expect(imported.job.jobType).toBe("target_import");
		expect(detail.fitSummary?.provenance.promptVersion).toBe("target.import.parse.v1");
		expect(updated.locationText).toBe("San Francisco, CA");

		const summary = callSummary(fetchSpy);
		expect(summary.map(({ method, path }) => `${method} ${path}`)).toEqual([
			"GET /api/v1/targets?pageSize=12",
			"POST /api/v1/uploads/presign",
			"POST /api/v1/targets/import",
			`GET /api/v1/targets/${TARGET_JOB_ID}`,
			`PATCH /api/v1/targets/${TARGET_JOB_ID}`,
		]);
		expect(fetchSpy.mock.calls.map(([input]) => String(input))).toEqual([
			`${API_BASE_URL}/targets?pageSize=12`,
			`${API_BASE_URL}/uploads/presign`,
			`${API_BASE_URL}/targets/import`,
			`${API_BASE_URL}/targets/${TARGET_JOB_ID}`,
			`${API_BASE_URL}/targets/${TARGET_JOB_ID}`,
		]);

		for (const call of summary) {
			expect(call.credentials).toBe("include");
			expect(call.headers.get("Prefer")).toBeNull();
			expect(`${call.path}${String(call.body ?? "")}`).not.toContain(SOURCE_URL);
		}
		expect(summary[1]?.headers.get("Idempotency-Key")).toBe("ik_real_upload");
		expect(summary[2]?.headers.get("Idempotency-Key")).toBe("ik_real_import");
		expect(summary[4]?.headers.get("Idempotency-Key")).toBe("ik_real_update");
		expect(String(summary[0]?.body ?? "")).not.toContain(RAW_JD_TEXT);
		expect(summary[2]?.path).not.toContain(RAW_JD_TEXT);
		expect(JSON.parse(String(summary[2]?.body))).toMatchObject({
			source: { type: "manual_text", rawText: RAW_JD_TEXT },
			targetLanguage: "en",
		});
	});
});
