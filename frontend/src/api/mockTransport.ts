import { ALL_ROUTES, type OperationId, type Route } from "./generated/client";

export interface FixtureResponse {
	status: number;
	headers?: Record<string, string>;
	body?: unknown;
}

export interface FixtureScenario {
	response: FixtureResponse;
}

export interface OperationFixture {
	operationId: string;
	scenarios: Record<string, FixtureScenario>;
}

export type FixtureRegistry = Partial<Record<OperationId, OperationFixture>>;

export function createFixtureRegistry(fixtures: readonly OperationFixture[]): FixtureRegistry {
	const registry: FixtureRegistry = {};
	for (const fixture of fixtures) {
		if (!isOperationId(fixture.operationId)) {
			throw new Error(`fixture operationId is not generated: ${fixture.operationId}`);
		}
		registry[fixture.operationId] = fixture;
	}
	return registry;
}

export interface FixtureBackedFetchOptions {
	scenario?: string;
}

export function createFixtureBackedFetch(
	registry: FixtureRegistry,
	options: FixtureBackedFetchOptions = {},
): typeof fetch {
	return async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
		const request = normalizeRequest(input, init);
		const route = matchRoute(request.method, request.url);
		if (!route) {
			throw new Error(`no fixture route matched ${request.method} ${request.url}`);
		}
		const fixture = registry[route.operationId];
		if (!fixture) {
			throw new Error(`missing fixture for operationId: ${route.operationId}`);
		}
		const scenarioName = selectScenario(request.headers, options.scenario);
		const scenario = fixture.scenarios[scenarioName];
		if (!scenario) {
			throw new Error(`unknown fixture scenario ${scenarioName} for operationId: ${route.operationId}`);
		}
		const headers = new Headers(scenario.response.headers ?? {});
		if (scenario.response.body !== undefined && !headers.has("Content-Type")) {
			headers.set("Content-Type", "application/json");
		}
		return new Response(
			scenario.response.body === undefined ? null : JSON.stringify(scenario.response.body),
			{ status: scenario.response.status, headers },
		);
	};
}

function normalizeRequest(input: RequestInfo | URL, init?: RequestInit): { method: string; url: string; headers: Headers } {
	if (input instanceof Request) {
		const headers = new Headers(input.headers);
		new Headers(init?.headers).forEach((value, key) => headers.set(key, value));
		return {
			method: init?.method ?? input.method,
			url: input.url,
			headers,
		};
	}
	return {
		method: init?.method ?? "GET",
		url: input instanceof URL ? input.href : input,
		headers: new Headers(init?.headers),
	};
}

function isOperationId(value: string): value is OperationId {
	return (ALL_ROUTES as readonly Route[]).some((route) => route.operationId === value);
}

function matchRoute(method: string, url: string): Route | undefined {
	const requestPath = stripBasePath(new URL(url, "http://fixture.local").pathname);
	return (ALL_ROUTES as readonly Route[]).find((route) => {
		return route.method === method.toUpperCase() && pathTemplateMatches(route.path, requestPath);
	});
}

function stripBasePath(path: string): string {
	return path.startsWith("/api/v1/") ? path.slice("/api/v1".length) : path;
}

function pathTemplateMatches(template: string, actual: string): boolean {
	const templateParts = template.split("/").filter(Boolean);
	const actualParts = actual.split("/").filter(Boolean);
	if (templateParts.length !== actualParts.length) return false;
	return templateParts.every((part, index) => {
		return part.startsWith("{") && part.endsWith("}") || part === actualParts[index];
	});
}

function selectScenario(headers: Headers, fallback = "default"): string {
	const prefer = headers.get("Prefer");
	if (!prefer) return fallback;
	for (const part of prefer.split(",")) {
		const trimmed = part.trim();
		if (trimmed.startsWith("example=")) {
			return trimmed.slice("example=".length).trim();
		}
	}
	return fallback;
}
