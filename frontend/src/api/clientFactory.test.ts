import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import { createAppClient } from "./clientFactory";

describe("frontend API client factory", () => {
	it("uses fixture-backed mock client by default in Vite dev", async () => {
		const client = createAppClient({ DEV: true });

		await expect(client.getRuntimeConfig()).resolves.toMatchObject({
			appVersion: "1.0.0+dev.0428",
		});
	});

	it("requires an explicit API base when dev opts into real API", () => {
		expect(() =>
			createAppClient({
				DEV: true,
				VITE_EI_API_MODE: "real",
			}),
		).toThrow(/VITE_EI_API_BASE_URL/);
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
