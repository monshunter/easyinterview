// @vitest-environment jsdom
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ApiClientError, EasyInterviewClient } from "../../api/generated/client";
import type { PrivacyRequestWithJob, UserContext } from "../../api/generated/types";
import { DisplayPreferencesProvider } from "../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../navigation/NavigationProvider";
import {
  AppRuntimeContext,
  type AppRuntimeValue,
} from "../runtime/AppRuntimeProvider";
import { SettingsScreen } from "./SettingsScreen";

function acceptedDeletion(): PrivacyRequestWithJob {
  return {
    privacyRequestId: "privacy-1",
    job: {
      id: "job-1",
      jobType: "privacy_delete",
      status: "queued",
      resourceType: "privacy_request",
      resourceId: "privacy-1",
      errorCode: null,
      createdAt: "2026-07-15T00:00:00Z",
      updatedAt: "2026-07-15T00:00:00Z",
    },
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, reject, resolve };
}

function themedUser(theme: "ocean" | "plum" = "plum"): UserContext {
  return {
    id: "user-1",
    displayName: "Alice Candidate",
    email: "alice@example.com",
    profileCompletionRequired: false,
    displayPreferences: { theme, customAccent: null },
  };
}

function renderSettings() {
  const client = new EasyInterviewClient({
    fetch: vi.fn() as unknown as typeof fetch,
  });
  const getMe = vi.spyOn(client, "getMe");
  const deleteMe = vi.spyOn(client, "deleteMe");
  const updateMe = vi.spyOn(client, "updateMe");
  const refreshAuth = vi.fn().mockResolvedValue({ status: "unauthenticated" });
  const navigate = vi.fn();
  const replaceRoute = vi.fn();
  const runtime: AppRuntimeValue = {
    client,
    runtime: { status: "loading" },
    auth: {
      status: "authenticated",
      user: {
        id: "user-1",
        displayName: "Alice Candidate",
        email: "alice@example.com",
      profileCompletionRequired: false,
      displayPreferences: { theme: "ocean", customAccent: null },
      },
    },
    refreshAuth,
  };
  const view = render(
    <DisplayPreferencesProvider initial={{ lang: "en" }}>
      <AppRuntimeContext.Provider value={runtime}>
        <NavigationProvider value={{ navigate, replaceRoute }}>
          <SettingsScreen route={{ name: "settings", params: {} }} />
        </NavigationProvider>
      </AppRuntimeContext.Provider>
    </DisplayPreferencesProvider>,
  );
  return { ...view, deleteMe, getMe, navigate, refreshAuth, replaceRoute, updateMe };
}

