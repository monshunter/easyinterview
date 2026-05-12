import { describe, expect, it } from "vitest";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import {
  mapResumeAssetToUiSource,
  mapResumeVersionToUi,
} from "./adapters/resume";

import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import listResumeVersionsFixture from "../../../../../openapi/fixtures/Resumes/listResumeVersions.json";
import getResumeVersionFixture from "../../../../../openapi/fixtures/Resumes/getResumeVersion.json";

const FIXTURES = [
  listResumesFixture,
  listResumeVersionsFixture,
  getResumeVersionFixture,
];

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario,
    }),
  });
}

const someAssetId = "01918fa0-0000-7000-8000-000000001000";
const otherAssetId = "deadbeef-0000-7000-8000-000000009999";

describe("Resumes fixture parity (counts derived from fixture body)", () => {
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

  it("listResumeVersions master-only scenario yields a single MASTER (structured_master) version", async () => {
    const client = buildClient("master-only");
    const result = await client.listResumeVersions(someAssetId);
    expect(result.items).toHaveLength(
      listResumeVersionsFixture.scenarios["master-only"].response.body.items
        .length,
    );
    expect(result.items.every((v) => v.versionType === "structured_master")).toBe(
      true,
    );
    const ui = result.items.map(mapResumeVersionToUi);
    expect(ui.every((v) => v.tag === "MASTER")).toBe(true);
  });

  it("listResumeVersions with-targeted-branches scenario yields TARGETED versions with parentVersionId set", async () => {
    const client = buildClient("with-targeted-branches");
    const result = await client.listResumeVersions(someAssetId);
    expect(result.items.length).toBeGreaterThan(0);
    expect(result.items.every((v) => v.versionType === "targeted")).toBe(true);
    expect(result.items.every((v) => v.parentVersionId !== null)).toBe(true);
    const ui = result.items.map(mapResumeVersionToUi);
    expect(ui.every((v) => v.tag === "TARGETED")).toBe(true);
  });

  it("mock transport is scenario-scoped, NOT request-aware: listResumeVersions returns the same scenario body regardless of resumeAssetId path param", async () => {
    const client = buildClient("default");
    const a = await client.listResumeVersions(someAssetId);
    const b = await client.listResumeVersions(otherAssetId);
    expect(a.items.length).toBe(b.items.length);
    expect(a.items.map((v) => v.id)).toEqual(b.items.map((v) => v.id));
  });

  it("listResumes adapter mapping preserves count derived from fixture body", async () => {
    const client = buildClient("default");
    const result = await client.listResumes();
    const ui = result.items.map(mapResumeAssetToUiSource);
    expect(ui.length).toBe(
      listResumesFixture.scenarios.default.response.body.items.length,
    );
  });
});
