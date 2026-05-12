type ToastTone = "ok" | "warn" | "danger" | "neutral";
type EiToast = (
  message: string,
  opts?: { tone?: ToastTone; duration?: number },
) => void;

export function fireResumeWorkshopToast(
  message: string,
  tone: ToastTone = "neutral",
): void {
  if (typeof window === "undefined") return;
  const fn = (window as unknown as { eiToast?: EiToast }).eiToast;
  if (typeof fn === "function") fn(message, { tone });
}
