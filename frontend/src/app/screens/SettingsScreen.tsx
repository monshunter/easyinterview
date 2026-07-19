import {
  useEffect,
  useRef,
  useState,
  type FC,
  type KeyboardEvent as ReactKeyboardEvent,
} from "react";

import { ApiClientError } from "../../api/generated/client";
import { generateIdempotencyKey } from "../../lib/conventions/idempotency";
import { useI18n } from "../i18n/messages";
import { useDisplayPreferences, type CustomAccent, type Theme } from "../display/DisplayPreferencesProvider";
import { useNavigation } from "../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../runtime/AppRuntimeProvider";
import type { Route } from "../routes";
import { THEME_METADATA } from "../theme/themes.data";

const THEME_OPTIONS = ["ocean", "plum"] as const satisfies readonly Theme[];
const CUSTOM_ACCENT_SEEDS: Record<Theme, CustomAccent> = {
  ocean: { h: 255, c: 0.16 },
  plum: { h: 340, c: 0.15 },
};

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

  useEffect(() => () => prefs.restoreConfirmedAccountPreferences(), [prefs.restoreConfirmedAccountPreferences]);

  const themeDirty =
    prefs.theme !== prefs.confirmedTheme ||
    JSON.stringify(prefs.customAccent) !== JSON.stringify(prefs.confirmedCustomAccent);

  const saveTheme = async () => {
    if (!runtime || !user || themePending || !themeDirty) return;
    setThemePending(true);
    setThemeError(false);
    try {
      const updated = await runtime.client.updateMe({
        displayPreferences: {
          theme: prefs.theme,
          customAccent: prefs.customAccent,
        },
      });
      runtime.refreshAuth(updated);
      prefs.commitAccountPreferences(updated.displayPreferences);
    } catch {
      setThemeError(true);
    } finally {
      setThemePending(false);
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
      <header>
        <h1 className="ei-text-display">{t("settings.title")}</h1>
        <p className="ei-text-body">{t("settings.subtitle")}</p>
      </header>

      <section data-testid="settings-appearance" className="ei-screen-card">
        <div className="ei-settings-section-heading">
          <span className="ei-text-label">{t("settings.section.appearance")}</span>
          <h2 className="ei-text-title">{t("settings.appearance")}</h2>
        </div>
        <p className="ei-text-body ei-settings-appearance-copy">{t("settings.appearanceDescription")}</p>
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
              <input data-testid="settings-custom-accent-hue" type="range" min={0} max={359} step={1} value={prefs.customAccent.h} onChange={(event) => prefs.setCustomAccent({ ...prefs.customAccent!, h: Number(event.target.value) })} />
            </label>
            <label className="ei-settings-accent-row ei-text-label">
              <span>{t("settings.themeChroma")}</span>
              <input data-testid="settings-custom-accent-chroma" type="range" min={0} max={0.28} step={0.005} value={prefs.customAccent.c} onChange={(event) => prefs.setCustomAccent({ ...prefs.customAccent!, c: Number(event.target.value) })} />
            </label>
          </div>
        ) : null}
        {themeError ? <p role="alert" className="ei-settings-theme-error ei-text-body">{t("settings.themeSaveError")}</p> : null}
        <div className="ei-settings-actions">
          <button type="button" data-testid="settings-theme-save" className="ei-settings-primary-action" disabled={!user || !themeDirty || themePending} onClick={() => void saveTheme()}>
            {themePending ? t("settings.themeSaving") : t("settings.themeSave")}
          </button>
        </div>
      </section>

      <section data-testid="settings-account" className="ei-screen-card">
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
      </section>

      <section data-testid="settings-privacy" className="ei-screen-card">
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
