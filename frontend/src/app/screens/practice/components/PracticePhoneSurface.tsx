import {
  useEffect,
  useMemo,
  useState,
  type CSSProperties,
  type FC,
} from "react";

import type { Lang } from "../../../i18n/messages";
import type { PracticeVoicePlaybackState } from "../hooks/usePracticeVoicePlayback";
import type { PracticeVoiceTurnState } from "../hooks/usePracticeVoiceTurn";
import { PhoneIcon } from "./PhoneIcon";
import { Transcript, type TranscriptMessage } from "./Transcript";

export interface PracticePhoneSurfaceProps {
  lang: Lang;
  active: boolean;
  captureState: PracticeVoiceTurnState["kind"];
  playbackState: PracticeVoicePlaybackState["kind"];
  voiceError: string | null;
  playbackError: string | null;
  messages: TranscriptMessage[];
  aiLabel: string;
  userLabel: string;
  followUpLabel: string;
  onHangUp: () => void | Promise<void>;
}

const WAVEFORM_BAR_COUNT = 66;

export const PracticePhoneSurface: FC<PracticePhoneSurfaceProps> = ({
  lang,
  active,
  captureState,
  playbackState,
  voiceError,
  playbackError,
  messages,
  aiLabel,
  userLabel,
  followUpLabel,
  onHangUp,
}) => {
  const [captionsShown, setCaptionsShown] = useState(false);
  const copy = phoneCopy(lang);
  const errorMessage = voiceError || playbackError;

  return (
    <div
      data-testid="practice-phone-surface"
      style={{
        flex: 1,
        minHeight: 0,
        display: "flex",
        flexDirection: "column",
        background: "var(--ei-color-bg-canvas)",
      }}
    >
      <div
        style={{
          flex: captionsShown ? "0 0 auto" : 1,
          padding: "46px 56px 28px",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          gap: 22,
        }}
      >
        <div
          style={{
            width: 96,
            height: 96,
            borderRadius: 48,
            background: "var(--ei-color-accent-soft)",
            border: "1px solid var(--ei-color-accent)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            color: "var(--ei-color-accent)",
          }}
        >
          <MicrophoneIcon />
        </div>
          <div
            data-testid="practice-phone-call-state"
            data-state={active ? "connected" : "paused"}
            data-capture-state={captureState}
            data-playback-state={playbackState}
            style={{ textAlign: "center" }}
          >
            <div
              className="ei-serif"
              style={{
                fontSize: 28,
                color: "var(--ei-color-fg-primary)",
                letterSpacing: 0,
              }}
            >
              {copy.liveTitle}
            </div>
            <div
              style={{
                marginTop: 8,
                fontSize: 13,
                color: "var(--ei-color-fg-tertiary)",
                lineHeight: 1.5,
              }}
            >
              {copy.liveSub}
            </div>
            {errorMessage ? (
              <div
                data-testid="practice-phone-error"
                role="alert"
                style={{
                  marginTop: 8,
                  fontSize: 12,
                  color: "var(--ei-color-danger)",
                  lineHeight: 1.45,
                }}
              >
                {errorMessage}
              </div>
            ) : null}
          </div>
          <div
            data-testid="practice-phone-waveform"
            style={{
              display: "flex",
              alignItems: "center",
              gap: 16,
              width: "min(720px, 100%)",
              padding: "18px 22px",
              background: "var(--ei-color-bg-card)",
              border: "1px solid var(--ei-color-rule-strong)",
              borderRadius: 4,
            }}
          >
            <div
              data-testid="practice-phone-waveform-status"
              data-icon={active ? "mic" : "pause"}
              className={active ? "ei-pulse" : ""}
              style={{
                width: 34,
                height: 34,
                borderRadius: 17,
                background: active
                  ? "var(--ei-color-accent)"
                  : "var(--ei-color-fg-muted)",
                color: "#fff",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                flexShrink: 0,
                fontSize: 16,
              }}
            >
              {active ? <MicrophoneIcon size={15} /> : <PauseIcon size={15} />}
            </div>
            <PhoneWaveformBars active={active} />
          </div>
          <div
            style={{
              display: "flex",
              gap: 10,
              flexWrap: "wrap",
              justifyContent: "center",
            }}
          >
            <button
              data-testid="practice-phone-captions-toggle"
              type="button"
              onClick={() => setCaptionsShown((v) => !v)}
              style={phoneButtonStyle(captionsShown ? "primary" : "secondary")}
            >
              <ChatIcon size={13} />
              {captionsShown ? copy.hideCaptions : copy.showCaptions}
            </button>
            <button
              data-testid="practice-phone-hangup"
              type="button"
              aria-label={copy.hangUpLabel}
              title={copy.hangUpTitle}
              onClick={() => void onHangUp()}
              style={hangUpButtonStyle}
            >
              <PhoneIcon
                size={22}
                strokeWidth={1.8}
                style={{ transform: "rotate(135deg)" }}
              />
            </button>
          </div>
        </div>
        {captionsShown && (
          <div
            data-testid="practice-phone-captions"
            style={{
              borderTop: "1px solid var(--ei-color-rule-strong)",
              background: "var(--ei-color-bg-card)",
              minHeight: 220,
              maxHeight: "42vh",
              display: "flex",
              flexDirection: "column",
            }}
          >
            <div
              style={{
                padding: "12px 22px 0",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
              }}
            >
              <div
                className="ei-label"
                style={{ color: "var(--ei-color-fg-tertiary)" }}
              >
                {copy.captions}
              </div>
              <span
                data-testid="practice-phone-captions-session-tag"
                className="ei-mono"
                style={{
                  display: "inline-flex",
                  alignItems: "center",
                  padding: "3px 8px",
                  borderRadius: 3,
                  fontSize: 11.5,
                  background: "var(--ei-color-bg-soft)",
                  color: "var(--ei-color-fg-tertiary)",
                  border: "1px solid var(--ei-color-rule-strong)",
                }}
              >
                {copy.sameSessionTranscript}
              </span>
            </div>
            <Transcript
              messages={messages}
              helperText=""
              aiLabel={aiLabel}
              userLabel={userLabel}
              followUpLabel={followUpLabel}
            />
          </div>
        )}
    </div>
  );
};

