import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import {
	DEFAULT_DEV_REAL_API_BASE_URL,
	createAppClient,
} from "./clientFactory";

describe("frontend API client factory", () => {
	it("uses fixture-backed mock client by default in Vite dev", async () => {
		const client = createAppClient({ DEV: true });

		await expect(client.getRuntimeConfig()).resolves.toMatchObject({
			appVersion: "1.0.0+dev.0428",
		});
	});

	it("uses backend port 8080 by default when dev explicitly opts into real API", () => {
		const client = createAppClient({
			DEV: true,
			VITE_EI_API_MODE: "real",
		});

		expect(client.baseUrl).toBe(DEFAULT_DEV_REAL_API_BASE_URL);
		expect(client.baseUrl).not.toBe("/api/v1");
	});

	it("keeps production default on same-origin API and honors explicit real API base", () => {
		const production = createAppClient({ PROD: true });
		const customDevReal = createAppClient({
			DEV: true,
			VITE_EI_API_MODE: "real",
			VITE_EI_API_BASE_URL: "http://localhost:9090/api/v1/",
		});

		expect(production.baseUrl).toBe("/api/v1");
		expect(customDevReal.baseUrl).toBe("http://localhost:9090/api/v1");
	});

	it("main bootstrap creates the client through the factory", () => {
		const mainPath = fileURLToPath(new URL("../main.tsx", import.meta.url));
		const mainSource = readFileSync(mainPath, "utf8");

		expect(mainSource).toContain("createAppClient");
		expect(mainSource).not.toContain("new EasyInterviewClient()");
	});
});
