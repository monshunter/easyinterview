import type { FC } from "react";

interface ReportPageIllustrationProps {
  testId: string;
}

/** Decorative geometry shared by the report index and readonly transcript. */
export const ReportPageIllustration: FC<ReportPageIllustrationProps> = ({
  testId,
}) => (
  <svg
    aria-hidden="true"
    className="ei-report-page-illustration"
    data-testid={testId}
    viewBox="0 0 300 150"
    fill="none"
  >
    <path className="ei-report-page-illustration-plane" d="M38 129 133 26l130 104Z" />
    <path className="ei-report-page-illustration-plane" d="m91 128 69-71 84 71Z" />
    <rect x="132" y="22" width="112" height="22" rx="10" fill="currentColor" opacity=".23" />
    <rect
      className="ei-report-page-illustration-card"
      x="121"
      y="31"
      width="112"
      height="103"
      rx="8"
    />
    <circle cx="151" cy="60" r="10" fill="currentColor" opacity=".34" />
    <circle cx="151" cy="57" r="4" fill="currentColor" opacity=".62" />
    <path d="M143 67c2-6 14-6 16 0" fill="currentColor" opacity=".62" />
    <path className="ei-report-page-illustration-line" d="M171 55h36M171 68h25M143 84h64" opacity=".35" />
    <rect x="145" y="99" width="8" height="19" rx="3" fill="currentColor" opacity=".28" />
    <rect x="160" y="91" width="8" height="27" rx="3" fill="currentColor" opacity=".42" />
    <rect x="175" y="103" width="8" height="15" rx="3" fill="currentColor" opacity=".32" />
    <rect x="190" y="82" width="8" height="36" rx="3" fill="currentColor" opacity=".62" />
    <path d="M64 88h48a9 9 0 0 1 9 9v20a9 9 0 0 1-9 9H93l-12 10 2-10H64a9 9 0 0 1-9-9V97a9 9 0 0 1 9-9Z" fill="currentColor" opacity=".46" />
    <circle cx="76" cy="107" r="3" fill="#fff" opacity=".9" />
    <circle cx="88" cy="107" r="3" fill="#fff" opacity=".9" />
    <circle cx="100" cy="107" r="3" fill="#fff" opacity=".9" />
    <path d="m77 48 4 9 9 4-9 4-4 9-4-9-9-4 9-4Z" fill="currentColor" opacity=".42" />
  </svg>
);
