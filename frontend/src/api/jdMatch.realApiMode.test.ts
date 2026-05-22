import { afterEach, describe, expect, it, vi } from "vitest";

import { createAppClient } from "./clientFactory";
import type {
	AgentScanStatus,
	JobMatchProfile,
	JobMatchRecommendation,
	MarketSignalsResponse,
	PaginatedJobMatchRecommendation,
	SavedSearch,
	SavedSearchesResponse,
	SearchJobsResponse,
	WatchlistItem,
	WatchlistResponse,
} from "./generated/types";

const API_BASE_URL = "http://api.test/api/v1";
const JOB_MATCH_ID = "01918fa0-0000-7000-8000-00000000a001";
const PROVENANCE = {
	dataSourceVersion: "jdmatch-live-test-2026-05-22",
	featureFlag: "jd_match.real_backend",
	language: "en",
	modelId: "model-profile:jd-match-live",
	promptVersion: "jd-match-search.v1",
	rubricVersion: "jd-match-rubric.v1",
};

function buildRecommendation(overrides: Partial<JobMatchRecommendation> = {}): JobMatchRecommendation {
	return {
		id: JOB_MATCH_ID,
		title: "Senior Frontend Engineer",
		company: "Acme AI",
		companyTag: "AI productivity",
		level: "Senior",
		location: "Remote",
		comp: "$180k",
		posted: "2 days ago",
		score: 91,
		fit: { must: 5, total: 6, plus: 2, totalPlus: 3 },
		reasons: ["React and product engineering depth"],
		risks: ["Needs systems design depth"],
		highlights: ["Owns high-traffic UI systems"],
		seen: false,
		saved: false,
		sourceUrl: "https://jobs.example.test/acme-frontend",
		sourceLabel: "jobs.example.test",
		networkNote: "Public interview-review, JD, and company-source signals.",
		similarInterviewers: 3,
		interviewHypotheses: ["Expect design-system tradeoff questions."],
		provenance: PROVENANCE,
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
		const recommendation = buildRecommendation();

		if (route === "/jd-match/profile" && method === "GET") {
			const body: JobMatchProfile = {
				displayName: "Alice Example",
				headline: "Frontend platform lead",
				locationText: "San Francisco, CA",
				skills: ["React", "TypeScript", "Design systems"],
				sources: { resumes: 2, jds: 5, mocks: 4, debriefs: 1 },
				yearsOfExperience: 8,
			};
			return jsonResponse(body);
		}
		if (route === "/jd-match/agent-status" && method === "GET") {
			const body: AgentScanStatus = {
				status: "idle",
				lastScanAt: "2026-05-22T02:00:00Z",
				nextScanAt: "2026-05-22T06:00:00Z",
			};
			return jsonResponse(body);
		}
		if (route === "/jd-match/recommendations" && method === "GET") {
			const body: PaginatedJobMatchRecommendation = {
				items: [recommendation],
				pageInfo: { hasMore: false, nextCursor: null, pageSize: 20 },
			};
			return jsonResponse(body);
		}
		if (route === `/jd-match/recommendations/${JOB_MATCH_ID}` && method === "GET") {
			return jsonResponse(recommendation);
		}
		if (route === `/jd-match/recommendations/${JOB_MATCH_ID}/dismiss` && method === "POST") {
			return jsonResponse({
				jobMatchId: JOB_MATCH_ID,
				dismissedAt: "2026-05-22T03:00:00Z",
			});
		}
		if (route === "/jd-match/watchlist" && method === "GET") {
			const body: WatchlistResponse = {
				items: [
					{
						id: "01918fa0-0000-7000-8000-00000000b001",
						linkedJobMatchId: JOB_MATCH_ID,
						title: recommendation.title,
						company: recommendation.company,
						tone: "ok",
						addedAt: "2026-05-22T03:10:00Z",
					},
				],
			};
			return jsonResponse(body);
		}
		if (route === "/jd-match/watchlist" && method === "POST") {
			const body: WatchlistItem = {
				id: "01918fa0-0000-7000-8000-00000000b001",
				linkedJobMatchId: JOB_MATCH_ID,
				title: recommendation.title,
				company: recommendation.company,
				tone: "ok",
				addedAt: "2026-05-22T03:10:00Z",
			};
			return jsonResponse(body);
		}
		if (route === `/jd-match/watchlist/${JOB_MATCH_ID}` && method === "DELETE") {
			return new Response(null, { status: 204 });
		}
		if (route === "/jd-match/search" && method === "POST") {
			const body: SearchJobsResponse = {
				searchRunId: "01918fa0-0000-7000-8000-00000000c001",
				items: [buildRecommendation({ id: "01918fa0-0000-7000-8000-00000000a002" })],
			};
			return jsonResponse(body);
		}
		if (route === "/jd-match/saved-searches" && method === "GET") {
			const body: SavedSearchesResponse = {
				items: [
					{
						id: "01918fa0-0000-7000-8000-00000000d001",
						label: "Frontend platform",
						query: "frontend platform remote",
						createdAt: "2026-05-22T03:20:00Z",
					},
				],
			};
			return jsonResponse(body);
		}
		if (route === "/jd-match/saved-searches" && method === "POST") {
			const body: SavedSearch = {
				id: "01918fa0-0000-7000-8000-00000000d001",
				label: "Frontend platform",
				query: "frontend platform remote",
				filters: { remote: true },
				createdAt: "2026-05-22T03:20:00Z",
			};
			return jsonResponse(body);
		}
		if (route === "/jd-match/market-signals" && method === "GET") {
			const body: MarketSignalsResponse = {
				signals: [
					{ k: "Active roles", v: "14", d: "+3", tone: "ok" },
					{ k: "Remote fit", v: "High", d: null, tone: "muted" },
				],
				asOf: "2026-05-22T03:30:00Z",
			};
			return jsonResponse(body);
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

describe("JobMatch real API mode", () => {
	it("routes all 12 JobMatch operations through the real generated client with auth, IK, and provenance", async () => {
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

		const profile = await client.getJobMatchProfile();
		const agentStatus = await client.getAgentScanStatus();
		const recommendations = await client.listJobRecommendations({ query: { pageSize: 20 } });
		const detail = await client.getJobRecommendation(JOB_MATCH_ID);
		const addResult = await client.addToWatchlist(
			{ jobMatchId: JOB_MATCH_ID },
			{ idempotencyKey: "ik_real_add" },
		);
		await client.removeFromWatchlist(JOB_MATCH_ID, { idempotencyKey: "ik_real_remove" });
		const dismissResult = await client.markJobNotRelevant(
			JOB_MATCH_ID,
			{ reason: "not_relevant" },
			{ idempotencyKey: "ik_real_dismiss" },
		);
		const searchResult = await client.searchJobs(
			{ query: "frontend platform remote", filters: { remote: true } },
			{ idempotencyKey: "ik_real_search" },
		);
		const savedSearches = await client.listSavedSearches();
		const savedSearch = await client.createSavedSearch(
			{ label: "Frontend platform", query: "frontend platform remote", filters: { remote: true } },
			{ idempotencyKey: "ik_real_saved" },
		);
		const watchlist = await client.listWatchlist();
		const marketSignals = await client.getMarketSignals({ query: { window: "7d" } });

		expect(profile.displayName).toBe("Alice Example");
		expect(agentStatus.status).toBe("idle");
		expect(recommendations.items[0]?.provenance).toEqual(PROVENANCE);
		expect(detail.provenance.modelId).toBe("model-profile:jd-match-live");
		expect(addResult.linkedJobMatchId).toBe(JOB_MATCH_ID);
		expect(dismissResult.jobMatchId).toBe(JOB_MATCH_ID);
		expect(searchResult.items[0]?.provenance.promptVersion).toBe("jd-match-search.v1");
		expect(savedSearches.items[0]?.label).toBe("Frontend platform");
		expect(savedSearch.filters?.remote).toBe(true);
		expect(watchlist.items[0]?.linkedJobMatchId).toBe(JOB_MATCH_ID);
		expect(marketSignals.signals.length).toBe(2);

		const summary = callSummary(fetchSpy);
		expect(summary.map(({ method, path }) => `${method} ${path}`)).toEqual([
			"GET /api/v1/jd-match/profile",
			"GET /api/v1/jd-match/agent-status",
			"GET /api/v1/jd-match/recommendations?pageSize=20",
			`GET /api/v1/jd-match/recommendations/${JOB_MATCH_ID}`,
			"POST /api/v1/jd-match/watchlist",
			`DELETE /api/v1/jd-match/watchlist/${JOB_MATCH_ID}`,
			`POST /api/v1/jd-match/recommendations/${JOB_MATCH_ID}/dismiss`,
			"POST /api/v1/jd-match/search",
			"GET /api/v1/jd-match/saved-searches",
			"POST /api/v1/jd-match/saved-searches",
			"GET /api/v1/jd-match/watchlist",
			"GET /api/v1/jd-match/market-signals?window=7d",
		]);
		expect(fetchSpy.mock.calls.map(([input]) => String(input))).toEqual([
			`${API_BASE_URL}/jd-match/profile`,
			`${API_BASE_URL}/jd-match/agent-status`,
			`${API_BASE_URL}/jd-match/recommendations?pageSize=20`,
			`${API_BASE_URL}/jd-match/recommendations/${JOB_MATCH_ID}`,
			`${API_BASE_URL}/jd-match/watchlist`,
			`${API_BASE_URL}/jd-match/watchlist/${JOB_MATCH_ID}`,
			`${API_BASE_URL}/jd-match/recommendations/${JOB_MATCH_ID}/dismiss`,
			`${API_BASE_URL}/jd-match/search`,
			`${API_BASE_URL}/jd-match/saved-searches`,
			`${API_BASE_URL}/jd-match/saved-searches`,
			`${API_BASE_URL}/jd-match/watchlist`,
			`${API_BASE_URL}/jd-match/market-signals?window=7d`,
		]);

		for (const call of summary) {
			expect(call.credentials).toBe("include");
			expect(call.headers.get("Prefer")).toBeNull();
		}
		expect(summary[4]?.headers.get("Idempotency-Key")).toBe("ik_real_add");
		expect(summary[5]?.headers.get("Idempotency-Key")).toBe("ik_real_remove");
		expect(summary[6]?.headers.get("Idempotency-Key")).toBe("ik_real_dismiss");
		expect(summary[7]?.headers.get("Idempotency-Key")).toBe("ik_real_search");
		expect(summary[9]?.headers.get("Idempotency-Key")).toBe("ik_real_saved");
		expect(JSON.parse(String(summary[7]?.body))).toMatchObject({
			query: "frontend platform remote",
			filters: { remote: true },
		});
	});
});
