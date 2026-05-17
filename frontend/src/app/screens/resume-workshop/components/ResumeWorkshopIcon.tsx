import type { FC } from "react";

export type ResumeWorkshopIconName =
  | "arrowLeft"
  | "arrowRight"
  | "briefcase"
  | "chat"
  | "check"
  | "chevronDown"
  | "chevronRight"
  | "file"
  | "layers"
  | "plus"
  | "resume"
  | "sparkle"
  | "upload";

interface ResumeWorkshopIconProps {
  name: ResumeWorkshopIconName;
  size?: number;
  "data-testid"?: string;
}

export const ResumeWorkshopIcon: FC<ResumeWorkshopIconProps> = ({
  name,
  size = 13,
  "data-testid": testId,
}) => {
  const paths: Record<ResumeWorkshopIconName, JSX.Element> = {
    arrowLeft: <path d="M19 12H5M11 18l-6-6 6-6" />,
    arrowRight: <path d="M5 12h14M13 6l6 6-6 6" />,
    briefcase: (
      <>
        <path d="M4 8h16v11H4z" />
        <path d="M9 8V5h6v3M4 13h16" />
      </>
    ),
    chat: <path d="M4 5h16v11H9l-5 4V5z" />,
    check: <path d="M5 12l5 5L20 7" />,
    chevronDown: <path d="M6 9l6 6 6-6" />,
    chevronRight: <path d="M9 6l6 6-6 6" />,
    file: <path d="M7 3h8l4 4v14H7z M15 3v5h4" />,
    layers: <path d="M12 3l9 5-9 5-9-5 9-5z M3 12l9 5 9-5M3 16l9 5 9-5" />,
    plus: <path d="M12 5v14M5 12h14" />,
    resume: <path d="M7 3h8l4 4v14H7z M9 12h6M9 16h6M15 3v5h4" />,
    sparkle: <path d="M12 2l1.8 5.5L19 9l-5.2 1.5L12 16l-1.8-5.5L5 9l5.2-1.5z" />,
    upload: <path d="M12 16V4M7 9l5-5 5 5M5 20h14" />,
  };

  return (
    <svg
      aria-hidden="true"
      className="ei-resume-workshop-icon"
      data-icon={name}
      data-testid={testId}
      fill="none"
      height={size}
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={1.8}
      viewBox="0 0 24 24"
      width={size}
    >
      {paths[name]}
    </svg>
  );
};
