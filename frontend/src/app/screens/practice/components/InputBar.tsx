import { type FC, type ReactNode } from "react";

export interface InputBarProps {
  value: string;
  onChange: (next: string) => void;
  helperText: string;
  placeholder: string;
  sendLabel: string;
  disabled: boolean;
  sendDisabled?: boolean;
  recovery?: ReactNode;
  onSend: () => void;
}

export const InputBar: FC<InputBarProps> = ({ value, onChange, helperText, placeholder, sendLabel, disabled, sendDisabled = false, recovery, onSend }) => {
  const isSendDisabled = disabled || sendDisabled || !value.trim();
  return (
    <div data-testid="practice-input" className="ei-practice-input">
      {recovery}
      <div data-testid="practice-input-helper" className="ei-practice-input-helper"><SparkleIcon /><span>{helperText}</span></div>
      <div data-testid="practice-input-shell" className="ei-practice-input-shell">
        <textarea data-testid="practice-input-textarea" value={value} onChange={(event) => onChange(event.target.value)} onKeyDown={(event) => { if (!isSendDisabled && (event.metaKey || event.ctrlKey) && event.key === "Enter") onSend(); }} placeholder={placeholder} disabled={disabled} className="ei-practice-input-textarea" />
        <div className="ei-practice-input-actions">
          <button data-testid="practice-input-send" type="button" onClick={onSend} disabled={isSendDisabled} className="ei-practice-input-send">{sendLabel}<span aria-hidden="true">↵</span></button>
        </div>
      </div>
    </div>
  );
};

const SparkleIcon: FC = () => (
  <svg aria-hidden="true" viewBox="0 0 24 24" width="17" height="17" fill="currentColor"><path d="M12 2.8c.8 4.8 3.3 7.4 8.2 8.2-4.9.8-7.4 3.4-8.2 8.2-.8-4.8-3.3-7.4-8.2-8.2 4.9-.8 7.4-3.4 8.2-8.2Z" /><path d="M19 2.5c.25 1.55 1.05 2.35 2.5 2.5-1.45.25-2.25 1.05-2.5 2.5-.25-1.45-1.05-2.25-2.5-2.5 1.45-.15 2.25-.95 2.5-2.5Z" /></svg>
);
