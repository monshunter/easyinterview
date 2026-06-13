// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

import { ProfileScreen } from "./ProfileScreen";
import { SettingsScreen } from "./SettingsScreen";

describe("ProfileScreen", () => {
  it("renders the user-profile shell with documented sections", () => {
    render(
      <ProfileScreen route={{ name: "profile", params: {} }} />,
    );
    expect(screen.getByTestId("route-profile")).toBeInTheDocument();
    expect(screen.getByTestId("profile-identity-summary")).toBeInTheDocument();
    expect(screen.getByTestId("profile-sections")).toBeInTheDocument();
    expect(screen.getByTestId("profile-insight-cards")).toBeInTheDocument();
    expect(screen.getByTestId("profile-used-by")).toBeInTheDocument();
    expect(screen.getByTestId("profile-recent-evidence")).toBeInTheDocument();
  });

  it("does not surface legacy Growth / Experiences / Mistakes / Drill modules", () => {
    render(<ProfileScreen route={{ name: "profile", params: {} }} />);
    for (const legacy of [
      "profile-growth",
      "profile-experiences",
      "profile-mistakes",
      "profile-drill",
      "profile-followup",
    ]) {
      expect(screen.queryByTestId(legacy)).not.toBeInTheDocument();
    }
    expect(screen.queryByText(/错题本|成长中心|经历库/)).not.toBeInTheDocument();
  });
});

describe("SettingsScreen", () => {
  it("renders the profile tab sections by default and privacy on its own tab (D-21)", () => {
    render(
      <SettingsScreen route={{ name: "settings", params: {} }} />,
    );
    expect(screen.getByTestId("route-settings")).toBeInTheDocument();
    expect(screen.getByTestId("settings-account")).toBeInTheDocument();
    expect(
      screen.getByTestId("settings-login-security"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("settings-login-security").textContent,
    ).toMatch(/邮箱验证码 · 无密码|Email sign-in code · no password/);
    expect(screen.getByTestId("settings-font-preset")).toBeInTheDocument();
    expect(screen.getByTestId("settings-app-info")).toHaveTextContent("v1.0");
    expect(screen.queryByTestId("settings-privacy")).not.toBeInTheDocument();

    fireEvent.click(screen.getByTestId("settings-tab-privacy"));
    expect(screen.getByTestId("settings-privacy")).toBeInTheDocument();
    expect(screen.queryByTestId("settings-account")).not.toBeInTheDocument();
  });

  it("does not restore legacy Growth / Experiences / Mistakes / Drill modules", () => {
    render(<SettingsScreen route={{ name: "settings", params: {} }} />);
    for (const legacy of [
      "settings-growth",
      "settings-experiences",
      "settings-mistakes",
      "settings-drill",
      "settings-target-role",
      "settings-skills",
    ]) {
      expect(screen.queryByTestId(legacy)).not.toBeInTheDocument();
    }
    expect(
      screen.queryByText(/错题本|成长中心|经历库|目标角色|技能标签/),
    ).not.toBeInTheDocument();
  });

  it("does not keep notifications / subscription placeholder tabs (D-21)", () => {
    render(<SettingsScreen route={{ name: "settings", params: {} }} />);
    expect(
      screen.queryByTestId("settings-notifications-placeholder"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("settings-subscription-placeholder"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText(/通知（P1 占位）|订阅（P1 占位）/),
    ).not.toBeInTheDocument();
  });
});
