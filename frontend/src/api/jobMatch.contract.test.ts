import { describe, expect, it, vi } from "vitest";

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
import { EasyInterviewClient } from "./generated/client";
import type {
	AgentScanStatus,
	JobMatchProfile,
	JobMatchRecommendation,
	MarkNotRelevantResult,
	MarketSignalsResponse,
	PaginatedJobMatchRecommendation,
	SavedSearch,
	SavedSearchesResponse,
	SearchJobsResponse,
	WatchlistItem,
	WatchlistResponse,
} from "./generated/types";
import { createFixtureBackedFetch, createFixtureRegistry } from "./mockTransport";

const ALL_JOB_MATCH_FIXTURES = [
	getJobMatchProfileFixture,
	getAgentScanStatusFixture,
	listJobRecommendationsFixture,
	getJobRecommendationFixture,
	addToWatchlistFixture,
	removeFromWatchlistFixture,
	markJobNotRelevantFixture,
	searchJobsFixture,
	listSavedSearchesFixture,
	createSavedSearchFixture,
	listWatchlistFixture,
	getMarketSignalsFixture,
];

const SIDE_EFFECT_FIXTURES: Array<{ id: string; fixture: typeof addToWatchlistFixture }> = [
	{ id: "addToWatchlist", fixture: addToWatchlistFixture },
	{ id: "removeFromWatchlist", fixture: removeFromWatchlistFixture },
	{ id: "markJobNotRelevant", fixture: markJobNotRelevantFixture },
	{ id: "searchJobs", fixture: searchJobsFixture },
	{ id: "createSavedSearch", fixture: createSavedSearchFixture },
];

function buildClientWithSpy(): { client: EasyInterviewClient; spy: ReturnType<typeof vi.fn> } {
	const inner = createFixtureBackedFetch(createFixtureRegistry(ALL_JOB_MATCH_FIXTURES));
	const spy = vi.fn(inner);
	return { client: new EasyInterviewClient({ fetch: spy }), spy };
}

