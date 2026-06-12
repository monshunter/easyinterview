// Auth screens · unified email login, verification, profile setup, logout

const AuthShell = ({ T, lang, eyebrow, title, sub, children, side }) => (
  <div className="ei-fadein" style={{ maxWidth: 1160, margin: "0 auto", padding: "54px 48px 96px" }}>
    <div style={{ display: "grid", gridTemplateColumns: "0.88fr 1.12fr", gap: 44, alignItems: "start" }}>
      <div style={{ position: "sticky", top: 92 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{eyebrow}</div>
        <h1 className="ei-serif" style={{ margin: 0, color: T.ink, fontSize: 42, lineHeight: 1.1, letterSpacing: "-0.026em" }}>{title}</h1>
        <div style={{ marginTop: 14, color: T.ink3, fontSize: 14, lineHeight: 1.7, maxWidth: 430 }}>{sub}</div>
        <div style={{ marginTop: 26, padding: 18, border: `1px solid ${T.rule}`, background: T.bgSoft, borderRadius: 3 }}>
          {side || (
            <>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "AUTH PRINCIPLE" : "认证原则"}</div>
              <div style={{ color: T.ink2, fontSize: 13, lineHeight: 1.65 }}>
                {lang === "en"
                  ? "The app can be browsed first. Authentication is requested only when a user needs to save, sync, export, or continue across devices."
                  : "用户可以先进入产品浏览和开始准备；只有当需要保存、同步、导出或跨设备继续时，才进入邮箱验证码登录。首次使用的邮箱会在验证后创建账号。"}
              </div>
            </>
          )}
        </div>
      </div>
      <Card T={T} pad={0} style={{ overflow: "hidden" }}>
        {children}
      </Card>
    </div>
  </div>
);

const AuthField = ({ T, label, value, setValue, type = "text", placeholder }) => (
  <label style={{ display: "block" }}>
    <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{label}</div>
    <input type={type} value={value} onChange={(e) => setValue(e.target.value)} placeholder={placeholder} style={{
      width: "100%", height: 42, border: `1px solid ${T.rule}`, background: T.bg,
      color: T.ink, padding: "0 12px", fontSize: 14, outline: "none", borderRadius: 2,
    }} />
  </label>
);

const AuthActionLink = ({ T, children, onClick }) => (
  <button onClick={onClick} style={{ background: "transparent", border: "none", color: T.accent, padding: 0, fontSize: 12.5, cursor: "pointer" }}>
    {children}
  </button>
);

// Reusable hook: countdown after sending an email code
const useResendCountdown = () => {
  const [seconds, setSeconds] = React.useState(0);
  React.useEffect(() => {
    if (seconds <= 0) return;
    const t = setTimeout(() => setSeconds((s) => s - 1), 1000);
    return () => clearTimeout(t);
  }, [seconds]);
  return [seconds, () => setSeconds(60)];
};

const PendingActionPanel = ({ T, lang, pendingAction }) => {
  if (!pendingAction) return null;
  return (
    <>
      <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>{lang === "en" ? "PENDING ACTION" : "待恢复操作"}</div>
      <div style={{ color: T.ink, fontSize: 14, fontWeight: 600, lineHeight: 1.45 }}>{pendingAction.label}</div>
      <div style={{ color: T.ink3, fontSize: 12.5, lineHeight: 1.6, marginTop: 8 }}>
        {lang === "en" ? "After sign-in, EasyInterview returns to this exact action instead of sending you home." : "登录成功后会回到刚才的动作，不会统一回首页。"}
      </div>
    </>
  );
};

const AuthLoginScreen = ({ T, lang, nav, onSignIn, pendingAction }) => {
  const [email, setEmail] = React.useState("name@example.com");
  const [cooldown, startCooldown] = useResendCountdown();
  const sendCode = () => {
    if (cooldown > 0) return;
    startCooldown();
    window.eiToast && window.eiToast(
      lang === "en" ? `Sign-in code sent to ${email}` : `登录验证码已发送到 ${email}`,
      { tone: "ok" }
    );
  };

  return (
    <AuthShell
      T={T}
      lang={lang}
      eyebrow={lang === "en" ? "LOGIN" : "登录"}
      title={lang === "en" ? "Continue with your email." : "用邮箱继续。"}
      sub={lang === "en" ? "Use one email code entry for both new and existing accounts. If this is your first time, we ask for your name after verification." : "新用户和已有用户都从这里输入邮箱验证码。首次使用该邮箱时，验证后再补充显示姓名。"}
      side={pendingAction && <PendingActionPanel T={T} lang={lang} pendingAction={pendingAction} />}
    >
      <div style={{ padding: 28 }}>
        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <AuthField T={T} label={lang === "en" ? "Email" : "邮箱"} value={email} setValue={setEmail} placeholder="name@example.com" />
          <div style={{ padding: 14, border: `1px solid ${T.rule}`, background: T.bgSoft, color: T.ink3, fontSize: 12.5, lineHeight: 1.6 }}>
            {lang === "en"
              ? "We do not ask whether you are registering or signing in. Existing emails open the existing account; new emails create an account after verification."
              : "不需要选择注册或登录。已有邮箱会进入原账号，新邮箱会在验证后创建账号。"}
          </div>
        </div>

        <Btn T={T} variant="accent" icon="send" style={{ width: "100%", marginTop: 22 }} onClick={() => { sendCode(); nav("auth_verify", { email, pendingAction }); }}>
          {cooldown > 0
            ? (lang === "en" ? `Code sent · ${cooldown}s` : `验证码已发送 · ${cooldown}s`)
            : (lang === "en" ? "Send code and continue" : "发送验证码并继续")}
        </Btn>

        <div style={{ marginTop: 16, fontSize: 12.5, color: T.ink3, lineHeight: 1.6 }}>
          <div>{lang === "en" ? "One email can only own one account." : "一个邮箱只能对应一个账号。"}</div>
          <div style={{ marginTop: 4 }}>{lang === "en" ? "Code not arriving? You can resend it or switch email on the next step." : "收不到验证码？下一步可以重新发送，或换一个邮箱重试。"}</div>
        </div>
      </div>
    </AuthShell>
  );
};

const AuthProfileSetupScreen = ({ T, lang, nav, onCompleteProfile, pendingAction }) => {
  const [name, setName] = React.useState("");
  const [accepted, setAccepted] = React.useState(true);

  return (
    <AuthShell
      T={T}
      lang={lang}
      eyebrow={lang === "en" ? "PROFILE SETUP" : "完善资料"}
      title={lang === "en" ? "Tell us what to call you." : "告诉我们怎么称呼你。"}
      sub={lang === "en" ? "This appears in your interview workspace and reports. It is not used to decide account uniqueness; your email is the account identity." : "这个名称会显示在训练工作台和报告里。账号唯一性只由邮箱决定，显示姓名不参与去重。"}
      side={pendingAction && <PendingActionPanel T={T} lang={lang} pendingAction={pendingAction} />}
    >
      <div style={{ padding: 28 }}>
        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <AuthField T={T} label={lang === "en" ? "Display name" : "显示姓名"} value={name} setValue={setName} />
        </div>
        <label style={{ display: "flex", gap: 8, alignItems: "flex-start", fontSize: 12.5, color: T.ink3, lineHeight: 1.55, marginTop: 16 }}>
          <input type="checkbox" checked={accepted} onChange={(e) => setAccepted(e.target.checked)} style={{ marginTop: 3 }} />
          <span>{lang === "en" ? "I agree to the Terms and Privacy Policy. Sensitive interview data can be exported or deleted later." : "我同意服务条款和隐私政策。敏感的面试数据之后可以导出或删除。"}</span>
        </label>
        <Btn T={T} variant="accent" icon="check" style={{ width: "100%", marginTop: 22 }} disabled={!name.trim() || !accepted} onClick={() => onCompleteProfile && onCompleteProfile(name)}>
          {lang === "en" ? "Finish setup and continue" : "完成资料并继续"}
        </Btn>
        <div style={{ marginTop: 14, color: T.ink3, fontSize: 12.5, lineHeight: 1.6 }}>
          {lang === "en"
            ? "If you close this browser before finishing, the next sign-in with the same email returns here first."
            : "如果你在完成前关闭浏览器，下次用同一邮箱登录后仍会先回到这里。"}
        </div>
      </div>
    </AuthShell>
  );
};

const AuthVerifyScreen = ({ T, lang, nav, email, onSignIn, pendingAction }) => {
  const [code, setCode] = React.useState("");
  const [cooldown, startCooldown] = useResendCountdown();
  const resend = () => {
    if (cooldown > 0) return;
    startCooldown();
    window.eiToast && window.eiToast(
      lang === "en" ? `New sign-in code sent to ${email || "your email"}` : `新的登录验证码已发送到 ${email || "你的邮箱"}`,
      { tone: "ok" }
    );
  };

  return (
    <AuthShell
      T={T}
      lang={lang}
      eyebrow={lang === "en" ? "EMAIL VERIFICATION" : "邮箱验证"}
      title={lang === "en" ? "Confirm this is your email." : "确认这是你的邮箱。"}
      sub={lang === "en" ? `We sent a 6-digit sign-in code to ${email || "your email"}. It expires in 5 minutes.` : `我们已向 ${email || "你的邮箱"} 发送 6 位登录验证码。验证码 5 分钟内有效。`}
      side={pendingAction && <PendingActionPanel T={T} lang={lang} pendingAction={pendingAction} />}
    >
      <div style={{ padding: 28 }}>
        <AuthField T={T} label={lang === "en" ? "6-digit code" : "6 位验证码"} value={code} setValue={setCode} placeholder="123456" />
        <Btn T={T} variant="accent" icon="check" style={{ width: "100%", marginTop: 22 }} onClick={onSignIn}>
          {lang === "en" ? "Verify and continue" : "验证并继续"}
        </Btn>
        <div style={{ marginTop: 18, display: "flex", justifyContent: "space-between", fontSize: 12.5 }}>
          <AuthActionLink T={T} onClick={resend}>
            {cooldown > 0
              ? (lang === "en" ? `Resend in ${cooldown}s` : `${cooldown}s 后可重发`)
              : (lang === "en" ? "Resend code" : "重新发送验证码")}
          </AuthActionLink>
          <AuthActionLink T={T} onClick={() => nav("auth_login", { pendingAction })}>{lang === "en" ? "Use another email" : "换一个邮箱"}</AuthActionLink>
        </div>
      </div>
    </AuthShell>
  );
};

const AuthLogoutScreen = ({ T, lang, nav, signedIn, onSignOut }) => {
  const [done, setDone] = React.useState(!signedIn);

  const confirm = () => {
    onSignOut && onSignOut();
    setDone(true);
  };

  return (
    <AuthShell
      T={T}
      lang={lang}
      eyebrow={lang === "en" ? "LOGOUT" : "退出登录"}
      title={done ? (lang === "en" ? "You are signed out." : "你已退出登录。") : (lang === "en" ? "Sign out of this device?" : "从这台设备退出？")}
      sub={done
        ? (lang === "en" ? "You can still browse the product. Saving or syncing will ask you to sign in again." : "你仍然可以浏览产品。需要保存或同步时，系统会再次引导登录。")
        : (lang === "en" ? "This only clears the local session. Saved resumes, JDs, reports, and debrief records stay in your account." : "这只会清除本机登录态。已保存的简历、JD、报告和复盘记录仍保留在你的账号里。")}
    >
      <div style={{ padding: 28 }}>
        {done ? (
          <>
            <div style={{ padding: 18, background: T.bgSoft, border: `1px solid ${T.rule}`, color: T.ink2, fontSize: 13.5, lineHeight: 1.65, marginBottom: 20 }}>
              {lang === "en" ? "Local session cleared. Account data was not deleted." : "本地会话已清除，账号数据没有被删除。"}
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10 }}>
              <Btn T={T} variant="accent" icon="arrow_right" onClick={() => nav("auth_login")}>{lang === "en" ? "Sign in again" : "重新登录"}</Btn>
              <Btn T={T} variant="secondary" onClick={() => nav("home")}>{lang === "en" ? "Back home" : "返回首页"}</Btn>
            </div>
          </>
        ) : (
          <>
            <div style={{ padding: 18, background: T.warnSoft, border: `1px solid ${T.warn}`, color: T.ink2, fontSize: 13.5, lineHeight: 1.65, marginBottom: 20 }}>
              {lang === "en" ? "After signing out, unsaved edits on this device may be lost." : "退出后，这台设备上尚未保存的编辑可能丢失。"}
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10 }}>
              <Btn T={T} variant="danger" icon="logout" onClick={confirm}>{lang === "en" ? "Sign out" : "确认退出"}</Btn>
              <Btn T={T} variant="secondary" onClick={() => nav("home")}>{lang === "en" ? "Cancel" : "取消"}</Btn>
            </div>
          </>
        )}
      </div>
    </AuthShell>
  );
};

Object.assign(window, { AuthLoginScreen, AuthProfileSetupScreen, AuthVerifyScreen, AuthLogoutScreen });
