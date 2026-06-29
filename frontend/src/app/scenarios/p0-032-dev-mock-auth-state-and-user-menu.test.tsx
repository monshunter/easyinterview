// @vitest-environment jsdom
/**
 * E2E.P0.032 — Dev mock auth state and user menu parity scenario.
 *
 * Truth source: docs/spec/frontend-shell/plans/001-app-shell-auth-settings/bdd-plan.md
 *               + bdd-checklist.md.
 *
 * Given the default fixture-backed Vite dev mock client, the mounted App must
 * start signed out, sign in through passwordless auth, expose the ui-design
 * aligned avatar dropdown, route settings/logout from that dropdown,
 * and return to signed out after logout.
 */
import { describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { createDevMockClient } from "../../api/devMockClient";
import { App } from "../App";

describe("E2E.P0.032 dev mock auth state and user menu", () => {
  it("keeps /me stateful across login, settings menu action, and logout", async () => {
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
    expect(screen.queryByTestId("topbar-user-chip")).not.toBeInTheDocument();

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
    expect(screen.getByTestId("topbar-user-chip")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-user-avatar")).toHaveTextContent("AE");
    expect(screen.queryByTestId("topbar-user-profile")).not.toBeInTheDocument();
    expect(screen.queryByTestId("topbar-user-settings")).not.toBeInTheDocument();
    expect(screen.queryByTestId("topbar-user-logout")).not.toBeInTheDocument();

    await user.click(screen.getByTestId("topbar-user-chip"));
    expect(screen.getByTestId("topbar-user-menu")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-user-menu-header")).toHaveTextContent(
      "Alice Example",
    );
    expect(screen.getByTestId("topbar-user-email")).toHaveTextContent(
      "ali***@example.com",
    );

    await user.click(screen.getByTestId("topbar-user-settings"));
    expect(screen.getByTestId("route-settings")).toBeInTheDocument();

    await user.click(screen.getByTestId("topbar-user-chip"));
    await user.click(screen.getByTestId("topbar-user-logout"));
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
    expect(screen.queryByTestId("topbar-user-chip")).not.toBeInTheDocument();

    console.log(
      "E2E.P0.032 evidence: dev mock unauthenticated login avatar dropdown settings logout Alice Example ali***@example.com topbar-user-chip topbar-user-avatar",
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
