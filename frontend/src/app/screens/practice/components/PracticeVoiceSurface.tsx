import { useMemo, type CSSProperties, type FC } from "react";

import type { PracticeVoiceTTSError } from "../../../../api/generated/types";
import type { Lang } from "../../../i18n/messages";
import type { TranscriptMessage } from "./Transcript";
import type { PracticeVoiceTurnState } from "../hooks/usePracticeVoiceTurn";

export interface PracticeVoiceSurfaceProps {
  lang: Lang;
  questionBadge: string;
  topic: string;
  prompt: string;
  recording: boolean;
  messages: TranscriptMessage[];
  captureState: PracticeVoiceTurnState["kind"];
  manualTranscriptFallback: string;
  onManualTranscriptFallbackChange: (next: string) => void;
  onStartRecording: () => void;
  onSubmitRecording: () => void;
  controlsDisabled: boolean;
  voiceError: string | null;
  ttsError: PracticeVoiceTTSError | null;
  ttsChunkCount: number | null;
}

type VoiceTranscriptMessage = TranscriptMessage | {
  role: "note";
  text: string;
  t: string;
};

interface VoiceAnnotation {
  at: number;
  kind: "pause" | "filler" | "pace";
  dur?: number;
  label: string;
}

const WAVEFORM_BAR_COUNT = 70;
const ANNOTATED_SAMPLE_COUNT = 200;

/**
 * Formal voice-mode surface for PracticeScreen. It mirrors
 * `ui-design/src/screen-practice.jsx` lines 451-523 without importing
 * prototype runtime helpers or creating a standalone voice route.
 */
export const PracticeVoiceSurface: FC<PracticeVoiceSurfaceProps> = ({
  lang,
  questionBadge,
  topic,
  prompt,
  recording,
  messages,
  captureState,
  manualTranscriptFallback,
  onManualTranscriptFallbackChange,
  onStartRecording,
  onSubmitRecording,
  controlsDisabled,
  voiceError,
  ttsError,
  ttsChunkCount,
}) => {
  const samples = useMemo(buildAnnotatedSamples, []);
  const annotations = useMemo(() => buildAnnotations(lang), [lang]);
  const transcript = useMemo(
    () => buildVoiceTranscript(lang, messages),
    [lang, messages],
  );
  const copy = voiceSurfaceCopy(lang);

  return (
    <>
      <div
        data-testid="practice-voice-question"
        style={{
          padding: "24px 34px 18px",
          borderBottom: "1px solid var(--ei-color-rule-strong)",
          background: "var(--ei-color-bg-card)",
        }}
      >
        <div
          style={{
            display: "flex",
            gap: 8,
            marginBottom: 10,
            flexWrap: "wrap",
            alignItems: "center",
          }}
        >
          <span
            className="ei-mono"
            style={{
              padding: "3px 8px",
              borderRadius: 3,
              fontSize: 11.5,
              background: "var(--ei-color-accent-soft)",
              color: "var(--ei-color-accent)",
            }}
          >
            {questionBadge} · {topic}
          </span>
        </div>
        <div
          className="ei-serif"
          style={{
            fontSize: 22,
            color: "var(--ei-color-fg-primary)",
            lineHeight: 1.35,
            letterSpacing: 0,
          }}
        >
          {prompt}
        </div>
      </div>

      <div
        data-testid="practice-voice-surface"
        style={{
          flex: 1,
          overflowY: "auto",
          padding: "22px 34px",
          display: "flex",
          flexDirection: "column",
          gap: 20,
        }}
      >
        <div>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: 10,
            }}
          >
            <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)" }}>
              {copy.nowSpeaking}
            </div>
            <div
              style={{
                fontFamily: "var(--ei-font-mono)",
                fontSize: 11,
                color: "var(--ei-color-fg-tertiary)",
              }}
            >
              -12 dB
            </div>
          </div>
          <div
            data-testid="practice-voice-waveform"
            style={{
              display: "flex",
              alignItems: "center",
              gap: 14,
              padding: "16px 18px",
              background: "var(--ei-color-bg-card)",
              border: "1px solid var(--ei-color-rule-strong)",
              borderRadius: 3,
            }}
          >
            <div
              aria-hidden="true"
              style={{
                width: 38,
                height: 38,
                borderRadius: 19,
                background: "var(--ei-color-accent)",
                color: "#fff",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                flexShrink: 0,
                fontSize: 18,
              }}
            >
              ●
            </div>
            <VoiceWaveformBars active={recording} />
          </div>
        </div>

        <VoiceCaptureControls
          lang={lang}
          captureState={captureState}
          manualTranscriptFallback={manualTranscriptFallback}
          onManualTranscriptFallbackChange={onManualTranscriptFallbackChange}
          onStartRecording={onStartRecording}
          onSubmitRecording={onSubmitRecording}
          disabled={controlsDisabled}
          voiceError={voiceError}
          ttsError={ttsError}
          ttsChunkCount={ttsChunkCount}
        />

        <div>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: 10,
              gap: 12,
            }}
          >
            <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)" }}>
              {copy.annotatedAnswer}
            </div>
            <div
              style={{
                display: "flex",
                gap: 12,
                fontFamily: "var(--ei-font-mono)",
                fontSize: 10.5,
                flexWrap: "wrap",
                justifyContent: "flex-end",
              }}
            >
              <LegendSwatch
                label={copy.pause}
                color="var(--ei-color-warn)"
                background="var(--ei-color-amber-soft)"
              />
              <LegendSwatch
                label={copy.filler}
                color="var(--ei-color-danger)"
                background="var(--ei-color-danger)"
                rounded
              />
              <LegendSwatch
                label={copy.pace}
                color="var(--ei-color-cool)"
                background="var(--ei-color-cool-soft)"
              />
            </div>
          </div>
          <div
            data-testid="practice-voice-annotated-waveform"
            style={{
              padding: "16px 18px 4px",
              background: "var(--ei-color-bg-card)",
              border: "1px solid var(--ei-color-rule-strong)",
              borderRadius: 3,
            }}
          >
            <PracticeVoiceAnnotatedWaveform
              samples={samples}
              annotations={annotations}
            />
          </div>
        </div>

        <div data-testid="practice-voice-live-transcript">
          <div
            className="ei-label"
            style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}
          >
            {copy.liveTranscript}
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {transcript.map((message, idx) => (
              <VoiceTranscriptRow
                key={`${message.role}-${idx}`}
                message={message}
                lang={lang}
              />
            ))}
            <div style={{ display: "flex", gap: 12 }}>
              <div style={{ width: 58, flexShrink: 0 }} />
              <div
                className={recording ? "ei-pulse" : ""}
                style={{
                  width: 8,
                  height: 16,
                  background: recording
                    ? "var(--ei-color-accent)"
                    : "var(--ei-color-fg-muted)",
                }}
              />
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

