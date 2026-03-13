import type { HttpClient } from '@/lib/api/http/client';
import { getMockAirbaseDetails, getMockAirbases } from '@/lib/api/mock/map';
import type { MapServiceClient } from '@/lib/api/types';

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function createMapServiceClient(_httpClient: HttpClient): MapServiceClient {
  return {
    async getAirbases() {
      return getMockAirbases();
    },

    async getAirbaseDetails(idOrUrl: string) {
      return getMockAirbaseDetails(idOrUrl);
    },
  };
}
