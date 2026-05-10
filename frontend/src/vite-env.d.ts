/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_EI_API_MODE?: "mock" | "real";
	readonly VITE_EI_API_BASE_URL?: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