const VoiceCaptureControls: FC<{
  lang: Lang;
  captureState: PracticeVoiceTurnState["kind"];
  manualTranscriptFallback: string;
  onManualTranscriptFallbackChange: (next: string) => void;
  onStartRecording: () => void;
  onSubmitRecording: () => void;
  disabled: boolean;
  voiceError: string | null;
  ttsError: PracticeVoiceTTSError | null;
  ttsChunkCount: number | null;
}> = ({
  lang,
  captureState,
  manualTranscriptFallback,
  onManualTranscriptFallbackChange,
  onStartRecording,
  onSubmitRecording,
  disabled,
  voiceError,
  ttsError,
  ttsChunkCount,
}) => {
  const copy = voiceControlCopy(lang);
  const recording = captureState === "recording";
  const submitting = captureState === "submitting";
  return (
    <div
      data-testid="practice-voice-capture"
      style={{
        display: "grid",
        gap: 10,
        padding: "14px 16px",
        background: "var(--ei-color-bg-card)",
        border: "1px solid var(--ei-color-rule-strong)",
        borderRadius: 3,
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          gap: 12,
          flexWrap: "wrap",
        }}
      >
        <div
          data-testid="practice-voice-capture-status"
          data-state={captureState}
          className="ei-mono"
          style={{
            fontSize: 11,
            color: recording
              ? "var(--ei-color-accent)"
              : "var(--ei-color-fg-tertiary)",
          }}
        >
          {copy.status[captureState]}
        </div>
        <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
          <button
            data-testid="practice-voice-record-toggle"
            data-state={captureState}
            type="button"
            disabled={disabled || recording || submitting}
            onClick={onStartRecording}
            style={voiceButtonStyle("primary", disabled || recording || submitting)}
          >
            {copy.start}
          </button>
          <button
            data-testid="practice-voice-submit"
            type="button"
            disabled={disabled || !recording || submitting}
            onClick={onSubmitRecording}
            style={voiceButtonStyle("secondary", disabled || !recording || submitting)}
          >
            {copy.submit}
          </button>
        </div>
      </div>
      <textarea
        data-testid="practice-voice-manual-fallback"
        value={manualTranscriptFallback}
        disabled={submitting}
        onChange={(event) => onManualTranscriptFallbackChange(event.target.value)}
        placeholder={copy.manualPlaceholder}
        style={{
          width: "100%",
          minHeight: 56,
          resize: "vertical",
          background: "var(--ei-color-bg-soft)",
          border: "1px solid var(--ei-color-rule-strong)",
          borderRadius: 2,
          padding: "8px 10px",
          color: "var(--ei-color-fg-primary)",
          fontFamily: "var(--ei-font-sans)",
          fontSize: 13,
          lineHeight: 1.45,
          boxSizing: "border-box",
        }}
      />
      {voiceError ? (
        <div
          data-testid="practice-voice-error"
          style={{
            color: "var(--ei-color-danger)",
            fontSize: 12,
            fontFamily: "var(--ei-font-mono)",
          }}
        >
          {voiceError}
        </div>
      ) : null}
      {ttsError ? (
        <div
          data-testid="practice-voice-tts-error"
          style={{
            color: "var(--ei-color-warn)",
            fontSize: 12,
            fontFamily: "var(--ei-font-mono)",
          }}
        >
          {ttsError.code} · {ttsError.message}
        </div>
      ) : null}
      {ttsChunkCount !== null ? (
        <div
          data-testid="practice-voice-tts-status"
          style={{
            color: "var(--ei-color-fg-tertiary)",
            fontSize: 12,
            fontFamily: "var(--ei-font-mono)",
          }}
        >
          {copy.ttsStatus.replace("{n}", String(ttsChunkCount))}
        </div>
      ) : null}
    </div>
  );
};

