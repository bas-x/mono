import type { PropsWithChildren, ReactNode } from 'react';

type FormFieldProps = PropsWithChildren<{
  label: ReactNode;
  htmlFor?: string;
  hint?: ReactNode;
  className?: string;
}>;

function mergeClassNames(...parts: Array<string | undefined>) {
  return parts.filter(Boolean).join(' ');
}

export function FormField({ label, htmlFor, hint, className, children }: FormFieldProps) {
  return (
    <div className={mergeClassNames('space-y-2', className)}>
      <div className="space-y-1">
        <label htmlFor={htmlFor} className="shell-field-label block text-sm font-medium">
          {label}
        </label>
        {hint ? <p className="shell-field-hint m-0 text-xs">{hint}</p> : null}
      </div>
      {children}
    </div>
  );
}
