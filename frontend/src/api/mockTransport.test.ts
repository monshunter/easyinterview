import { describe, expect, it } from "vitest";

import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getPracticeSessionFixture from "../../../openapi/fixtures/PracticeSessions/getPracticeSession.json";
import listTargetJobsFixture from "../../../openapi/fixtures/TargetJobs/listTargetJobs.json";
import { EasyInterviewClient } from "./generated/client";
import type { PaginatedTargetJob, PracticeSession, RuntimeConfig, UserContext } from "./generated/types";
import { createFixtureBackedFetch, createFixtureRegistry } from "./mockTransport";

describe("fixture-backed generated client transport", () => {
	it("returns typed responses from OpenAPI fixtures", async () => {
		const client = new EasyInterviewClient({
			fetch: createFixtureBackedFetch(
				createFixtureRegistry([
					getRuntimeConfigFixture,
					getMeFixture,
					listTargetJobsFixture,
					getPracticeSessionFixture,
				]),
			),
		});

		const runtimeConfig: RuntimeConfig = await client.getRuntimeConfig();
		const me: UserContext = await client.getMe();
		const targetJobs: PaginatedTargetJob = await client.listTargetJobs();
		const session: PracticeSession = await client.getPracticeSession(
			"01918fa0-0000-7000-8000-000000005000",
		);

		expect(runtimeConfig.appVersion).toBe("1.0.0+dev.0428");
		expect(me.emailMasked).toBe("ali***@example.com");
		expect(targetJobs.items[0]?.title).toBe("Senior Frontend Engineer");
		expect(session.currentTurn?.questionIntent).toBe("behavioral.leadership.design_system");
	});
});
