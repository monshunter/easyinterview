import { useMemo, useState, type FC } from "react";

import { useI18n } from "../../i18n/messages";
import { useInterviewContext } from "../../interview-context/InterviewContext";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { TopBar } from "./components/TopBar";
import { SessionMap } from "./components/SessionMap";
import { LiveNotes } from "./components/LiveNotes";
import { QuestionCard } from "./components/QuestionCard";
import { Transcript } from "./components/Transcript";
import { InputBar } from "./components/InputBar";
import { HintBanner } from "./components/HintBanner";
import { RightPanel } from "./components/RightPanel";
import { FinishCta } from "./components/FinishCta";
import { VoiceSurfaceComingSoon } from "./components/VoiceSurfaceComingSoon";
import { PracticeSessionLostState } from "./components/PracticeSessionLostState";
import { ErrorState } from "./components/ErrorState";

interface PracticeScreenProps {
  route: Route;
}

/**
 * Item 1.1 — PracticeScreen static shell.
 *
 * Source-level mirror of `ui-design/src/screen-practice.jsx::PracticeScreen`
 * text branch (lines 184-326). Phase 1 only renders the static skeleton;
 * data hooks land in 1.3+ and event mutations land in Phase 2.
 *
 * sessionId guard: when the InterviewContext / route lacks a sessionId,
 * render PracticeSessionLostState instead. mode='voice' renders the scoped
 * VoiceSurfaceComingSoon placeholder instead of the text surface.
 */
