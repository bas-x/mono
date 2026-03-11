import type { InputHTMLAttributes } from 'react';

function mergeClassNames(...parts: Array<string | undefined>) {
  return parts.filter(Boolean).join(' ');
}

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return <input {...props} className={mergeClassNames('shell-input', className)} />;
}
