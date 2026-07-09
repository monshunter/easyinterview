import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type FC,
} from "react";

import type { Lang } from "../../../i18n/messages";
import type { PracticeVoicePlaybackState } from "../hooks/usePracticeVoicePlayback";
import type { PracticeVoiceTurnState } from "../hooks/usePracticeVoiceTurn";
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
  onStartCall: () => void | Promise<void>;
  onPauseCapture: () => void | Promise<void>;
  onHangUp: () => void | Promise<void>;
  onRestartCall: () => void | Promise<void>;
}

const WAVEFORM_BAR_COUNT = 70;

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
  onStartCall,
  onPauseCapture,
  onHangUp,
  onRestartCall,
}) => {
  const [captionsShown, setCaptionsShown] = useState(false);
  const [callEnded, setCallEnded] = useState(false);
  const autoStartedRef = useRef(false);
  const copy = phoneCopy(lang);
  const errorMessage = voiceError || playbackError;

  useEffect(() => {
    if (!active || callEnded || autoStartedRef.current) return;
    autoStartedRef.current = true;
    void onStartCall();
  }, [active, callEnded, onStartCall]);

  useEffect(() => {
    if (active || callEnded || captureState !== "recording") return;
    autoStartedRef.current = false;
    void onPauseCapture();
  }, [active, callEnded, captureState, onPauseCapture]);

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
            data-testid="practice-phone-call-state"
            data-state={callEnded ? "ended" : active ? "connected" : "paused"}
            data-capture-state={captureState}
            data-playback-state={playbackState}
            style={{
              width: 96,
              height: 96,
              borderRadius: 48,
              background: callEnded
                ? "var(--ei-color-bg-soft)"
                : "var(--ei-color-accent-soft)",
              border: `1px solid ${
                callEnded
                  ? "var(--ei-color-rule-strong)"
                  : "var(--ei-color-accent)"
              }`,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              color: callEnded
                ? "var(--ei-color-fg-tertiary)"
                : "var(--ei-color-accent)",
              fontSize: 30,
            }}
          >
            {callEnded ? "■" : "●"}
          </div>
          <div style={{ textAlign: "center" }}>
            <div
              className="ei-serif"
              style={{
                fontSize: 28,
                color: "var(--ei-color-fg-primary)",
                letterSpacing: 0,
              }}
            >
              {callEnded ? copy.endedTitle : copy.liveTitle}
            </div>
            <div
              style={{
                marginTop: 8,
                fontSize: 13,
                color: "var(--ei-color-fg-tertiary)",
                lineHeight: 1.5,
              }}
            >
              {callEnded ? copy.endedSub : copy.liveSub}
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
              className={!callEnded && active ? "ei-pulse" : ""}
              style={{
                width: 34,
                height: 34,
                borderRadius: 17,
                background: callEnded
                  ? "var(--ei-color-fg-muted)"
                  : "var(--ei-color-accent)",
                color: "#fff",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                flexShrink: 0,
                fontSize: 16,
              }}
            >
              {callEnded ? "■" : "●"}
            </div>
            <PhoneWaveformBars active={!callEnded && active} />
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
              {captionsShown ? copy.hideCaptions : copy.showCaptions}
            </button>
            <button
              data-testid="practice-phone-hangup"
              type="button"
              onClick={() => {
                setCallEnded(true);
                void onHangUp();
              }}
              disabled={callEnded || captureState === "submitting"}
              style={phoneButtonStyle(
                "danger",
                callEnded || captureState === "submitting",
              )}
            >
              {copy.hangUp}
            </button>
            <button
              data-testid="practice-phone-restart"
              type="button"
              onClick={() => {
                autoStartedRef.current = true;
                setCallEnded(false);
                setCaptionsShown(false);
                void onRestartCall();
              }}
              disabled={captureState === "submitting"}
              style={phoneButtonStyle("secondary", captureState === "submitting")}
            >
              {copy.restart}
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
              className="ei-label"
              style={{
                color: "var(--ei-color-fg-tertiary)",
                padding: "12px 22px 0",
              }}
            >
              {copy.captions}
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
  const heights = useMemo(
    () =>
      Array.from({ length: WAVEFORM_BAR_COUNT }, (_, index) => {
        const seed = Math.sin(index * 1.3) * 0.5 + 0.5;
        const accent = index > WAVEFORM_BAR_COUNT - 10;
        return { height: Math.max(6, (seed * 0.55 + 0.28) * 58), accent };
      }),
    [],
  );
  return (
    <div
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
            height: active ? item.height : Math.max(5, item.height * 0.45),
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
  variant: "primary" | "secondary" | "danger",
  disabled = false,
): CSSProperties {
  const color =
    variant === "primary"
      ? "var(--ei-color-accent)"
      : variant === "danger"
        ? "var(--ei-color-danger)"
        : "var(--ei-color-fg-secondary)";
  return {
    background:
      variant === "primary" ? "var(--ei-color-accent-soft)" : "transparent",
    border: `1px solid ${
      variant === "primary"
        ? "var(--ei-color-accent)"
        : variant === "danger"
          ? "var(--ei-color-danger)"
          : "var(--ei-color-rule-strong)"
    }`,
    color: disabled ? "var(--ei-color-fg-muted)" : color,
    padding: "8px 12px",
    borderRadius: 2,
    fontSize: 12.5,
    cursor: disabled ? "default" : "pointer",
  };
}

function phoneCopy(lang: Lang): {
  liveTitle: string;
  liveSub: string;
  endedTitle: string;
  endedSub: string;
  showCaptions: string;
  hideCaptions: string;
  captions: string;
  hangUp: string;
  restart: string;
} {
  if (lang === "en") {
    return {
      liveTitle: "Phone interview in progress",
      liveSub: "Listen and answer naturally. Captions are optional.",
      endedTitle: "Call ended",
      endedSub: "Restart when you are ready to continue this round.",
      showCaptions: "Show captions",
      hideCaptions: "Hide captions",
      captions: "Captions",
      hangUp: "Hang up",
      restart: "Restart",
    };
  }
  return {
    liveTitle: "电话模式进行中",
    liveSub: "像真实电话一样听题并回答；字幕可按需显示。",
    endedTitle: "通话已切断",
    endedSub: "准备好后可重新开始本轮通话。",
    showCaptions: "显示字幕",
    hideCaptions: "隐藏字幕",
    captions: "字幕",
    hangUp: "切断",
    restart: "重新开始",
  };
}
