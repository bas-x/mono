import { describe, expect, it } from 'vitest';

import type { SimulationThreat } from '@/lib/api/types';

const threat: SimulationThreat = {
  id: 'threat-1',
  position: { x: 210, y: 420 },
  createdAt: '2026-03-12T03:15:05Z',
  createdTick: 4,
};

describe('ThreatOverlayLayer fixture contract', () => {
  it('keeps threat positions available for map projection', () => {
    expect(threat.position).toEqual({ x: 210, y: 420 });
    expect(threat.id).toBe('threat-1');
  });
});
