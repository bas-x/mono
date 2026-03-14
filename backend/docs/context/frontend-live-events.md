# Frontend Integration: Live Simulation Events

This file is intended to give the frontend implementation agent a compact map of how to integrate with the backend simulation APIs and live event stream.

## Current Simulation Flow

The backend currently supports a base simulation with ID `base` plus service-generated clone IDs created from the base simulation.

Recommended frontend flow:

1. `GET /simulations`
   - discover existing simulations and whether they are already running
2. `POST /simulations/base`
   - create the base simulation if it does not exist yet; request can include `seed`, `untilTick`, and `simulationOptions`
3. `GET /simulations/base/airbases`
   - load initial airbase data
4. `GET /simulations/base/aircrafts`
   - load initial aircraft data
5. `GET /ws/simulations/base/events`
   - connect websocket for live updates
6. `POST /simulations/start`
   - start the continuous simulation runner if `running=false`

## REST Endpoints

### List simulations

- `GET /simulations`

Response:

```json
{
  "simulations": [
    {
      "id": "base",
      "running": false,
      "paused": false,
      "tick": 42,
      "timestamp": "2026-03-12T03:15:05Z",
      "parentId": null,
      "splitTick": null,
      "splitTimestamp": null
    },
    {
      "id": "7f3c2d1a9b8e6f10",
      "running": false,
      "paused": false,
      "tick": 42,
      "timestamp": "2026-03-12T03:15:05Z",
      "parentId": "base",
      "splitTick": 42,
      "splitTimestamp": "2026-03-12T03:15:05Z",
      "sourceEvent": {
        "id": "timeline-evt-17",
        "type": "landing_assignment",
        "tick": 41
      }
    }
  ]
}
```

Branch entries may include optional `sourceEvent`. `splitTick` and `splitTimestamp` are the actual fork coordinates. `sourceEvent.id`, `sourceEvent.type`, and `sourceEvent.tick` are metadata-only clicked-anchor fields.

### Get branch metadata

- `GET /simulations/:simulationId`

Branch detail response example:

```json
{
  "id": "7f3c2d1a9b8e6f10",
  "running": false,
  "paused": false,
  "tick": 42,
  "timestamp": "2026-03-12T03:15:05Z",
  "parentId": "base",
  "splitTick": 42,
  "splitTimestamp": "2026-03-12T03:15:05Z",
  "sourceEvent": {
    "id": "timeline-evt-17",
    "type": "landing_assignment",
    "tick": 41
  }
}
```

If a branch has no stored source-event anchor, `sourceEvent` is omitted rather than sent as `null`.

### Branch simulation

- `POST /simulations/:simulationId/branch`

Optional request body:

```json
{
  "sourceEvent": {
    "id": "timeline-evt-17",
    "type": "landing_assignment",
    "tick": 41
  }
}
```

Notes:

- V1 supports branching from `base` only.
- The request body is optional.
- If `sourceEvent` is provided, `id`, `type`, and `tick` are all required. Partial or malformed values return `400`.
- `splitTick` and `splitTimestamp` still define the actual branch split. `sourceEvent` is metadata only.
- Source-event persistence is runtime-only. A backend restart loses in-memory branch metadata.

### Create base simulation

- `POST /simulations/base`

Optional request fields:

- `seed`
- `untilTick`
- `simulationOptions`
  - supports `constellationOpts`, `fleetOpts`, `threatOpts`, and `lifecycleOpts`
  - each group is optional; omitted groups use backend defaults

Response:

```json
{
  "id": "base"
}
```

### Start simulation

- `POST /simulations/start`
- `202` on success
- `404` if missing
- `409` if already running

### Read models

- `GET /simulations/:simulationId/airbases`
- `GET /simulations/:simulationId/aircrafts`

### Override landing assignment

- `POST /simulations/:simulationId/aircraft/:tailNumber/assignment-override`

Request:

```json
{
  "baseId": "3a5f..."
}
```

Response:

```json
{
  "aircraft": {
    "tailNumber": "9b2e...",
    "state": "Inbound",
    "needs": [],
    "assignedTo": "3a5f...",
    "position": {"x": 132.45, "y": 611.08}
  },
  "assignment": {
    "base": "3a5f...",
    "source": "human"
  }
}
```

- `409` means the aircraft is already past the pre-commit override window.

These endpoints are useful for initial page hydration before live updates arrive.

## WebSocket Endpoint

- `GET /ws/simulations/:simulationId/events`

The socket emits **all** event types for the selected simulation.

For the base stream, this now includes `branch_created` when a new V1 branch is created from `base`. The event stays on `/ws/simulations/base/events` because websocket routing filters by `simulationId` and the payload keeps `simulationId="base"`.

Current event types:

- `simulation_step`
- `simulation_ended`
- `simulation_closed`
- `aircraft_state_change`
- `landing_assignment`
- `all_aircraft_positions`
- `threat_spawned`
- `threat_targeted`
- `threat_despawned`
- `branch_created` (base stream only)

## Event Shapes

### Simulation step

```json
{
  "type": "simulation_step",
  "simulationId": "base",
  "tick": 42,
  "timestamp": "2026-03-11T18:00:00Z"
}
```

### Aircraft state change

```json
{
  "type": "aircraft_state_change",
  "simulationId": "base",
  "tailNumber": "9b2e...",
  "oldState": "Outbound",
  "newState": "Engaged",
  "aircraft": {
    "tailNumber": "9b2e...",
    "state": "Engaged",
    "needs": [
      {
        "type": "fuel",
        "severity": 60,
        "requiredCapability": "fuel",
        "blocking": false
      }
    ],
    "assignedTo": null
  },
  "timestamp": "2026-03-11T18:00:05Z"
}
```

