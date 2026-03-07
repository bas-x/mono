import { SiOpenlayers } from 'react-icons/si';

export function Navbar() {
  return (
    <aside className="h-full w-full max-w-20 shrink-0 border-r border-zinc-800 bg-zinc-950 py-4 text-zinc-100 shadow-[18px_0_40px_-28px_rgba(0,0,0,0.9)]">
      <div className="flex h-full min-h-0 flex-col items-center gap-8">
        <div className="flex w-full flex-col items-center gap-3">
          <div
            aria-label="Bas X"
            className="flex h-12 w-12 text-center items-center justify-center rounded-md border border-zinc-800 bg-zinc-900/80 text-[0.68rem] font-semibold uppercase tracking-[0.24em] text-zinc-100"
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
            className="group flex h-12 w-12 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-zinc-400 transition-all duration-200 hover:border-teal-500/60 hover:bg-zinc-800 hover:text-teal-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-teal-400 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950"
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
