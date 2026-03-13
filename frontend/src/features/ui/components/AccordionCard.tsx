import type { ReactNode } from 'react';

type AccordionCardProps = {
  title: string;
  isOpen: boolean;
  onToggle: () => void;
  flexRatio: number;
  children: ReactNode;
};

export function AccordionCard({
  title,
  isOpen,
  onToggle,
  flexRatio: _flexRatio,
  children,
}: AccordionCardProps) {
  return (
    <section className="shell-panel pointer-events-auto flex flex-col rounded-xl border shadow-[0_24px_80px_-28px_rgba(0,0,0,0.85)]">
      <div className="flex flex-none items-center justify-between rounded-t-xl bg-inherit px-5 py-4">
        <h2
          className={`m-0 text-xl/6 font-medium tracking-tight text-[color:var(--color-shell-text)] origin-left transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] ${
            isOpen ? 'scale-100 opacity-100' : 'scale-[0.96] opacity-80'
          }`}
        >
          {title}
        </h2>
        <button
          onClick={onToggle}
          className="cursor-pointer text-xs font-medium uppercase tracking-wider text-[color:var(--color-shell-text-muted)] transition-all duration-200 ease-out hover:text-[color:var(--color-primary)] active:scale-95"
        >
          {isOpen ? 'HIDE DETAILS' : 'SEE DETAILS'}
        </button>
      </div>

      <div
        className="grid transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)]"
        style={{ gridTemplateRows: isOpen ? '1fr' : '0fr', opacity: isOpen ? 1 : 0 }}
      >
        <div className="overflow-hidden">
          <div
            className="w-full border-b border-[color:var(--color-shell-panel-border)] transition-opacity duration-500"
            style={{ opacity: isOpen ? 0.3 : 0 }}
          />
          <div
            className={`flex flex-col gap-5 overflow-y-auto px-5 py-5 transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] ${
              isOpen ? 'translate-y-0 scale-100 blur-none' : '-translate-y-4 scale-[0.98] blur-[2px]'
            }`}
          >
            {children}
          </div>
        </div>
      </div>
    </section>
  );
}