const VoiceWaveformBars: FC<{ active: boolean }> = ({ active }) => {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: 2,
        height: 48,
        flex: 1,
        minWidth: 0,
      }}
    >
      {Array.from({ length: WAVEFORM_BAR_COUNT }).map((_, idx) => {
        const phase = idx * 0.18;
        const seed = Math.sin(idx * 1.3) * 0.5 + 0.5;
        const wobble = active
          ? Math.sin(phase) * 0.28 + Math.cos(phase * 1.7) * 0.18
          : 0;
        const height = Math.max(3, (seed * 0.6 + 0.2 + wobble) * 48);
        const recent = idx > WAVEFORM_BAR_COUNT - 8;
        return (
          <div
            key={idx}
            style={{
              flex: 1,
              height,
              minWidth: 2,
              borderRadius: 1,
              background: recent && active
                ? "var(--ei-color-accent)"
                : "var(--ei-color-fg-muted)",
              opacity: recent ? 1 : 0.55,
            }}
          />
        );
      })}
    </div>
  );
};

const PracticeVoiceAnnotatedWaveform: FC<{
  samples: number[];
  annotations: VoiceAnnotation[];
}> = ({ samples, annotations }) => {
  const width = 880;
  const height = 72;
  const mid = height / 2;
  const step = width / samples.length;
  return (
    <svg
      aria-hidden="true"
      width="100%"
      viewBox={`0 0 ${width} ${height + 30}`}
      style={{ display: "block" }}
    >
      <line
        x1="0"
        y1={mid}
        x2={width}
        y2={mid}
        stroke="var(--ei-color-rule-strong)"
        strokeWidth="1"
      />
      {samples.map((value, idx) => {
        const barHeight = Math.max(1, Math.abs(value) * (height / 2 - 4));
        return (
          <rect
            key={idx}
            x={idx * step}
            y={mid - barHeight}
            width={Math.max(1, step - 0.5)}
            height={barHeight * 2}
            fill="var(--ei-color-fg-tertiary)"
            opacity={value < 0 ? 0.45 : 0.85}
          />
        );
      })}
      {annotations.map((annotation, idx) => {
        const x = annotation.at * width;
        if (annotation.kind === "pause") {
          const pauseWidth = (annotation.dur ?? 0.02) * width;
          return (
            <g key={idx}>
              <rect
                x={x}
                y={4}
                width={pauseWidth}
                height={height - 8}
                fill="var(--ei-color-amber-soft)"
                opacity="0.55"
              />
              <line
                x1={x}
                y1={4}
                x2={x}
                y2={height - 4}
                stroke="var(--ei-color-warn)"
                strokeWidth="1"
                strokeDasharray="2 2"
              />
              <line
                x1={x + pauseWidth}
                y1={4}
                x2={x + pauseWidth}
                y2={height - 4}
                stroke="var(--ei-color-warn)"
                strokeWidth="1"
                strokeDasharray="2 2"
              />
              <text
                x={x + pauseWidth / 2}
                y={height + 15}
                textAnchor="middle"
                fontSize="9.5"
                fill="var(--ei-color-warn)"
                fontFamily="var(--ei-font-mono)"
                letterSpacing="0"
              >
                PAUSE {annotation.label}
              </text>
            </g>
          );
        }
        if (annotation.kind === "filler") {
          return (
            <g key={idx}>
              <circle cx={x} cy={mid} r="4" fill="var(--ei-color-danger)" />
              <line
                x1={x}
                y1={mid + 6}
                x2={x}
                y2={height + 3}
                stroke="var(--ei-color-danger)"
                strokeWidth="1"
              />
              <text
                x={x}
                y={height + 15}
                textAnchor="middle"
                fontSize="9.5"
                fill="var(--ei-color-danger)"
                fontFamily="var(--ei-font-mono)"
              >
                {annotation.label}
              </text>
            </g>
          );
        }
        return (
          <g key={idx}>
            <line
              x1={x}
              y1={0}
              x2={x}
              y2={height}
              stroke="var(--ei-color-cool)"
              strokeWidth="1"
              strokeDasharray="3 3"
            />
            <rect
              x={x - 20}
              y={height + 3}
              width="40"
              height="14"
              fill="var(--ei-color-cool-soft)"
              rx="2"
            />
            <text
              x={x}
              y={height + 14}
              textAnchor="middle"
              fontSize="9.5"
              fill="var(--ei-color-cool)"
              fontFamily="var(--ei-font-mono)"
            >
              {annotation.label}
            </text>
          </g>
        );
      })}
    </svg>
  );
};

