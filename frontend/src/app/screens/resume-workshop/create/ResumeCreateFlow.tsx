import {
  useCallback,
  useMemo,
  useReducer,
  type FC,
  type KeyboardEvent,
} from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { ResumeWorkshopIcon } from "../components/ResumeWorkshopIcon";
import { ParsingStage } from "./ParsingStage";
import { PreviewStage } from "./PreviewStage";
import type { PreviewDraft } from "./ResumePreviewConfirm";
import { GuidedTab } from "./GuidedTab";
import { PasteTab } from "./PasteTab";
import { UploadTab } from "./UploadTab";
import { deriveDefaultTitle, type CreateMode } from "./util/title";
import type { ResumeParseState } from "./ResumeParseFlow";
import type { ResumeAsset } from "../../../../api/generated/types";

export type CreateStage = "input" | "parsing" | "preview";

export interface GuidedAnswers {
  recentRole: string;
  direction: string;
  proofProject: string;
  metrics: string;
  target: string;
}

const EMPTY_GUIDED_ANSWERS: GuidedAnswers = {
  recentRole: "",
  direction: "",
  proofProject: "",
  metrics: "",
  target: "",
};

const GUIDED_STEPS: Array<{
  key: keyof GuidedAnswers;
  label: MessageKey;
  question: MessageKey;
  placeholder: MessageKey;
}> = [
  {
    key: "recentRole",
    label: "resumeWorkshop.create.guided.step.recentRole.label",
    question: "resumeWorkshop.create.guided.step.recentRole.question",
    placeholder: "resumeWorkshop.create.guided.step.recentRole.placeholder",
  },
  {
    key: "direction",
    label: "resumeWorkshop.create.guided.step.direction.label",
    question: "resumeWorkshop.create.guided.step.direction.question",
    placeholder: "resumeWorkshop.create.guided.step.direction.placeholder",
  },
  {
    key: "proofProject",
    label: "resumeWorkshop.create.guided.step.proofProject.label",
    question: "resumeWorkshop.create.guided.step.proofProject.question",
    placeholder: "resumeWorkshop.create.guided.step.proofProject.placeholder",
  },
  {
    key: "metrics",
    label: "resumeWorkshop.create.guided.step.metrics.label",
    question: "resumeWorkshop.create.guided.step.metrics.question",
    placeholder: "resumeWorkshop.create.guided.step.metrics.placeholder",
  },
  {
    key: "target",
    label: "resumeWorkshop.create.guided.step.target.label",
    question: "resumeWorkshop.create.guided.step.target.question",
    placeholder: "resumeWorkshop.create.guided.step.target.placeholder",
  },
];

const TAB_DESCRIPTORS: Array<{
  mode: CreateMode;
  label: MessageKey;
  icon: "upload" | "file" | "chat";
}> = [
  { mode: "upload", label: "resumeWorkshop.create.tabs.upload", icon: "upload" },
  { mode: "paste", label: "resumeWorkshop.create.tabs.paste", icon: "file" },
  { mode: "guided", label: "resumeWorkshop.create.tabs.guided", icon: "chat" },
];

interface CreateState {
  stage: CreateStage;
  createMode: CreateMode;
  pickedFile: File | null;
  rawText: string;
  guidedAnswers: GuidedAnswers;
  guidedStep: number;
  resumeAssetId: string | null;
  sourceLabel: string | null;
  parseState: ResumeParseState | null;
  previewDraft: PreviewDraft | null;
  previewAsset: ResumeAsset | null;
  submitting: boolean;
  inlineError: string | null;
}

type CreateAction =
  | { type: "set_mode"; mode: CreateMode }
  | { type: "set_picked_file"; file: File | null }
  | { type: "set_raw_text"; text: string }
  | { type: "set_guided_answer"; key: keyof GuidedAnswers; value: string }
  | { type: "set_guided_step"; index: number }
  | { type: "advance_guided_step" }
  | { type: "back_guided_step" }
  | {
      type: "submit_registered";
      resumeAssetId: string;
      sourceLabel: string;
    }
  | { type: "set_parse_state"; parseState: ResumeParseState }
  | { type: "parse_ready"; draft: PreviewDraft; asset: ResumeAsset }
  | { type: "cancel_to_input" }
  | { type: "go_to_preview" }
  | { type: "back_to_input"; preserveResumeAssetId?: boolean }
  | { type: "set_submitting"; submitting: boolean }
  | { type: "set_inline_error"; error: string | null }
  | { type: "reset_after_success" };

