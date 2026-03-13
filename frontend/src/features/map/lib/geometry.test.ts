import { describe, expect, it } from 'vitest';

import { DEFAULT_MAP_VIEW_BOX } from '@/features/map/types';
import {
  createFocusedViewBox,
  getRenderedViewBoxMetrics,
  pointToViewBoxPercent,
  projectPointToPercent,
} from '@/features/map/lib/geometry';
import { getPlacementAnchor, normalizeLiveAirbase, normalizeSimulationAirbase } from '@/features/map/lib/placement';
import { MOCK_AIRBASES } from '@/lib/api/mock/map';
import { createMockSimulationServiceClient } from '@/lib/api/mock/simulation';

describe('geometry', () => {
  it('keeps full-map projection stable for lulea and the default mock simulation anchor', async () => {
    const liveAnchor = getPlacementAnchor(normalizeLiveAirbase(MOCK_AIRBASES[0]!));
    const simulationAirbases = await createMockSimulationServiceClient().getAirbases('base');
    const simulationAnchor = getPlacementAnchor(normalizeSimulationAirbase(simulationAirbases[0]!));

    const livePercent = pointToViewBoxPercent(liveAnchor, DEFAULT_MAP_VIEW_BOX);
    const simulationPercent = pointToViewBoxPercent(simulationAnchor, DEFAULT_MAP_VIEW_BOX);

    expect(livePercent.x).toBeCloseTo(82.7, 1);
    expect(livePercent.y).toBeCloseTo(22.5, 1);
    expect(simulationPercent.x).toBeCloseTo(31.7, 1);
    expect(simulationPercent.y).toBeCloseTo(95.0, 1);
  });

  it('computes rendered metrics for landscape and portrait containers', () => {
    const landscape = getRenderedViewBoxMetrics(DEFAULT_MAP_VIEW_BOX, { width: 1200, height: 800 });
    const portrait = getRenderedViewBoxMetrics(DEFAULT_MAP_VIEW_BOX, { width: 600, height: 1000 });

    expect(landscape).not.toBeNull();
    expect(portrait).not.toBeNull();

    expect(landscape?.renderedWidth).toBeCloseTo(348.89, 2);
    expect(landscape?.renderedHeight).toBe(800);
    expect(landscape?.offsetX).toBeCloseTo(425.56, 2);
    expect(landscape?.offsetY).toBe(0);

    expect(portrait?.renderedWidth).toBeCloseTo(436.11, 2);
    expect(portrait?.renderedHeight).toBe(1000);
    expect(portrait?.offsetX).toBeCloseTo(81.95, 2);
    expect(portrait?.offsetY).toBe(0);
  });

  it('uses rendered projection when container metrics are available', () => {
    const point = { x: 100, y: 100 };
    const projected = projectPointToPercent(point, DEFAULT_MAP_VIEW_BOX, { width: 1200, height: 800 });

    expect(projected.x).toBeCloseTo(43.88, 2);
    expect(projected.y).toBeCloseTo(12.62, 2);
  });

  it('falls back to viewBox projection for zero-size container', () => {
    const point = { x: 100, y: 100 };

    expect(projectPointToPercent(point, DEFAULT_MAP_VIEW_BOX, { width: 0, height: 0 })).toEqual(
      pointToViewBoxPercent(point, DEFAULT_MAP_VIEW_BOX),
    );
  });

  it('keeps focus clamp inside the default map bounds near the edge', () => {
    const focused = createFocusedViewBox(
      {
        minX: 1,
        maxX: 7,
        minY: 786,
        maxY: 791,
      },
      DEFAULT_MAP_VIEW_BOX,
    );

    expect(focused.minX).toBe(DEFAULT_MAP_VIEW_BOX.minX);
    expect(focused.minY).toBeGreaterThanOrEqual(DEFAULT_MAP_VIEW_BOX.minY);
    expect(focused.minX + focused.width).toBeLessThanOrEqual(
      DEFAULT_MAP_VIEW_BOX.minX + DEFAULT_MAP_VIEW_BOX.width,
    );
    expect(focused.minY + focused.height).toBeLessThanOrEqual(
      DEFAULT_MAP_VIEW_BOX.minY + DEFAULT_MAP_VIEW_BOX.height,
    );
  });
});
