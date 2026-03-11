import { SiOpenlayers } from 'react-icons/si';
import { HiOutlineLightningBolt, HiOutlineServer } from 'react-icons/hi';
import { useApi } from '@/lib/api';

export function Navbar() {
  const { config, setUseMock } = useApi();

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
            className="shell-nav-item-active group flex h-12 w-12 items-center justify-center rounded-md border transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-bg"
          >
            <SiOpenlayers
              aria-hidden="true"
              className="shell-nav-icon h-5 w-5 transition-[color,transform] duration-200 group-hover:scale-105"
            />
            <span className="sr-only">Simulation</span>
          </a>
        </nav>

        <div className="flex w-full flex-col items-center gap-4">
          <button
            type="button"
            onClick={() => setUseMock(!config.useMock)}
            aria-label={config.useMock ? "Switch to Real API" : "Switch to Mock API"}
            title={config.useMock ? "Using Mocks (Click for Real)" : "Using Real API (Click for Mocks)"}
            className={`flex h-12 w-12 items-center justify-center rounded-md border transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-bg ${
              config.useMock 
                ? 'shell-panel-soft border-yellow-500/50 text-yellow-500' 
                : 'shell-panel-soft border-green-500/50 text-green-500'
            }`}
          >
            {config.useMock ? (
              <HiOutlineLightningBolt className="h-6 w-6" />
            ) : (
              <HiOutlineServer className="h-6 w-6" />
            )}
          </button>
        </div>
      </div>
    </aside>
  );
}
