import { describe, expect, it } from "vitest";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { mapResumeToUiSource } from "./adapters/resume";

import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import getResumeFixture from "../../../../../openapi/fixtures/Resumes/getResume.json";

const FIXTURES = [listResumesFixture, getResumeFixture];

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario,
    }),
  });
}

const knownResumeId = "01918fa0-0000-7000-8000-000000001000";
const otherResumeId = "deadbeef-0000-7000-8000-000000009999";

describe("Resumes fixture parity (flat D-20 model, counts derived from fixture body)", () => {
  it("listResumes default scenario count matches fixture items length", async () => {
    const client = buildClient("default");
    const result = await client.listResumes();
    expect(result.items).toHaveLength(
      listResumesFixture.scenarios.default.response.body.items.length,
    );
  });

  it("listResumes empty scenario yields zero items and hasMore=false", async () => {
    const client = buildClient("empty");
    const result = await client.listResumes();
    expect(result.items).toEqual([]);
    expect(result.pageInfo.hasMore).toBe(false);
  });

  it("listResumes paginated scenario yields hasMore=true and a non-empty cursor", async () => {
    const client = buildClient("paginated");
    const result = await client.listResumes();
    expect(result.items.length).toBeGreaterThan(0);
    expect(result.pageInfo.hasMore).toBe(true);
    expect(result.pageInfo.nextCursor).toBeTruthy();
  });

  it("listResumes default scenario items are all flat resumes without tree fields", async () => {
    const client = buildClient("default");
    const result = await client.listResumes();
    expect(result.items.every((r) => typeof r.id === "string")).toBe(true);
    expect(
      result.items.every(
        (r) => !("resumeAssetId" in r) && !("versionType" in r),
      ),
    ).toBe(true);
  });

  it("getResume default scenario returns the flat resume body verbatim", async () => {
    const client = buildClient("default");
    const result = await client.getResume(knownResumeId);
    const { body } = getResumeFixture.scenarios.default.response;
    expect(result.id).toBe(body.id);
    expect(result.displayName).toBe(body.displayName);
    expect(result.parseStatus).toBe(body.parseStatus);
  });

  it("mock transport is scenario-scoped, NOT request-aware: getResume returns the same scenario body regardless of resumeId path param", async () => {
    const client = buildClient("default");
    const a = await client.getResume(knownResumeId);
    const b = await client.getResume(otherResumeId);
    expect(a.id).toBe(b.id);
    expect(a.displayName).toBe(b.displayName);
  });

  it("listResumes adapter mapping preserves count derived from fixture body", async () => {
    const client = buildClient("default");
    const result = await client.listResumes();
    const ui = result.items.map(mapResumeToUiSource);
    expect(ui.length).toBe(
      listResumesFixture.scenarios.default.response.body.items.length,
    );
  });
});
