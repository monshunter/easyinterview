// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  AuthLoginScreen,
  AuthLogoutScreen,
  AuthRegisterScreen,
  AuthResetScreen,
  AuthVerifyScreen,
} from "./index";
import type { Route } from "../routes";

const baseRoute = (name: Route["name"]): Route => ({ name, params: {} });

describe("AuthLoginScreen", () => {
  it("renders the login route shell with email + password / OAuth stubs and inline links", () => {
    render(
      <AuthLoginScreen
        route={baseRoute("auth_login")}
        onNavigate={() => {}}
        onStartChallenge={async () => {}}
      />,
    );
    expect(screen.getByTestId("route-auth_login")).toBeInTheDocument();
    expect(screen.getByTestId("auth-login-email")).toBeInTheDocument();
    // Password is rendered as a stub-only input (no real submit wire); presence
    // is enforced so 4.x can later assert it does not call any new API.
    expect(screen.getByTestId("auth-login-password-stub")).toBeInTheDocument();
    expect(screen.getByTestId("auth-login-oauth-stub")).toBeInTheDocument();
    expect(
      screen.getByTestId("auth-login-link-register"),
    ).toBeInTheDocument();
    expect(screen.getByTestId("auth-login-link-reset")).toBeInTheDocument();
  });

  it("submits the email challenge through onStartChallenge with returnTo from pending action when provided", async () => {
    const onStartChallenge = vi.fn().mockResolvedValue(undefined);
    const onNavigate = vi.fn();
    render(
      <AuthLoginScreen
        route={{ name: "auth_login", params: { returnTo: "/practice" } }}
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
      purpose: "login",
      returnTo: "/practice",
    });
    await waitFor(() =>
      expect(onNavigate).toHaveBeenCalledWith({
        name: "auth_verify",
        params: expect.objectContaining({ email: "alice@example.com" }),
      }),
    );
  });

  it("navigates to register / reset routes from inline links", async () => {
    const onNavigate = vi.fn();
    render(
      <AuthLoginScreen
        route={baseRoute("auth_login")}
        onNavigate={onNavigate}
        onStartChallenge={async () => {}}
      />,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("auth-login-link-register"));
    await user.click(screen.getByTestId("auth-login-link-reset"));
    expect(onNavigate).toHaveBeenNthCalledWith(1, {
      name: "auth_register",
      params: {},
    });
    expect(onNavigate).toHaveBeenNthCalledWith(2, {
      name: "auth_reset",
      params: {},
    });
  });

  it("preserves pending action params when switching from login to register", async () => {
    const onNavigate = vi.fn();
    render(
      <AuthLoginScreen
        route={{
          name: "auth_login",
          params: {
            pendingRoute: "practice",
            pendingType: "start_practice",
            pendingLabel: "立即面试",
            planId: "plan-tj-1",
          },
        }}
        onNavigate={onNavigate}
        onStartChallenge={async () => {}}
      />,
    );

    const user = userEvent.setup();
    await user.click(screen.getByTestId("auth-login-link-register"));

    expect(onNavigate).toHaveBeenCalledWith({
      name: "auth_register",
      params: expect.objectContaining({
        pendingRoute: "practice",
        pendingType: "start_practice",
        pendingLabel: "立即面试",
        planId: "plan-tj-1",
      }),
    });
  });
});

describe("AuthRegisterScreen", () => {
  it("renders register form fields and routes to verify on submit (passwordless wire)", async () => {
    const onStartChallenge = vi.fn().mockResolvedValue(undefined);
    const onNavigate = vi.fn();
    render(
      <AuthRegisterScreen
        route={baseRoute("auth_register")}
        onNavigate={onNavigate}
        onStartChallenge={onStartChallenge}
      />,
    );
    expect(screen.getByTestId("route-auth_register")).toBeInTheDocument();
    expect(screen.getByTestId("auth-register-name")).toBeInTheDocument();
    expect(screen.getByTestId("auth-register-email")).toBeInTheDocument();
    expect(
      screen.getByTestId("auth-register-password-stub"),
    ).toBeInTheDocument();
    expect(screen.getByTestId("auth-register-terms")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.type(screen.getByTestId("auth-register-name"), "Alice");
    await user.type(
      screen.getByTestId("auth-register-email"),
      "alice@example.com",
    );
    await user.click(screen.getByTestId("auth-register-terms"));
    await user.click(screen.getByTestId("auth-register-submit"));

    expect(onStartChallenge).toHaveBeenCalledWith({
      email: "alice@example.com",
      purpose: "signup",
      displayName: "Alice",
    });
    await waitFor(() =>
      expect(onNavigate).toHaveBeenCalledWith({
        name: "auth_verify",
        params: expect.objectContaining({ email: "alice@example.com" }),
      }),
    );
  });

  it("shows an inline error and stays on register when signup email already exists", async () => {
    const onStartChallenge = vi
      .fn()
      .mockRejectedValue(new Error("email registered"));
    const onNavigate = vi.fn();
    render(
      <AuthRegisterScreen
        route={baseRoute("auth_register")}
        onNavigate={onNavigate}
        onStartChallenge={onStartChallenge}
      />,
    );

    const user = userEvent.setup();
    await user.type(screen.getByTestId("auth-register-name"), "Alice");
    await user.type(
      screen.getByTestId("auth-register-email"),
      "alice@example.com",
    );
    await user.click(screen.getByTestId("auth-register-terms"));
    await user.click(screen.getByTestId("auth-register-submit"));

    await expect(
      screen.findByTestId("auth-register-status"),
    ).resolves.toHaveTextContent("already registered");
    expect(onNavigate).not.toHaveBeenCalled();
  });
});

describe("AuthVerifyScreen", () => {
  it("renders verify route shell, accepts six-digit code input, and triggers onVerify with token + onSuccess returnTo", async () => {
    const onVerify = vi.fn().mockResolvedValue(undefined);
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

  it("restores interview context from raw returnTo query params", async () => {
    const onVerify = vi.fn().mockResolvedValue(undefined);
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
