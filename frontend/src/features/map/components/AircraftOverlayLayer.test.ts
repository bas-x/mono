import { describe, expect, it } from 'vitest';

import { getAircraftStrokeColor } from '@/features/map/components/AircraftOverlayLayer';

describe('AircraftOverlayLayer state colors', () => {
  it('maps the main aircraft states to distinct high-contrast stroke colors', () => {
    expect(getAircraftStrokeColor('Inbound')).toBe('#2563eb');
    expect(getAircraftStrokeColor('Landing')).toBe('#f59e0b');
    expect(getAircraftStrokeColor('Servicing')).toBe('#d946ef');
    expect(getAircraftStrokeColor('Ready')).toBe('#16a34a');
  });

  it('supports the additional mock scenario states used in the UI', () => {
    expect(getAircraftStrokeColor('Turnaround')).toBe('#06b6d4');
    expect(getAircraftStrokeColor('Holding')).toBe('#7c3aed');
    expect(getAircraftStrokeColor('Assessment')).toBe('#f97316');
    expect(getAircraftStrokeColor('Repair')).toBe('#dc2626');
    expect(getAircraftStrokeColor('Taxi')).toBe('#eab308');
  });

  it('normalizes state casing and falls back for unknown states', () => {
    expect(getAircraftStrokeColor('  inbound  ')).toBe('#2563eb');
    expect(getAircraftStrokeColor('Unknown')).toBe('#94a3b8');
  });
});
