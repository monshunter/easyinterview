import {
  useCallback,
  useReducer,
  type FC,
  type KeyboardEvent,
} from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { ResumeWorkshopIcon } from "../components/ResumeWorkshopIcon";
import { PasteTab } from "./PasteTab";
import { UploadTab } from "./UploadTab";
import type { CreateMode } from "./util/title";

const TAB_DESCRIPTORS: Array<{
  mode: CreateMode;
  label: MessageKey;
  icon: "upload" | "file";
}> = [
  { mode: "upload", label: "resumeWorkshop.create.tabs.upload", icon: "upload" },
  { mode: "paste", label: "resumeWorkshop.create.tabs.paste", icon: "file" },
];

interface CreateState {
  createMode: CreateMode;
  pickedFile: File | null;
  rawText: string;
  submitting: boolean;
  inlineError: string | null;
}

type CreateAction =
  | { type: "set_mode"; mode: CreateMode }
  | { type: "set_picked_file"; file: File | null }
  | { type: "set_raw_text"; text: string }
  | { type: "set_submitting"; submitting: boolean }
  | { type: "set_inline_error"; error: string | null };

function initialState(initialMode: CreateMode): CreateState {
  return {
    createMode: initialMode,
    pickedFile: null,
    rawText: "",
    submitting: false,
    inlineError: null,
  };
}

function reducer(state: CreateState, action: CreateAction): CreateState {
  switch (action.type) {
    case "set_mode":
      return { ...state, createMode: action.mode, inlineError: null };
    case "set_picked_file":
      return { ...state, pickedFile: action.file, inlineError: null };
    case "set_raw_text":
      return { ...state, rawText: action.text };
    case "set_submitting":
      return { ...state, submitting: action.submitting };
    case "set_inline_error":
      return { ...state, inlineError: action.error };
    default:
      return state;
  }
}

export interface ResumeCreateFlowProps {
  /** Optional initial mode (from route param `createMode`). */
  initialMode?: CreateMode;
}

const isCreateMode = (value: unknown): value is CreateMode =>
  value === "upload" || value === "paste";

