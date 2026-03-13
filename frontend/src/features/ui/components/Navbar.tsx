import { SiOpenlayers } from 'react-icons/si';
import { HiOutlineLightningBolt, HiOutlineServer, HiOutlineDesktopComputer } from 'react-icons/hi';
import { useApi } from '@/lib/api';
import type { ApiMode } from '@/lib/api/types';
import logoSrc from '@/assets/logo.png';

export function Navbar() {
  const { config, setMode } = useApi();

  const cycleMode = () => {
    const modes: ApiMode[] = ['mock', 'remote', 'localhost'];
    const currentIndex = modes.indexOf(config.mode);
    const nextIndex = (currentIndex + 1) % modes.length;
    setMode(modes[nextIndex]);
  };

  const getModeInfo = (mode: ApiMode) => {
    switch (mode) {
      case 'mock':
        return {
          icon: <HiOutlineLightningBolt className="h-6 w-6" />,
          label: 'Using Mocks',
          next: 'Switch to Remote API',
          colorClass:
            'border-[color:var(--color-primary)]/45 text-[color:var(--color-primary)] bg-[color:var(--color-shell-button-hover)]',
        };
      case 'remote':
        return {
          icon: <HiOutlineServer className="h-6 w-6" />,
          label: 'Using Remote API',
          next: 'Switch to Localhost API',
          colorClass:
            'border-[color:var(--color-shell-button-border)] text-[color:var(--color-primary-strong)] bg-[color:var(--color-shell-panel)]',
        };
      case 'localhost':
        return {
          icon: <HiOutlineDesktopComputer className="h-6 w-6" />,
          label: 'Using Localhost (8080)',
          next: 'Switch to Mock API',
          colorClass:
            'border-[color:var(--color-accent)]/45 text-[color:var(--color-accent)] bg-[color:var(--color-shell-button-hover)]',
        };
    }
  };

  const modeInfo = getModeInfo(config.mode);

  return (
    <aside className="shell-panel h-full w-full max-w-20 shrink-0 border-r py-4 shadow-[18px_0_40px_-28px_rgba(0,0,0,0.9)]">
      <div className="flex h-full min-h-0 flex-col items-center gap-8">
        <div className="flex w-full flex-col items-center gap-3">
          <div
            aria-label="Bas X"
            className="flex h-12 w-12 items-center justify-center rounded-lg border border-[color:var(--color-primary)]/30 bg-[radial-gradient(circle_at_30%_30%,rgba(245,158,11,0.28),rgba(120,53,15,0.92)_68%)] shadow-[0_0_20px_-8px_rgba(217,119,6,0.8),inset_0_1px_0_rgba(255,255,255,0.08)]"
          >
            <img
              src={logoSrc}
              alt="Bas X"
              className="h-full w-full object-contain opacity-95"
              style={{
                filter:
                  'sepia(1) saturate(4.2) hue-rotate(345deg) brightness(0.88) contrast(1.12) drop-shadow(0 0 6px rgba(245,158,11,0.35))',
              }}
            />
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
            onClick={cycleMode}
            aria-label={modeInfo.next}
            title={`${modeInfo.label} (${modeInfo.next})`}
            className={`flex h-12 w-12 items-center justify-center rounded-md border transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-bg shell-panel-soft ${modeInfo.colorClass}`}
          >
            {modeInfo.icon}
          </button>
        </div>
      </div>
    </aside>
  );
}
