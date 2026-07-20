import {
  useEffect,
  useId,
  useRef,
  type FC,
  type KeyboardEvent as ReactKeyboardEvent,
  type RefObject,
} from "react";
import { createPortal } from "react-dom";

interface DestructiveActionDialogProps {
  eyebrow: string;
  title: string;
  description: string;
  cancelLabel: string;
  confirmLabel: string;
  pendingLabel: string;
  retryLabel: string;
  errorMessage: string | null;
  errorTestId?: string;
  pending: boolean;
  restoreFocusRef: RefObject<HTMLElement | null>;
  onCancel: () => void;
  onConfirm: () => void;
}

export const DestructiveActionDialog: FC<DestructiveActionDialogProps> = ({
  eyebrow,
  title,
  description,
  cancelLabel,
  confirmLabel,
  pendingLabel,
  retryLabel,
  errorMessage,
  errorTestId,
  pending,
  restoreFocusRef,
  onCancel,
  onConfirm,
}) => {
  const titleId = useId();
  const descriptionId = useId();
  const cancelRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    cancelRef.current?.focus();
  }, []);

  const cancel = () => {
    if (pending) return;
    onCancel();
    queueMicrotask(() => restoreFocusRef.current?.focus());
  };

  const handleKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      cancel();
      return;
    }
    if (event.key !== "Tab") return;
    const buttons = Array.from(
      event.currentTarget.querySelectorAll<HTMLButtonElement>("button:not(:disabled)"),
    );
    if (buttons.length === 0) return;
    const first = buttons[0];
    const last = buttons[buttons.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last?.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first?.focus();
    }
  };

  return createPortal(
    <div
      className="ei-destructive-dialog-layer"
      data-testid="destructive-action-dialog-layer"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) cancel();
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={descriptionId}
        className="ei-destructive-dialog"
        onKeyDown={handleKeyDown}
      >
        <span className="ei-text-label">{eyebrow}</span>
        <h2 id={titleId} className="ei-text-title">
          {title}
        </h2>
        <p id={descriptionId} className="ei-text-body">
          {description}
        </p>
        {errorMessage ? (
          <p
            role="alert"
            className="ei-destructive-dialog-error ei-text-body"
            data-testid={errorTestId}
          >
            {errorMessage}
          </p>
        ) : null}
        <div className="ei-destructive-dialog-actions">
          <button
            ref={cancelRef}
            type="button"
            disabled={pending}
            className="ei-destructive-dialog-cancel"
            onClick={cancel}
          >
            {cancelLabel}
          </button>
          <button
            type="button"
            disabled={pending}
            className="ei-destructive-dialog-confirm"
            onClick={onConfirm}
          >
            {pending ? pendingLabel : errorMessage ? retryLabel : confirmLabel}
          </button>
        </div>
      </div>
    </div>,
    document.body,
  );
};
