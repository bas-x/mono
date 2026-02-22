import type { PropsWithChildren, ReactNode } from 'react';

type CardProps = PropsWithChildren<{
  title: ReactNode;
  className?: string;
  ariaLabel?: string;
}>;

function mergeClassNames(...parts: Array<string | undefined>) {
  return parts.filter(Boolean).join(' ');
}

export function Card({ title, className, ariaLabel, children }: CardProps) {
  return (
    <section
      className={mergeClassNames(
        'min-h-55 bg-surface p-4 dark:border-border dark:bg-surface',
        className,
      )}
      aria-label={ariaLabel}
    >
      <h2 className="m-0 text-base font-semibold text-text-muted dark:text-text-muted">{title}</h2>
      {children}
    </section>
  );
}
