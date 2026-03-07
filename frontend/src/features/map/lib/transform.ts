import {
  IDENTITY_COORDINATE_TRANSFORM,
  type AirbasePoint,
  type MapCoordinateTransform,
} from '@/features/map/types';

export function resolveCoordinateTransform(
  transform?: Partial<MapCoordinateTransform>,
): MapCoordinateTransform {
  if (!transform) {
    return IDENTITY_COORDINATE_TRANSFORM;
  }

  return {
    scaleX: transform.scaleX ?? IDENTITY_COORDINATE_TRANSFORM.scaleX,
    scaleY: transform.scaleY ?? IDENTITY_COORDINATE_TRANSFORM.scaleY,
    offsetX: transform.offsetX ?? IDENTITY_COORDINATE_TRANSFORM.offsetX,
    offsetY: transform.offsetY ?? IDENTITY_COORDINATE_TRANSFORM.offsetY,
  };
}

export function applyCoordinateTransform(
  point: AirbasePoint,
  transform: MapCoordinateTransform,
): AirbasePoint {
  return {
    x: point.x * transform.scaleX + transform.offsetX,
    y: point.y * transform.scaleY + transform.offsetY,
  };
}

export function applyTransformToPolygon(
  points: AirbasePoint[],
  transform: MapCoordinateTransform,
): AirbasePoint[] {
  return points.map((point) => applyCoordinateTransform(point, transform));
}
