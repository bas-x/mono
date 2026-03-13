import { describe, expect, it } from 'vitest';

import { calculatePolygonBounds, calculatePolygonCentroid } from '@/features/map/lib/geometry';
import {
  getPlacementAnchor,
  getPlacementBounds,
  normalizeLiveAirbase,
  normalizeSimulationAirbase,
  resolveActivePlacementSources,
} from '@/features/map/lib/placement';
import { MOCK_AIRBASES } from '@/lib/api/mock/map';
import { createMockSimulationServiceClient } from '@/lib/api/mock/simulation';

describe('placement', () => {
  it('uses centroid math for polygon-backed anchors', () => {
    const source = normalizeLiveAirbase({
      id: 'skewed-live-base',
      area: [
        { x: 0, y: 0 },
        { x: 8, y: 0 },
        { x: 6, y: 8 },
        { x: 1, y: 6 },
      ],
    });
    const anchor = getPlacementAnchor(source);
    const centroid = calculatePolygonCentroid(source.area);
    const bounds = calculatePolygonBounds(source.area);
    const boundsCenter = {
      x: (bounds.minX + bounds.maxX) / 2,
      y: (bounds.minY + bounds.maxY) / 2,
    };

    expect(anchor).toEqual(centroid);
    expect(anchor.y).not.toBeCloseTo(boundsCenter.y, 4);
  });

  it('uses raw point anchors for point-backed simulation airbases', async () => {
    const simulationAirbases = await createMockSimulationServiceClient().getAirbases('base');
    const source = normalizeSimulationAirbase(simulationAirbases[0]!);

    expect(getPlacementAnchor(source)).toEqual({
      x: 109.44765839799018,
      y: 753.1689567645848,
    });
  });

  it('uses a deterministic 12x12 footprint for point-backed bounds', async () => {
    const simulationAirbases = await createMockSimulationServiceClient().getAirbases('base');
    const source = normalizeSimulationAirbase(simulationAirbases[0]!);

    expect(getPlacementBounds(source)).toEqual({
      minX: 103.44765839799018,
      minY: 747.1689567645848,
      maxX: 115.44765839799018,
      maxY: 759.1689567645848,
    });
  });

  it('keeps active mode source resolution mode-correct', async () => {
    const simulationAirbases = await createMockSimulationServiceClient().getAirbases('base');

    const simulateSources = resolveActivePlacementSources({
      viewMode: 'simulate',
      liveAirbases: MOCK_AIRBASES,
      simulationAirbases,
      hasRunningSimulation: true,
    });
    const liveSources = resolveActivePlacementSources({
      viewMode: 'live',
      liveAirbases: MOCK_AIRBASES,
      simulationAirbases,
      hasRunningSimulation: true,
    });

    expect(simulateSources[0]).toMatchObject({
      id: 'd397eeeddbfae33e',
      point: { x: 109.44765839799018, y: 753.1689567645848 },
    });
    expect(simulateSources[0]).not.toHaveProperty('area');
    expect(liveSources[0]).toMatchObject({ id: 'lulea' });
    expect(liveSources[0]).toHaveProperty('area');
  });

  it('falls back to live sources when simulate mode has no running simulation', async () => {
    const simulationAirbases = await createMockSimulationServiceClient().getAirbases('base');

    const sources = resolveActivePlacementSources({
      viewMode: 'simulate',
      liveAirbases: MOCK_AIRBASES,
      simulationAirbases,
      hasRunningSimulation: false,
    });

    expect(sources[0]).toMatchObject({ id: 'lulea' });
    expect(sources[0]).toHaveProperty('area');
    expect(sources[0]).not.toHaveProperty('point');
  });
});
