# Constellation Map API Context

## Endpoints
- `GET /map`
  - Response: `{ airbases: Array<{ id: string; area: Array<{ x: number; y: number }>; infoUrl?: string }> }`
- `GET /map/airbase/:id`
  - Response: object containing at least `{ id: string }` plus optional metadata fields.
- `infoUrl`
  - If `infoUrl` is present on an airbase, detail requests may use that URL/path directly.

## Frontend Adapter
- Map API is exposed through `MapServiceClient` in `frontend/src/lib/api/types.ts`.
- Methods:
  - `getAirbases(signal?) => Promise<ApiAirbase[]>`
  - `getAirbaseDetails(idOrUrl, signal?) => Promise<ApiAirbaseDetails>`
- Source wiring:
  - Real mode uses `createMapServiceClient` with HTTP transport.
  - Mock mode uses `createMockApiClients` map handlers.

## Mock-First Demo Behavior
- `ConstellationMap` supports `dataSource: 'mock' | 'api' | 'hybrid'`.
- V1 default is `mock` for predictable hackathon demos.
- `hybrid` tries API first and falls back to mock dataset when request fails.

## Request Shaping
- Airbase detail key can be `id` or `infoUrl`.
- Relative detail paths are routed through API base URL.
- Absolute `http(s)` detail URLs are fetched directly.
