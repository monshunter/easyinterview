import { EasyInterviewClient } from "./generated/client";
import { createDevMockClient } from "./devMockClient";

const DEFAULT_PRODUCTION_API_BASE_URL = "/api/v1";

export interface AppClientEnv {
	readonly DEV?: boolean;
	readonly PROD?: boolean;
	readonly VITE_EI_API_MODE?: string;
	readonly VITE_EI_API_BASE_URL?: string;
}

export function createAppClient(env: AppClientEnv = import.meta.env): EasyInterviewClient {
	const mode = resolveApiMode(env);
	if (mode === "mock") {
		return createDevMockClient();
	}
	return new EasyInterviewClient({
		baseUrl: normalizeBaseUrl(resolveRealApiBaseUrl(env)),
	});
}

function resolveApiMode(env: AppClientEnv): "mock" | "real" {
	const requested = env.VITE_EI_API_MODE?.trim().toLowerCase();
	if (!requested) {
		return env.DEV ? "mock" : "real";
	}
	if (requested === "mock" || requested === "real") {
		return requested;
	}
	throw new Error(`unsupported VITE_EI_API_MODE: ${env.VITE_EI_API_MODE}`);
}

function resolveRealApiBaseUrl(env: AppClientEnv): string {
	const configured = env.VITE_EI_API_BASE_URL?.trim();
	if (configured) {
		return configured;
	}
	if (env.DEV) {
		throw new Error("VITE_EI_API_BASE_URL is required when VITE_EI_API_MODE=real in dev");
	}
	return DEFAULT_PRODUCTION_API_BASE_URL;
}

function normalizeBaseUrl(baseUrl: string): string {
	const trimmed = baseUrl.trim();
	const withoutTrailingSlash = trimmed.replace(/\/+$/, "");
	return withoutTrailingSlash === "" ? "/" : withoutTrailingSlash;
}
