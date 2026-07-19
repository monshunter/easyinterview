import { useEffect, useRef, type FC } from "react";
import { createPortal } from "react-dom";

import { useI18n } from "../i18n/messages";
import { AsyncTransitionScene } from "../transition/AsyncTransitionScene";

interface BackgroundState {
  element: HTMLElement;
  inert: string | null;
  ariaHidden: string | null;
}

export const PracticeLaunchTransition: FC = () => {
  const { t } = useI18n();
  const transitionRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const transition = transitionRef.current;
    if (!transition) return;

    const appMain = document.querySelector<HTMLElement>(
      '[data-testid="app-root"] > main',
    );
    const backgroundCandidates = appMain
      ? [appMain]
      : Array.from(document.body.children).filter(
          (element): element is HTMLElement => (
            element instanceof HTMLElement && !element.contains(transition)
          ),
        );
    const background: BackgroundState[] = backgroundCandidates
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
    <AsyncTransitionScene
      ref={transitionRef}
      variant="brand"
      testId="practice-launch-transition"
      className="ei-practice-launch-transition"
      tabIndex={-1}
      eyebrow={t("practice.launch.eyebrow")}
      title={t("practice.launch.title")}
      body={t("practice.launch.body")}
      hint={t("practice.launch.hint")}
      showProgress
    />
  );

  return createPortal(content, document.body);
};
