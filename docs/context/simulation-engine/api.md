# Simulation Engine API Context

## Endpoints
- `POST /simulations/base`
  - Request: `{ seed?: string, untilTick?: number, simulationOptions?: { constellationOpts?: {...}, fleetOpts?: {...}, threatOpts?: {...}, lifecycleOpts?: {...} } }`
  - Response: `{ id: string }`
- `GET /simulations/:simulationId`
  - Response: `{ id: string, running: boolean, paused: boolean, tick: number, timestamp: string, untilTick?: number }`

## Request Shaping
- `seed` remains top-level and is optional; an empty frontend seed should be omitted so the backend can apply its default seed handling.
- The simulation setup UI accepts duration in seconds and converts it to `untilTick` using the backend's default `64` ticks/second runner rate before submitting `POST /simulations/base`.
- Current frontend controls map into `simulationOptions.constellationOpts` and `simulationOptions.fleetOpts`.
- Region filters are sent as trimmed string arrays in `includeRegions` and `excludeRegions`.
- Percent-based UI controls are sent as ratio objects with `{ numerator, denominator }` using denominator `100`.
- `threatOpts` and `lifecycleOpts` are reserved in the request contract for future frontend controls and may be omitted today.

## Frontend Adapter
- Request/response types live in `frontend/src/lib/api/types.ts`.
- Real HTTP transport is implemented in `frontend/src/lib/api/services/simulation.ts`.
- Mock parity for create simulation requests lives in `frontend/src/lib/api/mock/simulation.ts`.
- The timeline should treat `untilTick` as the preferred replay boundary when present, and fall back to the latest observed tick or `simulation_ended.tick` when it is not present yet.
