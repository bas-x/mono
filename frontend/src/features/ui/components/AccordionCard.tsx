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
  flexRatio,
  children,
}: AccordionCardProps) {
  return (
    <section
      className="shell-panel pointer-events-auto flex flex-col rounded-xl border shadow-[0_24px_80px_-28px_rgba(0,0,0,0.85)] transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)]"
      style={{
        flexGrow: isOpen ? flexRatio : 0,
        flexShrink: isOpen ? 1 : 0,
        flexBasis: 'auto',
      }}
    >
      <div className="rounded-xl flex flex-none items-center justify-between rounded-t-xl bg-inherit px-5 py-4 z-10">
        <h2
          className={`m-0 text-lg font-semibold text-[color:var(--color-shell-text)] origin-left transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] ${
            isOpen ? 'scale-100 opacity-100' : 'scale-[0.96] opacity-80'
          }`}
        >
          {title}
        </h2>
        <button
          onClick={onToggle}
          className="cursor-pointer text-[10px] font-bold uppercase tracking-wider text-white transition-all duration-200 ease-out hover:text-white/70 active:scale-95"
        >
          {isOpen ? 'HIDE DETAILS' : 'SEE DETAILS'}
        </button>
      </div>

      <div
        className="flex min-h-0 flex-col overflow-hidden transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)]"
        style={{
          flexGrow: isOpen ? 1 : 0,
          flexBasis: '0px',
          opacity: isOpen ? 1 : 0,
        }}
      >
        <div
          className="w-full flex-none border-b border-[color:var(--color-shell-panel-border)] transition-opacity duration-500"
          style={{ opacity: isOpen ? 0.3 : 0 }}
        />
        <div
          className={`flex min-h-0 flex-1 flex-col gap-5 overflow-y-auto px-5 py-5 transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] ${
            isOpen ? 'translate-y-0 scale-100 blur-none' : '-translate-y-4 scale-[0.98] blur-[2px]'
          }`}
        >
          {children}
        </div>
      </div>
    </section>
  );
}
