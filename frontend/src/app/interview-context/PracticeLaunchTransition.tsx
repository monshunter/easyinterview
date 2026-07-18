import { useEffect, useRef, type FC } from "react";
import { createPortal } from "react-dom";

import { useI18n } from "../i18n/messages";

interface BackgroundState {
  element: HTMLElement;
  inert: string | null;
  ariaHidden: string | null;
}

export const PracticeLaunchTransition: FC = () => {
  const { t } = useI18n();
  const transitionRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const transition = transitionRef.current;
    if (!transition) return;

    const background: BackgroundState[] = Array.from(document.body.children)
      .filter((element): element is HTMLElement => (
        element instanceof HTMLElement && !element.contains(transition)
      ))
      .map((element) => ({
        element,
        inert: element.getAttribute("inert"),
        ariaHidden: element.getAttribute("aria-hidden"),
      }));
    const previousOverflow = document.body.style.overflow;
    const previousFocus = document.activeElement instanceof HTMLElement
      ? document.activeElement
      : null;

    for (const state of background) {
      state.element.setAttribute("inert", "");
      state.element.setAttribute("aria-hidden", "true");
    }
    document.body.style.overflow = "hidden";
    transition.focus();

    return () => {
      for (const state of background) {
        if (state.inert === null) state.element.removeAttribute("inert");
        else state.element.setAttribute("inert", state.inert);
        if (state.ariaHidden === null) state.element.removeAttribute("aria-hidden");
        else state.element.setAttribute("aria-hidden", state.ariaHidden);
      }
      document.body.style.overflow = previousOverflow;
      if (previousFocus?.isConnected) previousFocus.focus();
    };
  }, []);

  const content = (
    <div
      ref={transitionRef}
      className="ei-practice-launch-transition"
      data-testid="practice-launch-transition"
      role="status"
      aria-live="polite"
      aria-busy="true"
      tabIndex={-1}
    >
      <div className="ei-practice-launch-panel">
        <div className="ei-practice-launch-visual" aria-hidden="true">
          <span className="ei-practice-launch-orbit ei-practice-launch-orbit-outer" />
          <span className="ei-practice-launch-orbit ei-practice-launch-orbit-inner" />
          <span className="ei-practice-launch-core">E</span>
        </div>
        <div className="ei-label ei-practice-launch-eyebrow">
          {t("practice.launch.eyebrow")}
        </div>
        <h1 className="ei-serif ei-practice-launch-title">
          {t("practice.launch.title")}
        </h1>
        <p className="ei-practice-launch-body">
          {t("practice.launch.body")}
        </p>
        <div className="ei-practice-launch-rule" aria-hidden="true">
          <span />
        </div>
        <p className="ei-practice-launch-hint">
          {t("practice.launch.hint")}
        </p>
      </div>
    </div>
  );

  return createPortal(content, document.body);
};
