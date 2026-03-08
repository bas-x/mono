import { SheetHeader } from '@/features/ui/components/Sheet';

type SimulationSetupHeaderProps = {
  onClose: () => void;
};

export function SimulationSetupHeader({ onClose }: SimulationSetupHeaderProps) {
  return (
    <SheetHeader>
      <div className="flex items-center justify-between gap-4">
        <div className="space-y-1">
          <p className="shell-text-muted m-0 text-[0.68rem] font-semibold uppercase tracking-[0.22em]">
            Simulate
          </p>
          <h2 className="m-0 text-xl font-semibold text-[color:var(--color-shell-text)]">
            Create simulation run
          </h2>
          <p className="shell-field-hint m-0 max-w-2xl text-sm">
            Parameters mirror the backend simulation setup: seed, airbase generation, fleet sizing,
            needs pool, and constraint severity.
          </p>
        </div>

        <button
          type="button"
          onClick={onClose}
          className="shell-button cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors"
        >
          Close
        </button>
      </div>
    </SheetHeader>
  );
}