export const ResumeCreateFlow: FC<ResumeCreateFlowProps> = ({
  initialMode,
}) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const startMode: CreateMode = isCreateMode(initialMode) ? initialMode : "upload";
  const [state, dispatch] = useReducer(reducer, startMode, initialState);

  const handleBackToWorkshop = useCallback(() => {
    navigate({ name: "resume_versions", params: {} });
  }, [navigate]);

  const handleRegistered = useCallback(
    (resumeId: string) => {
      navigate({ name: "resume_versions", params: { resumeId } });
    },
    [navigate],
  );

  const onModeTabKeyDown = (
    event: KeyboardEvent<HTMLButtonElement>,
    currentIndex: number,
  ) => {
    if (event.key !== "ArrowRight" && event.key !== "ArrowLeft") return;
    event.preventDefault();
    const direction = event.key === "ArrowRight" ? 1 : -1;
    const nextIndex =
      (currentIndex + direction + TAB_DESCRIPTORS.length) %
      TAB_DESCRIPTORS.length;
    const nextMode = TAB_DESCRIPTORS[nextIndex]!.mode;
    dispatch({ type: "set_mode", mode: nextMode });
    const nextEl = document.querySelector<HTMLElement>(
      `[data-testid="resume-create-tab-${nextMode}"]`,
    );
    nextEl?.focus();
  };

  return (
    <div
      className="ei-resume-create-flow"
      data-testid="resume-create-flow"
      data-stage="input"
      data-create-mode={state.createMode}
    >
      <button
        type="button"
        className="ei-resume-create-back"
        data-testid="resume-create-back"
        onClick={handleBackToWorkshop}
      >
        <ResumeWorkshopIcon name="arrowLeft" size={14} />
        {t("resumeWorkshop.create.back")}
      </button>

      <header className="ei-resume-create-header">
        <div className="ei-resume-create-header-copy">
          <span className="ei-text-label ei-resume-create-eyebrow">
            {t("resumeWorkshop.create.eyebrow")}
          </span>
          <h1 className="ei-text-display">{t("resumeWorkshop.create.title")}</h1>
          <p className="ei-resume-create-subtitle">
            {t("resumeWorkshop.create.subtitle")}
          </p>
        </div>
        <ResumeCreateHeaderArt />
      </header>

      <div className="ei-resume-create-grid">
        <section
          className="ei-resume-create-card"
          data-testid="resume-create-card"
        >
          <div className="ei-resume-create-tablist" role="tablist">
            {TAB_DESCRIPTORS.map((descriptor, index) => {
              const active = state.createMode === descriptor.mode;
              return (
                <button
                  key={descriptor.mode}
                  type="button"
                  role="tab"
                  aria-selected={active}
                  tabIndex={active ? 0 : -1}
                  className="ei-resume-create-tab"
                  data-testid={`resume-create-tab-${descriptor.mode}`}
                  data-active={active}
                  onClick={() =>
                    dispatch({ type: "set_mode", mode: descriptor.mode })
                  }
                  onKeyDown={(event) => onModeTabKeyDown(event, index)}
                >
                  <ResumeWorkshopIcon name={descriptor.icon} size={14} />
                  {t(descriptor.label)}
                </button>
              );
            })}
          </div>

          <div
            className="ei-resume-create-tab-panel"
            role="tabpanel"
            data-mode={state.createMode}
          >
            {state.createMode === "upload" ? (
              <UploadTab
                pickedFile={state.pickedFile}
                submitting={state.submitting}
                inlineError={state.inlineError}
                onPickFile={(file) =>
                  dispatch({ type: "set_picked_file", file })
                }
                onValidationError={(message) =>
                  dispatch({ type: "set_inline_error", error: message })
                }
                onRegistered={handleRegistered}
                setSubmitting={(value) =>
                  dispatch({ type: "set_submitting", submitting: value })
                }
                setInlineError={(message) =>
                  dispatch({ type: "set_inline_error", error: message })
                }
              />
            ) : null}
            {state.createMode === "paste" ? (
              <PasteTab
                rawText={state.rawText}
                submitting={state.submitting}
                inlineError={state.inlineError}
                onRawTextChange={(text) =>
                  dispatch({ type: "set_raw_text", text })
                }
                onRegistered={handleRegistered}
                setSubmitting={(value) =>
                  dispatch({ type: "set_submitting", submitting: value })
                }
                setInlineError={(message) =>
                  dispatch({ type: "set_inline_error", error: message })
                }
              />
            ) : null}
          </div>
        </section>
      </div>
    </div>
  );
};

const ResumeCreateHeaderArt: FC = () => (
  <svg
    aria-hidden="true"
    className="ei-resume-create-header-art"
    viewBox="0 0 250 170"
  >
    <path d="M74 29h83l34 34v83H74z" fill="currentColor" opacity=".08" />
    <path d="M157 29v36h35" fill="none" stroke="currentColor" strokeWidth="3" opacity=".13" />
    <path d="M94 82h66M94 101h48" fill="none" stroke="currentColor" strokeLinecap="round" strokeWidth="7" opacity=".09" />
    <path d="M42 92h137a16 16 0 0 1 16 16v40H42z" fill="currentColor" opacity=".1" />
    <rect x="28" y="92" width="59" height="39" rx="13" fill="currentColor" opacity=".34" />
    <circle cx="48" cy="112" r="3" fill="white" />
    <circle cx="59" cy="112" r="3" fill="white" />
    <circle cx="70" cy="112" r="3" fill="white" />
    <path d="M18 48l4 10 10 4-10 4-4 10-4-10-10-4 10-4zM217 62l2 6 6 2-6 2-2 6-2-6-6-2 6-2z" fill="currentColor" opacity=".15" />
  </svg>
);
