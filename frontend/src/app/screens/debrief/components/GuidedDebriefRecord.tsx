import { useCallback, useMemo, useState, type FC } from "react";

import type { MessageKey } from "../../../i18n/locales/zh";
import { useI18n } from "../../../i18n/messages";
import type {
  DebriefEntry,
  DebriefEntrySource,
  SuggestedDebriefQuestion,
} from "../types";

interface GuidedDebriefRecordProps {
  suggestions: SuggestedDebriefQuestion[] | null;
  loading: boolean;
  errorCode: string | null;
  entries: DebriefEntry[];
  setEntries: (next: DebriefEntry[]) => void;
  activeGuide: number;
  setActiveGuide: (next: number) => void;
  onRegenerate: () => void;
}

let entryIdCounter = 0;
function makeEntryId(): string {
  entryIdCounter += 1;
  return `entry-${Date.now().toString(36)}-${entryIdCounter}`;
}

function suggestionToEntry(
  guide: SuggestedDebriefQuestion,
  source: DebriefEntrySource,
  questionText: string,
  myAnswerSummary: string,
  interviewerReaction?: string,
): DebriefEntry {
  return {
    id: makeEntryId(),
    stage: guide.stage ?? undefined,
    questionText,
    myAnswerSummary,
    interviewerReaction,
    source,
  };
}

type EditorMode = "occurred" | "edit" | "manual";
type ReactionKey = NonNullable<DebriefEntry["reaction"]>;

const REACTION_KEYS: ReactionKey[] = [
  "positive",
  "neutral",
  "probed",
  "skeptical",
];

function sourceBadgeMessage(source: DebriefEntrySource): MessageKey {
  if (source === "manual") return "debrief.record.entries.source.manual";
  if (source === "voice_extracted") return "debrief.record.entries.source.voice";
  return "debrief.record.entries.source.recorded";
}

function reactionMessage(reaction: ReactionKey): MessageKey {
  return `debrief.record.entries.reaction.${reaction}` as MessageKey;
}

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx::GuidedDebriefRecord
 * (lines 519-619). Left panel renders the current guide question + 4 CTAs
 * (occurred / skip / edit / manual). Right panel renders the cross-mode
 * shared entries list. Entries write source attribution per Phase 3.3.
 */
