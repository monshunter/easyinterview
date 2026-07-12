import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import test from "node:test";

const readUiFile = (path) => readFileSync(new URL(path, import.meta.url), "utf8");
const readUiSources = () => readdirSync(new URL("./src/", import.meta.url))
  .filter((name) => name.endsWith(".jsx"))
  .map((name) => [name, readUiFile(`./src/${name}`)]);
const readUiJsxTree = (directory = new URL("./", import.meta.url), prefix = "") =>
  readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const relative = `${prefix}${entry.name}`;
    const entryUrl = new URL(`./${relative}${entry.isDirectory() ? "/" : ""}`, import.meta.url);
    if (entry.isDirectory()) {
      return readUiJsxTree(entryUrl, `${relative}/`);
    }
    return entry.name.endsWith(".jsx") ? [[relative, readFileSync(entryUrl, "utf8")]] : [];
  });

test("prototype JSX sources do not duplicate whole files", () => {
  const sourceByHash = new Map();
  for (const [name, source] of readUiJsxTree()) {
    const hash = createHash("sha256").update(source).digest("hex");
    assert.equal(sourceByHash.get(hash), undefined, `${name} duplicates ${sourceByHash.get(hash)}`);
    sourceByHash.set(hash, name);
  }
});

test("prototype runner uses the repository Python 3 toolchain only", () => {
  const runner = readUiFile("./run.sh");
  const toolchainCheck = runner.indexOf("if ! command -v python3");
  const urlEncoding = runner.indexOf("url_encode() {");

  assert.ok(toolchainCheck >= 0 && toolchainCheck < urlEncoding);
  assert.match(runner, /SCRIPT_PATH="\$SCRIPT_DIR\/\$\(basename "\$0"\)"/);
  assert.match(runner, /sed -n '2,7p' "\$SCRIPT_PATH"/);
  assert.match(runner, /exec python3 -m http\.server/);
  assert.doesNotMatch(runner, /SimpleHTTPServer|npx --yes serve|elif command -v python/);
});

test("prototype primitive globals expose only current consumers", () => {
  const primitives = readUiFile("./src/primitives.jsx");

  assert.doesNotMatch(primitives, /const (?:Sparkline|KV) =/);
  assert.doesNotMatch(primitives, /\b(?:Sparkline|KV)\b/);
  assert.match(
    primitives,
    /Object\.assign\(window, \{ Icon, Tag, Btn, Card, SectionHeader, ReadinessDial \}\);/,
  );
});

