import { useMemo, useState, type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { DebriefEntry } from "../types";

interface VoiceDebriefRecordProps {
  entries: DebriefEntry[];
  setEntries: (next: DebriefEntry[]) => void;
}

type VoicePhase = "intro" | "chat" | "review" | "committed";

interface VoiceMessage {
  id: string;
  role: "ai" | "user";
  text: string;
}

interface VoiceCard {
  id: string;
  stage: string;
  questionText: string;
  myAnswerSummary: string;
  interviewerReaction: string;
  reflection: string;
  confidenceLabel: string;
}

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx::VoiceDebriefRecord
 * (lines 656-1229). This is a fixture-backed local conversation prototype:
 * it mirrors the frontend interaction and extracted-card flow, while real
 * microphone / STT / LLM / TTS integration remains owned by the voice backend
 * contract.
 */
export const VoiceDebriefRecord: FC<VoiceDebriefRecordProps> = ({
  entries,
  setEntries,
}) => {
  const { t } = useI18n();
  const [phase, setPhase] = useState<VoicePhase>("intro");
  const [paused, setPaused] = useState(false);
  const [saved, setSaved] = useState(false);
  const voiceEntries = entries.filter((e) => e.source === "voice_extracted");
  const topics = [
    t("debrief.record.voice.topicOverall"),
    t("debrief.record.voice.topicOpening"),
    t("debrief.record.voice.topicMissed"),
    t("debrief.record.voice.topicClose"),
  ];
  const messages = useMemo<VoiceMessage[]>(
    () => [
      {
        id: "a1",
        role: "ai",
        text: t("debrief.record.voice.chat.aiMissed"),
      },
      {
        id: "u1",
        role: "user",
        text: t("debrief.record.voice.chat.userDesignSystem"),
      },
      {
        id: "a2",
        role: "ai",
        text: t("debrief.record.voice.chat.aiClose"),
      },
      {
        id: "u2",
        role: "user",
        text: t("debrief.record.voice.chat.userReverse"),
      },
      {
        id: "a3",
        role: "ai",
        text: t("debrief.record.voice.chat.aiWrap"),
      },
    ],
    [t],
  );
  const extractedCards = useMemo<VoiceCard[]>(
    () => [
      {
        id: "vc1",
        stage: "Q1",
        questionText: t("debrief.record.voice.card1.question"),
        myAnswerSummary: t("debrief.record.voice.card1.summary"),
        interviewerReaction: t("debrief.record.voice.card1.followup"),
        reflection: t("debrief.record.voice.card1.reflection"),
        confidenceLabel: t("debrief.record.voice.confidence.high"),
      },
      {
        id: "vc2",
        stage: "Q2",
        questionText: t("debrief.record.voice.card2.question"),
        myAnswerSummary: t("debrief.record.voice.card2.summary"),
        interviewerReaction: t("debrief.record.voice.card2.followup"),
        reflection: t("debrief.record.voice.card2.reflection"),
        confidenceLabel: t("debrief.record.voice.confidence.medium"),
      },
      {
        id: "vc3",
        stage: "Q3",
        questionText: t("debrief.record.voice.card3.question"),
        myAnswerSummary: t("debrief.record.voice.card3.summary"),
        interviewerReaction: t("debrief.record.voice.card3.followup"),
        reflection: t("debrief.record.voice.card3.reflection"),
        confidenceLabel: t("debrief.record.voice.confidence.high"),
      },
    ],
    [t],
  );

  const commitVoiceCards = () => {
    if (saved) {
      setPhase("committed");
      return;
    }
    setEntries([
      ...entries,
      ...extractedCards.map((card): DebriefEntry => ({
        id: `voice-${card.id}`,
        stage: t("debrief.record.voice.entryStage"),
        questionText: card.questionText,
        myAnswerSummary: card.myAnswerSummary,
        interviewerReaction: card.interviewerReaction,
        reflection: card.reflection,
        reaction: "neutral",
        source: "voice_extracted",
        tag: t("debrief.record.voice.entryTag"),
      })),
    ]);
    setSaved(true);
    setPhase("committed");
  };

  if (phase === "chat") {
    return (
      <section
        className="ei-debrief-voice ei-debrief-voice--active"
        data-testid="debrief-voice-record"
      >
        <div className="ei-debrief-voice__chat" data-testid="debrief-voice-chat">
          <div className="ei-debrief-voice__main" data-testid="debrief-voice-thread">
            <div className="ei-debrief-voice__status" data-testid="debrief-voice-status">
              <div>
                <span className="ei-debrief-voice__status-dot" aria-hidden="true" />
                <span>{paused ? t("debrief.record.voice.statusPaused") : t("debrief.record.voice.statusWrapping")}</span>
              </div>
              <span className="ei-mono">00:45</span>
            </div>
            <div className="ei-debrief-voice__messages">
              {messages.map((message) => (
                <div
                  key={message.id}
                  className={`ei-debrief-voice__message ei-debrief-voice__message--${message.role}`}
                >
                  <div className="ei-debrief-voice__speaker">
                    {message.role === "ai"
                      ? t("debrief.record.voice.speakerAi")
                      : t("debrief.record.voice.speakerUser")}
                  </div>
                  <div>{message.text}</div>
                </div>
              ))}
            </div>
            <div className="ei-debrief-voice__control">
              <button
                type="button"
                className="ei-debrief-voice__pause"
                data-testid="debrief-voice-pause"
                aria-pressed={paused}
                onClick={() => setPaused((current) => !current)}
              >
                {paused ? (
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                    <path d="M8 5v14l11-7z" />
                  </svg>
                ) : (
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                    <rect x="6" y="5" width="4" height="14" rx="1" />
                    <rect x="14" y="5" width="4" height="14" rx="1" />
                  </svg>
                )}
              </button>
              <div>
                <div>{paused ? t("debrief.record.voice.controlPaused") : t("debrief.record.voice.controlListening")}</div>
                <small>{paused ? t("debrief.record.voice.controlResumeHint") : t("debrief.record.voice.controlHint")}</small>
              </div>
              <button
                type="button"
                className="ei-debrief-voice__end"
                data-testid="debrief-voice-end-review"
                onClick={() => setPhase("review")}
              >
                ✓ {t("debrief.record.voice.endReview")}
              </button>
            </div>
          </div>

          <aside className="ei-debrief-voice__extract" data-testid="debrief-voice-live-extract">
            <div className="ei-label">
              {t("debrief.record.voice.extracting")}
              <span>{extractedCards.length}</span>
            </div>
            <div className="ei-debrief-voice__extract-list">
              {extractedCards.map((card) => (
                <article key={card.id} data-testid="debrief-voice-extracted-card">
                  <div className="ei-mono">{card.stage} · {card.confidenceLabel}</div>
                  <p>{card.questionText}</p>
                </article>
              ))}
            </div>
            <button type="button" className="ei-debrief-voice__manual">
              + {t("debrief.record.voice.addManual")}
            </button>
            <p>{t("debrief.record.voice.confirmLater")}</p>
          </aside>
        </div>
      </section>
    );
  }

  if (phase === "review") {
    return (
      <section
        className="ei-debrief-voice ei-debrief-voice--active"
        data-testid="debrief-voice-record"
      >
        <div className="ei-debrief-voice__review" data-testid="debrief-voice-review">
          <div className="ei-debrief-voice__review-head">
            <div>
              <div className="ei-label">{t("debrief.record.voice.reviewTitle")}</div>
              <p>
                {t("debrief.record.voice.reviewStats").replace(
                  "{count}",
                  String(extractedCards.length),
                )}
              </p>
            </div>
            <div>
              <button type="button" onClick={() => setPhase("chat")}>
                {t("debrief.record.voice.backToChat")}
              </button>
              <button
                type="button"
                className="ei-debrief-voice__end"
                data-testid="debrief-voice-save"
                onClick={commitVoiceCards}
              >
                ✓ {t("debrief.record.voice.saveCards").replace("{count}", String(extractedCards.length))}
              </button>
            </div>
          </div>
          <div className="ei-debrief-voice__review-list">
            {extractedCards.map((card) => (
              <article key={card.id}>
                <div className="ei-mono">{card.stage} · {card.confidenceLabel}</div>
                <h4>{card.questionText}</h4>
                <p>{card.myAnswerSummary}</p>
                <p>{card.interviewerReaction}</p>
              </article>
            ))}
          </div>
        </div>
      </section>
    );
  }

  if (phase === "committed") {
    return (
      <section
        className="ei-debrief-voice ei-debrief-voice--active"
        data-testid="debrief-voice-record"
      >
        <div className="ei-debrief-voice__committed" data-testid="debrief-voice-committed">
          <div className="ei-label">{t("debrief.record.voice.committedTitle")}</div>
          <p>{t("debrief.record.voice.committedBody")}</p>
          <button type="button" onClick={() => setPhase("chat")}>
            {t("debrief.record.voice.backToChat")}
          </button>
        </div>
      </section>
    );
  }

  return (
    <section
      className="ei-debrief-voice"
      data-testid="debrief-voice-record"
    >
      <div className="ei-debrief-voice__card" data-testid="debrief-voice-intro-card">
        <div className="ei-debrief-voice__intro">
          <div className="ei-debrief-voice__icon" aria-hidden="true">
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z" />
              <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
            </svg>
          </div>
          <div>
            <h3>{t("debrief.record.voice.title")}</h3>
            <p>{t("debrief.record.voice.description")}</p>
          </div>
        </div>
        <div className="ei-debrief-voice__topics">
          <div className="ei-label">{t("debrief.record.voice.topicsLabel")}</div>
          <ol data-testid="debrief-voice-topic-list">
            {topics.map((topic, index) => (
              <li key={topic}>
                <span>{String(index + 1).padStart(2, "0")}</span>
                {topic}
              </li>
            ))}
          </ol>
        </div>
        <button
          type="button"
          className="ei-debrief-voice__start"
          data-testid="debrief-voice-start"
          onClick={() => setPhase("chat")}
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z" />
            <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
          </svg>
          {t("debrief.record.voice.start")}
        </button>
        <div className="ei-debrief-voice__tip">{t("debrief.record.voice.tip")}</div>
      </div>
      {voiceEntries.length > 0 && (
        <ul className="ei-debrief-voice__pending" data-testid="debrief-voice-pending-list">
          {voiceEntries.map((entry) => (
            <li key={entry.id} data-testid={`debrief-voice-pending-${entry.id}`}>
              {entry.questionText}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
};
