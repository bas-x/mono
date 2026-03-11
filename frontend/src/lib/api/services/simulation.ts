import type { HttpClient } from '@/lib/api/http/client';
import type {
  SimulationAirbase,
  SimulationAircraft,
  SimulationServiceClient,
} from '@/lib/api/types';

type CreateBaseSimulationResponse = {
  id: string;
};

type GetAirbasesResponse = {
  airbases: SimulationAirbase[];
};

type GetAircraftsResponse = {
  aircrafts: SimulationAircraft[];
};

export function createSimulationServiceClient(
  httpClient: HttpClient,
): SimulationServiceClient {
  return {
    async createBaseSimulation(seed: string, signal?: AbortSignal) {
      return httpClient.requestJson<CreateBaseSimulationResponse>('/simulation', {
        method: 'POST',
        body: JSON.stringify({ seed }),
        signal,
      });
    },

    async getAirbases(simulationId: string, signal?: AbortSignal) {
      const response = await httpClient.requestJson<GetAirbasesResponse>(
        `/simulation/${encodeURIComponent(simulationId)}/airbases`,
        { signal },
      );
      return response.airbases || [];
    },

    async getAircrafts(simulationId: string, signal?: AbortSignal) {
      const response = await httpClient.requestJson<GetAircraftsResponse>(
        `/simulation/${encodeURIComponent(simulationId)}/aircrafts`,
        { signal },
      );
      return response.aircrafts || [];
    },
  };
}
