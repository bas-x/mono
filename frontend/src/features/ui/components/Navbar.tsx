export function Navbar() {
  return (
    <aside className="w-full max-w-25 shrink-0 bg-surface px-2 py-3 text-text shadow-[0_12px_28px_-18px_rgba(15,23,42,0.75)] dark:shadow-[0_14px_30px_-18px_rgba(2,6,23,0.95)]">
      <div className="flex h-full min-h-[calc(100vh-2rem)] flex-col items-center gap-6">
        <div className="pt-1 text-center text-sm font-semibold tracking-[0.18em] text-primary">
          bas x
        </div>
        <nav aria-label="Primary" className="w-full">
          <a
            href="/simulation"
            onClick={(event) => event.preventDefault()}
            className="block rounded-lg border border-border/80 bg-bg px-2 py-2 text-center text-xs font-medium text-text-muted transition-colors hover:border-primary hover:text-text focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-surface"
          >
            Simulation
          </a>
        </nav>
      </div>
    </aside>
  );
}