const LegendSwatch: FC<{
  label: string;
  color: string;
  background: string;
  rounded?: boolean;
}> = ({ label, color, background, rounded = false }) => (
  <span style={{ color, display: "flex", gap: 4, alignItems: "center" }}>
    <span
      style={{
        width: 8,
        height: 8,
        borderRadius: rounded ? 4 : 0,
        background,
        border: rounded ? "none" : `1px solid ${color}`,
      }}
    />
    {label}
  </span>
);

const VoiceTranscriptRow: FC<{
  message: VoiceTranscriptMessage;
  lang: Lang;
}> = ({ message, lang }) => {
  if (message.role === "note") {
    return (
      <div
        style={{
          display: "flex",
          gap: 10,
          padding: "6px 12px",
          background: "var(--ei-color-warn-soft)",
          borderLeft: "2px solid var(--ei-color-warn)",
          fontSize: 12,
          color: "var(--ei-color-warn)",
          fontFamily: "var(--ei-font-mono)",
        }}
      >
        <span aria-hidden="true">i</span>
        {message.text}
      </div>
    );
  }
  const isAi = message.role === "ai";
  return (
    <div style={{ display: "flex", gap: 12 }}>
      <div
        style={{
          fontFamily: "var(--ei-font-mono)",
          fontSize: 11,
          color: "var(--ei-color-fg-muted)",
          width: 58,
          flexShrink: 0,
          paddingTop: 2,
        }}
      >
        {message.t}
      </div>
      <div
        style={{
          fontSize: 14,
          lineHeight: 1.55,
          color: isAi
            ? "var(--ei-color-fg-tertiary)"
            : "var(--ei-color-fg-primary)",
          fontStyle: isAi ? "italic" : "normal",
          flex: 1,
        }}
      >
        {message.text}
        {!isAi && (
          <span
            style={{
              display: "inline-block",
              marginLeft: 8,
              padding: "1px 6px",
              background: "var(--ei-color-amber-soft)",
              color: "var(--ei-color-warn)",
              fontSize: 11,
              fontFamily: "var(--ei-font-mono)",
              borderRadius: 2,
            }}
          >
            II {lang === "en" ? "1.2s" : "1.2 秒"}
          </span>
        )}
      </div>
    </div>
  );
};

function buildAnnotatedSamples(): number[] {
  const out: number[] = [];
  for (let idx = 0; idx < ANNOTATED_SAMPLE_COUNT; idx += 1) {
    const env = Math.sin((idx / ANNOTATED_SAMPLE_COUNT) * Math.PI * 3) * 0.5 + 0.5;
    const signal = (
      Math.sin(idx * 0.9) +
      Math.sin(idx * 0.3) * 0.7 +
      Math.sin(idx * 2.1) * 0.18
    ) * 0.5;
    let value = signal * (0.3 + env * 0.7);
    if (idx > 58 && idx < 70) value *= 0.05;
    if (idx > 120 && idx < 138) value *= 0.08;
    out.push(value);
  }
  return out;
}