export const PracticeScreen: FC<PracticeScreenProps> = ({ route }) => {
  const { t, lang } = useI18n();
  const { navigate } = useNavigation();
  const { ctx } = useInterviewContext();

  const sessionId = route.params.sessionId || ctx.sessionId || "";
  const mode = route.params.mode || ctx.mode || "text";
  const modality = route.params.modality || ctx.modality || mode;
  const practiceMode =
    route.params.practiceMode || ctx.practiceMode || "strict";
  const isStrict = practiceMode === "strict";
  const activeMode = modality === "voice" ? "voice" : "text";

  // Phase 1: skeleton state. Phase 2 replaces with real session data.
  const [paused, setPaused] = useState(false);
  const [input, setInput] = useState("");

  const handleBackToWorkspace = () => {
    navigate({
      name: "workspace",
      params: {
        targetJobId: route.params.targetJobId || ctx.targetJobId,
        jdId: route.params.jdId || ctx.jdId || "",
        planId: route.params.planId || ctx.planId || "",
        resumeVersionId:
          route.params.resumeVersionId || ctx.resumeVersionId || "",
      },
    });
  };

  if (!sessionId) {
    return <PracticeSessionLostState onBack={handleBackToWorkspace} />;
  }

  const handleSwitchMode = (k: "text" | "voice") => {
    navigate({
      name: "practice",
      params: {
        ...route.params,
        sessionId,
        mode: k,
        modality: k,
      },
    });
  };

  const skeletonQuestions = useMemo(
    () =>
      Array.from({ length: 5 }, (_, idx) => ({
        id: `q-skeleton-${idx}`,
        topic: t("practice.sessionMap.itemTopicSkeleton"),
        duration: "—",
        status: idx === 0 ? ("active" as const) : ("pending" as const),
      })),
    [t],
  );

  return (
    <div
      data-testid="practice-screen"
      data-session-id={sessionId}
      data-plan-id={route.params.planId || ctx.planId || ""}
      data-target-job-id={route.params.targetJobId || ctx.targetJobId || ""}
      data-jd-id={route.params.jdId || ctx.jdId || ""}
      data-resume-version-id={
        route.params.resumeVersionId || ctx.resumeVersionId || ""
      }
      data-round-id={route.params.roundId || ctx.roundId || ""}
      data-mode={mode}
      data-modality={modality}
      data-practice-mode={practiceMode}
      data-practice-goal={
        route.params.practiceGoal || ctx.practiceGoal || "baseline"
      }
      className="ei-fadein"
      style={{
        height: "100vh",
        display: "flex",
        flexDirection: "column",
        background: "var(--ei-color-bg)",
      }}
    >
      <TopBar
        company={t("practice.toolbar.companySkeleton")}
        title={t("practice.toolbar.titleSkeleton")}
        questionIndex={1}
        questionTotal={5}
        elapsed="00:00"
        budget="25:00"
        paused={paused}
        onTogglePause={() => setPaused((p) => !p)}
        activeMode={activeMode}
        onSwitchMode={handleSwitchMode}
        strict={isStrict}
        onToggleStrict={() => undefined}
      />

      <div
        data-testid="practice-main"
        style={{
          flex: 1,
          display: "grid",
          gridTemplateColumns: "260px 1fr 280px",
          minHeight: 0,
        }}
      >
        <div
          data-testid="practice-sessionmap"
          style={{
            borderRight: "1px solid var(--ei-color-rule)",
            padding: "20px 18px",
            overflowY: "auto",
            background: "var(--ei-color-bgSoft)",
          }}
        >
          <SessionMap
            label={t("practice.sessionMap.label")}
            items={skeletonQuestions}
            activeIndex={0}
          />
          {!isStrict && (
            <LiveNotes
              label={t("practice.sessionMap.liveNotes")}
              okText={t("practice.sessionMap.liveNotesOk")}
              warnText={t("practice.sessionMap.liveNotesWarn")}
              note={t("practice.sessionMap.liveNotesNote")}
            />
          )}
        </div>

        <div
          data-testid="practice-center"
          style={{ display: "flex", flexDirection: "column", minHeight: 0 }}
        >
          {activeMode === "voice" ? (
            <VoiceSurfaceComingSoon
              title={t("practice.voiceComingSoon.title")}
              desc={t("practice.voiceComingSoon.desc")}
              backLabel={t("practice.voiceComingSoon.backToText")}
              onBackToText={() => handleSwitchMode("text")}
            />
          ) : (
            <>
              <QuestionCard
                badgeText={t("practice.question.tagPrefix").replace(
                  "{n}",
                  "1",
                )}
                topic={t("practice.sessionMap.itemTopicSkeleton")}
                tags={[]}
                prompt={t("practice.question.skeletonPrompt")}
              />
              <Transcript
                messages={[]}
                helperText={t("practice.transcript.helper")}
                aiLabel={t("practice.transcript.aiLabel")}
                userLabel={t("practice.transcript.userLabel")}
                followUpLabel={t("practice.transcript.followUp")}
              />
              <InputBar
                value={input}
                onChange={setInput}
                placeholder={t("practice.input.placeholder")}
                hintLabel={t("practice.input.hint")}
                skipLabel={t("practice.input.skip")}
                sendLabel={t("practice.input.send")}
                dictateLabel={t("practice.input.dictateOn")}
                showHintButton={!isStrict}
                disabled={paused}
                onHint={() => undefined}
                onSkip={() => undefined}
                onSend={() => undefined}
                onDictate={() => undefined}
                hintBanner={null}
              />
            </>
          )}
        </div>

        <RightPanel
          jdLinkLabel={t("practice.rightpanel.jdLink")}
          jdProbesLabel={t("practice.rightpanel.jdProbes")}
          jdProbesText={t("practice.rightpanel.jdProbesSkeleton")}
          experienceLabel={t("practice.rightpanel.experienceLabel")}
          aiTransparencyLabel={t("practice.rightpanel.aiTransparency")}
          aiTransparencyMeta={{
            promptVersion: "v1.0.4",
            rubricVersion: "v0.9",
            modelId: "haiku-4.5",
            language: lang,
          }}
          strict={isStrict}
          strictBannerText={t("practice.rightpanel.strictBanner")}
          experiences={[]}
          finishCta={
            <FinishCta
              label={t("practice.rightpanel.finishCta")}
              hintCount={0}
              hintUsageNote={t("practice.rightpanel.hintUsageNote")}
              onFinish={() => undefined}
            />
          }
        />
      </div>
      <ErrorState message={null} />
      <HintBanner show={false} prefix={t("practice.hint.prefix")} text="" />
    </div>
  );
};