describe("JobMatch generated client × fixture contract", () => {
	it("12 JobMatch operations are present on the generated client", () => {
		const { client } = buildClientWithSpy();
		const surface = client as unknown as Record<string, unknown>;
		const required = [
			"getJobMatchProfile",
			"getAgentScanStatus",
			"listJobRecommendations",
			"getJobRecommendation",
			"addToWatchlist",
			"removeFromWatchlist",
			"markJobNotRelevant",
			"searchJobs",
			"listSavedSearches",
			"createSavedSearch",
			"listWatchlist",
			"getMarketSignals",
		];
		for (const name of required) {
			expect(typeof surface[name]).toBe("function");
		}
	});

	it("getJobMatchProfile returns the typed profile snapshot", async () => {
		const { client } = buildClientWithSpy();
		const profile: JobMatchProfile = await client.getJobMatchProfile();
		expect(profile.displayName).toBe("Alice Example");
		expect(profile.skills.length).toBeGreaterThan(0);
		expect(profile.sources).toEqual({ resumes: 2, jds: 5, mocks: 4, debriefs: 1 });

		const partial: JobMatchProfile = await client.getJobMatchProfile({
			headers: { Prefer: "example=partial-profile" },
		});
		expect(partial.skills).toEqual([]);
		expect(partial.sources.resumes).toBe(0);
	});

	it("getAgentScanStatus exposes idle / scanning / error variants", async () => {
		const { client } = buildClientWithSpy();
		const idle: AgentScanStatus = await client.getAgentScanStatus();
		expect(idle.status).toBe("idle");
		const scanning: AgentScanStatus = await client.getAgentScanStatus({
			headers: { Prefer: "example=scanning" },
		});
		expect(scanning.status).toBe("scanning");
		const errored: AgentScanStatus = await client.getAgentScanStatus({
			headers: { Prefer: "example=error" },
		});
		expect(errored.status).toBe("error");
		expect(errored.message).not.toBeNull();
	});

	it("listJobRecommendations covers default / empty / one / many variants", async () => {
		const { client } = buildClientWithSpy();
		const def: PaginatedJobMatchRecommendation = await client.listJobRecommendations();
		expect(def.items.length).toBeGreaterThan(1);
		const empty: PaginatedJobMatchRecommendation = await client.listJobRecommendations({
			headers: { Prefer: "example=empty" },
		});
		expect(empty.items).toEqual([]);
		const one: PaginatedJobMatchRecommendation = await client.listJobRecommendations({
			headers: { Prefer: "example=one" },
		});
		expect(one.items.length).toBe(1);
		const many: PaginatedJobMatchRecommendation = await client.listJobRecommendations({
			headers: { Prefer: "example=many" },
		});
		expect(many.items.length).toBeGreaterThanOrEqual(4);
		const first = many.items[0];
		expect(first?.provenance.modelId).toMatch(/^model-profile:/);
		expect(first?.fit.must).toBeGreaterThan(0);
	});

	it("getJobRecommendation default carries provenance and intel fields", async () => {
		const { client } = buildClientWithSpy();
		const detail: JobMatchRecommendation = await client.getJobRecommendation(
			"01918fa0-0000-7000-8000-00000000a001",
		);
		expect(detail.id).toBe("01918fa0-0000-7000-8000-00000000a001");
		expect(detail.networkNote).not.toBeNull();
		expect(detail.similarInterviewers).not.toBeNull();
		expect(detail.provenance.promptVersion).toBeTruthy();
		const noIntel: JobMatchRecommendation = await client.getJobRecommendation(
			"01918fa0-0000-7000-8000-00000000a003",
			{ headers: { Prefer: "example=network-intel-empty" } },
		);
		expect(noIntel.networkNote).toBeNull();
		expect(noIntel.similarInterviewers).toBeNull();
	});

	it("addToWatchlist sends Idempotency-Key and returns a watchlist row", async () => {
		const { client, spy } = buildClientWithSpy();
		const item: WatchlistItem = await client.addToWatchlist(
			{ jobMatchId: "01918fa0-0000-7000-8000-00000000a001" },
			{ idempotencyKey: "ik_test_add" },
		);
		expect(item.linkedJobMatchId).toBe("01918fa0-0000-7000-8000-00000000a001");
		expect(item.tone).toBe("ok");
		const init = (spy.mock.calls[0]?.[1] ?? {}) as RequestInit;
		const headers = new Headers(init.headers ?? {});
		expect(headers.get("Idempotency-Key")).toBe("ik_test_add");
		expect(init.method).toBe("POST");
	});

	it("removeFromWatchlist accepts Idempotency-Key and returns 204 (void)", async () => {
		const { client, spy } = buildClientWithSpy();
		await client.removeFromWatchlist(
			"01918fa0-0000-7000-8000-00000000a001",
			{ idempotencyKey: "ik_test_remove" },
		);
		const init = (spy.mock.calls[0]?.[1] ?? {}) as RequestInit;
		const headers = new Headers(init.headers ?? {});
		expect(headers.get("Idempotency-Key")).toBe("ik_test_remove");
		expect(init.method).toBe("DELETE");
	});

	it("markJobNotRelevant sends Idempotency-Key and request body with reason", async () => {
		const { client, spy } = buildClientWithSpy();
		const result: MarkNotRelevantResult = await client.markJobNotRelevant(
			"01918fa0-0000-7000-8000-00000000a002",
			{ reason: "not_relevant" },
			{ idempotencyKey: "ik_test_dismiss" },
		);
		expect(result.jobMatchId).toBe("01918fa0-0000-7000-8000-00000000a002");
		const init = (spy.mock.calls[0]?.[1] ?? {}) as RequestInit;
		const headers = new Headers(init.headers ?? {});
		expect(headers.get("Idempotency-Key")).toBe("ik_test_dismiss");
		expect(headers.get("Content-Type")).toBe("application/json");
		expect(JSON.parse(String(init.body))).toEqual({ reason: "not_relevant" });
	});

	it("searchJobs sends Idempotency-Key and request body with query", async () => {
		const { client, spy } = buildClientWithSpy();
		const response: SearchJobsResponse = await client.searchJobs(
			{ query: "Senior frontend roles with strong design-system culture" },
			{ idempotencyKey: "ik_test_search" },
		);
		expect(response.searchRunId).toBeTruthy();
		expect(response.items.length).toBeGreaterThan(0);
		const init = (spy.mock.calls[0]?.[1] ?? {}) as RequestInit;
		const headers = new Headers(init.headers ?? {});
		expect(headers.get("Idempotency-Key")).toBe("ik_test_search");
		expect(JSON.parse(String(init.body))).toMatchObject({
			query: "Senior frontend roles with strong design-system culture",
		});
	});

	it("listSavedSearches default returns saved-searches collection", async () => {
		const { client } = buildClientWithSpy();
		const response: SavedSearchesResponse = await client.listSavedSearches();
		expect(response.items.length).toBeGreaterThan(0);
		expect(response.items[0]?.label).toBeTruthy();
		const empty: SavedSearchesResponse = await client.listSavedSearches({
			headers: { Prefer: "example=empty" },
		});
		expect(empty.items).toEqual([]);
	});

	it("createSavedSearch sends Idempotency-Key and returns SavedSearch", async () => {
		const { client, spy } = buildClientWithSpy();
		const saved: SavedSearch = await client.createSavedSearch(
			{
				label: "Design-system leadership",
				query: "Senior frontend roles with strong design-system culture",
				filters: { remote: true, minScore: 80 },
			},
			{ idempotencyKey: "ik_test_create_saved" },
		);
		expect(saved.label).toBe("Design-system leadership");
		const init = (spy.mock.calls[0]?.[1] ?? {}) as RequestInit;
		const headers = new Headers(init.headers ?? {});
		expect(headers.get("Idempotency-Key")).toBe("ik_test_create_saved");
	});

	it("listWatchlist exposes default + empty + few variants", async () => {
		const { client } = buildClientWithSpy();
		const def: WatchlistResponse = await client.listWatchlist();
		expect(def.items.length).toBeGreaterThan(1);
		expect(def.items[0]?.linkedJobMatchId).toBeTruthy();
		const empty: WatchlistResponse = await client.listWatchlist({
			headers: { Prefer: "example=empty" },
		});
		expect(empty.items).toEqual([]);
		const few: WatchlistResponse = await client.listWatchlist({
			headers: { Prefer: "example=few" },
		});
		expect(few.items.length).toBe(1);
	});

	it("getMarketSignals returns 4-card grid in default and partial-data variants", async () => {
		const { client } = buildClientWithSpy();
		const def: MarketSignalsResponse = await client.getMarketSignals();
		expect(def.signals.length).toBe(4);
		expect(def.asOf).toBeTruthy();
		const partial: MarketSignalsResponse = await client.getMarketSignals({
			headers: { Prefer: "example=partial-data" },
		});
		expect(partial.signals.length).toBeLessThan(4);
		expect(partial.asOf).toBeNull();
	});

	it("five JobMatch side-effect fixtures declare Idempotency-Key requests", () => {
		for (const { id, fixture } of SIDE_EFFECT_FIXTURES) {
			const def = fixture.scenarios.default;
			expect(def, `${id} default scenario must be declared`).toBeDefined();
			const headers = (def?.request as { headers?: Record<string, string> } | undefined)?.headers ?? {};
			expect(headers["Idempotency-Key"], `${id}.default.request.headers.Idempotency-Key`).toMatch(/^ik_/);
		}
	});
});
