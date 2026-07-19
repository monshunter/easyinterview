import {
  forwardRef,
  type HTMLAttributes,
  type ReactNode,
} from "react";

export type AsyncTransitionVariant = "brand" | "resume" | "report" | "job";
export type AsyncTransitionStepState = "done" | "current" | "pending";

export interface AsyncTransitionStep {
  label: ReactNode;
  state: AsyncTransitionStepState;
  statusLabel?: ReactNode;
  testId?: string;
}

export interface AsyncTransitionAction {
  label: ReactNode;
  onClick: () => void;
  testId?: string;
  wrapperTestId?: string;
}

export interface AsyncTransitionSceneProps
  extends Omit<HTMLAttributes<HTMLElement>, "title"> {
  variant: AsyncTransitionVariant;
  testId: string;
  eyebrow: ReactNode;
  title: ReactNode;
  body?: ReactNode;
  hint?: ReactNode;
  steps?: readonly AsyncTransitionStep[];
  action?: AsyncTransitionAction;
  showProgress?: boolean;
  card?: boolean;
  contentTestId?: string;
  eyebrowTestId?: string;
  titleTestId?: string;
  bodyTestId?: string;
}

export const AsyncTransitionScene = forwardRef<HTMLElement, AsyncTransitionSceneProps>(
  function AsyncTransitionScene(
    {
      variant,
      testId,
      eyebrow,
      title,
      body,
      hint,
      steps,
      action,
      showProgress = false,
      card = false,
      contentTestId,
      eyebrowTestId,
      titleTestId,
      bodyTestId,
      className,
      ...rootProps
    },
    ref,
  ) {
    const classes = [
      "ei-transition-scene",
      `ei-transition-scene--${variant}`,
      card ? "ei-transition-scene--card" : "",
      className ?? "",
    ].filter(Boolean).join(" ");

    return (
      <section
        {...rootProps}
        ref={ref}
        className={classes}
        data-testid={testId}
        data-transition-variant={variant}
        role="status"
        aria-live="polite"
        aria-busy="true"
      >
        <div className="ei-transition-scene__atmosphere" aria-hidden="true">
          <span className="ei-transition-scene__blob ei-transition-scene__blob--left" />
          <span className="ei-transition-scene__blob ei-transition-scene__blob--right" />
          <span className="ei-transition-scene__dot ei-transition-scene__dot--one" />
          <span className="ei-transition-scene__dot ei-transition-scene__dot--two" />
          <span className="ei-transition-scene__dot ei-transition-scene__dot--three" />
        </div>

        <div className="ei-transition-scene__content" data-testid={contentTestId}>
          <TransitionIllustration variant={variant} />

          <div
            className="ei-label ei-transition-scene__eyebrow"
            data-testid={eyebrowTestId}
          >
            {eyebrow}
          </div>
          <h1
            className="ei-serif ei-transition-scene__title"
            data-testid={titleTestId}
          >
            {title}
          </h1>
          {body ? (
            <p className="ei-transition-scene__body" data-testid={bodyTestId}>
              {body}
            </p>
          ) : null}

          {steps?.length ? <TransitionSteps steps={steps} /> : null}

          {showProgress ? (
            <div
              className="ei-transition-scene__progress"
              data-testid="transition-progress"
              aria-hidden="true"
            >
              <span />
            </div>
          ) : null}

          {action ? (
            <span
              className="ei-transition-scene__action-wrap"
              data-testid={action.wrapperTestId}
            >
              <button
                type="button"
                className="ei-transition-scene__action"
                data-testid={action.testId}
                onClick={action.onClick}
              >
                {action.label}
              </button>
            </span>
          ) : null}

          {hint ? <p className="ei-transition-scene__hint">{hint}</p> : null}
        </div>
      </section>
    );
  },
);