function initialState(initialMode: CreateMode): CreateState {
  return {
    stage: "input",
    createMode: initialMode,
    pickedFile: null,
    rawText: "",
    guidedAnswers: EMPTY_GUIDED_ANSWERS,
    guidedStep: 0,
    resumeAssetId: null,
    sourceLabel: null,
    parseState: null,
    previewDraft: null,
    previewAsset: null,
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
    case "set_guided_answer":
      return {
        ...state,
        guidedAnswers: { ...state.guidedAnswers, [action.key]: action.value },
      };
    case "set_guided_step":
      return {
        ...state,
        guidedStep: Math.min(
          Math.max(0, action.index),
          GUIDED_STEPS.length - 1,
        ),
      };
    case "advance_guided_step":
      return {
        ...state,
        guidedStep: Math.min(state.guidedStep + 1, GUIDED_STEPS.length - 1),
      };
    case "back_guided_step":
      return {
        ...state,
        guidedStep: Math.max(0, state.guidedStep - 1),
      };
    case "submit_registered":
      return {
        ...state,
        stage: "parsing",
        resumeAssetId: action.resumeAssetId,
        sourceLabel: action.sourceLabel,
        parseState: { phase: "polling" },
        submitting: false,
        inlineError: null,
      };
    case "set_parse_state":
      return { ...state, parseState: action.parseState };
    case "parse_ready":
      return {
        ...state,
        stage: "preview",
        previewDraft: action.draft,
        previewAsset: action.asset,
        parseState: { phase: "ready" },
      };
    case "cancel_to_input":
      return {
        ...state,
        stage: "input",
        parseState: null,
        // Preserve user input (createMode, rawText, guidedAnswers, pickedFile).
      };
    case "go_to_preview":
      return {
        ...state,
        stage: "preview",
        submitting: false,
      };
    case "back_to_input":
      return {
        ...state,
        stage: "input",
        previewDraft: null,
        parseState: null,
        submitting: false,
        inlineError: null,
        resumeAssetId: action.preserveResumeAssetId
          ? state.resumeAssetId
          : null,
      };
    case "set_submitting":
      return { ...state, submitting: action.submitting };
    case "set_inline_error":
      return { ...state, inlineError: action.error };
    case "reset_after_success":
      return initialState(state.createMode);
    default:
      return state;
  }
}

export interface ResumeCreateFlowProps {
  /** Optional initial mode (from route param `createMode`). */
  initialMode?: CreateMode;
}

const isCreateMode = (value: unknown): value is CreateMode =>
  value === "upload" || value === "paste" || value === "guided";

