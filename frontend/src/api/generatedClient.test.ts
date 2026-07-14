import { describe, expect, it, vi } from "vitest";

import { ApiClientError, EasyInterviewClient } from "./generated/client";

describe("EasyInterviewClient response parsing", () => {
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

async function captureApiClientError(promise: Promise<unknown>): Promise<ApiClientError> {
	try {
		await promise;
	} catch (error) {
		expect(error).toBeInstanceOf(ApiClientError);
		return error as ApiClientError;
	}
	throw new Error("expected request to reject");
}
