import { type SimulationState } from '@/features/simulation/hooks/useSimulation';
import { createPortal } from 'react-dom';
import { useState } from 'react';
import { AccordionCard } from '@/features/ui/components/AccordionCard';
import type { SimulationInfo } from '@/lib/api/types';

type SimulationInfoCardProps = {
  simulationState: SimulationState;
  simulations?: SimulationInfo[];
  onSelectSimulation?: (simulationId: string) => void;
  onOverrideAssignment?: (tailNumber: string, baseId: string) => Promise<boolean>;
  portalRoot: Element | null;
};

export function SimulationInfoCard({
  simulationState,
  simulations = [],
  onSelectSimulation,
  onOverrideAssignment,
  portalRoot,
}: SimulationInfoCardProps) {
  const [isSimOpen, setIsSimOpen] = useState(true);
  const [isAirOpen, setIsAirOpen] = useState(false);
  const [overrideTargets, setOverrideTargets] = useState<Record<string, string>>({});

  if (typeof document === 'undefined' || simulationState.status !== 'running' || !portalRoot) {
    return null;
  }

  const timeString = simulationState.time
    ? new Date(simulationState.time).toLocaleString(undefined, {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
    : '00:00:00';

  const dateString = simulationState.time
    ? new Date(simulationState.time).toLocaleDateString()
    : 'N/A';

  const aircraftCount = simulationState.aircrafts?.length ?? 0;
  const airbaseCount = simulationState.airbases?.length ?? 0;

  const branches = simulations.map((simulation) => ({
    id: simulation.id,
    name: simulation.id === 'base' ? 'Base' : simulation.id.substring(0, 8),
    status: simulationState.simulationId === simulation.id ? 'active' : 'idle',
    detail: simulation.id === 'base'
      ? 'Canonical base run'
      : typeof simulation.splitTick === 'number'
        ? `Forked at tick ${simulation.splitTick}`
        : 'Branch',
    annotation: simulation.sourceEvent
      ? `${simulation.sourceEvent.type} @ ${simulation.sourceEvent.tick}`
      : simulation.parentId ?? 'No metadata',
  }));

  return createPortal(
    <div className="pointer-events-none absolute inset-4 z-20 flex items-start justify-start">
      <div className="pointer-events-auto flex w-full max-w-96 flex-col gap-4">
        <AccordionCard
          title="Simulation information"
          isOpen={isSimOpen}
          onToggle={() => setIsSimOpen(!isSimOpen)}
          flexRatio={7}
        >
        <div className="grid grid-cols-4 gap-2">
          <div className="flex flex-col">
            <span className="text-[10px] font-semibold uppercase tracking-widest text-[color:var(--color-shell-text-muted)]">
              Tick
            </span>
            <span className="mt-1 font-mono text-xl font-medium tracking-tight text-[color:var(--color-shell-text)]">
              {simulationState.tick ?? 0}
            </span>
          </div>
          <div className="flex flex-col">
            <span className="text-[10px] font-semibold uppercase tracking-widest text-[color:var(--color-shell-text-muted)]">
              Aircraft
            </span>
            <span className="mt-1 font-mono text-xl font-medium tracking-tight text-[color:var(--color-shell-text)]">
              {aircraftCount}
            </span>
          </div>
          <div className="flex flex-col">
            <span className="text-[10px] font-semibold uppercase tracking-widest text-[color:var(--color-shell-text-muted)]">
              Airbase
            </span>
            <span className="mt-1 font-mono text-xl font-medium tracking-tight text-[color:var(--color-shell-text)]">
              {airbaseCount}
            </span>
          </div>
          <div className="flex flex-col">
            <span className="text-[10px] font-semibold uppercase tracking-widest text-[color:var(--color-shell-text-muted)]">
              Events
            </span>
            <span className="mt-1 font-mono text-xl font-medium tracking-tight text-[color:var(--color-shell-text)]">
              4
            </span>
          </div>
        </div>

        <div className="flex flex-col justify-center transition-transform duration-500 hover:scale-[1.02]">
          <span className="text-xl font-medium tracking-tight text-[color:var(--color-primary)] drop-shadow-[0_0_8px_rgba(217,119,6,0.28)] transition-all duration-300">
            {timeString}
          </span>
          <span className="mt-1 text-xs text-[color:var(--color-shell-text-muted)]">{dateString}</span>
        </div>

        <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-[color:var(--color-shell-border)] bg-[color:var(--color-shell-panel-soft)]">
          <div className="flex items-center justify-between border-b border-[color:var(--color-shell-border)] bg-[color:var(--color-shell-panel)] px-2 py-3">
            <h3 className="m-0 flex items-center gap-2 text-sm font-medium text-[color:var(--color-shell-text)]">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="text-[color:var(--color-primary)]"
              >
                <path d="M6 3v12" />
                <circle cx="18" cy="6" r="3" />
                <circle cx="6" cy="18" r="3" />
                <path d="M18 9a9 9 0 0 1-9 9" />
              </svg>
              Branches
            </h3>
            <span className="rounded-full border border-[color:var(--color-shell-button-border)] bg-[color:var(--color-shell-button-bg)] px-2 py-0.5 text-[10px] uppercase tracking-wider text-[color:var(--color-shell-text-muted)] transition-colors hover:bg-[color:var(--color-shell-button-hover)] hover:text-[color:var(--color-shell-text)]">
              {branches.length} Available
            </span>
          </div>
          <div className="flex-1 space-y-1 overflow-y-auto p-2">
            {branches.map((branch, idx) => (
              <div key={branch.id}>
                <button
                  type="button"
                  onClick={() => onSelectSimulation?.(branch.id)}
                  className={`group flex w-full items-center justify-between rounded-md p-3 text-left transition-all duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] active:scale-[0.98] ${
                    branch.status === 'active'
                      ? 'border border-[color:var(--color-primary)]/35 bg-[color:var(--color-shell-nav-active-bg)] shadow-[inset_0_0_12px_rgba(217,119,6,0.12)]'
                      : 'border border-transparent hover:bg-[color:var(--color-shell-button-hover)]'
                  }`}
                >
                  <div className="flex items-center gap-3 transition-transform duration-300 ease-out group-hover:translate-x-1">
                    <div
                      className={`h-2 w-2 rounded-full transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] ${
                        branch.status === 'active'
                          ? 'scale-110 bg-[color:var(--color-primary)] shadow-[0_0_8px_rgba(217,119,6,0.55)]'
                          : 'bg-[color:var(--color-shell-text-muted)]/30 group-hover:scale-110 group-hover:bg-[color:var(--color-shell-text-muted)]/55'
                      }`}
                    />
                    <div>
                      <div
                        className={`text-sm font-medium transition-colors duration-300 ${
                          branch.status === 'active'
                            ? 'text-[color:var(--color-shell-text)]'
                            : 'text-[color:var(--color-shell-text-muted)] group-hover:text-[color:var(--color-shell-text)]'
                        }`}
                      >
                        {branch.name}
                      </div>
                      <div className="mt-0.5 text-[10px] text-[color:var(--color-shell-text-muted)] transition-colors duration-300 group-hover:text-[color:var(--color-shell-text)]">
                        {branch.detail}
                      </div>
                    </div>
                  </div>
                  <div className="font-mono text-[10px] font-medium text-[color:var(--color-primary)] transition-transform duration-300 ease-out group-hover:-translate-x-1">
                    {branch.annotation}
                  </div>
                </button>
                {idx < branches.length - 1 && (
                  <div className="my-1 h-px w-full bg-[color:var(--color-shell-border)]" />
                )}
              </div>
            ))}
          </div>
        </div>
        </AccordionCard>

        <AccordionCard
          title="Aircrafts"
          isOpen={isAirOpen}
          onToggle={() => setIsAirOpen(!isAirOpen)}
          flexRatio={3}
        >
        <div className="flex flex-col gap-0 overflow-hidden rounded-lg border border-[color:var(--color-shell-border)] bg-[color:var(--color-shell-panel-soft)]">
          {simulationState.aircrafts?.map((ac, idx) => (
            <div
              key={ac.tailNumber}
              className={`group flex cursor-pointer flex-col gap-3 p-3 transition-all duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] hover:bg-[color:var(--color-shell-button-hover)] ${
                idx !== simulationState.aircrafts.length - 1
                  ? 'border-b border-[color:var(--color-shell-border)]'
                  : ''
              }`}
            >
              <div className="flex items-center justify-between gap-3">
                <div className="flex items-center gap-3 transition-transform duration-400 ease-[cubic-bezier(0.32,0.72,0,1)] group-hover:translate-x-2">
                  <span className="font-mono text-sm font-semibold text-[color:var(--color-shell-text)] transition-colors duration-300 group-hover:text-[color:var(--color-primary)]">
                    {ac.tailNumber.substring(0, 8)}
                  </span>
                  <span className="rounded-full border border-[color:var(--color-shell-button-border)] bg-[color:var(--color-shell-button-bg)] px-2 py-0.5 text-[10px] uppercase tracking-wider text-[color:var(--color-shell-text-muted)] transition-all duration-300 group-hover:border-[color:var(--color-primary)]/40 group-hover:bg-[color:var(--color-shell-button-hover)] group-hover:text-[color:var(--color-shell-text)]">
                    {ac.state}
                  </span>
                </div>
                <div className="flex items-center gap-2 text-[10px] uppercase tracking-wider text-[color:var(--color-shell-text-muted)]">
                  <span>{ac.assignedTo ? `Base ${ac.assignedTo.slice(0, 6)}` : 'Unassigned'}</span>
                  <span className="rounded-full border border-[color:var(--color-shell-button-border)] px-2 py-0.5">
                    {ac.assignmentSource === 'human' ? 'Manual' : 'Auto'}
                  </span>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <select
                  value={overrideTargets[ac.tailNumber] ?? ac.assignedTo ?? simulationState.airbases[0]?.id ?? ''}
                  onChange={(event) => {
                    const nextBaseId = event.target.value;
                    setOverrideTargets((current) => ({
                      ...current,
                      [ac.tailNumber]: nextBaseId,
                    }));
                  }}
                  className="shell-input min-w-0 flex-1 rounded-sm border px-2 py-1.5 text-xs"
                >
                  {simulationState.airbases.map((airbase) => (
                    <option key={airbase.id} value={airbase.id}>
                      {airbase.region} ({airbase.id.slice(0, 6)})
                    </option>
                  ))}
                </select>
                <button
                  type="button"
                  onClick={() => {
                    const targetBaseId = overrideTargets[ac.tailNumber] ?? ac.assignedTo ?? simulationState.airbases[0]?.id;
                    if (!targetBaseId || !onOverrideAssignment) {
                      return;
                    }
                    void onOverrideAssignment(ac.tailNumber, targetBaseId);
                  }}
                  className="shell-button rounded-sm border px-3 py-1.5 text-[10px] font-bold uppercase tracking-wider text-[color:var(--color-primary)]"
                >
                  Override
                </button>
              </div>
            </div>
          ))}
          {(!simulationState.aircrafts || simulationState.aircrafts.length === 0) && (
            <div className="p-4 text-center text-xs italic text-[color:var(--color-shell-text-muted)]">
              No aircrafts active in simulation
            </div>
          )}
        </div>
        </AccordionCard>
      </div>
    </div>,
    portalRoot,
  );
}
