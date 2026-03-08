import { SIMULATION_NEED_TYPE_OPTIONS, type SimulationSetupFormValues } from '@/features/simulation/types';
import type { SimulationFormSectionProps } from '@/features/simulation/components/shared';
import { CheckboxGroup } from '@/features/ui/components/CheckboxGroup';
import { FormField } from '@/features/ui/components/FormField';
import { Input } from '@/features/ui/components/Input';
import { Textarea } from '@/features/ui/components/Textarea';

export function NeedsAndNotesSection({ values, onUpdateField }: SimulationFormSectionProps) {
  return (
    <section className="grid gap-4">
      <FormField
        label="Need types"
        hint="Maps to FleetOptions.NeedsPool. Leave selected options broad for more diverse scenarios."
      >
        <CheckboxGroup
          name="Need types"
          options={SIMULATION_NEED_TYPE_OPTIONS}
          values={values.needsPool}
          onChange={(nextValues) =>
            onUpdateField('needsPool', nextValues as SimulationSetupFormValues['needsPool'])
          }
        />
      </FormField>

      <FormField
        label="Blocking chance"
        hint="Maps to FleetOptions.BlockingChance as a percentage of generated needs."
      >
        <div className="grid grid-cols-[minmax(0,1fr)_auto] gap-2 md:max-w-56">
          <Input
            type="number"
            min={0}
            max={100}
            value={values.blockingChancePercent}
            onChange={(event) =>
              onUpdateField('blockingChancePercent', Number(event.target.value || 0))
            }
          />
          <div className="shell-panel-soft shell-divider flex items-center rounded-md border px-3 text-sm font-medium">
            %
          </div>
        </div>
      </FormField>

      <FormField
        label="Operator notes"
        hint="Optional planning notes for this simulation setup."
      >
        <Textarea
          value={values.notes}
          onChange={(event) => onUpdateField('notes', event.target.value)}
          placeholder="Priority assumptions, branch intent, special constraints..."
        />
      </FormField>
    </section>
  );
}
