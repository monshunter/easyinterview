// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  AuthLoginScreen,
  AuthLogoutScreen,
  AuthProfileSetupScreen,
  AuthResetScreen,
  AuthVerifyScreen,
} from "./index";
import type { Route } from "../routes";

const baseRoute = (name: Route["name"]): Route => ({ name, params: {} });

describe("AuthLoginScreen", () => {
  it("renders the single email-code login route shell", () => {
    render(
      <AuthLoginScreen
        route={baseRoute("auth_login")}
        onNavigate={() => {}}
        onStartChallenge={async () => {}}
      />,
    );
    expect(screen.getByTestId("route-auth_login")).toBeInTheDocument();
    expect(screen.getByTestId("auth-login-email")).toBeInTheDocument();
    expect(screen.queryByTestId("auth-login-password-stub")).not.toBeInTheDocument();
    expect(screen.queryByTestId("auth-login-oauth-stub")).not.toBeInTheDocument();
    expect(screen.queryByTestId("auth-login-link-register")).not.toBeInTheDocument();
    expect(screen.getByTestId("auth-login-link-reset")).toBeInTheDocument();
  });

  it("submits the email challenge with email only and carries pendingAction params to verify", async () => {
    const onStartChallenge = vi.fn().mockResolvedValue(undefined);
    const onNavigate = vi.fn();
    render(
      <AuthLoginScreen
        route={{
          name: "auth_login",
          params: {
            returnTo: "/practice?planId=legacy-return-to",
            pendingRoute: "practice",
            pendingType: "start_practice",
            pendingLabel: "立即面试",
            planId: "plan-tj-1",
          },
        }}
        onNavigate={onNavigate}
        onStartChallenge={onStartChallenge}
      />,
    );
    const user = userEvent.setup();
    await user.type(
      screen.getByTestId("auth-login-email"),
      "alice@example.com",
    );
    await user.click(screen.getByTestId("auth-login-submit-email"));

    expect(onStartChallenge).toHaveBeenCalledWith({
      email: "alice@example.com",
    });
    await waitFor(() =>
      expect(onNavigate).toHaveBeenCalledWith({
        name: "auth_verify",
        params: expect.objectContaining({
          email: "alice@example.com",
          pendingRoute: "practice",
          planId: "plan-tj-1",
        }),
      }),
    );
    expect(onNavigate.mock.calls[0]?.[0].params).not.toHaveProperty("returnTo");
  });

  it("navigates to reset route from the inline link", async () => {
    const onNavigate = vi.fn();
    render(
      <AuthLoginScreen
        route={baseRoute("auth_login")}
        onNavigate={onNavigate}
        onStartChallenge={async () => {}}
      />,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("auth-login-link-reset"));
    expect(onNavigate).toHaveBeenCalledWith({
      name: "auth_reset",
      params: {},
    });
  });
});

