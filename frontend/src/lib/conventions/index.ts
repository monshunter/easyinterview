// Conventions barrel: re-exports the generated truth source plus hand-written
// helpers (idempotency). Hand-edited downstream code should import from this
// module instead of reaching into generated files directly so future generator
// changes stay invisible to callers.

export * from './enums';
export * from './errors';
export * from './pagination';
export * from './idempotency';
