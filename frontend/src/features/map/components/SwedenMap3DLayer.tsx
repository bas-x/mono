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

type RegionMesh = {
  id: string;
  topPolygons: string[];
  sideFaces: ExtrudedFace[];
  shadowPolygons: string[];
  topFill: string;
  sideFill: string;
  shadowFill: string;
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
const TOP_FILL_VARIANTS = [
  'color-mix(in srgb, var(--color-map-surface) 88%, white 12%)',
  'color-mix(in srgb, var(--color-map-surface) 94%, white 6%)',
  'color-mix(in srgb, var(--color-map-surface) 82%, black 18%)',
] as const;
const SIDE_FILL_VARIANTS = [
  'color-mix(in srgb, var(--color-map-boundary) 70%, var(--color-map-surface) 30%)',
  'color-mix(in srgb, var(--color-map-boundary) 78%, var(--color-map-surface) 22%)',
  'color-mix(in srgb, var(--color-map-boundary) 66%, black 34%)',
] as const;
const SHADOW_FILL_VARIANTS = [
  'color-mix(in srgb, var(--color-map-boundary) 28%, transparent)',
  'color-mix(in srgb, var(--color-map-boundary) 34%, transparent)',
  'color-mix(in srgb, black 24%, transparent)',
] as const;
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

function createSideFaces(regionId: string, areaIndex: number, points: readonly MapPoint[]) {
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
      key: `${regionId}-${areaIndex}-${index}`,
      points: toPointsAttribute(facePoints),
    });
  }

  return faces;
}

function createRegionMeshes(): RegionMesh[] {
  return SWEDEN_REGIONS.map((region, index) => {
    const variantIndex = index % TOP_FILL_VARIANTS.length;
    const topPolygons = region.areas.map((area) => toPointsAttribute(area));
    const shadowPolygons = region.areas.map((area) => toPointsAttribute(offsetPoints(area)));
    const sideFaces = region.areas.flatMap((area, areaIndex) =>
      createSideFaces(region.id, areaIndex, area),
    );

    return {
      id: region.id,
      topPolygons,
      sideFaces,
      shadowPolygons,
      topFill: TOP_FILL_VARIANTS[variantIndex] ?? TOP_FILL_VARIANTS[0],
      sideFill: SIDE_FILL_VARIANTS[variantIndex] ?? SIDE_FILL_VARIANTS[0],
      shadowFill: SHADOW_FILL_VARIANTS[variantIndex] ?? SHADOW_FILL_VARIANTS[0],
    };
  });
}

function SwedenMap3DLayerComponent({ viewBox }: SwedenMap3DLayerProps) {
  const regionMeshes = useMemo(() => createRegionMeshes(), []);
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

      {regionMeshes.map((region) => (
        <g key={region.id}>
          {region.shadowPolygons.map((points, polygonIndex) => (
            <polygon
              key={`${region.id}-shadow-${polygonIndex}`}
              points={points}
              fill={region.shadowFill}
              stroke="none"
            />
          ))}

          {region.sideFaces.map((face) => (
            <polygon key={face.key} points={face.points} fill={region.sideFill} stroke="none" />
          ))}

          {region.topPolygons.map((points, polygonIndex) => (
            <polygon
              key={`${region.id}-top-${polygonIndex}`}
              points={points}
              fill={region.topFill}
              stroke="var(--color-map-boundary)"
              strokeWidth={0.9}
              vectorEffect="non-scaling-stroke"
              paintOrder="stroke fill"
            />
          ))}
        </g>
      ))}

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
