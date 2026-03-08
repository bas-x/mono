export type CheckboxOption = {
  label: string;
  value: string;
  description?: string;
};

type CheckboxGroupProps = {
  options: CheckboxOption[];
  values: string[];
  onChange: (nextValues: string[]) => void;
  name: string;
};

function toggleValue(values: string[], value: string) {
  if (values.includes(value)) {
    return values.filter((entry) => entry !== value);
  }

  return [...values, value];
}

export function CheckboxGroup({ options, values, onChange, name }: CheckboxGroupProps) {
  return (
    <div className="grid gap-2 sm:grid-cols-2" role="group" aria-label={name}>
      {options.map((option) => {
        const isSelected = values.includes(option.value);

        return (
          <button
            key={option.value}
            type="button"
            role="checkbox"
            aria-checked={isSelected}
            onClick={() => onChange(toggleValue(values, option.value))}
            className={`shell-check-option cursor-pointer rounded-lg border px-3 py-3 text-left transition-colors ${
              isSelected ? 'shell-check-option-selected' : ''
            }`}
          >
            <div className="flex items-start justify-between gap-3">
              <div className="space-y-1">
                <span className="block text-sm font-medium">{option.label}</span>
                {option.description ? (
                  <span className="shell-field-hint block text-xs">{option.description}</span>
                ) : null}
              </div>
              <span
                aria-hidden="true"
                className={`mt-0.5 size-3 rounded-full border ${
                  isSelected ? 'shell-check-indicator-selected' : 'shell-check-indicator'
                }`}
              />
            </div>
          </button>
        );
      })}
    </div>
  );
}
