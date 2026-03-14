import { useEffect, useMemo, useState } from 'react';

import {
  formatTerminalSummaryDuration,
  formatTerminalSummaryHeadline,
  type TerminalSimulationRecord,
} from '@/features/simulation/hooks/useSimulation';
import { Sheet, SheetBody, SheetFooter, SheetHeader } from '@/features/ui/components/Sheet';

type TerminalSummarySheetProps = {
  record: TerminalSimulationRecord | null;
  records: TerminalSimulationRecord[];
  onClose: () => void;
};

function getOutcomeBadge(record: TerminalSimulationRecord) {
  if (record.kind === 'ended') {
    return 'Completed';
  }

  return record.reason === 'reset' ? 'Reset' : 'Stopped';
}

function getSimulationLabel(simulationId: string) {
  return simulationId === 'base' ? 'Base run' : `Branch ${simulationId.slice(0, 8)}`;
}

function getSimulationCaption(record: TerminalSimulationRecord) {
  if (record.simulationId === 'base') {
    return 'Primary run';
  }

  return record.reason === 'cancel' ? 'Alternate branch' : 'Derived run';
}

function getOutcomeTone(record: TerminalSimulationRecord) {
  if (record.kind === 'ended') {
    return 'border-emerald-500/25 bg-emerald-500/10 text-emerald-200';
  }

  return record.reason === 'reset'
    ? 'border-amber-500/25 bg-amber-500/10 text-amber-200'
    : 'border-sky-500/25 bg-sky-500/10 text-sky-200';
}

function getSummaryIntro(record: TerminalSimulationRecord | null, count: number) {
  if (!record) {
    return 'Terminal summaries appear here when a run or branch reaches its end state.';
  }

  if (count <= 1) {
    return `${getSimulationLabel(record.simulationId)} has finished. Review the servicing outcome below.`;
  }

  return `${getSimulationLabel(record.simulationId)} just finished. Compare it with ${count - 1} other recorded run summaries below.`;
}

