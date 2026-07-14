import type { ContentLimits, RuntimeConfig } from "../api/generated/types";

export function utf8ByteLength(value: string): number {
  return new TextEncoder().encode(value).byteLength;
}

export function resolveContentLimits(
  runtimeConfig: RuntimeConfig | null | undefined,
): ContentLimits | null {
  return runtimeConfig?.contentLimits ?? null;
}

export function formatBinaryByteLimit(bytes: number, compact = false): string {
  const units = [
    { bytes: 1024 * 1024, label: "MiB" },
    { bytes: 1024, label: "KiB" },
    { bytes: 1, label: "bytes" },
  ];
  const unit = units.find((candidate) => bytes >= candidate.bytes) ?? {
    bytes: 1,
    label: "bytes",
  };
  const amount = bytes / unit.bytes;
  const value = Number.isInteger(amount)
    ? String(amount)
    : amount.toFixed(2).replace(/0+$/, "").replace(/\.$/, "");
  return `${value}${compact ? "" : " "}${unit.label}`;
}
