import type { SimulationSetupFormValues } from '@/features/simulation/types';

export type SimulationFormSectionProps = {
  values: SimulationSetupFormValues;
  onUpdateField: <T extends keyof SimulationSetupFormValues>(
    field: T,
    value: SimulationSetupFormValues[T],
  ) => void;
};