export function TerminalSummarySheet({ record, records, onClose }: TerminalSummarySheetProps) {
  const featuredRecord = record ?? records[0] ?? null;
  const featuredRecordKey = useMemo(() => {
    if (!featuredRecord) {
      return null;
    }

    return `${featuredRecord.simulationId}:${featuredRecord.timestamp}:${featuredRecord.tick}:${featuredRecord.kind}:${featuredRecord.reason ?? 'none'}`;
  }, [featuredRecord]);
  const [isOpen, setIsOpen] = useState(featuredRecord != null);

  useEffect(() => {
    if (featuredRecordKey) {
      setIsOpen(true);
    }
  }, [featuredRecordKey]);

  const handleClose = () => {
    setIsOpen(false);
    onClose();
  };

  return (
    <Sheet isOpen={isOpen && featuredRecord != null} onClose={handleClose} width="68rem">
      <SheetHeader>
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <p className="shell-text-muted m-0 text-[0.68rem] font-semibold uppercase tracking-[0.22em]">
              Run summaries
            </p>
            <h2 className="m-0 text-xl font-semibold text-[color:var(--color-shell-text)]">
              Completed run overview
            </h2>
            <p className="shell-field-hint m-0 max-w-3xl text-sm">
              {getSummaryIntro(featuredRecord, records.length)}
            </p>
          </div>

          <button
            type="button"
            onClick={handleClose}
            className="shell-button cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors"
          >
            Close
          </button>
        </div>
      </SheetHeader>

      {featuredRecord ? (
        <>
          <SheetBody>
            <div className="grid gap-6">
              <div className="grid gap-4 rounded-2xl border border-[color:var(--color-shell-border)] bg-[color:var(--color-shell-panel-soft)] p-5 lg:grid-cols-[1.4fr_0.9fr]">
                <div className="space-y-4">
                  <div className="space-y-1">
                    <div className="text-[10px] font-semibold uppercase tracking-[0.22em] text-[color:var(--color-shell-text-muted)]">
                      Latest outcome
                    </div>
                    <div className="text-2xl font-semibold text-[color:var(--color-shell-text)]">
                      {formatTerminalSummaryHeadline(featuredRecord)}
                    </div>
                    <div className="text-sm text-[color:var(--color-shell-text-muted)]">
                      {getSimulationLabel(featuredRecord.simulationId)} reached a terminal state at tick {featuredRecord.tick}.
                    </div>
                  </div>

                  <div className="grid gap-3 sm:grid-cols-3">
                    <div className="rounded-xl border border-[color:var(--color-shell-border)] bg-black/10 p-4">
                      <div className="text-[10px] font-semibold uppercase tracking-[0.22em] text-[color:var(--color-shell-text-muted)]">
                        Completed services
                      </div>
                      <div className="mt-2 font-mono text-2xl font-semibold text-[color:var(--color-shell-text)]">
                        {featuredRecord.summary.completedVisitCount}
                      </div>
                      <div className="mt-1 text-xs text-[color:var(--color-shell-text-muted)]">
                        Visits fully completed before the run stopped.
                      </div>
                    </div>

                    <div className="rounded-xl border border-[color:var(--color-shell-border)] bg-black/10 p-4">
                      <div className="text-[10px] font-semibold uppercase tracking-[0.22em] text-[color:var(--color-shell-text-muted)]">
                        Total service time
                      </div>
                      <div className="mt-2 font-mono text-2xl font-semibold text-[color:var(--color-shell-text)]">
                        {formatTerminalSummaryDuration(featuredRecord.summary.totalDurationMs)}
                      </div>
                      <div className="mt-1 text-xs text-[color:var(--color-shell-text-muted)]">
                        Combined servicing duration across completed visits.
                      </div>
                    </div>

                    <div className="rounded-xl border border-[color:var(--color-shell-border)] bg-black/10 p-4">
                      <div className="text-[10px] font-semibold uppercase tracking-[0.22em] text-[color:var(--color-shell-text-muted)]">
                        Average service time
                      </div>
                      <div className="mt-2 font-mono text-2xl font-semibold text-[color:var(--color-shell-text)]">
                        {formatTerminalSummaryDuration(featuredRecord.summary.averageDurationMs)}
                      </div>
                      <div className="mt-1 text-xs text-[color:var(--color-shell-text-muted)]">
                        Shows once at least one service is fully completed.
                      </div>
                    </div>
                  </div>
                </div>

                <div className="flex flex-col gap-3 rounded-xl border border-[color:var(--color-shell-border)] bg-black/10 p-4">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <div className="text-[10px] font-semibold uppercase tracking-[0.22em] text-[color:var(--color-shell-text-muted)]">
                        Run details
                      </div>
                      <div className="mt-1 text-base font-semibold text-[color:var(--color-shell-text)]">
                        {getSimulationLabel(featuredRecord.simulationId)}
                      </div>
                    </div>
                    <div className={`rounded-full border px-3 py-1 text-[10px] font-semibold uppercase tracking-[0.22em] ${getOutcomeTone(featuredRecord)}`}>
                      {getOutcomeBadge(featuredRecord)}
                    </div>
                  </div>

                  <div className="grid gap-3 sm:grid-cols-2">
                    <div>
                      <div className="text-[10px] font-semibold uppercase tracking-[0.22em] text-[color:var(--color-shell-text-muted)]">
                        Run type
                      </div>
                      <div className="mt-1 text-sm text-[color:var(--color-shell-text)]">
                        {getSimulationCaption(featuredRecord)}
                      </div>
                    </div>
                    <div>
                      <div className="text-[10px] font-semibold uppercase tracking-[0.22em] text-[color:var(--color-shell-text-muted)]">
                        Recorded at
                      </div>
                      <div className="mt-1 font-mono text-sm text-[color:var(--color-shell-text)]">
                        {new Date(featuredRecord.timestamp).toLocaleTimeString()}
                      </div>
                    </div>
                  </div>

                  <div className="rounded-lg border border-[color:var(--color-shell-border)] bg-[color:var(--color-shell-panel-bg)] p-3 text-sm text-[color:var(--color-shell-text-muted)]">
                    Use the comparison cards below to review how each branch finished and how servicing performance differed.
                  </div>
                </div>
              </div>

              <div className="space-y-3">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <div className="text-[10px] font-semibold uppercase tracking-[0.22em] text-[color:var(--color-shell-text-muted)]">
                      Run comparison
                    </div>
                    <div className="mt-1 text-sm text-[color:var(--color-shell-text-muted)]">
                      Compare all recorded base and branch summaries currently available in this session.
                    </div>
                  </div>
                  <div className="rounded-full border border-[color:var(--color-shell-button-border)] bg-[color:var(--color-shell-button-bg)] px-3 py-1 text-[10px] font-semibold uppercase tracking-[0.22em] text-[color:var(--color-primary)]">
                    {records.length} recorded summaries
                  </div>
                </div>

                <div className="grid gap-3 lg:grid-cols-2">
                  {records.map((summaryRecord) => {
                    const isFeatured = summaryRecord.simulationId === featuredRecord.simulationId
                      && summaryRecord.timestamp === featuredRecord.timestamp
                      && summaryRecord.tick === featuredRecord.tick;

                    return (
                      <section
                        key={`${summaryRecord.simulationId}:${summaryRecord.timestamp}:${summaryRecord.tick}`}
                        className={`rounded-2xl border p-4 transition-colors ${
                          isFeatured
                            ? 'border-[color:var(--color-primary)]/35 bg-[color:var(--color-shell-panel-soft)] shadow-[0_0_0_1px_rgba(217,119,6,0.12)]'
                            : 'border-[color:var(--color-shell-border)] bg-[color:var(--color-shell-panel-bg)]'
                        }`}
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div className="space-y-1">
                            <div className="text-[10px] font-semibold uppercase tracking-[0.22em] text-[color:var(--color-shell-text-muted)]">
                              {getSimulationCaption(summaryRecord)}
                            </div>
                            <div className="text-lg font-semibold text-[color:var(--color-shell-text)]">
                              {getSimulationLabel(summaryRecord.simulationId)}
                            </div>
                            <div className="text-sm text-[color:var(--color-shell-text-muted)]">
                              {formatTerminalSummaryHeadline(summaryRecord)}
                            </div>
                          </div>

                          <div className={`rounded-full border px-3 py-1 text-[10px] font-semibold uppercase tracking-[0.22em] ${getOutcomeTone(summaryRecord)}`}>
                            {getOutcomeBadge(summaryRecord)}
                          </div>
                        </div>

                        <div className="mt-4 grid gap-3 sm:grid-cols-3">
                          <div>
                            <div className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[color:var(--color-shell-text-muted)]">
                              Completed services
                            </div>
                            <div className="mt-1 font-mono text-lg font-semibold text-[color:var(--color-shell-text)]">
                              {summaryRecord.summary.completedVisitCount}
                            </div>
                          </div>
                          <div>
                            <div className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[color:var(--color-shell-text-muted)]">
                              Total time
                            </div>
                            <div className="mt-1 font-mono text-lg font-semibold text-[color:var(--color-shell-text)]">
                              {formatTerminalSummaryDuration(summaryRecord.summary.totalDurationMs)}
                            </div>
                          </div>
                          <div>
                            <div className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[color:var(--color-shell-text-muted)]">
                              Average time
                            </div>
                            <div className="mt-1 font-mono text-lg font-semibold text-[color:var(--color-shell-text)]">
                              {formatTerminalSummaryDuration(summaryRecord.summary.averageDurationMs)}
                            </div>
                          </div>
                        </div>

                        <div className="mt-4 flex flex-wrap gap-4 text-xs text-[color:var(--color-shell-text-muted)]">
                          <span>Tick {summaryRecord.tick}</span>
                          <span>{new Date(summaryRecord.timestamp).toLocaleTimeString()}</span>
                        </div>
                      </section>
                    );
                  })}
                </div>
              </div>
            </div>
          </SheetBody>

          <SheetFooter>
            <div className="flex items-center justify-between gap-4">
              <p className="m-0 text-sm text-[color:var(--color-shell-text-muted)]">
                These results come from terminal websocket events captured during this session for each run and branch.
              </p>
              <button
                type="button"
                onClick={handleClose}
                className="shell-button cursor-pointer rounded-sm border px-4 py-2 text-sm font-medium"
              >
                Done
              </button>
            </div>
          </SheetFooter>
        </>
      ) : null}
    </Sheet>
  );
}
