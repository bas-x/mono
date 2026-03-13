import type { HttpClient } from '@/lib/api/http/client';
import type {
  CreateBaseSimulationRequest,
  SimulationAirbase,
  SimulationAircraft,
  SimulationServiceClient,
} from '@/lib/api/types';

type GetSimulationsResponse = {
  simulations: Array<{ id: string }>;
};

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
    async getSimulations(signal?: AbortSignal) {
      const response = await httpClient.requestJson<GetSimulationsResponse>('/simulations', {
        signal,
      });
      return response.simulations || [];
    },

    async createBaseSimulation(request: CreateBaseSimulationRequest, signal?: AbortSignal) {
      return httpClient.requestJson<CreateBaseSimulationResponse>('/simulations/base', {
        method: 'POST',
        body: JSON.stringify(request),
        signal,
      });
    },

    async startSimulation(simulationId: string, signal?: AbortSignal) {
      return httpClient.requestJson<void>(
        `/simulations/${encodeURIComponent(simulationId)}/start`,
        {
          method: 'POST',
          signal,
        },
      );
    },

    async pauseSimulation(simulationId: string, signal?: AbortSignal) {
      return httpClient.requestJson<void>(
        `/simulations/${encodeURIComponent(simulationId)}/pause`,
        {
          method: 'POST',
          signal,
        },
      );
    },

    async resumeSimulation(simulationId: string, signal?: AbortSignal) {
      return httpClient.requestJson<void>(
        `/simulations/${encodeURIComponent(simulationId)}/resume`,
        {
          method: 'POST',
          signal,
        },
      );
    },

    async resetSimulation(simulationId: string, signal?: AbortSignal) {
      return httpClient.requestJson<void>(
        `/simulations/${encodeURIComponent(simulationId)}/reset`,
        {
          method: 'POST',
          signal,
        },
      );
    },

    async getAirbases(simulationId: string, signal?: AbortSignal) {
      const response = await httpClient.requestJson<GetAirbasesResponse>(
        `/simulations/${encodeURIComponent(simulationId)}/airbases`,
        { signal },
      );
      return response.airbases || [];
    },

    async getAircrafts(simulationId: string, signal?: AbortSignal) {
      const response = await httpClient.requestJson<GetAircraftsResponse>(
        `/simulations/${encodeURIComponent(simulationId)}/aircrafts`,
        { signal },
      );
      return response.aircrafts || [];
    },
  };
}
