// @vitest-environment jsdom
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";
import { render } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../display/DisplayPreferencesProvider";
import {
  AuthLoginScreen,
  AuthLogoutScreen,
  AuthProfileSetupScreen,
  AuthVerifyScreen,
} from "./index";
import type { Route } from "../routes";

const HERE = resolve(__dirname);
const REPO_ROOT = resolve(HERE, "..", "..", "..", "..");
const AUTH_CSS = resolve(HERE, "auth.css");
const AUTH_SCREEN_JSX = resolve(REPO_ROOT, "ui-design", "src", "screen-auth.jsx");

const noop = () => {};

function withProvider(node: React.ReactElement) {
  return <DisplayPreferencesProvider>{node}</DisplayPreferencesProvider>;
}

interface AuthScreenSpec {
  name: string;
  routeName: Route["name"];
  render: () => React.ReactElement;
}

const SCREENS: AuthScreenSpec[] = [
  {
    name: "AuthLoginScreen",
    routeName: "auth_login",
    render: () => (
      <AuthLoginScreen
        route={{ name: "auth_login", params: {} } as Route}
        onNavigate={noop}
        onStartChallenge={async () => {}}
      />
    ),
  },
  {
    name: "AuthProfileSetupScreen",
    routeName: "auth_profile_setup",
    render: () => (
      <AuthProfileSetupScreen
        route={{ name: "auth_profile_setup", params: {} } as Route}
        onNavigate={noop}
        onCompleteProfile={async () => ({ profileCompletionRequired: false })}
      />
    ),
  },
  {
    name: "AuthVerifyScreen",
    routeName: "auth_verify",
    render: () => (
      <AuthVerifyScreen
        route={{ name: "auth_verify", params: { email: "u@example.com" } } as Route}
        onNavigate={noop}
        onVerify={async () => ({ profileCompletionRequired: false })}
      />
    ),
  },
  {
    name: "AuthLogoutScreen",
    routeName: "auth_logout",
    render: () => (
      <AuthLogoutScreen
        route={{ name: "auth_logout", params: {} } as Route}
        onNavigate={noop}
        onLogout={async () => {}}
      />
    ),
  },
];

