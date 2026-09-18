import { forwardRef, useId, type InputHTMLAttributes } from "react";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { label, error, id, className = "", ...props },
  ref,
) {
  const generatedId = useId();
  const inputId = id ?? generatedId;
  const errorId = `${inputId}-error`;

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={inputId} className="text-label-md text-(--ink-primary)">
        {label}
      </label>
      <input
        ref={ref}
        id={inputId}
        className={`text-body-md rounded-md border px-3 py-2.5 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30 ${
          error ? "border-error-500" : "border-(--border-default)"
        } ${className}`}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errorId : undefined}
        {...props}
      />
      {error && (
        <p id={errorId} className="text-body-sm text-error-500">
          {error}
        </p>
      )}
    </div>
  );
});
