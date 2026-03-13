import { FormField } from '@/features/ui/components/FormField';
import { Input } from '@/features/ui/components/Input';
import { NumberField } from '@/features/ui/components/NumberField';

import type { SimulationFormSectionProps } from '@/features/simulation/components/shared';
import { durationSecondsToTicks } from '@/features/simulation/types';

export function SeedAndAirbaseSection({ values, onUpdateField }: SimulationFormSectionProps) {
  return (
    <section className="grid gap-6">
      <div className="grid gap-4 md:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)]">
        <FormField
          label="Seed"
          hint="Optional 32-byte seed in hex. Leave blank to let the backend default it."
        >
          <Input
            value={values.seedHex}
            onChange={(event) => onUpdateField('seedHex', event.target.value)}
            placeholder="001122..."
          />
        </FormField>

        <FormField
          label="Simulation duration"
          hint={`Sent to the backend as ticks at 64 ticks/sec. ${values.durationSeconds}s = ${durationSecondsToTicks(values.durationSeconds)} ticks.`}
        >
          <Input
            type="number"
            min={1}
            value={values.durationSeconds}
            onChange={(event) =>
              onUpdateField('durationSeconds', Number(event.target.value || 1))
            }
          />
        </FormField>

        <FormField
          label="Region probability"
          hint="Maps to ConstellationOptions.RegionProbability as a percentage."
        >
          <div className="grid grid-cols-[minmax(0,1fr)_auto] gap-2">
            <Input
              type="number"
              min={0}
              max={100}
              value={values.regionProbabilityPercent}
              onChange={(event) =>
                onUpdateField('regionProbabilityPercent', Number(event.target.value || 0))
              }
            />
            <div className="shell-panel-soft shell-divider flex items-center rounded-md border px-3 text-sm font-medium">
              %
            </div>
          </div>
        </FormField>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <FormField
          label="Include regions"
          hint="Comma-separated region names. Empty means all regions are eligible."
        >
          <Input
            value={values.includeRegions}
            onChange={(event) => onUpdateField('includeRegions', event.target.value)}
            placeholder="Norrbotten, Gotland"
          />
        </FormField>

        <FormField
          label="Exclude regions"
          hint="Comma-separated region names to remove from candidate generation."
        >
          <Input
            value={values.excludeRegions}
            onChange={(event) => onUpdateField('excludeRegions', event.target.value)}
            placeholder="Skane"
          />
        </FormField>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <NumberField
          label="Min bases / region"
          min={0}
          value={values.minPerRegion}
          onChange={(event) => onUpdateField('minPerRegion', Number(event.target.value || 0))}
        />
        <NumberField
          label="Max bases / region"
          min={0}
          value={values.maxPerRegion}
          onChange={(event) => onUpdateField('maxPerRegion', Number(event.target.value || 0))}
        />
        <NumberField
          label="Max total bases"
          min={1}
          value={values.maxTotal}
          onChange={(event) => onUpdateField('maxTotal', Number(event.target.value || 1))}
        />
      </div>
    </section>
  );
}
