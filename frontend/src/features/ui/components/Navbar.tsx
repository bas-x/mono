import { SiOpenlayers } from 'react-icons/si';

export function Navbar() {
  return (
    <aside className="shell-panel h-full w-full max-w-20 shrink-0 border-r py-4 shadow-[18px_0_40px_-28px_rgba(0,0,0,0.9)]">
      <div className="flex h-full min-h-0 flex-col items-center gap-8">
        <div className="flex w-full flex-col items-center gap-3">
          <div
            aria-label="Bas X"
            className="shell-panel-soft flex h-12 w-12 items-center justify-center rounded-md border border-border text-center text-[0.68rem] font-semibold uppercase tracking-[0.24em] text-text"
          >
            bas x
          </div>
        </div>

        <nav aria-label="Primary" className="flex w-full flex-1 flex-col items-center">
          <a
            href="/simulation"
            onClick={(event) => event.preventDefault()}
            aria-label="Simulation"
            title="Simulation"
            aria-current="page"
            className="shell-button-active group flex h-12 w-12 items-center justify-center rounded-md border transition-all duration-200 hover:bg-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-bg"
          >
            <SiOpenlayers
              aria-hidden="true"
              className="h-5 w-5 transition-transform duration-200 group-hover:scale-105"
            />
            <span className="sr-only">Simulation</span>
          </a>
        </nav>
      </div>
    </aside>
  );
}
