export interface PendingImportIntent {
  rawText: string;
  targetLanguage: string;
  resumeId: string;
  idempotencyKey: string;
}

interface StoredPendingImportIntent extends PendingImportIntent {
  expiresAt: number;
}

const PENDING_IMPORT_TTL_MS = 10 * 60 * 1000;
const pendingImports = new Map<string, StoredPendingImportIntent>();

function pruneExpired(now: number): void {
  for (const [id, intent] of pendingImports) {
    if (intent.expiresAt <= now) pendingImports.delete(id);
  }
}

export function storePendingImportIntent(intent: PendingImportIntent): string {
  const now = Date.now();
  pruneExpired(now);
  const id = `pending-import-${crypto.randomUUID()}`;
  pendingImports.set(id, {
    ...intent,
    expiresAt: now + PENDING_IMPORT_TTL_MS,
  });
  return id;
}

export function consumePendingImportIntent(
  id: string,
): PendingImportIntent | null {
  const stored = pendingImports.get(id) ?? null;
  pendingImports.delete(id);
  if (!stored || stored.expiresAt <= Date.now()) return null;

  return {
    rawText: stored.rawText,
    targetLanguage: stored.targetLanguage,
    resumeId: stored.resumeId,
    idempotencyKey: stored.idempotencyKey,
  };
}
