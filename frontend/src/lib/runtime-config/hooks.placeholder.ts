/**
 * Placeholder for the React `useRuntimeConfig()` hook.
 *
 * D1 frontend-shell will replace this file with the real implementation
 * (Provider + Suspense / useSyncExternalStore wiring). A4 only locks the
 * exported type signature so cross-plan handoff stays stable.
 *
 * IMPORTANT: do NOT import React from this file. The type-only signature
 * keeps the bundle clean for non-React consumers (vitest helpers, SSR
 * smoke tests).
 */

import type { RuntimeConfig } from './types';

/**
 * Placeholder signature. Returns the in-memory cached RuntimeConfig if
 * `fetchRuntimeConfig()` has resolved at least once, otherwise undefined.
 * The full hook will throw a Suspense promise instead — implementation
 * owned by D1.
 */
export type UseRuntimeConfig = () => RuntimeConfig | undefined;