### Simulation ended

```json
{
  "type": "simulation_ended",
  "simulationId": "base",
  "tick": 3,
  "timestamp": "2026-03-11T18:00:15Z",
  "summary": {
    "completedVisitCount": 0,
    "totalDurationMs": 0,
    "averageDurationMs": null
  }
}
```

### Simulation closed

```json
{
  "type": "simulation_closed",
  "simulationId": "base",
  "tick": 1,
  "timestamp": "2026-03-11T18:00:05Z",
  "reason": "reset",
  "summary": {
    "completedVisitCount": 0,
    "totalDurationMs": 0,
    "averageDurationMs": null
  }
}
```

Both terminal event types use the same direct `summary` object. `totalDurationMs` and `averageDurationMs` are milliseconds. `averageDurationMs` is `null` until at least one servicing visit completes. `simulation_closed` always carries `reason`, `reset` for base reset and `cancel` for branch reset. Use event type, not summary shape, to distinguish natural completion (`simulation_ended`) from non-natural closure (`simulation_closed`).

### Landing assignment

```json
{
  "type": "landing_assignment",
  "simulationId": "base",
  "tailNumber": "9b2e...",
  "baseId": "3a5f...",
  "source": "algorithm",
  "timestamp": "2026-03-11T18:00:08Z"
}
```

### All aircraft positions

```json
{
  "type": "all_aircraft_positions",
  "simulationId": "base",
  "tick": 42,
  "timestamp": "2026-03-11T18:00:00Z",
  "positions": [
    {
      "tailNumber": "9b2e...",
      "position": {"x": 132.45, "y": 611.08},
      "state": "Ready",
      "needs": []
    }
  ]
}
```

`all_aircraft_positions` is emitted every simulation tick and mirrors the simulation state's current aircraft coordinates. Aircraft positions are initialized from generated airbase locations during simulation init, so websocket position snapshots no longer start at `{x:0,y:0}` by default.

Notes:

- A landing override uses the existing `landing_assignment` websocket event.
- Frontend code should key off `source`:
  - `algorithm` = dispatcher-selected assignment
  - `human` = operator override applied through the API
- It is valid to receive an `algorithm` assignment event before a `human` override event for the same aircraft if the backend registers inbound and then applies the override.

### Branch created

```json
{
  "type": "branch_created",
  "simulationId": "base",
  "branchId": "7f3c2d1a9b8e6f10",
  "parentId": "base",
  "splitTick": 42,
  "splitTimestamp": "2026-03-12T03:15:05Z",
  "sourceEvent": {
    "id": "timeline-evt-17",
    "type": "landing_assignment",
    "tick": 41
  }
}
```

Notes:

- This is a base-stream event carrying branch lineage summary metadata for the newly created branch.
- V1 still supports branching from `base` only.
- `splitTick` and `splitTimestamp` are the canonical split coordinates.
- `sourceEvent` is metadata only, and is omitted from legacy branches or requests that did not provide it.

## Frontend State Strategy

Recommended state model:

- keep `airbases` as mostly static read model data
- keep `aircraftByTailNumber` as the main mutable live map
- apply websocket events incrementally instead of refetching on every tick
- treat `simulation_step` as a heartbeat/timeline signal

Suggested reducer behavior:

- `aircraft_state_change`
  - replace the aircraft entry from `event.aircraft`
  - optionally record the state transition in a timeline panel
- `landing_assignment`
  - update the aircraft assignment if the aircraft already exists in local state
  - optionally annotate the selected base in UI
- `branch_created`
  - treat it as a base-stream metadata event, not as an event on the new branch stream
  - add/update branch summary state from `branchId`, `parentId`, `splitTick`, and `splitTimestamp`
- `simulation_step`
  - update current tick / time cursor
- `simulation_ended`
  - mark the simulation as completed naturally, persist `summary`, and stop assuming further live updates will arrive unless restarted
- `simulation_closed`
  - treat the simulation as non-natural removal/closure, persist required `reason` plus `summary`, and stop assuming further live updates will arrive unless the simulation is recreated

## Operational Notes

- IDs are opaque lowercase hex strings
- the simulation package itself does **not** know about `simulationId`; the service injects it into outgoing events
- slow websocket clients are disconnected by the backend rather than allowed to block simulation progress
- branch creation is available via `POST /simulations/:simulationId/branch`
- branch lineage metadata (`parentId`, `splitTick`, `splitTimestamp`, optional `sourceEvent`) comes from REST simulation reads, the branch creation response, and the base-stream `branch_created` event
- `simulation_closed` is the websocket-only terminal non-natural removal event emitted by reset semantics; `reason=reset` is used for base removal and `reason=cancel` for branch removal
- terminal servicing summaries are shipped directly as `summary.completedVisitCount`, `summary.totalDurationMs`, and nullable `summary.averageDurationMs`
- `sourceEvent` is metadata-only clicked-anchor data. It does not change branch snapshot semantics, and it is only kept for the current backend runtime
- `branch_created` is emitted only on `/ws/simulations/base/events`; branch streams continue to receive only events tagged with their own `simulationId`
- first branch support is base simulation only; checkpoint-based branch creation and branch-from-branch workflows are not implemented
- determinism guarantee: branch creation copies current simulation state and RNG state, so equivalent future advancement keeps base and branch aligned until a later divergence decision is introduced
- the local tester auto-creates the base simulation at startup, shows `Base` as the initial tab, and switches the full tester context when a branch tab is selected
