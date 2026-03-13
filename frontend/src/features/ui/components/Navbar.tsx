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
          colorClass: 'border-yellow-500/50 text-yellow-500',
        };
      case 'remote':
        return {
          icon: <HiOutlineServer className="h-6 w-6" />,
          label: 'Using Remote API',
          next: 'Switch to Localhost API',
          colorClass: 'border-green-500/50 text-green-500',
        };
      case 'localhost':
        return {
          icon: <HiOutlineDesktopComputer className="h-6 w-6" />,
          label: 'Using Localhost (8080)',
          next: 'Switch to Mock API',
          colorClass: 'border-blue-500/50 text-blue-500',
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
            className="flex h-12 w-12 items-center justify-center rounded-lg bg-gradient-to-b from-white/[0.06] to-transparent shadow-[0_0_12px_-3px_rgba(255,255,255,0.1),inset_0_1px_0_rgba(255,255,255,0.06)] ring-1 ring-white/[0.04]"
          >
            <img
              src={logoSrc}
              alt="Bas X"
              className="h-full w-full object-contain drop-shadow-[0_0_4px_rgba(255,255,255,0.15)]"
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
