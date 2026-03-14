import type { HttpClient } from '@/lib/api/http/client';
import type {
  CreateBaseSimulationRequest,
  CreateBranchSimulationRequest,
  SimulationAirbase,
  SimulationAircraft,
  SimulationInfo,
  SimulationServiceClient,
} from '@/lib/api/types';

type GetSimulationsResponse = {
  simulations: SimulationInfo[];
};

type CreateBaseSimulationResponse = {
  id: string;
};

type GetSimulationResponse = SimulationInfo;

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

    async getSimulation(simulationId: string, signal?: AbortSignal) {
      return httpClient.requestJson<GetSimulationResponse>(
        `/simulations/${encodeURIComponent(simulationId)}`,
        { signal },
      );
    },

    async createBaseSimulation(request: CreateBaseSimulationRequest, signal?: AbortSignal) {
      return httpClient.requestJson<CreateBaseSimulationResponse>('/simulations/base', {
        method: 'POST',
        body: JSON.stringify(request),
        signal,
      });
    },

    async createBranchSimulation(
      simulationId: string,
      request?: CreateBranchSimulationRequest,
      signal?: AbortSignal,
    ) {
      return httpClient.requestJson<SimulationInfo>(
        `/simulations/${encodeURIComponent(simulationId)}/branch`,
        {
          method: 'POST',
          body: request ? JSON.stringify(request) : undefined,
          signal,
        },
      );
    },

    async startSimulation(simulationId: string, signal?: AbortSignal) {
      void simulationId;
      return httpClient.requestJson<void>('/simulations/start', {
        method: 'POST',
        signal,
      });
    },

    async pauseSimulation(simulationId: string, signal?: AbortSignal) {
      void simulationId;
      return httpClient.requestJson<void>('/simulations/pause', {
        method: 'POST',
        signal,
      });
    },

    async resumeSimulation(simulationId: string, signal?: AbortSignal) {
      void simulationId;
      return httpClient.requestJson<void>('/simulations/resume', {
        method: 'POST',
        signal,
      });
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
