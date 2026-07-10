export type PendingImportSource =
  | { source: "paste"; rawText: string }
  | { source: "upload" }
  | { source: "url"; url: string };

const pendingSources = new Map<string, PendingImportSource>();

export function storePendingImportSource(source: PendingImportSource): string {
  const id = `pending-import-${crypto.randomUUID()}`;
  pendingSources.set(id, source);
  return id;
}

export function consumePendingImportSource(
  id: string,
): PendingImportSource | null {
  const source = pendingSources.get(id) ?? null;
  pendingSources.delete(id);
  return source;
}