describe("AuthProfileSetupScreen", () => {
  it("renders profile setup fields and routes to the pending action after completion", async () => {
    const onCompleteProfile = vi.fn().mockResolvedValue({
      profileCompletionRequired: false,
    });
    const onNavigate = vi.fn();
    render(
      <AuthProfileSetupScreen
        route={{
          name: "auth_profile_setup",
          params: {
            pendingRoute: "practice",
            pendingType: "start_practice",
            pendingLabel: "立即面试",
            planId: "plan-tj-1",
          },
        }}
        onNavigate={onNavigate}
        onCompleteProfile={onCompleteProfile}
      />,
    );
    expect(screen.getByTestId("route-auth_profile_setup")).toBeInTheDocument();
    expect(screen.getByTestId("auth-profile-name")).toBeInTheDocument();
    expect(screen.getByTestId("auth-profile-terms")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.type(screen.getByTestId("auth-profile-name"), "Alice");
    await user.click(screen.getByTestId("auth-profile-terms"));
    await user.click(screen.getByTestId("auth-profile-submit"));

    expect(onCompleteProfile).toHaveBeenCalledWith({
      displayName: "Alice",
      acceptedTerms: true,
    });
    expect(onNavigate).toHaveBeenCalledWith({
      name: "practice",
      params: expect.objectContaining({
        planId: "plan-tj-1",
      }),
    });
  });

  it("shows an inline error and stays on profile setup when completion fails", async () => {
    const onCompleteProfile = vi.fn().mockRejectedValue(new Error("terms required"));
    const onNavigate = vi.fn();
    render(
      <AuthProfileSetupScreen
        route={baseRoute("auth_profile_setup")}
        onNavigate={onNavigate}
        onCompleteProfile={onCompleteProfile}
      />,
    );

    const user = userEvent.setup();
    await user.type(screen.getByTestId("auth-profile-name"), "Alice");
    await user.click(screen.getByTestId("auth-profile-terms"));
    await user.click(screen.getByTestId("auth-profile-submit"));

    await expect(
      screen.findByTestId("auth-profile-status"),
    ).resolves.toHaveTextContent("failed");
    expect(onNavigate).not.toHaveBeenCalled();
  });

  it("does not restore the pending route until completeMyProfile confirms the profile flag is cleared", async () => {
    const onCompleteProfile = vi.fn().mockResolvedValue({
      profileCompletionRequired: true,
    });
    const onNavigate = vi.fn();
    render(
      <AuthProfileSetupScreen
        route={{
          name: "auth_profile_setup",
          params: {
            pendingRoute: "practice",
            pendingType: "start_practice",
            pendingLabel: "立即面试",
            planId: "plan-tj-1",
          },
        }}
        onNavigate={onNavigate}
        onCompleteProfile={onCompleteProfile}
      />,
    );

    const user = userEvent.setup();
    await user.type(screen.getByTestId("auth-profile-name"), "Alice");
    await user.click(screen.getByTestId("auth-profile-terms"));
    await user.click(screen.getByTestId("auth-profile-submit"));

    expect(onCompleteProfile).toHaveBeenCalledWith({
      displayName: "Alice",
      acceptedTerms: true,
    });
    await expect(
      screen.findByTestId("auth-profile-status"),
    ).resolves.toHaveTextContent("failed");
    expect(onNavigate).not.toHaveBeenCalled();
  });
});

describe("AuthVerifyScreen", () => {
  it("renders verify route shell, accepts six-digit code input, and triggers onVerify with token + onSuccess returnTo", async () => {
    const onVerify = vi.fn().mockResolvedValue({
      profileCompletionRequired: false,
    });
    const onNavigate = vi.fn();
    render(
      <AuthVerifyScreen
        route={{
          name: "auth_verify",
          params: { email: "alice@example.com", returnTo: "/practice" },
        }}
        onNavigate={onNavigate}
        onVerify={onVerify}
      />,
    );
    expect(screen.getByTestId("route-auth_verify")).toBeInTheDocument();
    expect(screen.getByTestId("auth-verify-code")).toBeInTheDocument();
    expect(screen.getByTestId("auth-verify-resend")).toBeInTheDocument();

    const user = userEvent.setup();
    const codeInput = screen.getByTestId("auth-verify-code");
    expect(codeInput).toHaveAttribute("inputmode", "numeric");
    expect(codeInput).toHaveAttribute("maxlength", "6");

    await user.type(codeInput, "65a4321");
    expect(codeInput).toHaveValue("654321");
    await user.click(screen.getByTestId("auth-verify-submit"));

    expect(onVerify).toHaveBeenCalledWith({ token: "654321" });
    await waitFor(() => expect(onNavigate).toHaveBeenCalled());
    const navArg = onNavigate.mock.calls[0]![0];
    expect(navArg).toMatchObject({ name: "practice" });
  });

  it("routes to profile setup before restoring pending action when /me requires completion", async () => {
    const onVerify = vi.fn().mockResolvedValue({
      profileCompletionRequired: true,
    });
    const onNavigate = vi.fn();
    render(
      <AuthVerifyScreen
        route={{
          name: "auth_verify",
          params: {
            email: "alice@example.com",
            pendingRoute: "practice",
            pendingType: "start_practice",
            pendingLabel: "立即面试",
            planId: "plan_1",
          },
        }}
        onNavigate={onNavigate}
        onVerify={onVerify}
      />,
    );
    const user = userEvent.setup();
    await user.type(screen.getByTestId("auth-verify-code"), "654321");
    await user.click(screen.getByTestId("auth-verify-submit"));

    await waitFor(() =>
      expect(onNavigate).toHaveBeenCalledWith({
        name: "auth_profile_setup",
        params: expect.objectContaining({
          pendingRoute: "practice",
          planId: "plan_1",
        }),
      }),
    );
  });

  it("restores interview context from raw returnTo query params", async () => {
    const onVerify = vi.fn().mockResolvedValue({
      profileCompletionRequired: false,
    });
    const onNavigate = vi.fn();
    render(
      <AuthVerifyScreen
        route={{
          name: "auth_verify",
          params: {
            email: "alice@example.com",
            returnTo:
              "/practice?planId=plan_1&targetJobId=tj_1&jdId=jd_1&resumeVersionId=resume_1&roundId=round_1",
          },
        }}
        onNavigate={onNavigate}
        onVerify={onVerify}
      />,
    );

    const user = userEvent.setup();
    await user.type(screen.getByTestId("auth-verify-code"), "654321");
    await user.click(screen.getByTestId("auth-verify-submit"));

    await waitFor(() =>
      expect(onNavigate).toHaveBeenCalledWith({
        name: "practice",
        params: {
          planId: "plan_1",
          targetJobId: "tj_1",
          jdId: "jd_1",
          resumeVersionId: "resume_1",
          roundId: "round_1",
        },
      }),
    );
  });
});

describe("AuthResetScreen", () => {
  it("renders the reset shell with email input, a stub send button, and back-to-login link without wiring any API", async () => {
    const onStartChallenge = vi.fn();
    const onNavigate = vi.fn();
    render(
      <AuthResetScreen
        route={baseRoute("auth_reset")}
        onNavigate={onNavigate}
        onStartChallenge={onStartChallenge}
      />,
    );
    expect(screen.getByTestId("route-auth_reset")).toBeInTheDocument();
    expect(screen.getByTestId("auth-reset-email")).toBeInTheDocument();
    expect(screen.getByTestId("auth-reset-send-stub")).toBeInTheDocument();
    expect(screen.getByTestId("auth-reset-link-login")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByTestId("auth-reset-send-stub"));
    // Reset is intentionally a UI shell; no API call is dispatched.
    expect(onStartChallenge).not.toHaveBeenCalled();

    await user.click(screen.getByTestId("auth-reset-link-login"));
    expect(onNavigate).toHaveBeenCalledWith({
      name: "auth_login",
      params: {},
    });
  });
});

describe("AuthLogoutScreen", () => {
  it("calls onLogout and routes home after success", async () => {
    const onLogout = vi.fn().mockResolvedValue(undefined);
    const onNavigate = vi.fn();
    render(
      <AuthLogoutScreen
        route={baseRoute("auth_logout")}
        onNavigate={onNavigate}
        onLogout={onLogout}
      />,
    );
    expect(screen.getByTestId("route-auth_logout")).toBeInTheDocument();
    expect(screen.getByTestId("auth-logout-confirm")).toBeInTheDocument();
    expect(screen.getByTestId("auth-logout-cancel")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByTestId("auth-logout-confirm"));
    expect(onLogout).toHaveBeenCalled();
    await waitFor(() =>
      expect(onNavigate).toHaveBeenCalledWith({
        name: "home",
        params: {},
      }),
    );
  });

  it("returns to home without calling onLogout when cancel is clicked", async () => {
    const onLogout = vi.fn();
    const onNavigate = vi.fn();
    render(
      <AuthLogoutScreen
        route={baseRoute("auth_logout")}
        onNavigate={onNavigate}
        onLogout={onLogout}
      />,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("auth-logout-cancel"));
    expect(onLogout).not.toHaveBeenCalled();
    expect(onNavigate).toHaveBeenCalledWith({ name: "home", params: {} });
  });
});
