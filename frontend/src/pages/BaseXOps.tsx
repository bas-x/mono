import { MapPanel, Navbar } from '@/features';

export function BaseXOps() {
  return (
    <div className="flex h-dvh overflow-hidden">
      <a
        href="#main-content"
        className="sr-only rounded-md bg-surface px-3 py-2 text-text focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-bg"
      >
        Skip to main content
      </a>
      <Navbar />
      <main id="main-content" className="min-h-0 min-w-0 flex-1">
        <MapPanel />
      </main>
    </div>
  );
}
