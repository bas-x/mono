import {
  type Airbase,
  type AirbasePlacementSource,
  type AirbasePoint,
  type PolygonBackedAirbasePlacement,
  type PointBackedAirbasePlacement,
} from '@/features/map/types';
import type { SimulationAirbase } from '@/lib/api/types';
import {
  calculatePolygonBounds,
  calculatePolygonCentroid,
  type PolygonBounds,
} from '@/features/map/lib/geometry';

const POINT_BACKED_HALF_FOOTPRINT = 6;

type ActivePlacementMode = 'live' | 'simulate';

type ResolveActivePlacementSourcesOptions = {
  viewMode: ActivePlacementMode;
  liveAirbases: Airbase[];
  simulationAirbases: SimulationAirbase[];
  hasRunningSimulation: boolean;
};

function isPointBackedPlacement(
  source: AirbasePlacementSource,
): source is PointBackedAirbasePlacement {
  return 'point' in source;
}

export function normalizeLiveAirbase(airbase: Airbase): PolygonBackedAirbasePlacement {
  return {
    id: airbase.id,
    name: airbase.name,
    area: airbase.area,
  };
}

export function normalizeSimulationAirbase(
  airbase: SimulationAirbase,
): PointBackedAirbasePlacement {
  return {
    id: airbase.id,
    name: airbase.name,
    point: airbase.location,
    regionId: airbase.regionId,
    region: airbase.region,
  };
}

export function resolveActivePlacementSources({
  viewMode,
  liveAirbases,
  simulationAirbases,
  hasRunningSimulation,
}: ResolveActivePlacementSourcesOptions): AirbasePlacementSource[] {
  if (viewMode === 'simulate' && hasRunningSimulation) {
    return simulationAirbases.map(normalizeSimulationAirbase);
  }

  return liveAirbases.map(normalizeLiveAirbase);
}

export function getPlacementAnchor(source: AirbasePlacementSource): AirbasePoint {
  if (isPointBackedPlacement(source)) {
    return source.point;
  }

  return calculatePolygonCentroid(source.area);
}

export function getPlacementBounds(source: AirbasePlacementSource): PolygonBounds {
  if (isPointBackedPlacement(source)) {
    return {
      minX: source.point.x - POINT_BACKED_HALF_FOOTPRINT,
      minY: source.point.y - POINT_BACKED_HALF_FOOTPRINT,
      maxX: source.point.x + POINT_BACKED_HALF_FOOTPRINT,
      maxY: source.point.y + POINT_BACKED_HALF_FOOTPRINT,
    };
  }

  return calculatePolygonBounds(source.area);
}
