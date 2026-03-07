import type { Airbase, AirbasePoint, MapViewBox } from '@/features/map/types';

export type PolygonBounds = {
  minX: number;
  minY: number;
  maxX: number;
  maxY: number;
};

export function hasValidPolygon(points: AirbasePoint[]): boolean {
  return points.length >= 3;
}

export function polygonToPointsAttribute(points: AirbasePoint[]): string {
  return points.map((point) => `${point.x},${point.y}`).join(' ');
}

export function calculatePolygonBounds(points: AirbasePoint[]): PolygonBounds {
  let minX = Number.POSITIVE_INFINITY;
  let minY = Number.POSITIVE_INFINITY;
  let maxX = Number.NEGATIVE_INFINITY;
  let maxY = Number.NEGATIVE_INFINITY;

  for (const point of points) {
    minX = Math.min(minX, point.x);
    minY = Math.min(minY, point.y);
    maxX = Math.max(maxX, point.x);
    maxY = Math.max(maxY, point.y);
  }

  return { minX, minY, maxX, maxY };
}

function averagePoint(points: AirbasePoint[]): AirbasePoint {
  let sumX = 0;
  let sumY = 0;

  for (const point of points) {
    sumX += point.x;
    sumY += point.y;
  }

  return {
    x: sumX / points.length,
    y: sumY / points.length,
  };
}

export function calculatePolygonCentroid(points: AirbasePoint[]): AirbasePoint {
  if (points.length === 0) {
    return { x: 0, y: 0 };
  }

  let signedArea = 0;
  let centroidX = 0;
  let centroidY = 0;

  for (let index = 0; index < points.length; index += 1) {
    const current = points[index];
    const next = points[(index + 1) % points.length];
    if (!current || !next) {
      continue;
    }

    const cross = current.x * next.y - next.x * current.y;
    signedArea += cross;
    centroidX += (current.x + next.x) * cross;
    centroidY += (current.y + next.y) * cross;
  }

  const area = signedArea / 2;
  if (Math.abs(area) < 1e-6) {
    return averagePoint(points);
  }

  return {
    x: centroidX / (6 * area),
    y: centroidY / (6 * area),
  };
}

export function pointToViewBoxPercent(point: AirbasePoint, viewBox: MapViewBox) {
  return {
    x: ((point.x - viewBox.minX) / viewBox.width) * 100,
    y: ((point.y - viewBox.minY) / viewBox.height) * 100,
  };
}

export function toAriaLabel(airbase: Pick<Airbase, 'id'>): string {
  return `Airbase ${airbase.id}`;
}
