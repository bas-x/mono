import { useState, type FormEvent } from 'react';

import { FleetSection } from '@/features/simulation/components/FleetSection';
import { NeedsAndNotesSection } from '@/features/simulation/components/NeedsAndNotesSection';
import { SeedAndAirbaseSection } from '@/features/simulation/components/SeedAndAirbaseSection';
import { SimulationSetupActions } from '@/features/simulation/components/SimulationSetupActions';
import { SimulationSetupHeader } from '@/features/simulation/components/SimulationSetupHeader';
import {
  DEFAULT_SIMULATION_SETUP_FORM_VALUES,
  type SimulationSetupFormValues,
} from '@/features/simulation/types';
import { Sheet, SheetBody } from '@/features/ui/components/Sheet';

type SimulationSetupSheetProps = {
  isOpen: boolean;
  onClose: () => void;
  defaultValues?: SimulationSetupFormValues;
  onSubmit: (values: SimulationSetupFormValues) => void;
};

function clampPercent(value: number) {
  return Math.max(0, Math.min(100, value));
}

function normalizeSimulationSetupValues(
  values: SimulationSetupFormValues,
): SimulationSetupFormValues {
  return {
    ...values,
    durationSeconds: Math.max(1, values.durationSeconds),
    regionProbabilityPercent: clampPercent(values.regionProbabilityPercent),
    blockingChancePercent: clampPercent(values.blockingChancePercent),
    maxPerRegion: Math.max(values.maxPerRegion, values.minPerRegion),
    aircraftMax: Math.max(values.aircraftMax, values.aircraftMin),
    needsMax: Math.max(values.needsMax, values.needsMin),
    severityMax: Math.max(values.severityMax, values.severityMin),
  };
}

export function SimulationSetupSheet({
  isOpen,
  onClose,
  defaultValues = DEFAULT_SIMULATION_SETUP_FORM_VALUES,
  onSubmit,
}: SimulationSetupSheetProps) {
  const [values, setValues] = useState<SimulationSetupFormValues>(defaultValues);

  const handleUpdateField = <T extends keyof SimulationSetupFormValues>(
    field: T,
    value: SimulationSetupFormValues[T],
  ) => {
    setValues((current) => ({ ...current, [field]: value }));
  };

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    onSubmit(normalizeSimulationSetupValues(values));
  };

  return (
    <Sheet isOpen={isOpen} onClose={onClose} width="42rem">
      <SimulationSetupHeader onClose={onClose} />

      <form
        onSubmit={handleSubmit}
        className="min-h-0 flex flex-1 flex-col overflow-hidden border-t border-[color:var(--color-shell-border)]"
      >
        <SheetBody>
          <div className="grid gap-6">
            <SeedAndAirbaseSection values={values} onUpdateField={handleUpdateField} />
            <FleetSection values={values} onUpdateField={handleUpdateField} />
            <NeedsAndNotesSection values={values} onUpdateField={handleUpdateField} />
          </div>
        </SheetBody>

        <SimulationSetupActions onClose={onClose} />
      </form>
    </Sheet>
  );
}
