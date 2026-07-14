// @vitest-environment jsdom
/**
 * Code-level default Home shell regression.
 *
 * Truth source: docs/spec/frontend-shell/plans/001-app-shell-auth-settings/bdd-plan.md
 *               + bdd-checklist.md.
 *
 * Given a user without any saved session or saved route, opening the App must
 * render Home, the three primary nav entries (D-22), the single login entry, and the global
 * display controls. Welcome, standalone voice, Debrief, User Profile, and the out-of-scope
 * Growth / Mistakes / Drill modules must NOT be reachable.
 */
import { describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import getMeFixture from "../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import { EasyInterviewClient } from "../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../api/mockTransport";
import { App } from "../App";

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getRuntimeConfigFixture, getMeFixture]),
    ),
  });
}

describe("default home shell", () => {
  it("renders Home + three primary nav + login + display controls without out-of-scope entries", async () => {
    const client = buildClient();
    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );

    expect(screen.getByTestId("route-home")).toBeInTheDocument();

    const primaryNav = screen.getByTestId("topbar-primary-nav");
    expect(primaryNav).toBeInTheDocument();
    for (const name of [
      "home",
      "workspace",
      "resume_versions",
    ]) {
      expect(screen.getByTestId(`topbar-nav-${name}`)).toBeInTheDocument();
    }
    expect(
      primaryNav.querySelectorAll("button[data-testid^='topbar-nav-']"),
    ).toHaveLength(3);

    expect(screen.getByTestId("topbar-theme-button")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-dark-toggle")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-lang-toggle")).toBeInTheDocument();

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "false",
      ),
    );
    expect(screen.getByTestId("topbar-login")).toBeInTheDocument();
    expect(screen.queryByTestId("topbar-register")).not.toBeInTheDocument();

    expect(screen.queryByTestId("route-welcome")).not.toBeInTheDocument();
    for (const outOfScope of ["mistakes", "growth", "voice", "drill", "welcome", "debrief", "profile"]) {
      expect(
        screen.queryByTestId(`topbar-nav-${outOfScope}`),
      ).not.toBeInTheDocument();
    }
  });
});
