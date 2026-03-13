# HTTP API Context

## Simulations

### Create base simulation

- **Method**: `POST`
- **Path**: `/simulations/base`
- **Body**:

```json
{
  "seed": "<optional seed string>",
  "untilTick": 3,
  "simulationOptions": {
    "constellationOpts": {
      "includeRegions": ["Blekinge"],
      "minPerRegion": 1,
      "maxPerRegion": 1,
      "maxTotal": 1,
      "regionProbability": {"numerator": 1, "denominator": 1}
    },
    "fleetOpts": {
      "aircraftMin": 1,
      "aircraftMax": 1,
      "needsMin": 0,
      "needsMax": 0,
      "needsPool": ["fuel", "munitions"],
      "blockingChance": {"numerator": 1, "denominator": 2}
    },
    "threatOpts": {
      "spawnChance": {"numerator": 1, "denominator": 1},
      "maxActive": 1,
      "maxActiveTicks": 10
    },
    "lifecycleOpts": {
      "durations": {
        "outbound": 1500000000000,
        "engaged": 2100000000000,
        "inboundDecision": 480000000000,
        "commitApproach": 360000000000,
        "servicing": 4500000000000,
        "ready": 1200000000000
      },
      "returnThreshold": 85,
      "needRates": {
        "fuel": {
          "outboundMilliPerHour": 2600,
          "engagedMilliPerHour": 4200,
          "servicingMilliPerHour": 12000,
          "variancePermille": 450
        }
      }
    }
  }
}
```

- **Behavior**:
  - `untilTick` is optional and stops the created simulation once that tick is reached.
  - `simulationOptions` is optional; if omitted, the backend uses the current default demo options.
  - `needsPool` accepts: `fuel`, `charge`, `munitions`, `repairs`, `maintenance`, `mission_configuration`, `crew_support`, `emergency`, `weather_constraint`, `ground_support`, `protection`.
  - `lifecycleOpts.needRates` uses the same need keys as `needsPool`.

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

### Branch a simulation

- **Method**: `POST`
- **Path**: `/simulations/:simulationId/branch`
- **Response** `201`:

```json
{
  "id": "7f3c2d1a9b8e6f10"
}
```

- **Behavior**:
  - V1 only supports branching from `simulationId=base`.
  - The returned `id` is the new branch simulation ID.
- **Response** `404`: simulation not found

### Start a simulation

- **Method**: `POST`
- **Path**: `/simulations/start`
- **Response** `202`: no body
- **Behavior**:
  - Starts all non-running simulations.
- **Response** `404`: simulation not found
- **Response** `409`: simulation already running

### Pause a simulation

- **Method**: `POST`
- **Path**: `/simulations/pause`
- **Response** `202`: no body
- **Behavior**:
  - Pauses all running simulations.
- **Response** `404`: simulation not found
- **Response** `409`: simulation not running or already paused

### Resume a simulation

- **Method**: `POST`
- **Path**: `/simulations/resume`
- **Response** `202`: no body
- **Behavior**:
  - Resumes all paused running simulations.
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

### Override an aircraft landing assignment

- **Method**: `POST`
- **Path**: `/simulations/:simulationId/aircraft/:tailNumber/assignment-override`
- **Body**:

```json
{
  "baseId": "3a5f..."
}
```

- **Behavior**:
  - Applies a human landing assignment override for the specified aircraft tail number.
  - V1 supports **set override only**; clearing an override is not exposed over HTTP.
  - The override is allowed only before the aircraft is committed to landing/service flow.
  - The response returns both the aircraft read model and the assignment metadata.

- **Response** `200`:

```json
{
  "aircraft": {
    "tailNumber": "9b2e...",
    "state": "Inbound",
    "needs": [],
    "assignedTo": "3a5f...",
    "position": {"x": 0, "y": 0}
  },
  "assignment": {
    "base": "3a5f...",
    "source": "human"
  }
}
```

- **Response** `400`: invalid tail number, base ID, or target airbase
- **Response** `404`: simulation or aircraft not found
- **Response** `409`: assignment override too late

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
  "type": "simulation_ended",
  "simulationId": "base",
  "tick": 3,
  "timestamp": "2026-03-11T18:00:15Z"
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

- `simulationId` is part of the path for base and branch simulations.
- Current implementation supports `simulationId=base` plus service-generated branch IDs created from the base simulation.
- Branch creation is available over HTTP via `POST /simulations/:simulationId/branch`.
- First branch support is base simulation only; checkpoint-based branch creation and branch-from-branch workflows are not implemented.
- Determinism guarantee: branching copies the current simulation state and RNG state, so if base and branch advance equivalently after branching they produce the same future behavior.
- The local tester now auto-creates the base simulation at startup, shows `Base` as the initial tab, and adds separate tabs for created branches.
- Airbase IDs and tail numbers are serialized as lowercase hex strings.
