import { getMockAirbaseDetails, getMockAirbases } from '@/lib/api/mock/map';

import type { Airbase, AirbaseDetails } from '@/features/map/types';

export const MOCK_AIRBASES: Airbase[] = getMockAirbases();

export function resolveMockAirbaseDetails(idOrUrl: string): AirbaseDetails {
  return getMockAirbaseDetails(idOrUrl);
}
