import { HiOutlineExclamationCircle, HiOutlineXCircle } from 'react-icons/hi';
import type { ReactNode } from 'react';

type ErrorMessageProps = {
  title?: string;
  message: ReactNode;
  variant?: 'warning' | 'error';
  className?: string;
};

export function ErrorMessage({ title, message, variant = 'error', className = '' }: ErrorMessageProps) {
  const isWarning = variant === 'warning';
  
  return (
    <div
      className={`flex items-start gap-3 rounded-md border p-3 text-sm ${
        isWarning
          ? 'border-yellow-500/30 bg-yellow-500/10 text-yellow-200'
          : 'shell-error-surface shell-divider border text-red-400'
      } ${className}`}
      role="alert"
    >
      {isWarning ? (
        <HiOutlineExclamationCircle className="mt-0.5 h-4 w-4 shrink-0 text-yellow-500" />
      ) : (
        <HiOutlineXCircle className="mt-0.5 h-4 w-4 shrink-0 text-red-500" />
      )}
      <div className="flex flex-col gap-1">
        {title && <span className="font-semibold">{title}</span>}
        <span className="opacity-90 leading-snug">{message}</span>
      </div>
    </div>
  );
}