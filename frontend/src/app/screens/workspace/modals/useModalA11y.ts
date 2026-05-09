import { useCallback, useEffect, useRef } from "react";

export interface UseModalA11yOptions {
  open: boolean;
  onClose: () => void;
  /** ref to the trigger element for focus return on close */
  triggerRef?: React.RefObject<HTMLElement | null>;
}

/**
 * Shared modal a11y hook: ESC close, backdrop click close,
 * focus trap with Tab cycling, focus return on close.
 */
export function useModalA11y({
  open,
  onClose,
  triggerRef,
}: UseModalA11yOptions) {
  const modalRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  // Save focus on open, trap inside modal
  useEffect(() => {
    if (!open) return;

    previousFocusRef.current = document.activeElement as HTMLElement | null;

    // Focus first focusable element after a tick
    const timer = requestAnimationFrame(() => {
      if (modalRef.current) {
        const first = getFocusableElements(modalRef.current)[0];
        first?.focus();
      }
    });

    return () => cancelAnimationFrame(timer);
  }, [open]);

  // Restore focus on close
  useEffect(() => {
    if (!open && previousFocusRef.current) {
      previousFocusRef.current.focus();
      previousFocusRef.current = null;
    }
  }, [open]);

  // Keyboard handler: ESC close + Tab trap
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Escape") {
        e.stopPropagation();
        onClose();
        return;
      }

      if (e.key === "Tab" && modalRef.current) {
        const focusable = getFocusableElements(modalRef.current);
        if (focusable.length === 0) return;

        const first = focusable[0]!;
        const last = focusable[focusable.length - 1]!;

        if (e.shiftKey) {
          if (document.activeElement === first) {
            e.preventDefault();
            last.focus();
          }
        } else {
          if (document.activeElement === last) {
            e.preventDefault();
            first.focus();
          }
        }
      }
    },
    [onClose],
  );

  // Backdrop click handler
  const handleBackdropClick = useCallback(
    (e: React.MouseEvent) => {
      if (e.target === e.currentTarget) {
        onClose();
      }
    },
    [onClose],
  );

  return { modalRef, handleKeyDown, handleBackdropClick };
}

function getFocusableElements(container: HTMLElement): HTMLElement[] {
  const selectors = [
    "a[href]",
    "button:not([disabled])",
    "input:not([disabled])",
    "textarea:not([disabled])",
    "select:not([disabled])",
    '[tabindex]:not([tabindex="-1"])',
  ];
  const els = container.querySelectorAll<HTMLElement>(selectors.join(","));
  return Array.from(els).filter(
    (el) => !el.hasAttribute("disabled") && el.offsetParent !== null,
  );
}