export const ResumeCreateFlow: FC<ResumeCreateFlowProps> = ({
  initialMode,
}) => {
  const { t, lang } = useI18n();
  const { navigate } = useNavigation();
  const startMode: CreateMode = isCreateMode(initialMode) ? initialMode : "upload";
  const [state, dispatch] = useReducer(reducer, startMode, initialState);

  const handleBackToWorkshop = useCallback(() => {
    navigate({ name: "resume_versions", params: {} });
  }, [navigate]);

  const sourceLabel = useMemo(() => {
    if (state.sourceLabel) return state.sourceLabel;
    return deriveDefaultTitle(
      state.createMode,
      lang,
      state.pickedFile?.name ?? null,
    );
  }, [lang, state.createMode, state.pickedFile, state.sourceLabel]);

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

  if (state.stage === "parsing" && state.resumeAssetId) {
    return (
      <ParsingStage
        resumeAssetId={state.resumeAssetId}
        sourceLabel={sourceLabel}
        onReady={(asset: ResumeAsset, draft) =>
          dispatch({ type: "parse_ready", draft, asset })
        }
        onCancel={() => dispatch({ type: "cancel_to_input" })}
      />
    );
  }

  if (state.stage === "preview" && state.previewDraft && state.previewAsset) {
    return (
      <PreviewStage
        asset={state.previewAsset}
        draft={state.previewDraft}
        sourceLabel={sourceLabel}
        onBack={() =>
          dispatch({ type: "back_to_input", preserveResumeAssetId: true })
        }
        onSaved={() => dispatch({ type: "reset_after_success" })}
      />
    );
  }

  return (
    <div
      className="ei-resume-create-flow"
      data-testid="resume-create-flow"
      data-stage={state.stage}
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
        <span className="ei-text-label ei-resume-create-eyebrow">
          {t("resumeWorkshop.create.eyebrow")}
        </span>
        <h1 className="ei-text-display">{t("resumeWorkshop.create.title")}</h1>
        <p className="ei-resume-create-subtitle">
          {t("resumeWorkshop.create.subtitle")}
        </p>
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
                onRegistered={(resumeAssetId, label) =>
                  dispatch({
                    type: "submit_registered",
                    resumeAssetId,
                    sourceLabel: label,
                  })
                }
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
                onRegistered={(resumeAssetId, label) =>
                  dispatch({
                    type: "submit_registered",
                    resumeAssetId,
                    sourceLabel: label,
                  })
                }
                setSubmitting={(value) =>
                  dispatch({ type: "set_submitting", submitting: value })
                }
                setInlineError={(message) =>
                  dispatch({ type: "set_inline_error", error: message })
                }
              />
            ) : null}
            {state.createMode === "guided" ? (
              <GuidedTab
                guidedAnswers={state.guidedAnswers}
                guidedStep={state.guidedStep}
                steps={GUIDED_STEPS}
                submitting={state.submitting}
                inlineError={state.inlineError}
                onAnswerChange={(key, value) =>
                  dispatch({ type: "set_guided_answer", key, value })
                }
                onSelectStep={(index) =>
                  dispatch({ type: "set_guided_step", index })
                }
                onAdvanceStep={() => dispatch({ type: "advance_guided_step" })}
                onBackStep={() => dispatch({ type: "back_guided_step" })}
                onRegistered={(resumeAssetId, label) =>
                  dispatch({
                    type: "submit_registered",
                    resumeAssetId,
                    sourceLabel: label,
                  })
                }
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

        <aside
          className="ei-resume-create-sidebar"
          data-testid="resume-create-sidebar"
        >
          <div className="ei-resume-create-sidebar-card">
            <span className="ei-text-label ei-resume-create-sidebar-eyebrow">
              {t("resumeWorkshop.create.sidebar.whatSavedEyebrow")}
            </span>
            <ul className="ei-resume-create-sidebar-list">
              <li>
                <ResumeWorkshopIcon name="file" size={15} />
                <div>
                  <div className="ei-resume-create-sidebar-item-title">
                    {t(
                      "resumeWorkshop.create.sidebar.whatSaved.original.title",
                    )}
                  </div>
                  <p className="ei-resume-create-sidebar-item-body">
                    {t("resumeWorkshop.create.sidebar.whatSaved.original.body")}
                  </p>
                </div>
              </li>
              <li>
                <ResumeWorkshopIcon name="resume" size={15} />
                <div>
                  <div className="ei-resume-create-sidebar-item-title">
                    {t(
                      "resumeWorkshop.create.sidebar.whatSaved.structured.title",
                    )}
                  </div>
                  <p className="ei-resume-create-sidebar-item-body">
                    {t(
                      "resumeWorkshop.create.sidebar.whatSaved.structured.body",
                    )}
                  </p>
                </div>
              </li>
              <li>
                <ResumeWorkshopIcon name="layers" size={15} />
                <div>
                  <div className="ei-resume-create-sidebar-item-title">
                    {t(
                      "resumeWorkshop.create.sidebar.whatSaved.baseline.title",
                    )}
                  </div>
                  <p className="ei-resume-create-sidebar-item-body">
                    {t("resumeWorkshop.create.sidebar.whatSaved.baseline.body")}
                  </p>
                </div>
              </li>
            </ul>
          </div>
          <div className="ei-resume-create-sidebar-card">
            <span className="ei-text-label ei-resume-create-sidebar-eyebrow">
              {t("resumeWorkshop.create.sidebar.whatNextEyebrow")}
            </span>
            <p className="ei-resume-create-sidebar-item-body">
              {t("resumeWorkshop.create.sidebar.whatNextBody")}
            </p>
          </div>
        </aside>
      </div>
    </div>
  );
};
