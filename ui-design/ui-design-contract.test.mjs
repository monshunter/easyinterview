import assert from "node:assert/strict";
import { existsSync } from "node:fs";
import { readdirSync, readFileSync } from "node:fs";
import test from "node:test";

const readUiFile = (path) => readFileSync(new URL(path, import.meta.url), "utf8");
const readUiSources = () => readdirSync(new URL("./src/", import.meta.url))
  .filter((name) => name.endsWith(".jsx"))
  .map((name) => [name, readUiFile(`./src/${name}`)]);

test("workspace mock records are scoped to the active mock plan", () => {
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

test("D-22 keeps debrief and user profile outside current static screens", () => {
  const app = readUiFile("./src/app.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.ok(!existsSync(new URL("./src/screens-p1-depth.jsx", import.meta.url)), "screens-p1-depth.jsx must stay absent");
  assert.ok(!existsSync(new URL("./src/screen-profile.jsx", import.meta.url)), "screen-profile.jsx must stay absent");
  assert.match(app, /debrief:\s*"home"/);
  assert.match(app, /debrief_full:\s*"home"/);
  assert.match(app, /profile:\s*"home"/);
  assert.doesNotMatch(app, /debrief:\s*</);
  assert.doesNotMatch(app, /debrief_full:\s*</);
  assert.doesNotMatch(app, /profile:\s*</);
  assert.doesNotMatch(canvas, /<DCSection id="profile"|<DCArtboard id="profile-[^"]+"|route="profile"/);
  assert.doesNotMatch(canvas, /<DCSection id="p1-depth"|<DCArtboard id="debrief"|<DCSection id="p1-voice-debrief"|route="debrief"/);
});

test("current UI source does not expose out-of-scope inbox wording", () => {
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /Inbox|收件箱/, `${name} contains out-of-scope inbox wording`);
  }
});

test("resume workshop is a flat list without version-tree concepts", () => {
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.match(resume, /const ResumeListView = /);
  assert.match(resume, /const openResume = \(r\) => nav\("resume_versions", \{ resumeId: r\.id \}\);/);
  assert.doesNotMatch(resume, /ResumeTreeView|ResumeBranchFlow|ResumeBranchMap/);
  assert.doesNotMatch(resume, /"MASTER"|"TARGETED"|主版本|岗位定制|选为底稿|分叉/);
  assert.doesNotMatch(resume, /createMode === "guided"|轻量问答|guideSteps|guideAnswers/);
  assert.doesNotMatch(resume, /versionType|parentVersionId|originalId/);
});

test("resume workshop keeps rewrites and second-step save surfaces absent", () => {
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.doesNotMatch(resume, /const acceptBullet = \(id\) => \{/);
  assert.doesNotMatch(resume, /const RewriteSaveConfirmModal = /);
  assert.doesNotMatch(resume, /覆盖原简历|保存为新简历/);
  assert.doesNotMatch(resume, /onConfirm\(mode\)|saveRewriteResult/);
  assert.doesNotMatch(resume, /ResumePreviewConfirm|ResumeParseFlow/);
  assert.doesNotMatch(resume, /"rejected"|已拒绝|拒绝/);
  assert.doesNotMatch(resume, /保存人工改写|Save manual edit/);
});

test("resume workshop detail is read-only original content without source preview/export controls", () => {
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.match(resume, /const ResumeDetailView = /);
  assert.match(resume, /const lines = Array\.isArray\(resume\.text\) && resume\.text\.length > 0/);
  assert.match(resume, /background: "#f6f3ee"/);
  assert.match(resume, /minHeight: 664, margin: "0 auto", background: "#ffffff"/);
  assert.doesNotMatch(resume, /fontSize: 28, fontWeight: 600, letterSpacing: "-0\.02em" \}\}>\{resume\.name\}/);
  assert.doesNotMatch(resume, /fontSize: 14, color: "#666", marginTop: 4 \}\}>\{resume\.summary\}/);
  assert.doesNotMatch(resume, /sourcePreviewOpen|OriginalResumePreviewModal|onPreviewOriginal/);
  assert.doesNotMatch(resume, /onExport|exportPdf|onCopy|copyText/);
  assert.doesNotMatch(resume, /navigator\.clipboard\.writeText/);
  assert.doesNotMatch(resume, /icon="download"/);
});

