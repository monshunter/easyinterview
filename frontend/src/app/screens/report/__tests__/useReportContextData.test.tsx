/**
 * @vitest-environment jsdom
 *
 * Phase 2.9 — useReportContextData: pulls target + resume labels through the
 * generated client, falls back to IDs on failure, and never carries write
 * headers (read-only contract).
 */

import { describe, expect, it, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import type {
  Resume,
  TargetJob,
} from "../../../../api/generated/types";
import { EasyInterviewClient } from "../../../../api/generated/client";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import { useReportContextData } from "../hooks/useReportContextData";

const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const RESUME_ID = "01918fa0-0000-7000-8000-000000004000";

interface ClientOpts {
  targetJob?: TargetJob | { reject: unknown };
  resume?: Resume | { reject: unknown };
}

function makeClient(opts: ClientOpts = {}): EasyInterviewClient {
  const targetFn = vi.fn(async () => {
    const next = opts.targetJob;
    if (!next) {
      return {
        id: TARGET_JOB_ID,
        title: "Senior Frontend Engineer",
        companyName: "Acme",
      } as unknown as TargetJob;
    }
    if ("reject" in next) throw next.reject;
    return next;
  });
  const resumeFn = vi.fn(async () => {
    const next = opts.resume;
    if (!next) {
      return {
        id: RESUME_ID,
        displayName: "Resume v3",
      } as unknown as Resume;
    }
    if ("reject" in next) throw next.reject;
    return next;
  });
  return {
    async getRuntimeConfig() {
      return { aiProviderProfile: "stub" } as never;
    },
    async getMe() {
      throw new Error("HTTP 401 Unauthorized");
    },
    getTargetJob: targetFn,
    getResume: resumeFn,
  } as unknown as EasyInterviewClient;
}

function Wrapper({
  client,
  children,
}: {
  client: EasyInterviewClient;
  children: ReactNode;
}) {
  return <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>;
}

describe("useReportContextData", () => {
  it("loads target + resume labels through generated client (TestReportContextDataLoadsTargetJobAndResumeVersion)", async () => {
    const client = makeClient();
    const { result } = renderHook(
      () =>
        useReportContextData({
          targetJobId: TARGET_JOB_ID,
          resumeId: RESUME_ID,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.targetLabel).toBe("Acme · Senior Frontend Engineer");
    expect(result.current.resumeLabel).toBe("Resume v3");
  });

  it("falls back to the id label when one operation fails (TestReportContextDataFallsBackToIds)", async () => {
    const client = makeClient({
      targetJob: { reject: new Error("HTTP 500 Internal") },
    });
    const { result } = renderHook(
      () =>
        useReportContextData({
          targetJobId: TARGET_JOB_ID,
          resumeId: RESUME_ID,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.targetLabel).toBe(TARGET_JOB_ID);
    expect(result.current.resumeLabel).toBe("Resume v3");
  });

  it("does not read raw resume body or JD body fields (TestReportContextDataDoesNotReadRawBody)", async () => {
    const sensitiveResume = {
      id: RESUME_ID,
      displayName: "Resume v3",
      originalText: "PRIVATE: do-not-leak",
      parsedTextSnapshot: "PRIVATE: snapshot",
    } as unknown as Resume;
    const sensitiveJob = {
      id: TARGET_JOB_ID,
      title: "Senior Frontend Engineer",
      companyName: "Acme",
      jdText: "PRIVATE: JD body",
    } as unknown as TargetJob;
    const client = makeClient({
      targetJob: sensitiveJob,
      resume: sensitiveResume,
    });
    const { result } = renderHook(
      () =>
        useReportContextData({
          targetJobId: TARGET_JOB_ID,
          resumeId: RESUME_ID,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.targetLabel ?? "").not.toContain("PRIVATE");
    expect(result.current.resumeLabel ?? "").not.toContain("PRIVATE");
  });
});
