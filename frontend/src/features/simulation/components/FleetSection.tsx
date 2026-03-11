import { NumberField } from '@/features/ui/components/NumberField';

import type { SimulationFormSectionProps } from '@/features/simulation/components/shared';

export function FleetSection({ values, onUpdateField }: SimulationFormSectionProps) {
  return (
    <section className="shell-divider grid gap-4 border-t pt-5 md:grid-cols-2">
      <NumberField
        label="Min aircraft"
        min={0}
        value={values.aircraftMin}
        onChange={(event) => onUpdateField('aircraftMin', Number(event.target.value || 0))}
      />
      <NumberField
        label="Max aircraft"
        min={0}
        value={values.aircraftMax}
        onChange={(event) => onUpdateField('aircraftMax', Number(event.target.value || 0))}
      />
      <NumberField
        label="Min needs / aircraft"
        min={0}
        value={values.needsMin}
        onChange={(event) => onUpdateField('needsMin', Number(event.target.value || 0))}
      />
      <NumberField
        label="Max needs / aircraft"
        min={0}
        value={values.needsMax}
        onChange={(event) => onUpdateField('needsMax', Number(event.target.value || 0))}
      />
      <NumberField
        label="Min severity"
        min={0}
        max={100}
        value={values.severityMin}
        onChange={(event) => onUpdateField('severityMin', Number(event.target.value || 0))}
      />
      <NumberField
        label="Max severity"
        min={0}
        max={100}
        value={values.severityMax}
        onChange={(event) => onUpdateField('severityMax', Number(event.target.value || 0))}
      />
    </section>
  );
}
