# Frontend Integration: Live Simulation Events

This file is intended to give the frontend implementation agent a compact map of how to integrate with the backend simulation APIs and live event stream.

## Current Simulation Flow

The backend currently supports a single simulation with ID `base`.

Recommended frontend flow:

1. `GET /simulations`
   - discover existing simulations and whether they are already running
2. `POST /simulations/base`
   - create the base simulation if it does not exist yet
3. `GET /simulations/base/airbases`
   - load initial airbase data
4. `GET /simulations/base/aircrafts`
   - load initial aircraft data
5. `GET /ws/simulations/base/events`
   - connect websocket for live updates
6. `POST /simulations/base/start`
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
      "running": false
    }
  ]
}
```

### Create base simulation

- `POST /simulations/base`

Response:

```json
{
  "id": "base"
}
```

### Start simulation

- `POST /simulations/:simulationId/start`
- `202` on success
- `404` if missing
- `409` if already running

### Read models

- `GET /simulations/:simulationId/airbases`
- `GET /simulations/:simulationId/aircrafts`

These endpoints are useful for initial page hydration before live updates arrive.

## WebSocket Endpoint

- `GET /ws/simulations/:simulationId/events`

The socket emits **all** event types for the selected simulation.

Current event types:

- `simulation_step`
- `aircraft_state_change`
- `landing_assignment`

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
- `simulation_step`
  - update current tick / time cursor

## Operational Notes

- IDs are opaque lowercase hex strings
- the simulation package itself does **not** know about `simulationId`; the service injects it into outgoing events
- slow websocket clients are disconnected by the backend rather than allowed to block simulation progress
- current implementation is base-simulation-first, but the API paths already include `simulationId` for future branch support
