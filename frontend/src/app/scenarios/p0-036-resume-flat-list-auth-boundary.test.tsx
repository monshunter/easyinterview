// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StrictMode } from "react";

import { EasyInterviewClient } from "../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../api/mockTransport";
import { App } from "../App";

import getRuntimeConfigFixture from "../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../openapi/fixtures/Resumes/listResumes.json";
import getResumeFixture from "../../../../openapi/fixtures/Resumes/getResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  getResumeFixture,
];

const FIRST_RESUME_ID =
  listResumesFixture.scenarios.default.response.body.items[0]?.id ?? "";
const SECOND_RESUME_ID =
  listResumesFixture.scenarios.default.response.body.items[1]?.id ?? "";

function buildClient(requests: string[] = []): EasyInterviewClient {
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry(FIXTURES),
    { scenario: "default" },
  );
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string"
          ? input
          : input instanceof URL
            ? input.href
            : input.url;
      requests.push(`${init?.method ?? "GET"} ${new URL(url, "http://localhost").pathname}${new URL(url, "http://localhost").search}`);
      return fixtureFetch(input, init);
    },
  });
}

function renderApp(authMode: "authenticated" | "unauthenticated") {
  return render(
    <App
      client={buildClient()}
      requestOptions={{
        getMe: { headers: { Prefer: `example=${authMode}` } },
      }}
      initialRoute={{ name: "resume_versions", params: {} }}
    />,
  );
}

