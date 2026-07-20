import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const CSS_PATH = resolve(__dirname, "..", "..", "screens.css");
const FLOW_PATH = resolve(__dirname, "ResumeCreateFlow.tsx");
const UPLOAD_PATH = resolve(__dirname, "UploadTab.tsx");

describe("Resume create reference visual contract", () => {
  it("uses the wide create canvas and supplied desktop rhythm", () => {
    const css = readFileSync(CSS_PATH, "utf8");
    expect(css).toMatch(/\.ei-screen-shell\[data-testid="resume-workshop-screen"\]\[data-flow="create"\]\s*\{[^}]*max-width:\s*1470px/s);
    expect(css).toMatch(/\.ei-resume-create-flow\s*\{[^}]*min-height:\s*calc\(100vh - 76px\)/s);
    expect(css).toMatch(/\.ei-resume-create-card\s*\{[^}]*border-radius:\s*12px[^}]*box-shadow:/s);
    expect(css).toMatch(/\.ei-resume-create-upload-dropzone\s*\{[^}]*min-height:\s*366px[^}]*border-radius:\s*10px/s);
    expect(css).toMatch(/@media\s*\(max-width:\s*720px\)[\s\S]*\.ei-resume-create-header-art\s*\{[^}]*display:\s*none/s);
  });

  it("renders decorative header art and three upload capability labels", () => {
    const flow = readFileSync(FLOW_PATH, "utf8");
    const upload = readFileSync(UPLOAD_PATH, "utf8");
    expect(flow).toContain("ei-resume-create-header-art");
    expect(upload).toContain("resume-create-upload-capabilities");
    expect(upload.match(/className="ei-resume-create-upload-capability"/g)).toHaveLength(3);
  });

  it("removes the retired central upload icon and styles the drag-active state", () => {
    const css = readFileSync(CSS_PATH, "utf8");
    const upload = readFileSync(UPLOAD_PATH, "utf8");

    expect(css).not.toContain(".ei-resume-create-upload-icon");
    expect(upload).not.toContain("ei-resume-create-upload-icon");
    expect(css).toMatch(/\.ei-resume-create-upload-dropzone\[data-drag-active="true"\]\s*\{/);
  });
});
