import assert from "node:assert/strict";
import { readdirSync, readFileSync } from "node:fs";
import test from "node:test";

const readUiFile = (path) => readFileSync(new URL(path, import.meta.url), "utf8");
const readUiSources = () => readdirSync(new URL("./src/", import.meta.url))
  .filter((name) => name.endsWith(".jsx"))
  .map((name) => [name, readUiFile(`./src/${name}`)]);

test("workspace mock history is scoped to the active mock plan", () => {
  const workspace = readUiFile("./src/screen-workspace.jsx");
  assert.match(
    workspace,
    /const sessionHistory = getWorkspaceSessionHistory\(lang, job, currentRound\?\.name, interviewContext\);/,
  );

  const historyStart = workspace.indexOf("const getWorkspaceSessionHistory");
  const historyEnd = workspace.indexOf("const getWorkspacePlanOptions");
  const historySource = workspace.slice(historyStart, historyEnd);

  assert.notEqual(historyStart, -1);
  assert.notEqual(historyEnd, -1);
  assert.match(historySource, /getWorkspaceTargetLabel/);
  assert.doesNotMatch(historySource, /Lumen Labs · Frontend Platform Engineer/);
});

test("debrief context cards open local pickers instead of cross-page navigation", () => {
  const debrief = readUiFile("./src/screens-p1-depth.jsx");
  assert.doesNotMatch(debrief, /Inbox|收件箱/);
  assert.match(debrief, /DebriefContextPickerModal/);

  const stripStart = debrief.indexOf("const DebriefContextStrip");
  const stripEnd = debrief.indexOf("const GuidedDebriefRecord");
  const stripSource = debrief.slice(stripStart, stripEnd);

  assert.notEqual(stripStart, -1);
  assert.notEqual(stripEnd, -1);
  assert.match(stripSource, /onOpenPicker/);
  assert.doesNotMatch(stripSource, /nav\("workspace"|nav\("report"|nav\("resume_versions"/);
});

test("current UI source does not expose removed inbox wording", () => {
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /Inbox|收件箱/, `${name} contains removed inbox wording`);
  }
});

test("resume workshop opens targeted versions on rewrite decisions", () => {
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.match(resume, /const resumeDefaultTab = \(version\) => version && version\.tag === "TARGETED" \? "rewrites" : "preview";/);
  assert.match(resume, /const openVersion = \(v, tab = resumeDefaultTab\(v\)\) => nav\("resume_versions", \{ versionId: v\.id, tab \}\);/);
  assert.match(resume, /onClick=\{\(\) => onOpen\(v\)\}/);
  assert.doesNotMatch(resume, /onOpen\(v, "preview"\)/);
});

test("resume workshop source preview and export controls are wired", () => {
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.match(resume, /const \[sourcePreviewOpen, setSourcePreviewOpen\] = React\.useState\(false\);/);
  assert.match(resume, /onPreviewOriginal=\{\(\) => setSourcePreviewOpen\(true\)\}/);
  assert.match(resume, /const OriginalResumePreviewModal = /);
  assert.match(resume, /onClick=\{onPreviewOriginal\}/);
  assert.match(resume, /onExport=\{exportPdf\}/);
  assert.match(resume, /onCopy=\{copyText\}/);
  assert.match(resume, /navigator\.clipboard\.writeText\(text\)/);
  assert.match(resume, /icon="download" onClick=\{exportPdf\}/);
});

test("resume workshop create actions mutate local prototype data", () => {
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.match(resume, /const \[createdOriginals, setCreatedOriginals\] = React\.useState\(\[\]\);/);
  assert.match(resume, /const \[createdVersions, setCreatedVersions\] = React\.useState\(\[\]\);/);
  assert.match(resume, /setCreatedOriginals\(\(prev\) => \[\.\.\.prev, original\]\);/);
  assert.match(resume, /setCreatedVersions\(\(prev\) => \[\.\.\.prev, created\]\);/);
  assert.match(resume, /onCreateVersion=\{\(draft\) => addTargetedVersion\(sourceOriginal, sourceMaster, draft\)\}/);
  assert.match(resume, /onConfirm=\{\(label\) => onCreateOriginal \? onCreateOriginal\(label\) : onBack\(\)\}/);
  assert.match(resume, /onClick=\{createVersion\}/);
});

test("P0 context routes use InterviewContext instead of fixed tj-1 nav payloads", () => {
  const app = readUiFile("./src/app.jsx");
  const workspace = readUiFile("./src/screen-workspace.jsx");

  assert.match(app, /const DEFAULT_INTERVIEW_CONTEXT = /);
  assert.match(app, /const createInterviewContext = /);
  assert.match(app, /const shouldCarryInterviewContext = /);
  assert.doesNotMatch(app, /params:\s*\{\s*jobId:\s*"tj-1"\s*\}/);

  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(
      source,
      /nav\("[a-z_]+",\s*\{[^}]*jobId:\s*"tj-1"/,
      `${name} still hard-codes tj-1 in a nav payload`,
    );
  }

  assert.match(workspace, /const interviewContext = createWorkspaceInterviewContext\(/);
  assert.match(workspace, /planId/);
  assert.match(workspace, /targetJobId/);
  assert.match(workspace, /jdId/);
  assert.match(workspace, /resumeVersionId/);
  assert.match(workspace, /roundId/);
});

test("P0 report requires sessionId and separates replay from next-round payloads", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.match(report, /const ReportMissingSessionState = /);
  assert.match(report, /const ReportFailureState = /);
  assert.match(report, /if \(!params\?\.sessionId\)/);
  assert.match(report, /reportStatus === "failed"/);
  assert.match(report, /const replayContext = /);
  assert.match(report, /replayItems/);
  assert.match(report, /const nextRoundContext = /);
  assert.match(report, /nextRoundId/);
  assert.match(report, /sourceSessionId: params\.sessionId/);
  assert.doesNotMatch(report, /nav\("workspace",\s*\{\s*jobId:\s*"tj-1"\s*\}\)/);
});

test("P0 report replay and next-round CTAs start interview sessions directly", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.match(report, /const run = \(\) => nav\("practice", payload\);/);
  assert.match(report, /route: "practice"/);
  assert.match(report, /action: "replay-current-round"/);
  assert.match(report, /action: "start-next-round"/);
  assert.match(report, /`session-\$\{params\.planId\}-\$\{params\.roundId\}-replay`/);
  assert.match(report, /`session-\$\{params\.planId\}-\$\{nextRoundId\}-start`/);
  assert.doesNotMatch(report, /const run = \(\) => nav\("workspace", payload\);/);
  assert.doesNotMatch(report, /route: "workspace"/);
  assert.doesNotMatch(report, /action: "prepare-next-round"/);
  assert.doesNotMatch(report, /`session-\$\{params\.planId\}-\$\{nextRoundId\}-prep`/);
});

