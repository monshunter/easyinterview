import { describe, expect, it, vi } from "vitest";

import { EasyInterviewClient } from "./generated/client";

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
});
