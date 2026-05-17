import { useCallback, useState, type FC } from "react";

import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { DebriefContextStrip } from "./components/DebriefContextStrip";
import { DebriefHeader } from "./components/DebriefHeader";
import { DebriefStepper } from "./components/DebriefStepper";
import {
  EMPTY_SELECTED_CONTEXT,
  type DebriefInputMode,
  type DebriefPickerKind,
  type DebriefSelectedContext,
  type DebriefStep,
} from "./types";
import "./debrief.css";

interface DebriefScreenProps {
  route: Route;
}

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx::DebriefFullScreen
 * (lines 38-368). Phase 0+1 lands the shell + Header + ContextStrip +
 * Stepper, with three-step state plumbing ready for Phase 2-6 to layer in
 * picker modals, suggestions, submit/polling, analysis rendering, and the
 * replay-interview handoff.
 */
export const DebriefScreen: FC<DebriefScreenProps> = ({ route }) => {
  const { navigate } = useNavigation();
  const [step, setStep] = useState<DebriefStep>(0);
  // maxVisited tracks the furthest step the user has progressed to. Phase 5
  // bumps this when createDebrief lands and the runtime promotes the user
  // to step 1; Phase 6.2 bumps it again when analysis completes.
  const [maxVisited] = useState<DebriefStep>(0);
  const [inputMode] = useState<DebriefInputMode>("text");
  const [selectedContext] =
    useState<DebriefSelectedContext>(EMPTY_SELECTED_CONTEXT);
  const [pickerKind, setPickerKind] = useState<DebriefPickerKind | null>(null);

  const handleStep = useCallback(
    (next: DebriefStep) => {
      if (next > maxVisited) return;
      setStep(next);
    },
    [maxVisited],
  );

  const handleBack = useCallback(() => {
    navigate({ name: "home" });
  }, [navigate]);

  const handleOpenPicker = useCallback((kind: DebriefPickerKind) => {
    setPickerKind(kind);
  }, []);

  return (
    <section
      className="ei-screen-shell ei-debrief-screen"
      data-testid="route-debrief"
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
      data-step={String(step)}
      data-input-mode={inputMode}
      data-picker-kind={pickerKind ?? "none"}
    >
      <DebriefHeader
        selectedContext={selectedContext}
        onBack={handleBack}
      />
      <DebriefContextStrip
        selectedContext={selectedContext}
        onOpenPicker={handleOpenPicker}
      />
      <DebriefStepper
        step={step}
        maxVisited={maxVisited}
        onStep={handleStep}
      />
      <div
        className="ei-debrief-step-panel"
        data-testid={`debrief-step-panel-${step}`}
      >
        {/* Phase 1.1 ships only the shell; Phase 3 (Step 0 record),
            Phase 6.1 (Step 1 analysis), and Phase 6.2 (Step 2 replay
            launcher) layer their concrete content here. */}
      </div>
    </section>
  );
};
