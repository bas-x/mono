import type { InputHTMLAttributes, ReactNode } from 'react';

import { FormField } from '@/features/ui/components/FormField';
import { Input } from '@/features/ui/components/Input';

type NumberFieldProps = {
  label: ReactNode;
  hint?: ReactNode;
  htmlFor?: string;
  className?: string;
} & Omit<InputHTMLAttributes<HTMLInputElement>, 'type'>;

export function NumberField({ label, hint, htmlFor, className, ...props }: NumberFieldProps) {
  return (
    <FormField label={label} hint={hint} htmlFor={htmlFor} className={className}>
      <Input {...props} id={htmlFor} type="number" />
    </FormField>
  );
}
