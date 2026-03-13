import type {
  CreateBaseSimulationRequest,
  SimulationInfo,
  SimulationServiceClient,
} from '@/lib/api/types';
import {
  createMockSimulationFromRequest,
  createMockSimulationInfoUpdate,
  getMockSimulationScenario,
  listMockSimulationScenarios,
} from '@/lib/api/mock/scenarios';

export function createMockSimulationServiceClient(): SimulationServiceClient {
  const simulationStore = new Map<string, ReturnType<typeof getMockSimulationScenario>>();

  for (const scenario of listMockSimulationScenarios()) {
    simulationStore.set(scenario.info.id, scenario);
  }

  function ensureScenario(simulationId: string) {
    const existingScenario = simulationStore.get(simulationId);
    if (existingScenario) {
      return existingScenario;
    }

    const fallbackScenario = getMockSimulationScenario(simulationId);
    simulationStore.set(simulationId, fallbackScenario);
    return fallbackScenario;
  }

  function updateSimulationInfo(simulationId: string, nextInfo: SimulationInfo) {
    const scenario = ensureScenario(simulationId);
    simulationStore.set(simulationId, {
      ...scenario,
      info: nextInfo,
    });
  }

  return {
    async getSimulations() {
      console.log('Mock: Getting simulations');
      return Array.from(simulationStore.values()).map((scenario) => ({ ...scenario.info }));
    },

    async getSimulation(simulationId: string) {
      console.log('Mock: Getting simulation', simulationId);
      return { ...ensureScenario(simulationId).info };
    },

    async createBaseSimulation(request: CreateBaseSimulationRequest) {
      console.log('Mock: Creating base simulation', request);
      const scenario = createMockSimulationFromRequest(request);
      simulationStore.set(scenario.info.id, scenario);
      return { id: scenario.info.id };
    },

    async startSimulation(simulationId: string) {
      console.log('Mock: Starting simulation', simulationId);
      updateSimulationInfo(
        simulationId,
        createMockSimulationInfoUpdate(simulationId, {
          running: true,
          paused: false,
        }),
      );
    },

    async pauseSimulation(simulationId: string) {
      console.log('Mock: Pausing simulation', simulationId);
      updateSimulationInfo(
        simulationId,
        createMockSimulationInfoUpdate(simulationId, {
          running: true,
          paused: true,
        }),
      );
    },

    async resumeSimulation(simulationId: string) {
      console.log('Mock: Resuming simulation', simulationId);
      updateSimulationInfo(
        simulationId,
        createMockSimulationInfoUpdate(simulationId, {
          running: true,
          paused: false,
        }),
      );
    },

    async resetSimulation(simulationId: string) {
      console.log('Mock: Resetting simulation', simulationId);
      const scenario = getMockSimulationScenario(simulationId);
      simulationStore.set(simulationId, scenario);
    },

    async getAirbases(simulationId: string) {
      console.log('Mock: Getting airbases for simulation', simulationId);
      return ensureScenario(simulationId).airbases.map((airbase) => ({
        ...airbase,
        location: { ...airbase.location },
        metadata: airbase.metadata ? { ...airbase.metadata } : undefined,
      }));
    },

    async getAircrafts(simulationId: string) {
      console.log('Mock: Getting aircrafts for simulation', simulationId);
      return ensureScenario(simulationId).aircrafts.map((aircraft) => ({
        ...aircraft,
        needs: aircraft.needs.map((need) => ({ ...need })),
      }));
    },
  };
}
