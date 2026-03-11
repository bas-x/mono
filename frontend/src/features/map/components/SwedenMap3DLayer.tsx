import { memo, useMemo } from 'react';

import boundsData from '@backend-assets/bounds.json';
import regionsData from '@backend-assets/sweden.json';

import type { MapViewBox } from '@/features/map/types';

type MapPoint = {
  x: number;
  y: number;
};

type SwedenMap3DLayerProps = {
  viewBox: MapViewBox;
};

type ExtrudedFace = {
  key: string;
  points: string;
};

type CountryMesh = {
  topPolygons: string[];
  sideFaces: ExtrudedFace[];
  shadowPolygons: string[];
};

type SwedenBounds = {
  overall: {
    min: MapPoint;
    max: MapPoint;
    width: number;
    height: number;
  };
};

type SwedenRegion = {
  id: string;
  name: string;
  areas: MapPoint[][];
};

const EXTRUSION_DEPTH = { x: 8, y: 12 } as const;
const TOP_FILL = 'color-mix(in srgb, var(--color-map-surface) 88%, white 12%)';
const SIDE_FILL = 'color-mix(in srgb, var(--color-map-boundary) 70%, var(--color-map-surface) 30%)';
const SHADOW_FILL = 'color-mix(in srgb, var(--color-map-boundary) 28%, transparent)';

const SWEDEN_MAP_BOUNDS = (boundsData as SwedenBounds).overall;
const SWEDEN_REGIONS = regionsData as SwedenRegion[];

function toPointsAttribute(points: readonly MapPoint[]) {
  return points.map((point) => `${point.x},${point.y}`).join(' ');
}

function offsetPoints(points: readonly MapPoint[]) {
  return points.map((point) => ({
    x: point.x + EXTRUSION_DEPTH.x,
    y: point.y + EXTRUSION_DEPTH.y,
  }));
}

function createSideFaces(idPrefix: string, areaIndex: number, points: readonly MapPoint[]) {
  const faces: ExtrudedFace[] = [];

  for (let index = 0; index < points.length; index += 1) {
    const current = points[index];
    const next = points[(index + 1) % points.length];
    const dx = next.x - current.x;
    const dy = next.y - current.y;

    if (dx + dy < 0.75) {
      continue;
    }

    const facePoints = [
      current,
      next,
      { x: next.x + EXTRUSION_DEPTH.x, y: next.y + EXTRUSION_DEPTH.y },
      { x: current.x + EXTRUSION_DEPTH.x, y: current.y + EXTRUSION_DEPTH.y },
    ];

    faces.push({
      key: `${idPrefix}-${areaIndex}-${index}`,
      points: toPointsAttribute(facePoints),
    });
  }

  return faces;
}

function createCountryMesh(): CountryMesh {
  const allAreas = SWEDEN_REGIONS.flatMap((region) => region.areas);
  const topPolygons = allAreas.map((area) => toPointsAttribute(area));
  const shadowPolygons = allAreas.map((area) => toPointsAttribute(offsetPoints(area)));
  const sideFaces = allAreas.flatMap((area, index) => createSideFaces('sweden', index, area));

  return {
    topPolygons,
    sideFaces,
    shadowPolygons,
  };
}

function SwedenMap3DLayerComponent({ viewBox }: SwedenMap3DLayerProps) {
  const countryMesh = useMemo(() => createCountryMesh(), []);
  const baseShadowHeight = SWEDEN_MAP_BOUNDS.height + EXTRUSION_DEPTH.y;
  const baseShadowWidth = SWEDEN_MAP_BOUNDS.width + EXTRUSION_DEPTH.x;

  return (
    <g className="sweden-map-3d-layer pointer-events-none" aria-hidden="true">
      <rect
        x={SWEDEN_MAP_BOUNDS.min.x + EXTRUSION_DEPTH.x * 0.2}
        y={SWEDEN_MAP_BOUNDS.min.y + EXTRUSION_DEPTH.y * 0.55}
        width={baseShadowWidth}
        height={baseShadowHeight}
        rx={18}
        fill="color-mix(in srgb, var(--color-map-boundary) 10%, transparent)"
      />

      <g>
        {countryMesh.shadowPolygons.map((points, index) => (
          <polygon key={`shadow-${index}`} points={points} fill={SHADOW_FILL} stroke="none" />
        ))}

        {countryMesh.sideFaces.map((face) => (
          <polygon key={face.key} points={face.points} fill={SIDE_FILL} stroke="none" />
        ))}

        {countryMesh.topPolygons.map((points, index) => (
          <polygon
            key={`top-${index}`}
            points={points}
            fill={TOP_FILL}
            stroke={TOP_FILL}
            strokeWidth={0.5}
            vectorEffect="non-scaling-stroke"
            paintOrder="stroke fill"
          />
        ))}
      </g>

      <rect
        x={viewBox.minX}
        y={viewBox.minY}
        width={viewBox.width}
        height={viewBox.height}
        fill="none"
        stroke="color-mix(in srgb, var(--color-map-boundary) 12%, transparent)"
        strokeWidth={0.4}
        vectorEffect="non-scaling-stroke"
      />
    </g>
  );
}

export const SwedenMap3DLayer = memo(SwedenMap3DLayerComponent);
