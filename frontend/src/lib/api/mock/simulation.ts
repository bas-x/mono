import type {
  CreateBaseSimulationRequest,
  CreateBranchSimulationRequest,
  OverrideAssignmentRequest,
  SimulationServiceClient,
} from '@/lib/api/types';
import {
  createStoredMockBaseSimulation,
  createStoredMockBranchSimulation,
  createStoredMockSimulationInfoUpdate,
  getStoredMockSimulationScenario,
  listStoredMockSimulationScenarios,
  overrideStoredMockAssignment,
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

    async overrideAssignment(simulationId: string, tailNumber: string, request: OverrideAssignmentRequest) {
      console.log('Mock: Overriding assignment', simulationId, tailNumber, request);
      return overrideStoredMockAssignment(simulationId, tailNumber, request.baseId);
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
