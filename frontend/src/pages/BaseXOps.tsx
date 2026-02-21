import { MapPanel, Navbar, TimelinePanel } from '../features';

export function BaseXOps() {
  return (
    <div className="flex min-h-screen flex-col gap-4 p-4">
      <Navbar title="bas X" />
      <main className="grid flex-1 grid-cols-1 gap-4 min-[900px]:grid-cols-[2fr_1fr]">
        <MapPanel />
        <TimelinePanel />
      </main>
    </div>
  );
}
