import { SIMULATION_TICKS_PER_SECOND } from '@/lib/api/types';

export const SIMULATION_NEED_TYPE_OPTIONS = [
  { value: 'fuel', label: 'Fuel', description: 'Refuel and energy turnaround.' },
  { value: 'charge', label: 'Charge', description: 'Battery or power replenishment.' },
  { value: 'munitions', label: 'Munitions', description: 'Rearm and weapons loading.' },
  { value: 'repairs', label: 'Repairs', description: 'Damage or defect correction.' },
  { value: 'maintenance', label: 'Maintenance', description: 'Routine mechanical service.' },
  {
    value: 'mission_configuration',
    label: 'Mission Config',
    description: 'Mission package and configuration changes.',
  },
  { value: 'crew_support', label: 'Crew Support', description: 'Crew swap and readiness support.' },
  { value: 'emergency', label: 'Emergency', description: 'Urgent recovery actions.' },
  {
    value: 'weather_constraint',
    label: 'Weather Constraint',
    description: 'Weather-driven handling constraints.',
  },
  {
    value: 'ground_support',
    label: 'Ground Support',
    description: 'Ground crew and service equipment needs.',
  },
  {
    value: 'protection',
    label: 'Protection',
    description: 'Base security and protective measures.',
  },
] as const;

export type SimulationNeedType = (typeof SIMULATION_NEED_TYPE_OPTIONS)[number]['value'];

export type SimulationSetupFormValues = {
  seedHex: string;
  durationSeconds: number;
  includeRegions: string;
  excludeRegions: string;
  minPerRegion: number;
  maxPerRegion: number;
  maxTotal: number;
  regionProbabilityPercent: number;
  aircraftMin: number;
  aircraftMax: number;
  needsMin: number;
  needsMax: number;
  needsPool: SimulationNeedType[];
  severityMin: number;
  severityMax: number;
  blockingChancePercent: number;
  notes: string;
};

export const DEFAULT_SIMULATION_SETUP_FORM_VALUES: SimulationSetupFormValues = {
  seedHex: '',
  durationSeconds: 100,
  includeRegions: '',
  excludeRegions: '',
  minPerRegion: 1,
  maxPerRegion: 2,
  maxTotal: 6,
  regionProbabilityPercent: 100,
  aircraftMin: 3,
  aircraftMax: 6,
  needsMin: 1,
  needsMax: 3,
  needsPool: SIMULATION_NEED_TYPE_OPTIONS.map((option) => option.value),
  severityMin: 20,
  severityMax: 80,
  blockingChancePercent: 25,
  notes: '',
};

export function durationSecondsToTicks(durationSeconds: number): number {
  return Math.max(1, Math.round(durationSeconds * SIMULATION_TICKS_PER_SECOND));
}

export function ticksToDurationSeconds(ticks: number): number {
  return ticks / SIMULATION_TICKS_PER_SECOND;
}