test("resume workshop create flow keeps upload/paste only and opens detail directly", () => {
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.match(resume, /const \[createdResumes, setCreatedResumes\] = React\.useState\(\[\]\);/);
  assert.match(resume, /setCreatedResumes\(\(prev\) => \[\.\.\.prev, resume\]\);/);
  assert.match(resume, /nav\("resume_versions", \{ resumeId: resume\.id \}\);/);
  assert.match(resume, /onCreateResume\(sourceLabel, createMode === "paste" \? resumeText : ""\)/);
  assert.match(resume, /onCreateResume\(f\.name, ""\)/);
  assert.match(resume, /\{ k: "upload", icon: "upload"/);
  assert.match(resume, /\{ k: "paste", icon: "file"/);
  assert.doesNotMatch(resume, /onConfirm=|resume-preview-confirm|resume-parse-flow/);
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
  assert.match(workspace, /resumeId/);
  assert.match(workspace, /roundId/);
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /resumeVersionId/, `${name} still uses the out-of-scope resumeVersionId context key`);
  }
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

test("report CTA pair lives only in the header (D-19)", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.match(report, /const goReplay = /);
  assert.match(report, /const goNextRound = /);
  // Detail surface no longer receives session-start callbacks.
  assert.doesNotMatch(report, /onReplay|onNextRound/);
  // Question review marks items for the replay plan instead of starting a session.
  assert.match(report, /const \[replayQueued, setReplayQueued\] = React\.useState\(\{\}\);/);
  assert.match(report, /已加入本轮复练/);
  // The next-plan tab points back to the header CTA instead of duplicating it.
  assert.match(report, /开练入口在页面顶部/);
  // Former dead components stay absent.
  assert.doesNotMatch(report, /IssueRow|PerQBlock|KVInline/);
});

test("current UI source does not expose out-of-scope mistakes/growth/drill product surfaces", () => {
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
  assert.match(auth, /nav\("auth_verify", \{ email, pendingAction \}\)/);
  assert.match(auth, /AuthProfileSetupScreen/);
  assert.doesNotMatch(auth, /nav\("auth_register"/);
});

test("job picks module is outside current scope and jd_match aliases back home (D-17)", () => {
  const app = readUiFile("./src/app.jsx");

  assert.ok(!existsSync(new URL("./src/screen-jd-match.jsx", import.meta.url)), "screen-jd-match.jsx must stay absent");
  assert.match(app, /jd_match:\s*"home"/);
  assert.doesNotMatch(app, /jd_match:\s*</);
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /JDMatchScreen|岗位推荐|Job Picks|Job picks|job recommendations/, `${name} still references the out-of-scope job picks module`);
    if (name !== "app.jsx") {
      assert.doesNotMatch(source, /jd_match/, `${name} still references the out-of-scope jd_match route`);
    }
  }
});

