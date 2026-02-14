import type { PropsWithChildren, ReactNode } from 'react';

type CardProps = PropsWithChildren<{
  title: ReactNode;
  className?: string;
  ariaLabel?: string;
}>;

export function Card({ title, className = '', ariaLabel, children }: CardProps) {
  return (
    <section className={`panel ${className}`.trim()} aria-label={ariaLabel}>
      <h2>{title}</h2>
      {children}
    </section>
  );
}
