import type {
  CreateBaseSimulationRequest,
  CreateBranchSimulationRequest,
  SimulationServiceClient,
} from '@/lib/api/types';
import {
  createStoredMockBaseSimulation,
  createStoredMockBranchSimulation,
  createStoredMockSimulationInfoUpdate,
  getStoredMockSimulationScenario,
  listStoredMockSimulationScenarios,
  resetMockScenarioStore,
  resetStoredMockSimulation,
  updateStoredMockSimulationInfo,
} from '@/lib/api/mock/store';

export function createMockSimulationServiceClient(): SimulationServiceClient {
  resetMockScenarioStore();

  return {
    async getSimulations() {
      console.log('Mock: Getting simulations');
      return listStoredMockSimulationScenarios().map((scenario) => ({ ...scenario.info }));
    },

    async getSimulation(simulationId: string) {
      console.log('Mock: Getting simulation', simulationId);
      return { ...getStoredMockSimulationScenario(simulationId).info };
    },

    async createBaseSimulation(request: CreateBaseSimulationRequest) {
      console.log('Mock: Creating base simulation', request);
      const scenario = createStoredMockBaseSimulation(request);
      return { id: scenario.info.id };
    },

    async createBranchSimulation(simulationId: string, request?: CreateBranchSimulationRequest) {
      console.log('Mock: Branching simulation', simulationId, request);
      const scenario = createStoredMockBranchSimulation(simulationId, request);
      return { ...scenario.info };
    },

    async startSimulation(simulationId: string) {
      console.log('Mock: Starting simulation', simulationId);
      updateStoredMockSimulationInfo(
        simulationId,
        createStoredMockSimulationInfoUpdate(simulationId, {
          running: true,
          paused: false,
        }),
      );
    },

    async pauseSimulation(simulationId: string) {
      console.log('Mock: Pausing simulation', simulationId);
      updateStoredMockSimulationInfo(
        simulationId,
        createStoredMockSimulationInfoUpdate(simulationId, {
          running: true,
          paused: true,
        }),
      );
    },

    async resumeSimulation(simulationId: string) {
      console.log('Mock: Resuming simulation', simulationId);
      updateStoredMockSimulationInfo(
        simulationId,
        createStoredMockSimulationInfoUpdate(simulationId, {
          running: true,
          paused: false,
        }),
      );
    },

    async resetSimulation(simulationId: string) {
      console.log('Mock: Resetting simulation', simulationId);
      resetStoredMockSimulation(simulationId);
    },

    async getAirbases(simulationId: string) {
      console.log('Mock: Getting airbases for simulation', simulationId);
      return getStoredMockSimulationScenario(simulationId).airbases.map((airbase) => ({
        ...airbase,
        location: { ...airbase.location },
        metadata: airbase.metadata ? { ...airbase.metadata } : undefined,
      }));
    },

    async getAircrafts(simulationId: string) {
      console.log('Mock: Getting aircrafts for simulation', simulationId);
      return getStoredMockSimulationScenario(simulationId).aircrafts.map((aircraft) => ({
        ...aircraft,
        needs: aircraft.needs.map((need) => ({ ...need })),
      }));
    },
  };
}