test("home uses a resume dropdown and caps recent mocks at three", () => {
  const home = readUiFile("./src/screen-home.jsx");

  assert.match(home, /data-testid="home-jd-input-card"/);
  assert.match(home, /data-testid="home-jd-source-controls"/);
  assert.match(home, /data-testid="home-upload-trigger"/);
  assert.match(home, /data-testid="home-url-trigger"/);
  assert.doesNotMatch(home, /data-testid="home-source-layout"/);
  assert.doesNotMatch(home, /data-testid="home-upload-source-panel"/);
  assert.match(home, /data-testid="home-resume-row"/);
  assert.match(home, /data-testid="home-submit-row"/);
  assert.match(home, /<select[\s\S]*data-testid="home-resume-select"[\s\S]*value=\{selectedResumeId\}/);
  assert.match(home, /style=\{\{ width: 360, maxWidth: "100%"/);
  assert.match(home, /recentJobs\.slice\(0, 3\)/);
  assert.match(home, /hasMoreRecentJobs/);
  assert.match(home, /nav\("workspace", \{\}\)/);
  assert.doesNotMatch(home, /home-resume-option/);
});

test("TargetJob round assumptions use structured interview rounds across parse and recent mocks", () => {
  const home = readUiFile("./src/screen-home.jsx");
  const parse = readUiFile("./src/screens-p0-complete.jsx");
  const data = readUiFile("./src/data.jsx");

  assert.match(data, /const jdSampleInterviewRounds = \[/);
  assert.match(data, /interviewRounds: jdSampleInterviewRounds/);
  assert.match(data, /durationMinutes: 45/);
  assert.match(home, /round\.durationMinutes/);
  assert.match(parse, /roundId: currentRound \? `round-\$\{currentRound\.sequence\}-\$\{currentRound\.type\}` : ""/);
  assert.match(parse, /gridTemplateColumns: `repeat\(\$\{Math\.min\(parsed\.rounds\.length \|\| 1, 4\)\}, 1fr\)`/);
  assert.doesNotMatch(data, /interviewHypotheses/);
  assert.doesNotMatch(parse, /interviewHypotheses/);
  assert.doesNotMatch(data, /focus: "动机、求职节奏、薪资期望"/);
  assert.doesNotMatch(parse, /focus: lang === "en" \? "Motivation, timing, comp"/);
});

test("workspace insight card has no standalone route alias", () => {
  const app = readUiFile("./src/app.jsx");
  const insight = readUiFile("./src/screen-workspace-insight.jsx");
  const workspace = readUiFile("./src/screen-workspace.jsx");
  const outOfScopeInsightTerms = new RegExp([
    "company_" + "intel",
    "Company" + "Intel",
    "getCompany" + "Intel",
  ].join("|"));

  assert.ok(!existsSync(new URL("./src/screen-company-" + "intel.jsx", import.meta.url)), "standalone insight source must stay absent");
  assert.doesNotMatch(app, outOfScopeInsightTerms);
  assert.match(insight, /const WorkspaceInsightCard = /);
  assert.doesNotMatch(insight, outOfScopeInsightTerms);
  assert.doesNotMatch(insight, /打开情报|Open intel/);
  assert.match(workspace, /<WorkspaceInsightCard T=\{T\} lang=\{lang\} job=\{job\} \/>/);
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, outOfScopeInsightTerms, `${name} still references standalone company insight naming`);
  }
});

test("settings keeps only profile and privacy tabs (D-21)", () => {
  const p0 = readUiFile("./src/screens-p0-complete.jsx");

  assert.match(p0, /\{ k: "profile", t: "个人资料" \}, \{ k: "privacy", t: "隐私与数据" \}/);
  assert.doesNotMatch(p0, /SettingsNotif|SettingsBilling|"notifications"|"billing"|订阅/);
});

test("theme defaults to ocean and keeps the custom accent picker (D-21 v2.1)", () => {
  const app = readUiFile("./src/app.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.match(app, /"theme": "ocean"/);
  assert.match(app, /const CUSTOM_ACCENT_SEEDS = /);
  assert.match(app, /const AccentPicker = /);
  assert.match(app, /customAccent/);
  assert.match(app, /\? tweaks\.theme : "ocean"/);
  assert.match(canvas, /Ocean · 深海（默认）/);
  assert.match(canvas, /customAccent/);
});

test("home keeps the debrief auxiliary entry outside current scope (D-22)", () => {
  const home = readUiFile("./src/screen-home.jsx");

  assert.doesNotMatch(home, /POST-INTERVIEW|post-interview|nav\("debrief"\)|Open debrief|打开复盘/);
  assert.doesNotMatch(home, /JOB PICKS|jobPicks/);
});

test("user menu no longer exposes the user profile route (D-22)", () => {
  const app = readUiFile("./src/app.jsx");

  assert.doesNotMatch(app, /label:\s*"用户画像"|label:\s*"Profile"|nav\("profile"\)/);
  assert.match(app, /labelZh:\s*"设置与隐私"/);
  assert.match(app, /nav\("settings"\)/);
  assert.match(app, /nav\("auth_logout"\)/);
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
  assert.doesNotMatch(practice, /VoiceTranscriptionFailure|transcriptFailed/);
});

test("Home recent mock interviews are signed-in only", () => {
  const app = readUiFile("./src/app.jsx");
  const home = readUiFile("./src/screen-home.jsx");

  assert.match(app, /home:\s*<HomeScreen[^>]*signedIn=\{signedIn\}/);
  assert.match(home, /const HomeScreen = \(\{ T, lang, nav, role, signedIn/);
  assert.match(home, /\{signedIn && \(/);
  assert.match(home, /Recent mock interviews/);
  assert.match(home, /选择已有简历/);
  assert.match(home, /还没有简历？1 分钟创建 →/);
  assert.match(home, /立即面试/);
  assert.doesNotMatch(home, /粘贴 JD，或继续最近一次模拟面试。每一次练习都绑定具体岗位，而不是泛用题库。/);
  assert.doesNotMatch(home, /解析并确认面试/);
});

test("Home and workspace share action card behavior", () => {
  const home = readUiFile("./src/screen-home.jsx");
  const workspace = readUiFile("./src/screen-workspace.jsx");

  assert.match(home, /onClick=\{\(\) => nav\("parse", \{ targetJobId: j\.id \}\)\}/);
  assert.match(home, /onStart=\{\(\) => nav\("practice"/);
  assert.match(home, /Start interview now/);
  assert.doesNotMatch(home, /showDelete=\{true\}/);
  assert.match(workspace, /onClick=\{\(\) => openPlan\(job\)\}/);
  assert.match(workspace, /startInterview\(job\)/);
  assert.match(workspace, /Icon name="trash"/);
  assert.match(workspace, /position:\s*"absolute", top:\s*20, right:\s*20/);
  assert.doesNotMatch(workspace, /open:\s*"Open plan"|open:\s*"进入规划"|L\.open/);
});

test("P0 phone interview keeps the shared practice shell without deleted assistant surfaces", () => {
  const practice = readUiFile("./src/screen-practice.jsx");
  const phoneSurface = practice.slice(
    practice.indexOf("const PhoneSessionSurface = "),
    practice.indexOf("const TranscriptMsg = "),
  );

  assert.match(practice, /const PhoneSessionSurface = /);
  assert.match(practice, /const isPhone = activeMode === "phone";/);
  assert.match(practice, /<PhoneSessionSurface/);
  assert.match(practice, /电话模式|Phone/);
  assert.match(practice, /显示字幕|Show captions/);
  assert.match(practice, /切断|Hang up/);
  assert.match(practice, /重新开始|Restart/);
  assert.match(practice, /WaveformBars/);
  assert.match(practice, /gridTemplateColumns:\s*"260px minmax\(0, 1fr\)"/);
  assert.match(practice, /<QuestionHeader/);
  assert.match(practice, /<TranscriptPane/);
  assert.match(practice, /<PhoneSessionSurface/);
  assert.doesNotMatch(phoneSurface, /QuestionHeader|currentQ|qIdx/);
  assert.doesNotMatch(practice, /严格模拟|Strict|Speech-to-text|语音转文字|插入转写|Skip|跳过|表达层指标|口头禅|长停顿|语速|音量/);
  assert.doesNotMatch(practice, /if\s*\(\s*k\s*===\s*"voice"\s*\)\s*nav\("voice"/);
});

test("P0 report renders phone modality copy only for current phone params", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.match(report, /params\.modality === "phone"/);
  assert.doesNotMatch(report, /params\.modality === "voice"/);
  assert.match(report, /"Phone"/);
  assert.match(report, /"电话模式"/);
  assert.doesNotMatch(report, /modality:\s*params\.modality === "voice" \? "Voice" : "Text"/);
  assert.doesNotMatch(report, /modality:\s*params\.modality === "voice" \? "语音" : "文本"/);
});

test("parse confirm page is a readonly saved-plan receipt with direct launch", () => {
  const p0 = readUiFile("./src/screens-p0-complete.jsx");
  const app = readUiFile("./src/app.jsx");

  assert.match(p0, /const ParseScreen = \(\{ T, lang, nav, requestAuth \}\) =>/);
  assert.match(p0, /window\.getWorkspaceResumeOptions/);
  assert.match(p0, /立即面试/);
  assert.match(p0, /type: "create_session"/);
  assert.match(p0, /nav\("practice", startContext\)/);
  assert.doesNotMatch(p0, /ResumePickerModal/);
  assert.doesNotMatch(p0, /仅保存规划|Save plan only/);
  assert.doesNotMatch(p0, /nav\("workspace", buildParseInterviewContext\(\)\)/);
  assert.doesNotMatch(p0, /确认并进入面试前确认|Confirm & open interview setup/);
  assert.match(app, /parse:\s*<ParseScreen[^>]*requestAuth=\{requestAuth\}/);
});

test("D-22 does not leave debrief source or navigation hooks", () => {
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /DebriefFullScreen|DebriefContextPickerModal|GuidedDebriefRecord|VoiceDebriefRecord|DebriefReplayPlan/, `${name} still contains debrief components`);
    if (name !== "app.jsx") {
      assert.doesNotMatch(source, /nav\("debrief"|nav\("debrief_full"/, `${name} still navigates to debrief`);
    }
  }

  const app = readUiFile("./src/app.jsx");
  assert.doesNotMatch(app, /INTERVIEW_CONTEXT_ROUTES = new Set\(\[[^\]]*"debrief"/);
});

test("auth has no standalone reset page and aliases auth_reset back to login", () => {
  const app = readUiFile("./src/app.jsx");
  const auth = readUiFile("./src/screen-auth.jsx");

  assert.match(app, /auth_reset:\s*"auth_login"/);
  assert.doesNotMatch(app, /AuthResetScreen/);
  assert.doesNotMatch(app, /auth_reset:\s*</);
  assert.doesNotMatch(auth, /AuthResetScreen|PASSWORD RESET|找回密码|发送重置说明|auth_reset/);
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /忘记密码|两步验证|Two-step verification/, `${name} still references password-era auth wording`);
    if (name !== "app.jsx") {
      assert.doesNotMatch(source, /auth_reset/, `${name} still references the out-of-scope reset route`);
    }
  }
});

test("debrief thank-you letter draft stays absent", () => {
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /ThankYouLetter|感谢信/, `${name} still contains the out-of-scope thank-you letter draft`);
  }
});

test("out-of-scope resume versions screen stays absent", () => {
  assert.ok(!existsSync(new URL("./src/screens-p1-depth.jsx", import.meta.url)), "screens-p1-depth.jsx must stay absent");
  for (const [name, source] of readUiSources()) {
    const outOfScopeResumeSourcePattern = new RegExp(`_${"Leg" + "acy"}ResumeVersionsScreen|ResumeSourceMap`);
    assert.doesNotMatch(source, outOfScopeResumeSourcePattern, `${name} still contains out-of-scope resume version source`);
  }
});

test("phone interview only enters through explicit practice modality params", () => {
  const app = readUiFile("./src/app.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.doesNotMatch(app, /voice:\s*"practice"/);
  assert.doesNotMatch(app, /rawRoute === "voice"/);
  assert.doesNotMatch(app, /voice:\s*<VoicePracticeScreen/);
  assert.doesNotMatch(app, /route\.name === "voice"/);
  assert.match(canvas, /route="practice" mode="phone"/);
});
