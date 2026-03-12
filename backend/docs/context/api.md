# HTTP API Context

## Simulations

### Create base simulation

- **Method**: `POST`
- **Path**: `/simulations/base`
- **Body**:

```json
{
  "seed": "<optional 64-char hex seed>"
}
```

- **Response** `201`:

```json
{
  "id": "base"
}
```

### List simulations

- **Method**: `GET`
- **Path**: `/simulations`
- **Response** `200`:

```json
{
  "simulations": [
    {
      "id": "base",
      "running": true,
      "paused": false,
      "tick": 42,
      "timestamp": "2026-03-12T03:15:05Z"
    }
  ]
}
```

### Get a simulation

- **Method**: `GET`
- **Path**: `/simulations/:simulationId`
- **Response** `200`:

```json
{
  "id": "base",
  "running": true,
  "paused": false,
  "tick": 42,
  "timestamp": "2026-03-12T03:15:05Z"
}
```
- **Response** `404`: simulation not found

### Start a simulation

- **Method**: `POST`
- **Path**: `/simulations/:simulationId/start`
- **Response** `202`: no body
- **Response** `404`: simulation not found
- **Response** `409`: simulation already running

### Pause a simulation

- **Method**: `POST`
- **Path**: `/simulations/:simulationId/pause`
- **Response** `202`: no body
- **Response** `404`: simulation not found
- **Response** `409`: simulation not running or already paused

### Resume a simulation

- **Method**: `POST`
- **Path**: `/simulations/:simulationId/resume`
- **Response** `202`: no body
- **Response** `404`: simulation not found
- **Response** `409`: simulation not running or not paused

### Reset a simulation

- **Method**: `POST`
- **Path**: `/simulations/:simulationId/reset`
- **Response** `202`: no body
- **Response** `404`: simulation not found

### List airbases for a simulation

- **Method**: `GET`
- **Path**: `/simulations/:simulationId/airbases`
- **Response** `200`:

```json
{
  "airbases": [
    {
      "id": "3a5f...",
      "location": {"x": 0.0, "y": 0.0},
      "regionId": "SE-BLE",
      "region": "Blekinge",
      "metadata": {}
    }
  ]
}
```

### List aircrafts for a simulation

- **Method**: `GET`
- **Path**: `/simulations/:simulationId/aircrafts`
- **Response** `200`:

```json
{
  "aircrafts": [
    {
      "tailNumber": "9b2e...",
      "state": "Outbound",
      "needs": [
        {
          "type": "fuel",
          "severity": 75,
          "requiredCapability": "fuel",
          "blocking": true
        }
      ],
      "assignedTo": "3a5f..."
    }
  ]
}
```

### List threats for a simulation

- **Method**: `GET`
- **Path**: `/simulations/:simulationId/threats`
- **Response** `200`:

```json
{
  "threats": [
    {
      "id": "3a5f...",
      "regionId": "SE-K",
      "region": "Blekinge",
      "createdAt": "2026-03-12T03:15:05Z",
      "createdTick": 42
    }
  ]
}
```
- **Response** `404`: simulation not found

### Stream simulation events

- **Method**: `GET`
- **Path**: `/ws/simulations/:simulationId/events`
- **Transport**: WebSocket
- **Behavior**:
  - Streams all event types for the requested simulation.
  - Slow clients are disconnected instead of blocking simulation progress.

- **Example payloads**:

```json
{
  "type": "simulation_step",
  "simulationId": "base",
  "tick": 1,
  "timestamp": "2026-03-11T18:00:00Z"
}
```

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
    "needs": [],
    "assignedTo": null
  },
  "timestamp": "2026-03-11T18:00:05Z"
}
```

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

```json
{
  "type": "threat_spawned",
  "simulationId": "base",
  "threat": {
    "id": "3a5f...",
    "regionId": "SE-K",
    "region": "Blekinge",
    "createdAt": "2026-03-12T03:15:05Z",
    "createdTick": 42
  },
  "timestamp": "2026-03-12T03:15:05Z"
}
```

```json
{
  "type": "threat_claimed",
  "simulationId": "base",
  "threat": {
    "id": "3a5f...",
    "regionId": "SE-K",
    "region": "Blekinge",
    "createdAt": "2026-03-12T03:15:05Z",
    "createdTick": 42
  },
  "tailNumber": "9b2e...",
  "timestamp": "2026-03-12T03:15:10Z"
}
```

## Notes

- `simulationId` is already part of the path to support future branch simulations.
- Current implementation supports only `simulationId=base`.
- Airbase IDs and tail numbers are serialized as lowercase hex strings.
