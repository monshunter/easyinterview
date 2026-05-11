import { describe, expect, it } from "vitest";

import { ALL_OPERATION_IDS } from "./generated/client";
import {
	createDevMockClient,
	getDevMockFixtureOperationIds,
} from "./devMockClient";

describe("frontend dev fixture-backed mock client", () => {
	it("covers every generated operationId with a fixture", () => {
		expect(getDevMockFixtureOperationIds().sort()).toEqual(
			[...ALL_OPERATION_IDS].sort(),
		);
	});

	it("returns fixture-backed data for dev preview pages", async () => {
		const client = createDevMockClient();

		const runtime = await client.getRuntimeConfig();
		const me = await client.getMe({
			headers: { Prefer: "example=authenticated" },
		});
		const targets = await client.listTargetJobs();

		expect(runtime.appVersion).toBe("1.0.0+dev.0428");
		expect(me.displayName).toBe("Alice Example");
		expect(targets.items[0]?.title).toBe("Senior Frontend Engineer");
	});

	it("models dev auth session state across /me, verify, and logout", async () => {
		const client = createDevMockClient();

		await expect(client.getMe()).rejects.toThrow(/HTTP 401/);

		await client.verifyAuthEmailChallenge({
			query: { token: "654321" },
		});
		await expect(client.getMe()).resolves.toMatchObject({
			displayName: "Alice Example",
			emailMasked: "ali***@example.com",
		});

		await client.logout();
		await expect(client.getMe()).rejects.toThrow(/HTTP 401/);
	});
});