const PhoneWaveformBars: FC<{ active: boolean }> = ({ active }) => {
  const [tick, setTick] = useState(0);
  useEffect(() => {
    if (!active) return;
    const interval = window.setInterval(() => setTick((value) => value + 1), 90);
    return () => window.clearInterval(interval);
  }, [active]);

  const heights = useMemo(
    () =>
      Array.from({ length: WAVEFORM_BAR_COUNT }, (_, index) => {
        const phase = (tick + index * 3) * 0.18;
        const seed = Math.sin(index * 1.3) * 0.5 + 0.5;
        const wobble = active
          ? Math.sin(phase) * 0.5 + Math.cos(phase * 1.7) * 0.35
          : 0;
        const accent = index > WAVEFORM_BAR_COUNT - 10;
        return {
          height: Math.max(5, (seed * 0.55 + 0.24 + wobble * 0.38) * 58),
          accent,
        };
      }),
    [active, tick],
  );
  return (
    <div
      data-testid="practice-phone-waveform-bars"
      style={{
        display: "flex",
        alignItems: "center",
        gap: 3,
        height: 58,
        flex: 1,
        minWidth: 0,
      }}
    >
      {heights.map((item, index) => (
        <div
          key={index}
          style={{
            flex: 1,
            minWidth: 2,
            height: item.height,
            borderRadius: 1,
            background:
              item.accent && active
                ? "var(--ei-color-accent)"
                : "var(--ei-color-fg-muted)",
            opacity: item.accent ? 1 : 0.5,
            transition: "height .09s ease",
          }}
        />
      ))}
    </div>
  );
};

function phoneButtonStyle(
  variant: "primary" | "secondary",
  disabled = false,
): CSSProperties {
  const color =
    variant === "primary"
      ? "var(--ei-color-accent)"
      : "var(--ei-color-fg-secondary)";
  return {
    background:
      variant === "primary" ? "var(--ei-color-accent-soft)" : "transparent",
    border: `1px solid ${
      variant === "primary"
        ? "var(--ei-color-accent)"
        : "var(--ei-color-rule-strong)"
    }`,
    color: disabled ? "var(--ei-color-fg-muted)" : color,
    padding: "8px 12px",
    borderRadius: 2,
    fontSize: 12.5,
    display: "flex",
    alignItems: "center",
    gap: 6,
    cursor: disabled ? "default" : "pointer",
  };
}

const hangUpButtonStyle: CSSProperties = {
  width: 56,
  height: 56,
  padding: 0,
  background: "var(--ei-color-danger)",
  border: "none",
  color: "#fff",
  borderRadius: 28,
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  boxShadow: "0 6px 18px rgba(179,64,43,0.24)",
  cursor: "pointer",
};

const MicrophoneIcon: FC<{ size?: number }> = ({ size = 34 }) => (
  <svg
    aria-hidden="true"
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="none"
    stroke="currentColor"
    strokeWidth={1.7}
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <rect x="9" y="2" width="6" height="12" rx="3" />
    <path d="M5 10a7 7 0 0 0 14 0M12 17v5M8 22h8" />
  </svg>
);

const PauseIcon: FC<{ size?: number }> = ({ size = 15 }) => (
  <svg
    aria-hidden="true"
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="currentColor"
  >
    <path d="M6 5h4v14H6zM14 5h4v14h-4z" />
  </svg>
);

const ChatIcon: FC<{ size?: number }> = ({ size = 13 }) => (
  <svg
    aria-hidden="true"
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="none"
    stroke="currentColor"
    strokeWidth={1.7}
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="M4 5h16v11H9l-5 4V5z" />
  </svg>
);

function phoneCopy(lang: Lang): {
  liveTitle: string;
  liveSub: string;
  showCaptions: string;
  hideCaptions: string;
  captions: string;
  hangUpLabel: string;
  hangUpTitle: string;
  sameSessionTranscript: string;
} {
  if (lang === "en") {
    return {
      liveTitle: "Phone interview in progress",
      liveSub: "Listen and answer naturally. Captions are optional.",
      showCaptions: "Show captions",
      hideCaptions: "Hide captions",
      captions: "Captions",
      hangUpLabel: "Hang up and return to text mode",
      hangUpTitle: "Hang up",
      sameSessionTranscript: "same session transcript",
    };
  }
  return {
    liveTitle: "电话模式进行中",
    liveSub: "像真实电话一样听题并回答；字幕可按需显示。",
    showCaptions: "显示字幕",
    hideCaptions: "隐藏字幕",
    captions: "字幕",
    hangUpLabel: "挂断并返回文本模式",
    hangUpTitle: "挂断",
    sameSessionTranscript: "同一会话记录",
  };
}