test("current UI source does not expose removed mistakes/growth/drill product surfaces", () => {
  const report = readUiFile("./src/screen-report.jsx");
  const settings = readUiFile("./src/screens-p0-complete.jsx");
  const data = readUiFile("./src/data.jsx");

  assert.doesNotMatch(report, /错题|openDrill|addToMistakes|addedToMistakes/);
  assert.doesNotMatch(settings, /Mistakes derived|派生的错题/);
  assert.doesNotMatch(data, /\bmistakes\s*:|mistakesTotal|mistakesResolved|\bgrowth\s*:/);
});

test("P0 auth success resumes the pending action instead of always returning home", () => {
  const app = readUiFile("./src/app.jsx");
  const auth = readUiFile("./src/screen-auth.jsx");

  assert.match(app, /const requestAuth = /);
  assert.match(app, /pendingAction\?\.route/);
  assert.match(app, /pendingAction\.params/);
  assert.doesNotMatch(app, /setRoute\(\{\s*name:\s*"home",\s*params:\s*\{\}\s*\}\);/);

  assert.match(auth, /pendingAction/);
  assert.match(auth, /nav\("auth_register", \{ pendingAction \}\)/);
  assert.match(auth, /nav\("auth_verify", \{ email, pendingAction \}\)/);
});

test("P0 company intel copy uses compliant public-signal wording", () => {
  const jdMatch = readUiFile("./src/screen-jd-match.jsx");
  const companyIntel = readUiFile("./src/screen-company-intel.jsx");

  assert.doesNotMatch(jdMatch, /抓到\s*\$\{job\.similarInterviewers\}\s*位面试官公开信息/);
  assert.doesNotMatch(jdMatch, /interviewer profiles surfaced from public sources/);
  assert.match(jdMatch, /公开面经\s*\/\s*JD\s*\/\s*公司资料信号/);
  assert.match(jdMatch, /public interview-review, JD, and company-source signals/);
  assert.doesNotMatch(companyIntel, /nav\("workspace",\s*\{\s*jobId:\s*"tj-1"\s*\}\)/);
});

test("P0 empty and failure states avoid showing fake data", () => {
  const home = readUiFile("./src/screen-home.jsx");
  const workspace = readUiFile("./src/screen-workspace.jsx");
  const report = readUiFile("./src/screen-report.jsx");
  const practice = readUiFile("./src/screen-practice.jsx");

  assert.match(home, /const recentJobs = /);
  assert.match(home, /HomeEmptyState/);
  assert.match(workspace, /WorkspaceEmptyState/);
  assert.match(workspace, /WorkspaceMissingResumeState/);
  assert.match(report, /ReportMissingSessionState/);
  assert.match(report, /ReportFailureState/);
  assert.match(practice, /transcriptFailed/);
  assert.match(practice, /VoiceTranscriptionFailure/);
});

test("P0 voice interview keeps the shared practice shell and renders the voice surface", () => {
  const practice = readUiFile("./src/screen-practice.jsx");

  assert.match(practice, /const VoiceSessionSurface = /);
  assert.match(practice, /activeMode === "voice"\s*\?/);
  assert.match(practice, /<VoiceSessionSurface/);
  assert.match(practice, /WaveformBars/);
  assert.match(practice, /AnnotatedWaveform/);
  assert.match(practice, /表达层指标/);
  assert.match(practice, /实时转写/);
  assert.match(practice, /音频仅在本次会话缓存/);
  assert.match(practice, /VoiceTranscriptionFailure/);
  assert.doesNotMatch(practice, /if\s*\(\s*k\s*===\s*"voice"\s*\)\s*nav\("voice"/);
});

test("voice interview only enters through explicit practice modality params", () => {
  const app = readUiFile("./src/app.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.doesNotMatch(app, /voice:\s*"practice"/);
  assert.doesNotMatch(app, /rawRoute === "voice"/);
  assert.doesNotMatch(app, /voice:\s*<VoicePracticeScreen/);
  assert.doesNotMatch(app, /route\.name === "voice"/);
  assert.match(canvas, /route="practice" mode="voice"/);
});
