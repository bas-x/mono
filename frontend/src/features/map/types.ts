import type { ApiAirbase, ApiAirbaseDetails, ApiAirbasePoint } from '@/lib/api';
import type { SimulationAirbase } from '@/lib/api/types';

export type MapMode = 'static' | 'live' | 'replay';

export type MapDataSource = 'mock' | 'api' | 'hybrid';

export type MapCoordinateTransform = {
  scaleX: number;
  scaleY: number;
  offsetX: number;
  offsetY: number;
};

export type MapViewBox = {
  minX: number;
  minY: number;
  width: number;
  height: number;
};

export type AirbasePoint = ApiAirbasePoint;
export type Airbase = ApiAirbase;
export type AirbaseDetails = ApiAirbaseDetails;
export type PolygonBackedAirbasePlacement = Pick<Airbase, 'id' | 'name' | 'area'>;
export type PointBackedAirbasePlacement = {
  id: SimulationAirbase['id'];
  name: SimulationAirbase['name'];
  point: AirbasePoint;
  regionId?: SimulationAirbase['regionId'];
  region?: SimulationAirbase['region'];
};
export type AirbasePlacementSource =
  | PolygonBackedAirbasePlacement
  | PointBackedAirbasePlacement;

export type AirbaseDetailsState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; details: AirbaseDetails }
  | { status: 'error'; message: string };

export const DEFAULT_MAP_VIEW_BOX: MapViewBox = {
  minX: 0,
  minY: 0,
  width: 345.62482,
  height: 792.52374,
};

export const IDENTITY_COORDINATE_TRANSFORM: MapCoordinateTransform = {
  scaleX: 1,
  scaleY: 1,
  offsetX: 0,
  offsetY: 0,
};
