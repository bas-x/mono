import type {
  SimulationAirbase,
  SimulationAircraft,
  SimulationServiceClient,
} from '@/lib/api/types';

const MOCK_AIRBASES: SimulationAirbase[] = [
  {
    id: 'base-1',
    location: { x: 100, y: 100 },
    regionId: 'reg-1',
    region: 'North',
    metadata: { capacity: 10 },
  },
  {
    id: 'base-2',
    location: { x: 200, y: 300 },
    regionId: 'reg-2',
    region: 'South',
    metadata: { capacity: 5 },
  },
];

const MOCK_AIRCRAFTS: SimulationAircraft[] = [
  {
    tailNumber: 'AC001',
    needs: [
      {
        type: 'Refuel',
        severity: 2,
        requiredCapability: 'fuel-truck',
        blocking: true,
      },
    ],
    state: 'Landing',
  },
  {
    tailNumber: 'AC002',
    needs: [],
    state: 'Parked',
    assignedTo: 'base-1',
  },
];

export function createMockSimulationServiceClient(): SimulationServiceClient {
  return {
    async getSimulations() {
      console.log('Mock: Getting simulations');
      return [{ id: 'base' }];
    },

    async createBaseSimulation(seed: string) {
      console.log('Mock: Creating base simulation with seed', seed);
      return { id: 'base' };
    },

    async startSimulation(simulationId: string) {
      console.log('Mock: Starting simulation', simulationId);
    },

    async getAirbases(simulationId: string) {
      console.log('Mock: Getting airbases for simulation', simulationId);
      return MOCK_AIRBASES;
    },

    async getAircrafts(simulationId: string) {
      console.log('Mock: Getting aircrafts for simulation', simulationId);
      return MOCK_AIRCRAFTS;
    },
  };
}
