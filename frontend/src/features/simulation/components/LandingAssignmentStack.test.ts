import { describe, expect, it } from 'vitest';

import {
  getTopNeedCapabilities,
  getLandingAssignmentPrimaryCapability,
  getLandingAssignmentTone,
} from '@/features/simulation/components/LandingAssignmentStack';
import type { LandingAssignmentEvent, SimulationAirbase, SimulationAircraft } from '@/lib/api/types';

describe('LandingAssignmentStack helpers', () => {
  it('prefers the first required capability from needs for the card tone', () => {
    const event: LandingAssignmentEvent = {
      type: 'landing_assignment',
      simulationId: 'base',
      tailNumber: 'BX-101',
      baseId: 'base-a',
      source: 'algorithm',
      tick: 1,
      timestamp: '2026-03-12T03:15:05Z',
      needs: [
        { type: 'fuel', severity: 42, requiredCapability: 'fuel', blocking: false },
      ],
      capabilities: {
        fuel: { recoveryMultiplierPermille: 1300 },
      },
    };

    expect(getLandingAssignmentPrimaryCapability(event)).toBe('fuel');
    expect(getLandingAssignmentTone(event).cardClass).toContain('cyan');
  });

  it('falls back to capability keys when needs are empty', () => {
    const event: LandingAssignmentEvent = {
      type: 'landing_assignment',
      simulationId: 'base',
      tailNumber: 'BX-101',
      baseId: 'base-a',
      source: 'algorithm',
      tick: 1,
      timestamp: '2026-03-12T03:15:05Z',
      needs: [],
      capabilities: {
        repairs: { recoveryMultiplierPermille: 1400 },
      },
    };

    expect(getLandingAssignmentPrimaryCapability(event)).toBe('repairs');
    expect(getLandingAssignmentTone(event).badgeClass).toContain('orange');
  });

  it('returns the top 3 unique capabilities ranked by severity', () => {
    const event: LandingAssignmentEvent = {
      type: 'landing_assignment',
      simulationId: 'base',
      tailNumber: 'BX-101',
      baseId: 'base-a',
      source: 'algorithm',
      tick: 1,
      timestamp: '2026-03-12T03:15:05Z',
      needs: [
        { type: 'fuel', severity: 42, requiredCapability: 'fuel', blocking: false },
        { type: 'munitions', severity: 54, requiredCapability: 'munitions', blocking: false },
        { type: 'repairs', severity: 61, requiredCapability: 'repairs', blocking: true },
        { type: 'crew_support', severity: 18, requiredCapability: 'crew_support', blocking: false },
        { type: 'fuel', severity: 30, requiredCapability: 'fuel', blocking: false },
      ],
      capabilities: {
        fuel: { recoveryMultiplierPermille: 1300 },
        munitions: { recoveryMultiplierPermille: 1200 },
        repairs: { recoveryMultiplierPermille: 1400 },
        crew_support: { recoveryMultiplierPermille: 1100 },
      },
    };

    expect(getTopNeedCapabilities(event)).toEqual([
      { capability: 'repairs', severity: 61 },
      { capability: 'munitions', severity: 54 },
      { capability: 'fuel', severity: 42 },
    ]);
  });

  it('can resolve model and name-backed labels from the contract fields', () => {
    const aircrafts: SimulationAircraft[] = [
      {
        tailNumber: 'BX-101',
        model: 'Falcon HX-12',
        needs: [],
        state: 'Inbound',
      },
    ];
    const airbases: SimulationAirbase[] = [
      {
        id: 'base-a',
        name: 'Blekinge Forward Strip',
        location: { x: 0, y: 0 },
        regionId: 'SE-K',
        region: 'Blekinge',
      },
    ];

    expect(aircrafts[0]?.model).toBe('Falcon HX-12');
    expect(airbases[0]?.name).toBe('Blekinge Forward Strip');
  });
});
