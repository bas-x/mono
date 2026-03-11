import { SheetFooter } from '@/features/ui/components/Sheet';

type SimulationSetupActionsProps = {
  onClose: () => void;
};

export function SimulationSetupActions({ onClose }: SimulationSetupActionsProps) {
  return (
    <SheetFooter>
      <div className="flex items-center justify-end gap-2">
        <button
          type="button"
          onClick={onClose}
          className="shell-button cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          className="shell-button-active cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors"
        >
          Create simulation
        </button>
      </div>
    </SheetFooter>
  );
}
