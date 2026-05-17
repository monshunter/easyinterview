import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { DebriefEntry } from "../types";

interface VoiceDebriefRecordProps {
  entries: DebriefEntry[];
}

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx::VoiceDebriefRecord
 * (lines 656-870). UI shell only — spec D-6 P0 explicitly defers real STT
 * integration; we render the toggle highlight, idle placeholder, keyboard
 * hint copy, and a list of voice-extracted entries that stays empty until
 * the Phase 4+ STT integration lands.
 */
export const VoiceDebriefRecord: FC<VoiceDebriefRecordProps> = ({ entries }) => {
  const { t } = useI18n();
  const voiceEntries = entries.filter((e) => e.source === "voice_extracted");
  return (
    <section
      className="ei-debrief-voice"
      data-testid="debrief-voice-record"
    >
      <div
        className="ei-debrief-voice__placeholder"
        data-testid="debrief-voice-not-implemented"
      >
        {t("debrief.record.voice.notImplemented")}
      </div>
      <div className="ei-debrief-voice__idle">
        {t("debrief.record.voice.idle")}
      </div>
      <div className="ei-debrief-voice__shared">
        {t("debrief.record.voice.shared")} · {entries.length}
      </div>
      {voiceEntries.length > 0 && (
        <ul data-testid="debrief-voice-pending-list">
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
