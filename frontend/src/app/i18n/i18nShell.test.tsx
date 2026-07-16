// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { AuthLoginScreen } from "../auth/AuthLoginScreen";
import { DisplayPreferencesProvider } from "../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../navigation/NavigationProvider";
import { RouteShellScreen } from "../screens/RouteShellScreen";
import { SettingsScreen } from "../screens/SettingsScreen";
import { TopBar } from "../topbar/TopBar";

describe("D1 shell i18n", () => {
  it("renders TopBar, auth, settings, and route shell copy in English without localizing route keys", async () => {
    render(
      <DisplayPreferencesProvider initial={{ lang: "en" }}>
        <TopBar activeRoute="home" onNavigate={() => {}} signedIn={false} />
        <TopBar activeRoute="home" onNavigate={() => {}} signedIn={true} />
        <AuthLoginScreen
          route={{ name: "auth_login", params: {} }}
          onNavigate={() => {}}
          onStartChallenge={vi.fn()}
        />
        <NavigationProvider value={{ navigate: () => {}, replaceRoute: () => {} }}>
          <SettingsScreen route={{ name: "settings", params: {} }} />
        </NavigationProvider>
        <RouteShellScreen
          route={{ name: "workspace", params: { planId: "plan-tj-1" } }}
        />
      </DisplayPreferencesProvider>,
    );

    expect(screen.getAllByTestId("topbar-nav-home")[0]).toHaveTextContent(
      /^Home$/,
    );
    expect(screen.getAllByTestId("topbar-nav-workspace")[0]).toHaveTextContent(
      /^Interview$/,
    );
    expect(screen.getAllByTestId("topbar-login")[0]).toHaveTextContent(
      "Sign in",
    );
    expect(screen.queryByTestId("topbar-register")).not.toBeInTheDocument();
    for (const languageControl of screen.getAllByTestId("topbar-lang-toggle")) {
      expect(languageControl.tagName).toBe("BUTTON");
      expect(languageControl).toHaveTextContent("English");
    }
    expect(screen.getByTestId("topbar-settings")).toHaveAccessibleName(
      "Settings & privacy",
    );
    expect(screen.queryByTestId("topbar-user-menu")).not.toBeInTheDocument();

    expect(screen.getByTestId("route-auth_login")).toHaveTextContent(
      "Sign in",
    );
    expect(screen.getByTestId("auth-login-submit-email")).toHaveTextContent(
      "Send sign-in code",
    );
    expect(screen.getByTestId("route-settings")).toHaveTextContent(
      "Settings & privacy",
    );
    expect(screen.getByTestId("route-workspace")).toHaveTextContent(
      "Interview plans",
    );

    expect(screen.getByTestId("route-auth_login")).toHaveAttribute(
      "data-route-name",
      "auth_login",
    );
    expect(screen.getByTestId("route-workspace")).toHaveAttribute(
      "data-route-name",
      "workspace",
    );
    expect(screen.getByTestId("route-workspace")).toHaveAttribute(
      "data-route-params",
      JSON.stringify({ planId: "plan-tj-1" }),
    );
  });
});