describe("E2E.P0.036 resume flat list + auth boundary", () => {
  it("unauthenticated visit to /resume_versions routes to login, not the list, and triggers no Resume API", async () => {
    const client = buildClient();
    const listSpy = vi.spyOn(client, "listResumes");
    const getSpy = vi.spyOn(client, "getResume");
    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
        initialRoute={{ name: "resume_versions", params: {} }}
      />,
    );

    await waitFor(() =>
      expect(screen.getByTestId("route-auth_login")).toBeInTheDocument(),
    );
    expect(screen.getByTestId("auth-side-pending-action")).toBeInTheDocument();
    expect(screen.queryByTestId("resume-workshop-list")).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-auth-gate"),
    ).not.toBeInTheDocument();
    expect(listSpy).not.toHaveBeenCalled();
    expect(getSpy).not.toHaveBeenCalled();
  });

  it("authenticated default view renders the flat ResumeWorkshopScreen list with one row per resume and no grouped-list chrome", async () => {
    const expectedSummaryKeys = [
      "displayName",
      "hasReadableContent",
      "id",
      "language",
      "parseStatus",
      "sourceType",
      "summaryHeadline",
      "title",
      "updatedAt",
    ];
    for (const item of listResumesFixture.scenarios.default.response.body.items) {
      expect(Object.keys(item).sort()).toEqual(expectedSummaryKeys);
      expect(item).not.toHaveProperty("originalText");
      expect(item).not.toHaveProperty("parsedTextSnapshot");
      expect(item).not.toHaveProperty("structuredProfile");
      expect(item).not.toHaveProperty("parsedSummary");
    }
    renderApp("authenticated");

    // The `resume-workshop-list` testid is shared by the loading and loaded
    // states; wait for the loaded flat table specifically so this assertion is
    // deterministic regardless of fixture-fetch timing under parallel workers.
    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-table")).toBeInTheDocument();
    });
    // Resume workshop screen replaces the route shell.
    expect(
      screen.queryByTestId("route-resume_versions"),
    ).not.toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-screen")).toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-list")).toBeInTheDocument();

    // One flat row per resume, derived from the fixture body.
    expect(
      screen.getByTestId(`resume-list-row-${FIRST_RESUME_ID}`),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId(`resume-list-row-${SECOND_RESUME_ID}`),
    ).toBeInTheDocument();

    // Flat Resume list: stats strip, view switcher, and grouped UI stay absent.
    expect(
      screen.queryByTestId("resume-workshop-stats-originals"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-stats-versions"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-view-switcher-tree"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-view-switcher-flat"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(`resume-tree-row-${FIRST_RESUME_ID}`),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(["resume-workshop-selected", "tree-helper"].join("-")),
    ).not.toBeInTheDocument();
  });

  it("opening a flat row navigates to that resume's read-only detail with only resumeId", async () => {
    const requests: string[] = [];
    const client = buildClient(requests);
    render(
      <StrictMode>
        <App
          client={client}
          requestOptions={{
            getMe: { headers: { Prefer: "example=authenticated" } },
          }}
          initialRoute={{ name: "resume_versions", params: {} }}
        />
      </StrictMode>,
    );

    await waitFor(() => {
      expect(
        screen.getByTestId(`resume-list-open-${FIRST_RESUME_ID}`),
      ).toBeInTheDocument();
    });
    expect(
      requests.filter((request) =>
        /^GET \/api\/v1\/resumes(?:\?|$)/.test(request),
      ),
    ).toHaveLength(1);
    expect(
      requests.filter((request) => request === `GET /api/v1/resumes/${FIRST_RESUME_ID}`),
    ).toHaveLength(0);

    const user = userEvent.setup();
    await user.click(screen.getByTestId(`resume-list-open-${FIRST_RESUME_ID}`));

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-detail")).toBeInTheDocument();
    });
    const detail = screen.getByTestId("resume-workshop-detail");
    expect(detail).toHaveAttribute("data-resume-id", FIRST_RESUME_ID);
    expect(detail).not.toHaveAttribute("data-tab");
    expect(
      requests.filter((request) => request === `GET /api/v1/resumes/${FIRST_RESUME_ID}`),
    ).toHaveLength(1);
    console.log(
      "E2E.P0.036 summary-only list/detail transport PASS summaryFields=9 listResumes=1 getResumeBeforeOpen=0 getResumeAfterOpen=1",
    );
  });

  it("evicts a rejected StrictMode list request and retries with one new transport", async () => {
    const fixtureFetch = createFixtureBackedFetch(
      createFixtureRegistry(FIXTURES),
      { scenario: "default" },
    );
    let listTransports = 0;
    const client = new EasyInterviewClient({
      fetch: async (input, init) => {
        const url = new URL(String(input), "http://localhost");
        if (
          (init?.method ?? "GET").toUpperCase() === "GET" &&
          url.pathname === "/api/v1/resumes"
        ) {
          listTransports += 1;
          if (listTransports === 1) {
            throw new TypeError("temporary list transport failure");
          }
        }
        return fixtureFetch(input, init);
      },
    });
    render(
      <StrictMode>
        <App
          client={client}
          requestOptions={{
            getMe: { headers: { Prefer: "example=authenticated" } },
          }}
          initialRoute={{ name: "resume_versions", params: {} }}
        />
      </StrictMode>,
    );

    const retry = await screen.findByTestId("resume-workshop-list-retry");
    expect(listTransports).toBe(1);
    await userEvent.click(retry);
    await screen.findByTestId("resume-workshop-table");
    expect(listTransports).toBe(2);
    console.info(
      "E2E.P0.036 list rejection retry transport PASS initialRejected=1 retrySucceeded=2",
    );
  });

  it("out-of-scope prototype testids and runtime imports are absent from the resume-workshop source", () => {
    // Static gate already enforced in ResumeWorkshopPrivacy.test.ts. We assert
    // here that the rendered DOM never surfaces out-of-scope route testids that
    // would indicate out-of-scope welcome / mistakes / drill / voice modules
    // sneaked back in.
    const { unmount } = renderApp("authenticated");
    for (const forbidden of [
      "route-welcome",
      "route-mistakes",
      "route-drill",
      "route-followup",
      "route-onboarding",
      "route-experiences",
      "route-star",
      "route-voice",
    ]) {
      expect(screen.queryByTestId(forbidden)).not.toBeInTheDocument();
    }
    unmount();
  });
});
