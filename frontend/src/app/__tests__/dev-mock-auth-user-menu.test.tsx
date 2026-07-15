// @vitest-environment jsdom
/**
 * Code-level dev-mock auth state and settings-entry regression.
 *
 * Truth source: docs/spec/frontend-shell/plans/001-app-shell-auth-settings/bdd-plan.md
 *               + bdd-checklist.md.
 *
 * Given the default fixture-backed Vite dev mock client, the mounted App must
 * start signed out, sign in through email-code auth, expose the ui-design
 * aligned single settings entry, route logout from Settings,
 * and return to signed out after logout.
 */
import { describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { createDevMockClient } from "../../api/devMockClient";
import { App } from "../App";

describe("dev mock auth state and settings entry", () => {
  it("keeps /me stateful across login, settings action, and logout", async () => {
    setNavigatorLanguages("zh-CN", ["zh-CN", "en-US"]);
    render(<App client={createDevMockClient()} />);

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "false",
      ),
    );
    expect(screen.getByTestId("topbar-login")).toBeInTheDocument();
    expect(screen.queryByTestId("topbar-register")).not.toBeInTheDocument();
    expect(screen.queryByTestId("topbar-settings")).not.toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-login"));
    await user.type(screen.getByTestId("auth-login-email"), "alice@example.com");
    await user.click(screen.getByTestId("auth-login-submit-email"));
    await screen.findByTestId("route-auth_verify");
    await user.type(screen.getByTestId("auth-verify-code"), "654321");
    await user.click(screen.getByTestId("auth-verify-submit"));
    await screen.findByTestId("route-auth_profile_setup");
    await user.type(screen.getByTestId("auth-profile-name"), "Alice Example");
    await user.click(screen.getByTestId("auth-profile-terms"));
    await user.click(screen.getByTestId("auth-profile-submit"));

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "true",
      ),
    );
    expect(screen.getByTestId("topbar-settings")).toBeInTheDocument();
    expect(screen.queryByTestId("topbar-user-menu")).not.toBeInTheDocument();

    await user.click(screen.getByTestId("topbar-settings"));
    expect(screen.getByTestId("route-settings")).toBeInTheDocument();
    expect(screen.getByTestId("settings-account")).toHaveTextContent(
      "Alice Example",
    );
    expect(screen.getByTestId("settings-account")).toHaveTextContent(
      "alice@example.com",
    );

    await user.click(screen.getByRole("button", { name: "退出登录" }));
    await screen.findByTestId("route-auth_logout");
    await user.click(screen.getByTestId("auth-logout-confirm"));

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "false",
      ),
    );
    expect(screen.getByTestId("topbar-login")).toHaveTextContent("登录");
    expect(screen.queryByTestId("topbar-register")).not.toBeInTheDocument();
    expect(screen.queryByTestId("topbar-settings")).not.toBeInTheDocument();

    console.log(
      "dev mock regression: unauthenticated login settings logout Alice Example email=<redacted> topbar-settings",
    );
  });
});

function setNavigatorLanguages(language: string, languages = [language]) {
  Object.defineProperty(window.navigator, "language", {
    value: language,
    configurable: true,
  });
  Object.defineProperty(window.navigator, "languages", {
    value: languages,
    configurable: true,
  });
}
