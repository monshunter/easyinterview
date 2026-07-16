import { describe, expect, it, vi } from "vitest";

import { ApiClientError, EasyInterviewClient } from "./generated/client";

describe("EasyInterviewClient response parsing", () => {
	it("sends failed-report regeneration as a header-only POST and accepts the 202 envelope", async () => {
		const reportId = "01918fa0-0000-7000-8000-000000007102";
		const idempotencyKey =
			"v1.1784163600.01918fa0-0000-7000-8000-000000008102";
		const accepted = {
			reportId,
			job: {
				id: "01918fa0-0000-7000-8000-000000008103",
				jobType: "report_generate",
				resourceType: "feedback_report",
				resourceId: reportId,
				status: "queued",
				createdAt: "2026-07-16T09:00:00Z",
				updatedAt: "2026-07-16T09:00:00Z",
			},
		};
		const fetchSpy = vi.fn<typeof fetch>(async () => jsonResponse(accepted, 202));
		const client = new EasyInterviewClient({ fetch: fetchSpy });

		await expect(
			client.regenerateFeedbackReport(reportId, { idempotencyKey }),
		).resolves.toEqual(accepted);
		expect(fetchSpy).toHaveBeenCalledTimes(1);
		const [url, init] = fetchSpy.mock.calls[0]!;
		expect(url).toBe(`/api/v1/reports/${reportId}/regenerate`);
		expect(init).toMatchObject({
			method: "POST",
			credentials: "include",
		});
		expect(init?.body).toBeUndefined();
		expect(new Headers(init?.headers).get("Idempotency-Key")).toBe(
			idempotencyKey,
		);
	});

	it("treats a 202 response with an empty body as success", async () => {
		const fetchSpy = vi.fn<typeof fetch>(async () =>
			new Response(null, {
				status: 202,
				statusText: "Accepted",
			}),
		);
		const client = new EasyInterviewClient({ fetch: fetchSpy });

		await expect(
			client.startAuthEmailChallenge({ email: "alice@example.com" }),
		).resolves.toBeUndefined();
	});

	it("preserves a valid JSON ApiErrorResponse on HTTP failure", async () => {
		const apiError = {
			error: {
				code: "AI_PROVIDER_TIMEOUT" as const,
				message: "provider timeout",
				requestId: "req_typed_error",
				retryable: true,
				details: { phase: "reply" },
			},
		};
		const client = new EasyInterviewClient({
			fetch: vi.fn<typeof fetch>(async () =>
				new Response(JSON.stringify(apiError), {
					status: 502,
					statusText: "Bad Gateway",
					headers: { "Content-Type": "application/json" },
				}),
			),
		});

		const error = await captureApiClientError(
			client.sendPracticeMessage(
				"01918fa0-0000-7000-8000-000000005000",
				{
					clientMessageId: "01918fa0-0000-7000-8000-000000007001",
					text: "same message",
				},
			),
		);

		expect(error).toMatchObject({ kind: "http", status: 502, apiError });
		expect(error.message).toContain("AI_PROVIDER_TIMEOUT");
	});

	it("does not expose a non-JSON HTTP body", async () => {
		const rawBody = "private upstream body must not escape";
		const client = new EasyInterviewClient({
			fetch: vi.fn<typeof fetch>(async () =>
				new Response(rawBody, { status: 502, statusText: "Bad Gateway" }),
			),
		});

		const error = await captureApiClientError(
			client.getPracticeSession("01918fa0-0000-7000-8000-000000005000"),
		);

		expect(error).toMatchObject({ kind: "http", status: 502, apiError: null });
		expect(error.message).not.toContain(rawBody);
		expect(JSON.stringify(error)).not.toContain(rawBody);
	});

	it("represents an empty HTTP error body without fabricating an envelope", async () => {
		const client = new EasyInterviewClient({
			fetch: vi.fn<typeof fetch>(async () =>
				new Response(null, { status: 503, statusText: "Service Unavailable" }),
			),
		});

		const error = await captureApiClientError(
			client.getPracticeSession("01918fa0-0000-7000-8000-000000005000"),
		);

		expect(error).toMatchObject({ kind: "http", status: 503, apiError: null });
	});

	it("classifies AbortError separately from transport failure", async () => {
		const client = new EasyInterviewClient({
			fetch: vi.fn<typeof fetch>(async () => {
				throw new DOMException("request cancelled", "AbortError");
			}),
		});

		const error = await captureApiClientError(
			client.getPracticeSession("01918fa0-0000-7000-8000-000000005000"),
		);

		expect(error).toMatchObject({ kind: "abort", status: null, apiError: null });
	});

	it("classifies fetch rejection as transport failure", async () => {
		const client = new EasyInterviewClient({
			fetch: vi.fn<typeof fetch>(async () => {
				throw new TypeError("network unavailable");
			}),
		});

		const error = await captureApiClientError(
			client.getPracticeSession("01918fa0-0000-7000-8000-000000005000"),
		);

		expect(error).toMatchObject({ kind: "transport", status: null, apiError: null });
	});
});

