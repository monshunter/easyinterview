import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type CSSProperties,
  type FC,
  type KeyboardEvent as ReactKeyboardEvent,
} from "react";

import { ApiClientError } from "../../api/generated/client";
import { generateIdempotencyKey } from "../../lib/conventions/idempotency";
import { useI18n } from "../i18n/messages";
import {
  normalizeAccountDisplayPreferences,
  useDisplayPreferences,
  type CustomAccent,
  type Theme,
} from "../display/DisplayPreferencesProvider";
import { useNavigation } from "../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../runtime/AppRuntimeProvider";
import type { Route } from "../routes";
import { THEME_METADATA } from "../theme/themes.data";

const THEME_OPTIONS = ["ocean", "plum"] as const satisfies readonly Theme[];
const CUSTOM_ACCENT_SEEDS: Record<Theme, CustomAccent> = {
  ocean: { h: 255, c: 0.16 },
  plum: { h: 340, c: 0.15 },
};

const SettingsSectionIcon: FC<{ variant: "appearance" | "account" | "privacy" }> = ({
  variant,
}) => (
  <span className={`ei-settings-card-icon ei-settings-card-icon--${variant}`} aria-hidden="true">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      {variant === "appearance" ? (
        <><path d="M12 3a9 9 0 1 0 0 18c1.7 0 2.3-1 1.4-2.2-.8-1.2.1-2.8 1.6-2.8h2.2A3.8 3.8 0 0 0 21 12.2 9.2 9.2 0 0 0 12 3Z" /><circle cx="7.5" cy="10" r="1" /><circle cx="10" cy="6.8" r="1" /><circle cx="14.3" cy="7.2" r="1" /></>
      ) : variant === "account" ? (
        <><circle cx="12" cy="8" r="3.5" /><path d="M5 20c.7-4 3.1-6 7-6s6.3 2 7 6" /></>
      ) : (
        <><path d="M12 3 5 6v5c0 4.7 2.7 8 7 10 4.3-2 7-5.3 7-10V6l-7-3Z" /><path d="m9 12 2 2 4-4" /></>
      )}
    </svg>
  </span>
);

const SettingsHeaderArt: FC = () => (
  <svg
    className="ei-settings-header-art"
    data-settings-art="security-profile"
    data-testid="settings-header-art"
    aria-hidden="true"
    viewBox="0 0 360 200"
    fill="none"
  >
    <ellipse className="ei-settings-header-art__halo" cx="190" cy="112" rx="146" ry="82" />

    <g
      className="ei-settings-header-art__sparkle ei-settings-header-art__sparkle--left"
      data-settings-art-layer="sparkle"
    >
      <path d="M48 42c0 10-5 15-15 15 10 0 15 5 15 15 0-10 5-15 15-15-10 0-15-5-15-15Z" />
    </g>
    <g
      className="ei-settings-header-art__sparkle ei-settings-header-art__sparkle--right"
      data-settings-art-layer="sparkle"
    >
      <path d="M318 47c0 8-4 12-12 12 8 0 12 4 12 12 0-8 4-12 12-12-8 0-12-4-12-12Z" />
    </g>

    <g data-settings-art-layer="window">
      <rect className="ei-settings-header-art__window-frame" x="104" y="20" width="176" height="154" rx="16" />
      <path className="ei-settings-header-art__window-topbar" d="M104 36c0-8.8 7.2-16 16-16h144c8.8 0 16 7.2 16 16v18H104V36Z" />
      <circle className="ei-settings-header-art__window-dot" cx="121" cy="37" r="3" />
      <circle className="ei-settings-header-art__window-dot" cx="131" cy="37" r="3" />
      <rect className="ei-settings-header-art__window-pill" x="147" y="33" width="43" height="8" rx="4" />
    </g>

    <g className="ei-settings-header-art__profile" data-settings-art-layer="profile">
      <circle className="ei-settings-header-art__avatar-disc" cx="144" cy="87" r="22" />
      <circle className="ei-settings-header-art__avatar-line" cx="144" cy="81" r="7" />
      <path className="ei-settings-header-art__avatar-line" d="M132 99c2.4-7 6.4-10.5 12-10.5S153.6 92 156 99" />
      <rect className="ei-settings-header-art__detail-line ei-settings-header-art__detail-line--strong" x="177" y="70" width="52" height="8" rx="4" />
      <rect className="ei-settings-header-art__detail-line" x="177" y="88" width="67" height="7" rx="3.5" />
      <rect className="ei-settings-header-art__detail-line" x="177" y="104" width="43" height="7" rx="3.5" />
      <rect className="ei-settings-header-art__detail-line" x="124" y="126" width="68" height="7" rx="3.5" />
      <rect className="ei-settings-header-art__detail-line" x="124" y="142" width="50" height="7" rx="3.5" />
    </g>

    <g className="ei-settings-header-art__chart" data-settings-art-layer="chart">
      <rect x="203" y="139" width="10" height="22" rx="3" />
      <rect x="220" y="126" width="10" height="35" rx="3" />
      <rect x="237" y="112" width="10" height="49" rx="3" />
      <rect x="254" y="91" width="10" height="70" rx="3" />
    </g>

    <g className="ei-settings-header-art__lock" data-settings-art-layer="lock">
      <rect className="ei-settings-header-art__lock-tile" x="62" y="104" width="74" height="66" rx="17" />
      <path className="ei-settings-header-art__lock-mark" d="M85 130v-7c0-8 5-13 14-13s14 5 14 13v7" />
      <rect className="ei-settings-header-art__lock-mark" x="82" y="128" width="34" height="25" rx="7" />
      <circle className="ei-settings-header-art__lock-keyhole" cx="99" cy="139" r="3.5" />
      <path className="ei-settings-header-art__lock-keyhole" d="M99 142.5v4" />
    </g>

    <g className="ei-settings-header-art__shield" data-settings-art-layer="shield">
      <path className="ei-settings-header-art__shield-body" d="M286 104c11 7 21 8 30 10v21c0 17-10 29-30 38-20-9-30-21-30-38v-21c9-2 19-3 30-10Z" />
      <path className="ei-settings-header-art__shield-check" d="m273 137 9 9 18-22" />
    </g>
  </svg>
);

