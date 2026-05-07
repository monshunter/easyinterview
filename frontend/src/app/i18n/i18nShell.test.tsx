// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { AuthLoginScreen } from "../auth/AuthLoginScreen";
import { DisplayPreferencesProvider } from "../display/DisplayPreferencesProvider";
import { PlaceholderScreen } from "../screens/PlaceholderScreen";
import { ProfileScreen } from "../screens/ProfileScreen";
import { SettingsScreen } from "../screens/SettingsScreen";
import { TopBar } from "../topbar/TopBar";

describe("D1 shell i18n", () => {
  it("renders TopBar, auth, profile/settings, and placeholder shell copy in English without localizing route keys", () => {
    render(
      <DisplayPreferencesProvider initial={{ lang: "en" }}>
        <TopBar activeRoute="home" onNavigate={() => {}} signedIn={false} />
        <TopBar activeRoute="home" onNavigate={() => {}} signedIn={true} />
        <AuthLoginScreen
          route={{ name: "auth_login", params: {} }}
          onNavigate={() => {}}
          onStartChallenge={vi.fn()}
        />
        <ProfileScreen route={{ name: "profile", params: {} }} />
        <SettingsScreen route={{ name: "settings", params: {} }} />
        <PlaceholderScreen
          route={{ name: "workspace", params: { planId: "plan-tj-1" } }}
        />
      </DisplayPreferencesProvider>,
    );

    expect(screen.getAllByTestId("topbar-nav-home")[0]).toHaveTextContent(
      "Home",
    );
    expect(screen.getAllByTestId("topbar-nav-jd_match")[0]).toHaveTextContent(
      "Job Picks",
    );
    expect(screen.getAllByTestId("topbar-login")[0]).toHaveTextContent(
      "Sign in",
    );
    expect(screen.getAllByTestId("topbar-register")[0]).toHaveTextContent(
      "Register",
    );
    for (const languageControl of screen.getAllByLabelText("Language")) {
      expect(languageControl.tagName).toBe("SELECT");
    }
    expect(screen.getByTestId("topbar-user-profile")).toHaveTextContent(
      "User profile",
    );
    expect(screen.getByTestId("topbar-user-settings")).toHaveTextContent(
      "Settings & privacy",
    );
    expect(screen.getByTestId("topbar-user-logout")).toHaveTextContent(
      "Sign out",
    );

    expect(screen.getByTestId("route-auth_login")).toHaveTextContent(
      "Sign in",
    );
    expect(screen.getByTestId("auth-login-submit-email")).toHaveTextContent(
      "Send sign-in email",
    );
    expect(screen.getByTestId("route-profile")).toHaveTextContent(
      "User profile",
    );
    expect(screen.getByTestId("route-settings")).toHaveTextContent(
      "Settings & privacy",
    );
    expect(screen.getByTestId("route-workspace")).toHaveTextContent(
      "Workspace",
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