const TransitionSteps = ({ steps }: { steps: readonly AsyncTransitionStep[] }) => (
  <ol className="ei-transition-scene__steps">
    {steps.map((step, index) => (
      <li
        key={index}
        className={`ei-transition-scene__step ei-transition-scene__step--${step.state}`}
        data-testid={step.testId}
        data-step-state={step.state}
        aria-current={step.state === "current" ? "step" : undefined}
      >
        <span className="ei-transition-scene__step-marker" aria-hidden="true">
          {step.state === "done" ? (
            <svg viewBox="0 0 16 16" fill="none">
              <path d="m3.5 8 3 3 6-6" />
            </svg>
          ) : index + 1}
        </span>
        <span className="ei-transition-scene__step-label">{step.label}</span>
        {step.state === "current" && step.statusLabel ? (
          <span className="ei-transition-scene__step-status">
            <span aria-hidden="true">●</span> {step.statusLabel}
          </span>
        ) : null}
      </li>
    ))}
  </ol>
);

const TransitionIllustration = ({ variant }: { variant: AsyncTransitionVariant }) => (
  <svg
    className="ei-transition-scene__illustration"
    data-testid={`transition-illustration-${variant}`}
    viewBox="0 0 260 190"
    fill="none"
    aria-hidden="true"
  >
    <g className="ei-transition-scene__illustration-orbit">
      <ellipse cx="130" cy="95" rx="109" ry="57" />
      <ellipse cx="130" cy="95" rx="82" ry="82" />
      <circle cx="39" cy="65" r="4" className="ei-transition-scene__illustration-dot" />
      <circle cx="220" cy="126" r="5" className="ei-transition-scene__illustration-dot" />
      <circle cx="197" cy="39" r="3" className="ei-transition-scene__illustration-dot" />
    </g>
    {variant === "brand" ? <BrandIllustration /> : null}
    {variant === "resume" ? <ResumeIllustration /> : null}
    {variant === "report" ? <ReportIllustration /> : null}
    {variant === "job" ? <JobIllustration /> : null}
  </svg>
);

const BrandIllustration = () => (
  <g className="ei-transition-scene__illustration-main">
    <circle cx="130" cy="95" r="54" className="ei-transition-scene__halo" />
    <circle cx="130" cy="95" r="38" className="ei-transition-scene__surface" />
    <circle cx="130" cy="95" r="25" className="ei-transition-scene__accent" />
    <text x="130" y="104" textAnchor="middle" className="ei-transition-scene__brand-letter">E</text>
  </g>
);

const ResumeIllustration = () => (
  <g className="ei-transition-scene__illustration-main">
    <circle cx="130" cy="95" r="53" className="ei-transition-scene__halo" />
    <path d="M102 49h44l26 26v68H102z" className="ei-transition-scene__surface" />
    <path d="M146 49v27h26" className="ei-transition-scene__line" />
    <path d="M116 91h42M116 106h42M116 121h29" className="ei-transition-scene__line" />
    <circle cx="165" cy="57" r="8" className="ei-transition-scene__accent" />
  </g>
);

const ReportIllustration = () => (
  <g className="ei-transition-scene__illustration-main">
    <circle cx="130" cy="95" r="53" className="ei-transition-scene__halo" />
    <path d="M92 61h61l18 18v63H92z" className="ei-transition-scene__surface" />
    <path d="M153 61v20h18" className="ei-transition-scene__line" />
    <path d="M109 119v-13M124 119V92M139 119v-19" className="ei-transition-scene__accent-line" />
    <path d="M106 129h39" className="ei-transition-scene__line" />
    <circle cx="170" cy="61" r="8" className="ei-transition-scene__accent" />
  </g>
);

const JobIllustration = () => (
  <g className="ei-transition-scene__illustration-main">
    <circle cx="123" cy="95" r="53" className="ei-transition-scene__halo" />
    <path d="M89 57h63v82H89z" className="ei-transition-scene__surface" />
    <path d="M104 78h33M104 94h33M104 110h23" className="ei-transition-scene__line" />
    <circle cx="164" cy="118" r="22" className="ei-transition-scene__surface" />
    <circle cx="164" cy="118" r="12" className="ei-transition-scene__accent-line" />
    <path d="m180 134 16 16" className="ei-transition-scene__accent-line" />
  </g>
);
