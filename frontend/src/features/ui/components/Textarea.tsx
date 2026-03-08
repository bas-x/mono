import type { TextareaHTMLAttributes } from 'react';

function mergeClassNames(...parts: Array<string | undefined>) {
  return parts.filter(Boolean).join(' ');
}

export function Textarea({ className, ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea {...props} className={mergeClassNames('shell-input min-h-24 resize-y', className)} />;
}
