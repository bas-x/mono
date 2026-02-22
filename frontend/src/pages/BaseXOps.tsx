import { ApiStatusPanel, MapPanel, Navbar, TimelinePanel } from '@/features';

export function BaseXOps() {
  return (
    <div className="flex">
      <a
        href="#main-content"
        className="sr-only rounded-md bg-surface px-3 py-2 text-text focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-bg"
      >
        Skip to main content
      </a>
      <Navbar />
      <main
        id="main-content"
        className="grid min-w-0 flex-1 content-start grid-cols-1 gap-1 min-[900px]:grid-cols-[8fr_1fr]"
      >
        <MapPanel />

        <TimelinePanel />

        <ApiStatusPanel />
      </main>
    </div>
  );
}