test("workspace static prototype is a pure plan list with no unreachable detail branch", () => {
  const workspace = readUiFile("./src/screen-workspace.jsx");

  assert.match(workspace, /const WorkspaceScreen = \(\{ T, lang, nav \}\) => \{/);
  assert.match(workspace, /return <WorkspacePlanList T=\{T\} lang=\{lang\} nav=\{nav\} jobs=\{jobs\} \/>;/);
  assert.match(workspace, /Object\.assign\(window, \{ getWorkspaceResumeOptions \}\);/);
  assert.doesNotMatch(workspace, /^\s+updated:/m);

  for (const symbol of [
    "hasPlanContext",
    "WorkspaceEmptyState",
    "WorkspaceMissingResumeState",
    "createWorkspaceInterviewContext",
    "getWorkspaceRoundId",
    "getWorkspaceSessionHistory",
    "getWorkspaceTargetLabel",
    "getWorkspaceRoundLabel",
    "getWorkspacePlanOptions",
    "getWorkspaceJDSample",
    "ResumePickerModal",
    "PlanSwitcherModal",
    "WorkspaceInsightCard",
    "InterviewRoundRail",
    "BindingPill",
    "ReqBlock",
  ]) {
    assert.doesNotMatch(workspace, new RegExp(`\\b${symbol}\\b`), `${symbol} must stay absent from the pure list prototype`);
  }
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
  const app = readUiFile("./src/app.jsx");
  const resume = readUiFile("./src/screen-resume-workshop.jsx");

  assert.match(resume, /const \[createdResumes, setCreatedResumes\] = React\.useState\(\[\]\);/);
  assert.match(resume, /setCreatedResumes\(\(prev\) => \[\.\.\.prev, resume\]\);/);
  assert.match(resume, /setCreatedResumes\(\(prev\) => \[\.\.\.prev, resume\]\);\s+setFlow\("list"\);\s+nav\("resume_versions", \{ resumeId: resume\.id \}\);/);
  assert.match(resume, /nav\("resume_versions", \{ resumeId: resume\.id \}\);/);
  assert.match(resume, /setFlow\(params\.flow === "create" \? "create" : "list"\);/);
  assert.match(app, /<div key=\{route\.name \+ \(route\.params\.jobId \|\| ""\)\}>/);
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

  assert.match(workspace, /const openPlan = \(job\) => nav\("parse", \{/);
  assert.match(workspace, /targetJobId: job\.id/);
  for (const [name, source] of readUiSources()) {
    assert.doesNotMatch(source, /resumeVersionId/, `${name} still uses the out-of-scope resumeVersionId context key`);
  }
});

test("P0 report requires sessionId and uses conversation-level evidence", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.match(report, /const ReportMissingSessionState = /);
  assert.match(report, /const ReportFailureState = /);
  assert.match(report, /if \(!params\.sessionId\)/);
  assert.match(report, /reportStatus === "failed"/);
  assert.match(report, /const dimensions = report\.dimensions \|\| \[\]/);
  assert.match(report, /const highlights = report\.highlights \|\| \[\]/);
  assert.doesNotMatch(report, /perQuestion|replayItems|activeQuestion|qId/);
  assert.doesNotMatch(report, /nav\("workspace",\s*\{\s*jobId:\s*"tj-1"\s*\}\)/);
});

test("P0 report replay and next-round CTAs start interview sessions directly", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.match(report, /nav\("practice", \{ \.\.\.params, practiceGoal: "retry_current_round" \}\)/);
  assert.match(report, /const \{ nextRound \} = window\.eiResolveInterviewRoundContext/);
  assert.match(report, /practiceGoal: "next_round", roundId: nextRound\.id, roundName: nextRound\.name/);
  assert.match(report, /disabled=\{!nextRound\}/);
  assert.equal((report.match(/nav\("practice"/g) || []).length, 2);
});

test("report CTA pair lives only in the header (D-19)", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.equal((report.match(/practiceGoal: "retry_current_round"/g) || []).length, 1);
  assert.equal((report.match(/practiceGoal: "next_round"/g) || []).length, 1);
  assert.doesNotMatch(report, /ReportDetailSurface|QuestionsTab|题目回顾/);
});

test("current UI source does not expose out-of-scope mistakes/growth/drill product surfaces", () => {
  const report = readUiFile("./src/screen-report.jsx");
  const settings = readUiFile("./src/screens-p0-complete.jsx");
  const data = readUiFile("./src/data.jsx");

  assert.doesNotMatch(report, /错题|openDrill|addToMistakes|addedToMistakes/);
  assert.doesNotMatch(settings, /Mistakes derived|派生的错题/);
  assert.doesNotMatch(data, /\bmistakes\s*:|mistakesTotal|mistakesResolved|\bgrowth\s*:/);
});

test("prototype experience copy describes direct current actions", () => {
  const data = readUiFile("./src/data.jsx");

  assert.doesNotMatch(data, /兼容层/);
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

test("workspace insight source stays absent from the pure plan-list prototype", () => {
  const app = readUiFile("./src/app.jsx");
  const index = readUiFile("./index.html");
  const workspace = readUiFile("./src/screen-workspace.jsx");
  const outOfScopeInsightTerms = new RegExp([
    "company_" + "intel",
    "Company" + "Intel",
    "getCompany" + "Intel",
  ].join("|"));

  assert.ok(!existsSync(new URL("./src/screen-company-" + "intel.jsx", import.meta.url)), "standalone insight source must stay absent");
  assert.ok(!existsSync(new URL("./src/screen-workspace-insight.jsx", import.meta.url)), "workspace insight source must stay absent");
  assert.doesNotMatch(app, outOfScopeInsightTerms);
  assert.doesNotMatch(index, /screen-workspace-insight\.jsx/);
  assert.doesNotMatch(workspace, /WorkspaceInsightCard/);
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
  assert.match(workspace, /visibleJobs\.length === 0/);
  assert.match(workspace, /data-testid="workspace-plan-list-empty"/);
  assert.doesNotMatch(workspace, /WorkspaceEmptyState|WorkspaceMissingResumeState/);
  assert.match(report, /ReportMissingSessionState/);
  assert.match(report, /ReportFailureState/);
  assert.doesNotMatch(practice, /VoiceTranscriptionFailure|transcriptFailed/);
});

test("Home recent mock interviews are signed-in only", () => {
  const app = readUiFile("./src/app.jsx");
  const home = readUiFile("./src/screen-home.jsx");

  assert.match(app, /home:\s*<HomeScreen[^>]*signedIn=\{signedIn\}/);
  assert.match(home, /const HomeScreen = \(\{ T, lang, nav, signedIn/);
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

test("practice is one continuous text conversation with phone disabled", () => {
  const practice = readUiFile("./src/screen-practice.jsx");
  const primitives = readUiFile("./src/primitives.jsx");
  assert.match(practice, /data-testid="practice-topbar-phone-toggle"/);
  assert.equal((practice.match(/data-testid="practice-topbar-phone-toggle"/g) || []).length, 1);
  assert.match(practice, /data-testid="practice-topbar-phone-toggle"[\s\S]*disabled[\s\S]*aria-disabled="true"/);
  assert.match(practice, /电话模式暂未开放|Phone mode is temporarily unavailable/);
  assert.match(primitives, /phone:\s*<>/);
  assert.match(practice, /data-testid="practice-conversation"/);
  assert.match(practice, /width: "100%"/);
  assert.doesNotMatch(practice, /SESSION MAP|本轮题目|QuestionHeader|QuestionCard|qIdx|currentQ|questions\.map/);
  assert.doesNotMatch(practice, /Question[^a-z]|题\s*\{/);
  assert.doesNotMatch(practice, /PhoneSessionSurface|WaveformBars|practice-phone-surface|practice-phone-captions/);
  assert.match(practice, /const \{ currentRound \} = window\.eiResolveInterviewRoundContext/);
  assert.match(practice, /currentRound \? formatElapsed\(currentRound\.durationMinutes \* 60\) : "--:--"/);
  assert.doesNotMatch(practice, /25:00/);
});

test("P0 report has no voice or modality-specific report branch", () => {
  const report = readUiFile("./src/screen-report.jsx");

  assert.doesNotMatch(report, /params\.modality|practiceMode|hintUsed|hintCount/);
  assert.doesNotMatch(report, /Phone|电话模式|Voice|语音/);
});

test("parse confirm page is a readonly saved-plan receipt with direct launch", () => {
  const p0 = readUiFile("./src/screens-p0-complete.jsx");
  const app = readUiFile("./src/app.jsx");

  assert.match(p0, /const PlanBindingPill = /);
  assert.match(p0, /<PlanBindingPill T=\{T\}/);
  assert.doesNotMatch(p0, /window\.BindingPill/);
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

test("phone interview has no positive canvas or route entry while disabled", () => {
  const app = readUiFile("./src/app.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.doesNotMatch(app, /voice:\s*"practice"/);
  assert.doesNotMatch(app, /rawRoute === "voice"/);
  assert.doesNotMatch(app, /voice:\s*<VoicePracticeScreen/);
  assert.doesNotMatch(app, /route\.name === "voice"/);
  assert.doesNotMatch(canvas, /mode="phone"|practice-phone|phone-light|phone-dark/);
  assert.match(canvas, /电话入口暂不可用/);
});

test("design canvas component surface matches its only tracked consumer", () => {
  const wrapper = readUiFile("./design-canvas.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.match(wrapper, /function DesignCanvas\(\{ children \}\)/);
  assert.match(wrapper, /function DCViewport\(\{ children \}\)/);
  assert.match(wrapper, /function DCSection\(\{ id, title, subtitle, children \}\)/);
  assert.match(wrapper, /function DCPostIt\(\{ children, top, left, right, rotate = -2, width = 180 \}\)/);
  assert.doesNotMatch(wrapper, /\bminScale\b|\bmaxScale\b/);
  assert.doesNotMatch(wrapper, /const \{[^}]*\bstyle = \{\}[^}]*\} = artboard\.props/);

  assert.match(canvas, /const Screen = \(\{ route, mode, dark, theme, customAccent, fontPreset \}\) =>/);
  assert.doesNotMatch(canvas, /const PracticeScreen\b/);
  assert.doesNotMatch(canvas, /params\.set\("(?:nochrome|jobId|targetJobId|planId|jdId|resumeId|roundId|roundName)"/);
  assert.match(canvas, /params\.set\("lang", "zh"\)/);
  assert.match(canvas, /\["practice", "generating", "report"\]\.includes\(route\).*params\.set\("sessionId", "session-24"\)/);
});

test("design canvas keeps edits in memory without an unavailable sidecar bridge", () => {
  const wrapper = readUiFile("./design-canvas.jsx");

  assert.doesNotMatch(wrapper, /\.design-canvas\.state\.json|DC_STATE_FILE|window\.omelette/);
  assert.doesNotMatch(wrapper, /\b(?:didRead|skipNextWrite|setReady)\b/);
  assert.doesNotMatch(wrapper, /fetch\(['"]\.\//);
  assert.match(wrapper, /const \[state, setState\] = React\.useState\(\{ sections: \{\}, focus: null \}\)/);
  assert.match(wrapper, /patchSection: \(id, p\) => setState/);
  assert.match(wrapper, /<DCViewport>\{children\}<\/DCViewport>/);
});

test("prototype display controls do not depend on an unavailable edit-mode host", () => {
  const app = readUiFile("./src/app.jsx");
  const home = readUiFile("./src/screen-home.jsx");
  const canvas = readUiFile("./canvas.html");

  assert.doesNotMatch(app, /EDITMODE|__activate_edit_mode|__deactivate_edit_mode|__edit_mode_|window\.parent\.postMessage/);
  assert.doesNotMatch(app, /TweaksPanel|TweakRow|selectStyle|tweaksOpen|tweaksAvailable/);
  assert.doesNotMatch(app, /["']role["']\s*:|\["dark","role"|role=\{tweaks\.role\}|setRole=/);
  assert.doesNotMatch(home, /const HomeScreen = \(\{ T, lang, nav, role,/);
  assert.doesNotMatch(canvas, /\bTweaks\b|edit mode bridge/);

  assert.match(app, /<TopBar[^>]*setDark=\{\(v\) => updateTweak\("dark", v\)\}/);
  assert.match(app, /setCustomAccent=\{\(v\) => updateTweak\("customAccent", v\)\}/);
  assert.match(app, /settings:\s*<SettingsScreen[^>]*setFontPreset=\{setFontPreset\}/);
});

test("prototype app shell has no zero-read canvas mode binding", () => {
  const app = readUiFile("./src/app.jsx");

  assert.doesNotMatch(app, /\bconst isCanvasIframe\b/);
  assert.match(app, /const hideTopBar = [^;]*data-nochrome[^;]*;/);
});

test("auth prototype screens expose only consumed navigation and completion callbacks", () => {
  const app = readUiFile("./src/app.jsx");
  const auth = readUiFile("./src/screen-auth.jsx");

  assert.match(auth, /const AuthLoginScreen = \(\{ T, lang, nav, pendingAction \}\) =>/);
  assert.match(auth, /const AuthVerifyScreen = \(\{ T, lang, nav, email, onSignIn, pendingAction \}\) =>/);
  assert.match(auth, /const AuthProfileSetupScreen = \(\{ T, lang, onCompleteProfile, pendingAction \}\) =>/);
  assert.doesNotMatch(app, /auth_login:\s*<AuthLoginScreen[^>]*\bonSignIn=/);
  assert.match(app, /auth_verify:\s*<AuthVerifyScreen[^>]*\bonSignIn=\{completeSignIn\}/);
  assert.doesNotMatch(app, /auth_profile_setup:\s*<AuthProfileSetupScreen[^>]*\bnav=/);
  assert.match(app, /auth_profile_setup:\s*<AuthProfileSetupScreen[^>]*\bonCompleteProfile=\{completeProfile\}/);
});

test("settings prototype screen exposes only consumed display dependencies", () => {
  const app = readUiFile("./src/app.jsx");
  const screens = readUiFile("./src/screens-p0-complete.jsx");

  assert.match(screens, /const SettingsScreen = \(\{ T, lang, fontPreset, setFontPreset \}\) =>/);
  assert.doesNotMatch(app, /settings:\s*<SettingsScreen[^>]*\bnav=/);
  assert.match(app, /settings:\s*<SettingsScreen[^>]*\bfontPreset=\{tweaks\.fontPreset\}/);
  assert.match(app, /settings:\s*<SettingsScreen[^>]*\bsetFontPreset=\{setFontPreset\}/);
});

test("home mini round rail exposes only rendered structured-round dependencies", () => {
  const home = readUiFile("./src/screen-home.jsx");

  assert.match(home, /const MiniRoundRail = \(\{ T, rounds, currentIndex \}\) =>/);
  assert.match(home, /<MiniRoundRail T=\{T\} rounds=\{rounds\} currentIndex=\{currentRoundIndex\} \/>/);
  assert.match(home, /gridTemplateColumns: `repeat\(\$\{rounds\.length\}, 1fr\)`/);
  assert.match(home, /\{round\.name\}\{round\.durationMinutes \? ` · \$\{round\.durationMinutes\}m` : ""\}/);
  assert.match(home, /const current = i === currentIndex/);
});

test("report detail surface exposes dimensions, evidence, and actions only", () => {
  const report = readUiFile("./src/screen-report.jsx");
  assert.match(report, /CAPABILITY ASSESSMENT/);
  assert.match(report, /STRENGTH EVIDENCE/);
  assert.match(report, /NEXT ACTIONS/);
  assert.doesNotMatch(report, /question|Question|题目|qId|turnId/);
});

test("resume create prototype exposes only input-state owner callbacks", () => {
  const workshop = readUiFile("./src/screen-resume-workshop.jsx");
  const createCall = workshop.match(/<ResumeCreateFlow[\s\S]*?\/>/)?.[0];

  assert.ok(createCall, "ResumeCreateFlow call must remain present");
  assert.match(workshop, /const ResumeCreateFlow = \(\{ T, lang, onBack, onCreateResume \}\) =>/);
  assert.doesNotMatch(createCall, /\bnav=/);
  assert.match(createCall, /onBack=\{\(\) => setFlow\("list"\)\}/);
  assert.match(createCall, /onCreateResume=\{addCreatedResume\}/);
  assert.match(workshop, /\{ k: "upload"[\s\S]*\{ k: "paste"/);
  assert.match(workshop, /onCreateResume\(sourceLabel, createMode === "paste" \? resumeText : ""\)/);
});