export const SettingsScreen: FC<{ route: Route }> = ({ route }) => {
  const { t } = useI18n();
  const { navigate, replaceRoute } = useNavigation();
  const runtime = useAppRuntimeOptional();
  const prefs = useDisplayPreferences();
  const user = runtime?.auth.status === "authenticated" ? runtime.auth.user : null;
  const [themePending, setThemePending] = useState(false);
  const [themeError, setThemeError] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deletePending, setDeletePending] = useState(false);
  const [deleteError, setDeleteError] = useState(false);
  const deleteTriggerRef = useRef<HTMLButtonElement>(null);
  const cancelRef = useRef<HTMLButtonElement>(null);
  const deleteKeyRef = useRef<string | null>(null);
  const mountedRef = useRef(false);
  const saveGenerationRef = useRef(0);
  const currentUserIDRef = useRef<string | null>(user?.id ?? null);
  currentUserIDRef.current = user?.id ?? null;

  useLayoutEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      saveGenerationRef.current += 1;
      prefs.restoreConfirmedAccountPreferences();
    };
  }, [prefs.restoreConfirmedAccountPreferences]);

  const themeDirty =
    prefs.theme !== prefs.confirmedTheme ||
    JSON.stringify(prefs.customAccent) !== JSON.stringify(prefs.confirmedCustomAccent);

  const saveTheme = async () => {
    if (!runtime || !user || themePending || !themeDirty) return;
    const requestGeneration = ++saveGenerationRef.current;
    const requestUserID = user.id;
    setThemePending(true);
    setThemeError(false);
    try {
      const updated = await runtime.client.updateMe({
        displayPreferences: {
          theme: prefs.theme,
          customAccent: prefs.customAccent,
        },
      });
      if (
        !mountedRef.current ||
        saveGenerationRef.current !== requestGeneration ||
        currentUserIDRef.current !== requestUserID
      ) {
        return;
      }
      runtime.refreshAuth(updated);
      prefs.commitAccountPreferences(
        normalizeAccountDisplayPreferences(updated.displayPreferences),
      );
    } catch {
      if (
        mountedRef.current &&
        saveGenerationRef.current === requestGeneration &&
        currentUserIDRef.current === requestUserID
      ) {
        setThemeError(true);
      }
    } finally {
      if (
        mountedRef.current &&
        saveGenerationRef.current === requestGeneration &&
        currentUserIDRef.current === requestUserID
      ) {
        setThemePending(false);
      }
    }
  };

  useEffect(() => {
    if (deleteOpen) cancelRef.current?.focus();
  }, [deleteOpen]);

  const openDelete = () => {
    deleteKeyRef.current = generateIdempotencyKey();
    setDeleteError(false);
    setDeleteOpen(true);
  };

  const closeDelete = () => {
    if (deletePending) return;
    setDeleteOpen(false);
    setDeleteError(false);
    deleteKeyRef.current = null;
    queueMicrotask(() => deleteTriggerRef.current?.focus());
  };

  const handleDialogKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      closeDelete();
      return;
    }
    if (event.key !== "Tab") return;
    const buttons = Array.from(
      event.currentTarget.querySelectorAll<HTMLButtonElement>("button:not(:disabled)"),
    );
    if (buttons.length === 0) return;
    const first = buttons[0];
    const last = buttons[buttons.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last?.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first?.focus();
    }
  };

  const submitDelete = async () => {
    if (!runtime || deletePending || !deleteKeyRef.current) return;
    setDeletePending(true);
    setDeleteError(false);
    try {
      await runtime.client.deleteMe({ idempotencyKey: deleteKeyRef.current });
      setDeleteOpen(false);
      const nextAuth = await runtime.refreshAuth();
      if (nextAuth?.status === "unauthenticated") {
        replaceRoute({ name: "home", params: {} });
      }
    } catch (error: unknown) {
      if (error instanceof ApiClientError && error.status === 401) {
        setDeleteOpen(false);
        const nextAuth = await runtime.refreshAuth();
        if (nextAuth?.status === "unauthenticated") {
          replaceRoute({ name: "home", params: {} });
        }
        return;
      }
      setDeleteError(true);
    } finally {
      setDeletePending(false);
    }
  };

  return (
    <section
      data-testid={`route-${route.name}`}
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
      className="ei-screen-shell ei-settings-screen"
    >
      <header className="ei-settings-header">
        <div className="ei-settings-header-copy">
          <span className="ei-settings-eyebrow ei-text-label">{t("settings.eyebrow")}</span>
          <h1 className="ei-text-display">{t("settings.title")}</h1>
          <p className="ei-text-body">{t("settings.subtitle")}</p>
        </div>
        <SettingsHeaderArt />
      </header>

      <section data-testid="settings-appearance" className="ei-screen-card ei-settings-card">
        <SettingsSectionIcon variant="appearance" />
        <div className="ei-settings-card-body">
        <div className="ei-settings-section-heading">
          <span className="ei-text-label">{t("settings.section.appearance")}</span>
          <h2 className="ei-text-title">{t("settings.appearance")}</h2>
        </div>
        <p className="ei-text-body ei-settings-appearance-copy">{t("settings.appearanceDescription")}</p>
        <div
          data-testid="settings-theme-editor"
          className="ei-settings-theme-editor"
          style={{
            "--ei-settings-accent-hue": String(
              prefs.customAccent?.h ?? CUSTOM_ACCENT_SEEDS[prefs.theme].h,
            ),
          } as CSSProperties}
        >
          <div className="ei-settings-theme-options" role="group" aria-label={t("settings.themeLabel")}>
            {THEME_OPTIONS.map((theme) => {
              const selected = prefs.theme === theme && prefs.customAccent == null;
              const metadata = THEME_METADATA.find((item) => item.key === theme);
              return (
                <button
                  key={theme}
                  type="button"
                  data-testid={`settings-theme-${theme}`}
                  aria-pressed={selected}
                  className={selected ? "ei-settings-theme-option ei-settings-theme-option--selected" : "ei-settings-theme-option"}
                  onClick={() => {
                    prefs.setTheme(theme);
                    prefs.setCustomAccent(null);
                    setThemeError(false);
                  }}
                >
                  <span className="ei-settings-theme-swatch" style={{ background: metadata?.swatch }} aria-hidden="true" />
                  {t(theme === "ocean" ? "theme.ocean" : "theme.plum")}
                </button>
              );
            })}
            <button
              type="button"
              data-testid="settings-theme-custom"
              aria-pressed={prefs.customAccent != null}
              className={prefs.customAccent != null ? "ei-settings-theme-option ei-settings-theme-option--selected" : "ei-settings-theme-option"}
              onClick={() => {
                prefs.setCustomAccent(prefs.customAccent ?? { ...CUSTOM_ACCENT_SEEDS[prefs.theme] });
                setThemeError(false);
              }}
            >
              <span className="ei-settings-theme-swatch ei-settings-theme-swatch--custom" aria-hidden="true" />
              {t("settings.themeCustom")}
            </button>
          </div>
          {prefs.customAccent ? (
            <div data-testid="settings-custom-accent" className="ei-settings-custom-accent">
              <label className="ei-settings-accent-row ei-text-label">
                <span>{t("settings.themeHue")}</span>
                <input
                  data-testid="settings-custom-accent-hue"
                  className="ei-settings-accent-range ei-settings-accent-range--hue"
                  type="range"
                  min={0}
                  max={359}
                  step={1}
                  value={prefs.customAccent.h}
                  onChange={(event) => prefs.setCustomAccent({ ...prefs.customAccent!, h: Number(event.target.value) })}
                />
              </label>
              <label className="ei-settings-accent-row ei-text-label">
                <span>{t("settings.themeChroma")}</span>
                <input
                  data-testid="settings-custom-accent-chroma"
                  className="ei-settings-accent-range ei-settings-accent-range--chroma"
                  type="range"
                  min={0}
                  max={0.28}
                  step={0.005}
                  value={prefs.customAccent.c}
                  onChange={(event) => prefs.setCustomAccent({ ...prefs.customAccent!, c: Number(event.target.value) })}
                />
              </label>
            </div>
          ) : null}
          {themeError ? <p role="alert" className="ei-settings-theme-error ei-text-body">{t("settings.themeSaveError")}</p> : null}
        </div>
        <div className="ei-settings-actions">
          <button type="button" data-testid="settings-theme-save" className="ei-settings-primary-action" disabled={!user || !themeDirty || themePending} onClick={() => void saveTheme()}>
            {themePending ? t("settings.themeSaving") : t("settings.themeSave")}
          </button>
        </div>
        </div>
      </section>

      <section data-testid="settings-account" className="ei-screen-card ei-settings-card">
        <SettingsSectionIcon variant="account" />
        <div className="ei-settings-card-body">
        <div className="ei-settings-section-heading">
          <span className="ei-text-label">{t("settings.section.account")}</span>
          <h2 className="ei-text-title">{t("settings.account")}</h2>
        </div>
        <dl className="ei-settings-values">
          <div className="ei-settings-value-row">
            <dt className="ei-text-body">{t("settings.displayName")}</dt>
            <dd className="ei-text-body">{user?.displayName || "—"}</dd>
          </div>
          <div className="ei-settings-value-row">
            <dt className="ei-text-body">{t("settings.loginEmail")}</dt>
            <dd className="ei-text-body ei-settings-email">{user?.email || "—"}</dd>
          </div>
        </dl>
        <div className="ei-settings-actions">
          <button
            type="button"
            className="ei-settings-secondary-action"
            onClick={() => navigate({ name: "auth_logout", params: {} })}
          >
            {t("user.logout")}
          </button>
        </div>
        </div>
      </section>

      <section data-testid="settings-privacy" className="ei-screen-card ei-settings-card">
        <SettingsSectionIcon variant="privacy" />
        <div className="ei-settings-card-body">
        <div className="ei-settings-section-heading">
          <span className="ei-text-label">{t("settings.section.privacy")}</span>
          <h2 className="ei-text-title">{t("settings.privacy")}</h2>
        </div>
        <div className="ei-settings-privacy-row">
          <div>
            <h3 className="ei-text-subtitle">{t("settings.exportTitle")}</h3>
            <p className="ei-text-body">{t("settings.exportReason")}</p>
          </div>
          <span
            data-testid="settings-export-unavailable"
            className="ei-settings-unavailable ei-text-label"
          >
            {t("settings.exportUnavailable")}
          </span>
        </div>
        <div className="ei-settings-privacy-row ei-settings-danger-row">
          <div>
            <h3 className="ei-text-subtitle">{t("settings.deleteTitle")}</h3>
            <p className="ei-text-body">{t("settings.deleteDescription")}</p>
          </div>
          <button
            ref={deleteTriggerRef}
            type="button"
            className="ei-settings-danger-action"
            disabled={!user}
            onClick={openDelete}
          >
            {t("settings.deleteAction")}
          </button>
        </div>
        </div>
      </section>

      {deleteOpen ? (
        <div className="ei-settings-dialog-layer">
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="delete-account-title"
            aria-describedby="delete-account-description"
            className="ei-settings-dialog"
            onKeyDown={handleDialogKeyDown}
          >
            <span className="ei-text-label">{t("settings.deleteEyebrow")}</span>
            <h2 id="delete-account-title" className="ei-text-title">
              {t("settings.deleteQuestion")}
            </h2>
            <p id="delete-account-description" className="ei-text-body">
              {t("settings.deleteConfirmDescription")}
            </p>
            {deleteError ? (
              <p role="alert" className="ei-settings-dialog-error ei-text-body">
                {t("settings.deleteError")}
              </p>
            ) : null}
            <div className="ei-settings-dialog-actions">
              <button
                ref={cancelRef}
                type="button"
                disabled={deletePending}
                className="ei-settings-secondary-action"
                onClick={closeDelete}
              >
                {t("settings.cancel")}
              </button>
              <button
                type="button"
                disabled={deletePending}
                className="ei-settings-danger-action"
                onClick={() => void submitDelete()}
              >
                {deletePending
                  ? t("settings.deletePending")
                  : deleteError
                    ? t("settings.deleteRetry")
                    : t("settings.deleteConfirm")}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </section>
  );
};
