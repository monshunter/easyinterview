import type { FC } from "react";

export const AuthIllustration: FC<{ variant: "login" | "verify" }> = ({
  variant,
}) => (
  <svg
    aria-hidden="true"
    className={`ei-auth-illustration ei-auth-illustration--${variant}`}
    viewBox="0 0 360 210"
    fill="none"
  >
    <path d="M24 174 144 45l108 129M78 174l82-82 82 82M220 174l58-58 58 58" />
    {variant === "login" ? (
      <>
        <rect x="112" y="58" width="148" height="98" rx="16" />
        <path d="m122 75 64 48 64-48M122 143l45-40M250 143l-45-40" />
        <circle cx="278" cy="61" r="28" />
        <path d="m266 62 9 9 16-22" />
      </>
    ) : (
      <>
        <rect x="98" y="52" width="178" height="112" rx="16" />
        <circle cx="132" cy="93" r="9" />
        <circle cx="170" cy="93" r="9" />
        <circle cx="208" cy="93" r="9" />
        <circle cx="246" cy="93" r="9" />
        <path d="M125 132h124" />
        <circle cx="286" cy="53" r="28" />
        <path d="m274 54 9 9 16-22" />
      </>
    )}
  </svg>
);