describe("auth screen card visual contract (Phase 4.1)", () => {
  it.each(SCREENS)("$name renders the ei-auth-shell two-column layout", ({ render: renderScreen, routeName }) => {
    const { container } = render(withProvider(renderScreen()));
    const root = container.querySelector(`[data-testid="route-${routeName}"]`);
    expect(root, `route-${routeName} root missing`).toBeTruthy();
    expect(root!.className).toMatch(/\bei-auth-shell\b/);

    const side = root!.querySelector(".ei-auth-side");
    expect(side).toBeTruthy();
    const card = root!.querySelector(".ei-auth-card");
    expect(card).toBeTruthy();
  });

  it.each(SCREENS)("$name surfaces a serif title via ei-text-display + an eyebrow via ei-text-label", ({ render: renderScreen }) => {
    const { container } = render(withProvider(renderScreen()));
    const eyebrow = container.querySelector(".ei-auth-eyebrow");
    expect(eyebrow).toBeTruthy();
    expect(eyebrow!.className).toMatch(/\bei-text-label\b/);

    const heading = container.querySelector("h1");
    expect(heading).toBeTruthy();
    expect(heading!.className).toMatch(/\bei-text-display\b/);
  });

  it("auth.css transcribes the screen-auth.jsx AuthShell rhythm (max-width 1160, padding, gap, grid 0.88fr 1.12fr)", () => {
    const css = readFileSync(AUTH_CSS, "utf8");
    expect(css).toMatch(/\.ei-auth-shell\s*\{[^}]*max-width:\s*1160px/);
    expect(css).toMatch(/\.ei-auth-shell\s*\{[^}]*padding:\s*54px\s+48px\s+96px/);
    expect(css).toMatch(/\.ei-auth-shell\s*\{[^}]*display:\s*grid/);
    expect(css).toMatch(/\.ei-auth-shell\s*\{[^}]*grid-template-columns:\s*0\.88fr\s+1\.12fr/);
    expect(css).toMatch(/\.ei-auth-shell\s*\{[^}]*gap:\s*44px/);
    expect(css).toMatch(
      /\.ei-auth-card\s*\{[^}]*background:\s*var\(--ei-color-bg-card\)/,
    );
    expect(css).toMatch(
      /\.ei-auth-card\s*\{[^}]*border:\s*1px solid var\(--ei-color-rule-strong\)/,
    );
    expect(css).toMatch(/\.ei-auth-side-panel\s*\{[^}]*background:\s*var\(--ei-color-bg-soft\)/);
  });

  it("auth.css values can be traced back to ui-design/src/screen-auth.jsx", () => {
    const src = readFileSync(AUTH_SCREEN_JSX, "utf8");
    expect(src).toContain("maxWidth: 1160");
    expect(src).toContain('padding: "54px 48px 96px"');
    expect(src).toContain('gridTemplateColumns: "0.88fr 1.12fr"');
    expect(src).toContain("gap: 44");
    expect(src).toContain("padding: 18");
    expect(src).toContain("padding: 28");
  });

  it("global.css imports the auth.css module", () => {
    const global = readFileSync(
      resolve(HERE, "..", "theme", "global.css"),
      "utf8",
    );
    expect(global).toMatch(/@import\s+["']\.\.\/auth\/auth\.css["'];/);
  });
});

describe("auth screen D1 regression after visual parity (Phase 4.1)", () => {
  it("AuthLoginScreen retains the single email form testids", () => {
    const { container } = render(
      withProvider(
        <AuthLoginScreen
          route={{ name: "auth_login", params: {} } as Route}
          onNavigate={noop}
          onStartChallenge={async () => {}}
        />,
      ),
    );
    expect(container.querySelector("[data-testid='auth-login-email-form']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-login-email']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-login-submit-email']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-login-link-register']")).toBeFalsy();
    expect(container.querySelector("[data-testid='auth-login-link-reset']")).toBeFalsy();
  });

  it("AuthProfileSetupScreen retains form / submit testids", () => {
    const { container } = render(
      withProvider(
        <AuthProfileSetupScreen
          route={{ name: "auth_profile_setup", params: {} } as Route}
          onNavigate={noop}
          onCompleteProfile={async () => ({ profileCompletionRequired: false })}
        />,
      ),
    );
    expect(container.querySelector("[data-testid='auth-profile-form']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-profile-name']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-profile-terms']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-profile-submit']")).toBeTruthy();
  });

  it("AuthVerifyScreen retains code form testids", () => {
    const { container } = render(
      withProvider(
        <AuthVerifyScreen
          route={{ name: "auth_verify", params: { email: "u@example.com" } } as Route}
          onNavigate={noop}
          onVerify={async () => ({ profileCompletionRequired: false })}
        />,
      ),
    );
    expect(container.querySelector("[data-testid='auth-verify-form']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-verify-code']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-verify-submit']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-verify-email-hint']")).toBeTruthy();
  });

  it("AuthLoginScreen renders the static email-code help copy (D-16)", () => {
    const { container } = render(
      withProvider(
        <AuthLoginScreen
          route={{ name: "auth_login", params: {} } as Route}
          onNavigate={noop}
          onStartChallenge={async () => {}}
        />,
      ),
    );
    expect(container.querySelector("[data-testid='auth-login-help']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-login-link-reset']")).toBeNull();
  });

  it("AuthLogoutScreen retains confirm/cancel testids", () => {
    const { container } = render(
      withProvider(
        <AuthLogoutScreen
          route={{ name: "auth_logout", params: {} } as Route}
          onNavigate={noop}
          onLogout={async () => {}}
        />,
      ),
    );
    expect(container.querySelector("[data-testid='auth-logout-data-hint']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-logout-confirm']")).toBeTruthy();
    expect(container.querySelector("[data-testid='auth-logout-cancel']")).toBeTruthy();
  });
});