describe("Settings account and privacy contract", () => {
  it("previews locally and saves one theme update without a follow-up getMe", async () => {
    const { getMe, refreshAuth, updateMe } = renderSettings();
    updateMe.mockResolvedValue({
      id: "user-1",
      displayName: "Alice Candidate",
      email: "alice@example.com",
      profileCompletionRequired: false,
      displayPreferences: { theme: "plum", customAccent: null },
    });
    const user = userEvent.setup();

    expect(screen.getByTestId("settings-appearance")).toBeInTheDocument();
    await user.click(screen.getByTestId("settings-theme-plum"));
    expect(document.documentElement).toHaveAttribute("data-theme", "plum");
    expect(updateMe).not.toHaveBeenCalled();
    expect(getMe).not.toHaveBeenCalled();

    await user.click(screen.getByTestId("settings-theme-save"));
    await waitFor(() => expect(updateMe).toHaveBeenCalledTimes(1));
    expect(updateMe).toHaveBeenCalledWith({ displayPreferences: { theme: "plum", customAccent: null } });
    expect(refreshAuth).toHaveBeenCalledTimes(1);
    expect(getMe).not.toHaveBeenCalled();
  });

  it("keeps preset themes available while the custom color editor expands below them", async () => {
    const { getMe, updateMe } = renderSettings();
    const user = userEvent.setup();

    expect(screen.queryByTestId("settings-custom-accent")).not.toBeInTheDocument();
    await user.click(screen.getByTestId("settings-theme-custom"));

    const editor = screen.getByTestId("settings-theme-editor");
    const primaryRow = screen.getByTestId("settings-theme-primary-row");
    const options = screen.getByRole("group", { name: "Choose a theme color" });
    const save = screen.getByTestId("settings-theme-save");
    const customPanel = screen.getByTestId("settings-custom-accent");
    expect(editor).toContainElement(options);
    expect(editor).toContainElement(customPanel);
    expect(primaryRow).toContainElement(options);
    expect(primaryRow).toContainElement(save);
    expect(primaryRow).not.toContainElement(customPanel);
    expect(primaryRow.compareDocumentPosition(customPanel) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    for (const theme of ["ocean", "plum", "custom"]) {
      expect(screen.getByTestId(`settings-theme-${theme}`)).toBeVisible();
    }
    expect(screen.getByTestId("settings-custom-accent-hue")).toHaveClass(
      "ei-settings-accent-range--hue",
    );
    expect(screen.getByTestId("settings-custom-accent-chroma")).toHaveClass(
      "ei-settings-accent-range--chroma",
    );
    expect(editor.style.getPropertyValue("--ei-settings-accent-hue")).toBe("255");
    expect(updateMe).not.toHaveBeenCalled();
    expect(getMe).not.toHaveBeenCalled();

    await user.click(screen.getByTestId("settings-theme-plum"));
    expect(screen.queryByTestId("settings-custom-accent")).not.toBeInTheDocument();
    expect(screen.getByTestId("settings-theme-plum")).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByTestId("settings-theme-custom")).toHaveAttribute("aria-pressed", "false");
    expect(updateMe).not.toHaveBeenCalled();
    expect(getMe).not.toHaveBeenCalled();
  });

  it("keeps a rejected theme draft retryable and commits only the successful response", async () => {
    const { refreshAuth, updateMe } = renderSettings();
    updateMe
      .mockRejectedValueOnce(new Error("temporary failure"))
      .mockResolvedValueOnce(themedUser());
    const user = userEvent.setup();

    await user.click(screen.getByTestId("settings-theme-plum"));
    await user.click(screen.getByTestId("settings-theme-save"));

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(document.documentElement).toHaveAttribute("data-theme", "plum");
    expect(refreshAuth).not.toHaveBeenCalled();

    await user.click(screen.getByTestId("settings-theme-save"));
    await waitFor(() => expect(refreshAuth).toHaveBeenCalledTimes(1));
    expect(updateMe).toHaveBeenCalledTimes(2);
  });

  it("does not commit a late theme response after Settings unmounts", async () => {
    const pending = deferred<UserContext>();
    const { refreshAuth, unmount, updateMe } = renderSettings();
    updateMe.mockReturnValue(pending.promise);
    const user = userEvent.setup();

    await user.click(screen.getByTestId("settings-theme-plum"));
    await user.click(screen.getByTestId("settings-theme-save"));
    await waitFor(() => expect(updateMe).toHaveBeenCalledTimes(1));
    unmount();

    await act(async () => {
      pending.resolve(themedUser());
      await pending.promise;
    });

    expect(refreshAuth).not.toHaveBeenCalled();
  });

  it("uses the runtime user without a second getMe and routes sign-out", async () => {
    const { getMe, navigate } = renderSettings();

    expect(screen.getByTestId("settings-account")).toHaveTextContent("Alice Candidate");
    expect(screen.getByTestId("settings-account")).toHaveTextContent("alice@example.com");
    expect(getMe).not.toHaveBeenCalled();
    expect(screen.queryByTestId("settings-tabs")).not.toBeInTheDocument();
    expect(screen.queryByTestId("settings-login-security")).not.toBeInTheDocument();
    expect(screen.queryByTestId("settings-font-preset")).not.toBeInTheDocument();
    expect(screen.queryByTestId("settings-app-info")).not.toBeInTheDocument();

    await userEvent.setup().click(screen.getByRole("button", { name: "Sign out" }));
    expect(navigate).toHaveBeenCalledWith({ name: "auth_logout", params: {} });
  });

  it("presents export as unavailable without a trigger", () => {
    renderSettings();
    const exportStatus = screen.getByTestId("settings-export-unavailable");
    expect(exportStatus).toHaveTextContent(/not available/i);
    expect(exportStatus.tagName).not.toBe("BUTTON");
  });

  it("opens an accessible destructive dialog, cancels with Escape, and returns focus", async () => {
    renderSettings();
    const user = userEvent.setup();
    const trigger = screen.getByRole("button", { name: "Delete account" });
    await user.click(trigger);

    const dialog = screen.getByRole("dialog", { name: "Delete account?" });
    expect(dialog).toHaveAttribute("aria-describedby", "delete-account-description");
    expect(screen.getByRole("button", { name: "Cancel" })).toHaveFocus();
    await user.keyboard("{Shift>}{Tab}{/Shift}");
    expect(screen.getByRole("button", { name: "Confirm deletion" })).toHaveFocus();
    await user.keyboard("{Tab}");
    expect(screen.getByRole("button", { name: "Cancel" })).toHaveFocus();
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it("reuses one idempotency key for recoverable retry", async () => {
    const { deleteMe } = renderSettings();
    deleteMe
      .mockRejectedValueOnce(new Error("network unavailable"))
      .mockResolvedValueOnce(acceptedDeletion());
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Delete account" }));
    await user.click(screen.getByRole("button", { name: "Confirm deletion" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/try again/i);
    const firstKey = deleteMe.mock.calls[0]?.[0]?.idempotencyKey;
    expect(firstKey).toMatch(/^v1\./);
    await user.click(screen.getByRole("button", { name: "Try again" }));
    expect(deleteMe.mock.calls[1]?.[0]?.idempotencyKey).toBe(firstKey);
  });

  it("locks close and duplicate submit while deletion is pending", async () => {
    let resolveDelete!: (value: PrivacyRequestWithJob) => void;
    const pending = new Promise<PrivacyRequestWithJob>((resolve) => {
      resolveDelete = resolve;
    });
    const { deleteMe } = renderSettings();
    deleteMe.mockReturnValueOnce(pending);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Delete account" }));
    await user.click(screen.getByRole("button", { name: "Confirm deletion" }));

    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Deleting…" })).toBeDisabled();
    await user.keyboard("{Escape}");
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(deleteMe).toHaveBeenCalledTimes(1);

    resolveDelete(acceptedDeletion());
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("re-probes auth and replaces Home only after a 202 resolves as unauthenticated", async () => {
    const { deleteMe, refreshAuth, replaceRoute } = renderSettings();
    deleteMe.mockResolvedValue(acceptedDeletion());
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Delete account" }));
    await user.click(screen.getByRole("button", { name: "Confirm deletion" }));

    await waitFor(() => expect(refreshAuth).toHaveBeenCalledTimes(1));
    expect(replaceRoute).toHaveBeenCalledWith({ name: "home", params: {} });
  });

  it("treats a typed 401 as an auth re-probe without a recoverable retry alert", async () => {
    const { deleteMe, refreshAuth, replaceRoute } = renderSettings();
    deleteMe.mockRejectedValue(new ApiClientError("http", 401, null));
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Delete account" }));
    await user.click(screen.getByRole("button", { name: "Confirm deletion" }));

    await waitFor(() => expect(refreshAuth).toHaveBeenCalledTimes(1));
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(replaceRoute).toHaveBeenCalledWith({ name: "home", params: {} });
  });

  it("keeps probe failures honest and does not replace Home", async () => {
    const { deleteMe, refreshAuth, replaceRoute } = renderSettings();
    deleteMe.mockResolvedValue(acceptedDeletion());
    refreshAuth.mockResolvedValueOnce({ status: "error", error: new Error("probe failed") });
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Delete account" }));
    await user.click(screen.getByRole("button", { name: "Confirm deletion" }));

    await waitFor(() => expect(refreshAuth).toHaveBeenCalledTimes(1));
    expect(replaceRoute).not.toHaveBeenCalled();
  });
});
