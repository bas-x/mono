import { MapPanel, Navbar, TimelinePanel } from '../features';

export function BaseXOps() {
  return (
    <div className="app-shell">
      <Navbar title="bas X" />
      <main className="app-main">
        <MapPanel />
        <TimelinePanel />
      </main>
    </div>
  );
}
