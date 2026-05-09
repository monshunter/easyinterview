// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";

import getAgentScanStatusFixture from "../../../../../openapi/fixtures/JobMatch/getAgentScanStatus.json";
import getJobMatchProfileFixture from "../../../../../openapi/fixtures/JobMatch/getJobMatchProfile.json";

import { JDMatchScreen } from "./JDMatchScreen";

function buildClient(opts?: { profileScenario?: string; agentScenario?: string }) {
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url = typeof input === "string" ? input : (input as URL | Request).toString();
      const isAgent = url.includes("/jd-match/agent-status");
      const scenario = isAgent ? opts?.agentScenario : opts?.profileScenario;
      const inner = createFixtureBackedFetch(
        createFixtureRegistry([
          getJobMatchProfileFixture,
          getAgentScanStatusFixture,
        ]),
        scenario ? { scenario } : undefined,
      );
      return inner(input, init);
    },
  });
}

function renderJDMatch(client: EasyInterviewClient) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate }}>
            <JDMatchScreen route={{ name: "jd_match", params: {} }} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("JDMatchScreen data-driven Profile chip + AGENT badge (item 2.3)", () => {
  it("Profile chip avatar renders an <img> when avatarUrl is present", async () => {
    const client = buildClient({ profileScenario: "default", agentScenario: "default" });
    renderJDMatch(client);

    const avatar = await screen.findByTestId("jdmatch-profile-chip-avatar");
    const img = avatar.querySelector("img");
    expect(img).not.toBeNull();
    expect(img?.getAttribute("src")).toBe("https://avatar.example.com/alice.png");
  });

  it("Profile chip avatar renders initials when avatarUrl is missing", async () => {
    const client = buildClient({
      profileScenario: "partial-profile",
      agentScenario: "default",
    });
    renderJDMatch(client);

    const avatar = await screen.findByTestId("jdmatch-profile-chip-avatar");
    expect(avatar.querySelector("img")).toBeNull();
    // displayName "Alice Example" → initials "AE"
    expect(avatar.textContent ?? "").toMatch(/AE/);
  });

  it("Profile chip skills row renders one tag per skill from profile.skills", async () => {
    const client = buildClient({ profileScenario: "default", agentScenario: "default" });
    renderJDMatch(client);

    const skills = await screen.findByTestId("jdmatch-profile-chip-skills");
    const tags = skills.querySelectorAll("[data-testid^='jdmatch-profile-chip-skill-']");
    expect(tags.length).toBeGreaterThanOrEqual(6);
    expect(skills.textContent ?? "").toContain("React");
    expect(skills.textContent ?? "").toContain("TypeScript");
  });

  it("Profile chip sources renders aggregated counts from profile.sources", async () => {
    const client = buildClient({ profileScenario: "default", agentScenario: "default" });
    renderJDMatch(client);

    const sources = await screen.findByTestId("jdmatch-profile-chip-sources");
    const text = sources.textContent ?? "";
    // default fixture: resumes=2 jds=5 mocks=4 debriefs=1
    expect(text).toContain("2");
    expect(text).toContain("5");
    expect(text).toContain("4");
    expect(text).toContain("1");
  });

  it("Profile chip searching-as row reflects displayName from profile data", async () => {
    const client = buildClient({ profileScenario: "default", agentScenario: "default" });
    renderJDMatch(client);

    const row = await screen.findByTestId("jdmatch-profile-chip-searching-as");
    expect(row.textContent ?? "").toContain("Alice Example");
  });

  it("AGENT badge in idle state renders neutral tone, last-scan, next-scan", async () => {
    const client = buildClient({ profileScenario: "default", agentScenario: "idle" });
    renderJDMatch(client);

    const badge = await screen.findByTestId("jdmatch-agent-status-badge");
    await waitFor(() => {
      expect(badge.getAttribute("data-tone")).toBe("idle");
    });
    expect(screen.getByTestId("jdmatch-agent-status-last-scan")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-agent-status-next-scan")).toBeInTheDocument();
  });

  it("AGENT badge in scanning state renders accent tone and hides next-scan", async () => {
    const client = buildClient({ profileScenario: "default", agentScenario: "scanning" });
    renderJDMatch(client);

    const badge = await screen.findByTestId("jdmatch-agent-status-badge");
    await waitFor(() => {
      expect(badge.getAttribute("data-tone")).toBe("scanning");
    });
    expect(screen.queryByTestId("jdmatch-agent-status-next-scan")).toBeNull();
  });

  it("AGENT badge in error state renders warn tone and surfaces message", async () => {
    const client = buildClient({ profileScenario: "default", agentScenario: "error" });
    renderJDMatch(client);

    const badge = await screen.findByTestId("jdmatch-agent-status-badge");
    await waitFor(() => {
      expect(badge.getAttribute("data-tone")).toBe("error");
    });
    // error-variant message present in body
    expect(badge.textContent ?? "").toMatch(/error|失败|异常|fail/i);
  });

  it("renders fallback content when profile data is missing (no AppRuntimeProvider)", () => {
    render(
      <DisplayPreferencesProvider>
        <NavigationProvider value={{ navigate: vi.fn() }}>
          <JDMatchScreen route={{ name: "jd_match", params: {} }} />
        </NavigationProvider>
      </DisplayPreferencesProvider>,
    );

    // Hero + chip + tabs still render even without data
    expect(screen.getByTestId("jdmatch-hero-title")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-profile-chip")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-tab-recommended")).toBeInTheDocument();
  });

  it("does not render plan 001 placeholder testids", async () => {
    const client = buildClient({ profileScenario: "default", agentScenario: "default" });
    renderJDMatch(client);

    expect(screen.queryByTestId("jdmatch-placeholder")).toBeNull();
    expect(screen.queryByTestId("jdmatch-placeholder-cta")).toBeNull();
  });
});