export const GuidedDebriefRecord: FC<GuidedDebriefRecordProps> = ({
  suggestions,
  loading,
  errorCode,
  entries,
  setEntries,
  activeGuide,
  setActiveGuide,
  onRegenerate,
}) => {
  const { t } = useI18n();
  const total = suggestions?.length ?? 0;
  const currentGuide = useMemo(() => suggestions?.[activeGuide] ?? null, [
    suggestions,
    activeGuide,
  ]);
  const [questionValue, setQuestionValue] = useState("");
  const [answerValue, setAnswerValue] = useState("");
  const [reactionValue, setReactionValue] = useState("");
  const [editMode, setEditMode] = useState<EditorMode | null>(null);
  const [activeEntryId, setActiveEntryId] = useState<string | null>(null);
  const activeEntry = useMemo(
    () => entries.find((entry) => entry.id === activeEntryId) ?? entries[0] ?? null,
    [activeEntryId, entries],
  );

  const openEditor = useCallback(
    (mode: EditorMode) => {
      if ((mode === "occurred" || mode === "edit") && !currentGuide) return;
      setEditMode(mode);
      setQuestionValue(
        mode === "manual" ? "" : currentGuide?.questionText ?? "",
      );
      setAnswerValue("");
      setReactionValue("");
    },
    [currentGuide],
  );

  const skip = useCallback(() => {
    if (!currentGuide) return;
    setActiveGuide(activeGuide + 1);
  }, [activeGuide, currentGuide, setActiveGuide]);

  const saveEdit = useCallback(() => {
    const questionText = questionValue.trim();
    const myAnswerSummary = answerValue.trim();
    const interviewerReaction = reactionValue.trim();
    if (questionText === "" || myAnswerSummary === "" || !editMode) {
      return;
    }
    if (editMode === "manual") {
      const nextEntry: DebriefEntry = {
        id: makeEntryId(),
        questionText,
        myAnswerSummary,
        interviewerReaction: interviewerReaction || undefined,
        source: "manual",
      };
      setEntries([...entries, nextEntry]);
      setActiveEntryId(nextEntry.id);
    } else if (currentGuide) {
      const source: DebriefEntrySource =
        editMode === "edit" ? "ai_edited" : "ai_confirmed";
      const nextEntry = suggestionToEntry(
        currentGuide,
        source,
        questionText,
        myAnswerSummary,
        interviewerReaction || undefined,
      );
      setEntries([...entries, nextEntry]);
      setActiveEntryId(nextEntry.id);
      setActiveGuide(activeGuide + 1);
    }
    setQuestionValue("");
    setAnswerValue("");
    setReactionValue("");
    setEditMode(null);
  }, [
    activeGuide,
    answerValue,
    currentGuide,
    editMode,
    entries,
    questionValue,
    reactionValue,
    setActiveGuide,
    setEntries,
  ]);

  const cancelEdit = useCallback(() => {
    setEditMode(null);
    setQuestionValue("");
    setAnswerValue("");
    setReactionValue("");
  }, []);

  const saveDisabled = questionValue.trim() === "" || answerValue.trim() === "";
  const setEntryField = useCallback(
    (id: string, patch: Partial<DebriefEntry>) => {
      setEntries(entries.map((entry) => (entry.id === id ? { ...entry, ...patch } : entry)));
    },
    [entries, setEntries],
  );

  return (
    <section
      className="ei-debrief-guided"
      data-testid="debrief-guided-record"
      data-loading={loading}
      data-error-code={errorCode ?? "none"}
    >
      <div className="ei-debrief-guided__panel">
        <div className="ei-label">{t("debrief.record.guide.eyebrow")}</div>
        {loading && (
          <p data-testid="debrief-guided-loading">
            {t("debrief.record.guide.loading")}
          </p>
        )}
        {errorCode && !loading && (
          <div data-testid="debrief-guided-failure">
            <p>{t("debrief.record.guide.failure")}</p>
            <button
              type="button"
              data-testid="debrief-guided-regenerate"
              onClick={onRegenerate}
            >
              {t("debrief.record.guide.regenerate")}
            </button>
            <button
              type="button"
              data-testid="debrief-suggested-question-manual"
              onClick={() => openEditor("manual")}
            >
              {t("debrief.record.guide.ctaManual")}
            </button>
          </div>
        )}
        {!loading && !errorCode && currentGuide && (
          <article
            className="ei-debrief-guided__current"
            data-testid="debrief-guided-current"
          >
            <div className="ei-debrief-guided__current-copy">
              <div className="ei-debrief-guided__progress" data-testid="debrief-guided-progress">
                {t("debrief.record.guide.progress")
                  .replace("{current}", String(activeGuide + 1).padStart(2, "0"))
                  .replace("{total}", String(total).padStart(2, "0"))}
              </div>
              <div className="ei-debrief-guided__stage">{currentGuide.stage}</div>
              <h3 className="ei-serif">{currentGuide.questionText}</h3>
            </div>
            <div
              className="ei-debrief-guided__current-icon"
              data-testid="debrief-guided-current-icon"
              aria-hidden="true"
            >
              <svg
                width="28"
                height="28"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4z" />
              </svg>
            </div>
            <div className="ei-debrief-guided__why">
              <p>{currentGuide.whyLikelyAsked}</p>
              <div className="ei-debrief-guided__source">
                {t("debrief.record.guide.sourceLabel")} · {currentGuide.source}
              </div>
            </div>
            <div className="ei-debrief-guided__actions">
              <button
                type="button"
                data-testid="debrief-suggested-question-occurred"
                onClick={() => openEditor("occurred")}
              >
                ✓ {t("debrief.record.guide.ctaOccurred")}
              </button>
              <button
                type="button"
                data-testid="debrief-suggested-question-skip"
                onClick={skip}
              >
                {t("debrief.record.guide.ctaSkip")}
              </button>
              <button
                type="button"
                data-testid="debrief-suggested-question-edit"
                onClick={() => openEditor("edit")}
              >
                ✎ {t("debrief.record.guide.ctaEdit")}
              </button>
              <button
                type="button"
                data-testid="debrief-suggested-question-manual"
                onClick={() => openEditor("manual")}
              >
                {t("debrief.record.guide.ctaManual")}
              </button>
            </div>
          </article>
        )}
        {!loading && !errorCode && !currentGuide && (
          <div data-testid="debrief-guided-empty">
            <p>{t("debrief.record.guide.empty")}</p>
            <button
              type="button"
              data-testid="debrief-suggested-question-manual"
              onClick={() => openEditor("manual")}
            >
              {t("debrief.record.guide.ctaManual")}
            </button>
          </div>
        )}
        {editMode && (
          <div
            className="ei-debrief-guided__editor"
            data-testid="debrief-guided-editor"
            data-mode={editMode}
          >
            <textarea
              rows={3}
              value={questionValue}
              onChange={(e) => setQuestionValue(e.target.value)}
              placeholder={
                editMode === "manual"
                  ? t("debrief.record.guide.manualPlaceholder")
                  : t("debrief.record.guide.editPlaceholder")
              }
              data-testid="debrief-guided-editor-input"
            />
            <textarea
              rows={4}
              value={answerValue}
              onChange={(e) => setAnswerValue(e.target.value)}
              placeholder={t("debrief.record.guide.answerPlaceholder")}
              aria-label={t("debrief.record.guide.answerLabel")}
              data-testid="debrief-guided-editor-answer"
            />
            <textarea
              rows={2}
              value={reactionValue}
              onChange={(e) => setReactionValue(e.target.value)}
              placeholder={t("debrief.record.guide.reactionPlaceholder")}
              aria-label={t("debrief.record.guide.reactionLabel")}
              data-testid="debrief-guided-editor-reaction"
            />
            <div>
              <button
                type="button"
                data-testid="debrief-guided-editor-cancel"
                onClick={cancelEdit}
              >
                {t("debrief.record.guide.editCancel")}
              </button>
              <button
                type="button"
                data-testid="debrief-guided-editor-save"
                disabled={saveDisabled}
                onClick={saveEdit}
              >
                {editMode === "manual"
                  ? t("debrief.record.guide.manualSave")
                  : t("debrief.record.guide.editSave")}
              </button>
            </div>
          </div>
        )}
        <div className="ei-debrief-guided__detail-grid">
          <aside
            className="ei-debrief-guided__entries"
            data-testid="debrief-guided-entries"
          >
            <div className="ei-label">{t("debrief.record.entries.eyebrow")}</div>
            <div
              className="ei-debrief-guided__card-list"
              data-testid="debrief-guided-card-list"
            >
              {entries.length === 0 ? (
                <p data-testid="debrief-guided-entries-empty">
                  {t("debrief.record.entries.empty")}
                </p>
              ) : (
                entries.map((entry, index) => (
                  <button
                    key={entry.id}
                    type="button"
                    data-testid={`debrief-entry-${entry.id}`}
                    data-active={activeEntry?.id === entry.id}
                    onClick={() => setActiveEntryId(entry.id)}
                  >
                    <span className="ei-debrief-guided__entry-head">
                      <span className="ei-mono">Q{index + 1}</span>
                      <span>
                        {t(sourceBadgeMessage(entry.source))}
                      </span>
                    </span>
                    <span>{entry.stage || entry.tag || entry.questionText}</span>
                  </button>
                ))
              )}
            </div>
          </aside>

          <article
            className="ei-debrief-guided__active-card"
            data-testid="debrief-guided-active-card"
          >
            {activeEntry ? (
              <>
                <div className="ei-debrief-guided__active-head">
                  <div>
                    <div className="ei-label">
                      {(activeEntry.stage || t("debrief.record.entries.stageFallback"))}
                      {activeEntry.tag ? ` · ${activeEntry.tag}` : ""}
                    </div>
                    <h4 className="ei-serif">{activeEntry.questionText}</h4>
                  </div>
                  <div className="ei-debrief-guided__reactions">
                    {REACTION_KEYS.map((reaction) => (
                      <button
                        key={reaction}
                        type="button"
                        data-active={(activeEntry.reaction ?? "neutral") === reaction}
                        onClick={() => setEntryField(activeEntry.id, { reaction })}
                      >
                        {t(reactionMessage(reaction))}
                      </button>
                    ))}
                  </div>
                </div>
                <label className="ei-debrief-guided__field">
                  <span className="ei-label">{t("debrief.record.guide.answerLabel")}</span>
                  <textarea
                    rows={3}
                    value={activeEntry.myAnswerSummary ?? ""}
                    onChange={(event) =>
                      setEntryField(activeEntry.id, {
                        myAnswerSummary: event.target.value,
                      })
                    }
                  />
                </label>
                <label className="ei-debrief-guided__field">
                  <span className="ei-label">{t("debrief.record.guide.followLabel")}</span>
                  <textarea
                    rows={2}
                    value={activeEntry.interviewerReaction ?? ""}
                    onChange={(event) =>
                      setEntryField(activeEntry.id, {
                        interviewerReaction: event.target.value,
                      })
                    }
                  />
                </label>
                <label className="ei-debrief-guided__field">
                  <span className="ei-label">{t("debrief.record.guide.reflectionLabel")}</span>
                  <textarea
                    rows={2}
                    value={activeEntry.reflection ?? ""}
                    onChange={(event) =>
                      setEntryField(activeEntry.id, {
                        reflection: event.target.value,
                      })
                    }
                  />
                </label>
              </>
            ) : (
              <p>{t("debrief.record.entries.empty")}</p>
            )}
          </article>
        </div>
      </div>
    </section>
  );
};
