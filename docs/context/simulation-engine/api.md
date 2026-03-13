# Simulation Engine API Context

## Endpoints
- `POST /simulations/base`
  - Request: `{ seed?: string, untilTick?: number, simulationOptions?: { constellationOpts?: {...}, fleetOpts?: {...}, threatOpts?: {...}, lifecycleOpts?: {...} } }`
  - Response: `{ id: string }`

## Request Shaping
- `seed` remains top-level and is optional; an empty frontend seed should be omitted so the backend can apply its default seed handling.
- Current frontend controls map into `simulationOptions.constellationOpts` and `simulationOptions.fleetOpts`.
- Region filters are sent as trimmed string arrays in `includeRegions` and `excludeRegions`.
- Percent-based UI controls are sent as ratio objects with `{ numerator, denominator }` using denominator `100`.
- `threatOpts` and `lifecycleOpts` are reserved in the request contract for future frontend controls and may be omitted today.

## Frontend Adapter
- Request/response types live in `frontend/src/lib/api/types.ts`.
- Real HTTP transport is implemented in `frontend/src/lib/api/services/simulation.ts`.
- Mock parity for create simulation requests lives in `frontend/src/lib/api/mock/simulation.ts`.