describe("EasyInterviewClient safe-read single-flight", () => {
	it("shares only concurrent identical GETs and evicts after settlement", async () => {
		const fetchSpy = vi.fn<typeof fetch>(async () =>
			jsonResponse({ id: "target-1", analysisStatus: "ready" }),
		);
		const client = new EasyInterviewClient({ fetch: fetchSpy });

		const [first, second] = await Promise.all([
			client.getTargetJob("target-1"),
			client.getTargetJob("target-1"),
		]);

		expect(first).toEqual(second);
		expect(fetchSpy).toHaveBeenCalledTimes(1);

		await client.getTargetJob("target-1");
		expect(fetchSpy).toHaveBeenCalledTimes(2);
	});

	it("canonicalizes query, headers, and accepted statuses in the key", async () => {
		const fetchSpy = vi.fn<typeof fetch>(async () => jsonResponse({ items: [] }));
		const client = new EasyInterviewClient({ fetch: fetchSpy });
		const requestClient = client as unknown as RequestHarness;

		await Promise.all([
			client.listTargetJobs({
				query: { pageSize: "12", analysisStatus: "ready" },
				headers: { "X-Scope": "same", "Accept-Language": "zh" },
			}),
			client.listTargetJobs({
				query: { analysisStatus: "ready", pageSize: "12" },
				headers: { "Accept-Language": "zh", "X-Scope": "same" },
			}),
		]);
		expect(fetchSpy).toHaveBeenCalledTimes(1);

		await Promise.all([
			requestClient.request("GET", "/status-aware", undefined, undefined, [404, 202]),
			requestClient.request("GET", "/status-aware", undefined, undefined, [202, 404, 404]),
		]);
		expect(fetchSpy).toHaveBeenCalledTimes(2);
	});

	it("does not merge different query, header, or ok-status semantics", async () => {
		const fetchSpy = vi.fn<typeof fetch>(async () => jsonResponse({ items: [] }));
		const client = new EasyInterviewClient({ fetch: fetchSpy });
		const requestClient = client as unknown as RequestHarness;

		await Promise.all([
			client.listTargetJobs({ query: { pageSize: "12" } }),
			client.listTargetJobs({ query: { pageSize: "24" } }),
			client.listTargetJobs({
				query: { pageSize: "12" },
				headers: { "Accept-Language": "en" },
			}),
			requestClient.request("GET", "/status", undefined, undefined, [404]),
			requestClient.request("GET", "/status", undefined, undefined, []),
		]);

		expect(fetchSpy).toHaveBeenCalledTimes(5);
	});

	it("bypasses single-flight for caller AbortSignals and semantic GET mutations", async () => {
		const fetchSpy = vi.fn<typeof fetch>(async () => jsonResponse({ id: "session-1" }));
		const client = new EasyInterviewClient({ fetch: fetchSpy });
		const firstController = new AbortController();
		const secondController = new AbortController();

		await Promise.all([
			client.getTargetJob("target-1", { signal: firstController.signal }),
			client.getTargetJob("target-1", { signal: secondController.signal }),
			client.verifyAuthEmailChallenge({ query: { token: "same-token" } }),
			client.verifyAuthEmailChallenge({ query: { token: "same-token" } }),
		]);

		expect(fetchSpy).toHaveBeenCalledTimes(4);
	});

	it("fences reads both before mutation dispatch and after mutation settlement", async () => {
		const pending: Array<(response: Response) => void> = [];
		const fetchSpy = vi.fn<typeof fetch>(
			() => new Promise<Response>((resolve) => pending.push(resolve)),
		);
		const client = new EasyInterviewClient({ fetch: fetchSpy });

		const beforeMutation = client.getTargetJob("target-1");
		const mutation = client.logout();
		const duringMutation = client.getTargetJob("target-1");
		expect(fetchSpy).toHaveBeenCalledTimes(3);

		pending[1]?.(new Response(null, { status: 204 }));
		await mutation;

		const afterMutation = client.getTargetJob("target-1");
		expect(fetchSpy).toHaveBeenCalledTimes(4);

		pending[0]?.(jsonResponse({ id: "target-1" }));
		pending[2]?.(jsonResponse({ id: "target-1" }));
		pending[3]?.(jsonResponse({ id: "target-1" }));
		await Promise.all([beforeMutation, duringMutation, afterMutation]);
	});

	it("advances the post-settlement read fence when a mutation rejects", async () => {
		const pending: Array<{
			resolve: (response: Response) => void;
			reject: (error: unknown) => void;
		}> = [];
		const fetchSpy = vi.fn<typeof fetch>(
			() =>
				new Promise<Response>((resolve, reject) => {
					pending.push({ resolve, reject });
				}),
		);
		const client = new EasyInterviewClient({ fetch: fetchSpy });

		const beforeMutation = client.getTargetJob("target-1");
		const mutation = client.logout();
		const duringMutation = client.getTargetJob("target-1");
		expect(fetchSpy).toHaveBeenCalledTimes(3);

		pending[1]?.reject(new TypeError("logout transport failed"));
		await expect(mutation).rejects.toMatchObject({ kind: "transport" });

		const afterMutation = client.getTargetJob("target-1");
		expect(fetchSpy).toHaveBeenCalledTimes(4);

		pending[0]?.resolve(jsonResponse({ id: "target-1" }));
		pending[2]?.resolve(jsonResponse({ id: "target-1" }));
		pending[3]?.resolve(jsonResponse({ id: "target-1" }));
		await Promise.all([beforeMutation, duringMutation, afterMutation]);
	});

	it("advances auth scope only after successful email verification", async () => {
		const fetchSpy = vi
			.fn<typeof fetch>()
			.mockResolvedValueOnce(jsonResponse({ id: "verified-user" }))
			.mockResolvedValueOnce(
				jsonResponse(
					{
						error: {
							code: "AUTH_INVALID_OR_EXPIRED",
							message: "expired",
							requestId: "req_verify_expired",
							retryable: false,
						},
					},
					401,
				),
			);
		const client = new EasyInterviewClient({ fetch: fetchSpy });
		const epochs = client as unknown as { authEpoch: number };

		expect(epochs.authEpoch).toBe(0);
		await client.verifyAuthEmailChallenge({ query: { token: "valid-token" } });
		expect(epochs.authEpoch).toBe(1);

		await expect(
			client.verifyAuthEmailChallenge({ query: { token: "expired-token" } }),
		).rejects.toMatchObject({ kind: "http", status: 401 });
		expect(epochs.authEpoch).toBe(1);
	});

	it("shares a concurrent rejection and evicts it without an unhandled finally chain", async () => {
		const fetchSpy = vi
			.fn<typeof fetch>()
			.mockRejectedValueOnce(new TypeError("offline"))
			.mockResolvedValueOnce(jsonResponse({ id: "target-1" }));
		const client = new EasyInterviewClient({ fetch: fetchSpy });

		const first = client.getTargetJob("target-1");
		const second = client.getTargetJob("target-1");
		await expect(Promise.all([first, second])).rejects.toMatchObject({
			kind: "transport",
		});
		expect(fetchSpy).toHaveBeenCalledTimes(1);

		await expect(client.getTargetJob("target-1")).resolves.toMatchObject({
			id: "target-1",
		});
		expect(fetchSpy).toHaveBeenCalledTimes(2);
	});

	it("keeps single-flight state isolated per client instance", async () => {
		const fetchSpy = vi.fn<typeof fetch>(async () => jsonResponse({ id: "target-1" }));
		const firstClient = new EasyInterviewClient({ fetch: fetchSpy });
		const secondClient = new EasyInterviewClient({ fetch: fetchSpy });

		await Promise.all([
			firstClient.getTargetJob("target-1"),
			secondClient.getTargetJob("target-1"),
		]);
		expect(fetchSpy).toHaveBeenCalledTimes(2);
	});
});

interface RequestHarness {
	request<T>(
		method: string,
		path: string,
		body: unknown,
		opts?: Parameters<EasyInterviewClient["getTargetJob"]>[1],
		okStatuses?: readonly number[],
	): Promise<T>;
}

function jsonResponse(body: unknown, status = 200): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { "Content-Type": "application/json" },
	});
}

async function captureApiClientError(promise: Promise<unknown>): Promise<ApiClientError> {
	try {
		await promise;
	} catch (error) {
		expect(error).toBeInstanceOf(ApiClientError);
		return error as ApiClientError;
	}
	throw new Error("expected request to reject");
}