function buildAnnotations(lang: Lang): VoiceAnnotation[] {
  return [
    { at: 0.30, kind: "pause", dur: 0.06, label: "0.8s" },
    {
      at: 0.60,
      kind: "pause",
      dur: 0.09,
      label: lang === "en" ? "1.6s · long" : "1.6s · long",
    },
    { at: 0.44, kind: "filler", label: lang === "en" ? "um..." : "um..." },
    {
      at: 0.78,
      kind: "filler",
      label: lang === "en" ? "basically" : "basically",
    },
    { at: 0.15, kind: "pace", label: lang === "en" ? "steady" : "steady" },
    { at: 0.85, kind: "pace", label: lang === "en" ? "fast" : "fast" },
  ];
}

function buildVoiceTranscript(
  lang: Lang,
  messages: TranscriptMessage[],
): VoiceTranscriptMessage[] {
  if (messages.length > 0) return messages;
  return lang === "en"
    ? [
      {
        t: "00:02:14",
        role: "ai",
        text: "Take your time. When you're ready, start with the situation.",
      },
      {
        t: "00:02:18",
        role: "user",
        text: "OK. The project was our order repricing system...",
      },
      {
        t: "00:02:49",
        role: "note",
        text: "Long pause detected · 1.6s",
      },
    ]
    : [
      {
        t: "00:02:14",
        role: "ai",
        text: "别急，你准备好了就从当时的情境开始。",
      },
      {
        t: "00:02:18",
        role: "user",
        text: "好。那个项目是我们的订单改价系统。",
      },
      {
        t: "00:02:49",
        role: "note",
        text: "检测到长停顿 · 1.6 秒",
      },
    ];
}

function voiceSurfaceCopy(lang: Lang): {
  nowSpeaking: string;
  annotatedAnswer: string;
  pause: string;
  filler: string;
  pace: string;
  liveTranscript: string;
} {
  return lang === "en"
    ? {
      nowSpeaking: "NOW SPEAKING",
      annotatedAnswer: "CURRENT ANSWER · annotated",
      pause: "pause",
      filler: "filler",
      pace: "pace",
      liveTranscript: "Live transcript",
    }
    : {
      nowSpeaking: "正在说话",
      annotatedAnswer: "本次回答 · 已标注",
      pause: "停顿",
      filler: "口头禅",
      pace: "语速",
      liveTranscript: "实时转写",
    };
}

function voiceControlCopy(lang: Lang): {
  start: string;
  submit: string;
  manualPlaceholder: string;
  ttsStatus: string;
  status: Record<PracticeVoiceTurnState["kind"], string>;
} {
  return lang === "en"
    ? {
      start: "Start recording",
      submit: "Submit turn",
      manualPlaceholder: "Optional transcript fallback for STT recovery",
      ttsStatus: "TTS chunks ready: {n}",
      status: {
        idle: "VOICE CAPTURE · idle",
        recording: "VOICE CAPTURE · recording",
        submitting: "VOICE CAPTURE · submitting",
        success: "VOICE CAPTURE · submitted",
        error: "VOICE CAPTURE · needs attention",
      },
    }
    : {
      start: "开始录音",
      submit: "提交本轮",
      manualPlaceholder: "可选：语音识别不准时补充文字转写",
      ttsStatus: "TTS chunks ready: {n}",
      status: {
        idle: "VOICE CAPTURE · idle",
        recording: "VOICE CAPTURE · recording",
        submitting: "VOICE CAPTURE · submitting",
        success: "VOICE CAPTURE · submitted",
        error: "VOICE CAPTURE · needs attention",
      },
    };
}

function voiceButtonStyle(
  tone: "primary" | "secondary",
  disabled: boolean,
): CSSProperties {
  const primary = tone === "primary";
  return {
    background: primary
      ? "var(--ei-color-accent)"
      : "var(--ei-color-bg-soft)",
    border: primary
      ? "1px solid var(--ei-color-accent)"
      : "1px solid var(--ei-color-rule-strong)",
    color: primary ? "#fff" : "var(--ei-color-fg-secondary)",
    padding: "7px 11px",
    borderRadius: 2,
    fontSize: 12,
    fontFamily: "var(--ei-font-sans)",
    cursor: disabled ? "not-allowed" : "pointer",
    opacity: disabled ? 0.55 : 1,
  };
}
