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
  - `simulationOptions` can be partial; omitted option groups (`constellationOpts`, `fleetOpts`, `threatOpts`, `lifecycleOpts`) inherit backend defaults instead of zero-values.
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

- **Behavior**:
  - Branch entries return lineage metadata on the same `SimulationInfo` shape used by branch-create and detail reads.
  - `sourceEvent` is optional and is omitted when the branch has no clicked-anchor metadata, including legacy branches.
  - `splitTick` and `splitTimestamp` are the real fork coordinates. `sourceEvent.id`, `sourceEvent.type`, and `sourceEvent.tick` are metadata only.

### Get a simulation

- **Method**: `GET`
- **Path**: `/simulations/:simulationId`
- **Response** `200`:

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

- **Behavior**:
  - Base simulation reads return `parentId`, `splitTick`, and `splitTimestamp` as `null`, and omit `sourceEvent`.
  - Branch detail reads return the same lineage metadata surfaced by branch creation and list reads.
  - `sourceEvent` is optional and omitted when absent. It does not change the branch snapshot.
- **Response** `404`: simulation not found

### Branch a simulation

- **Method**: `POST`
- **Path**: `/simulations/:simulationId/branch`
- **Body**: optional

```json
{
  "sourceEvent": {
    "id": "timeline-evt-17",
    "type": "landing_assignment",
    "tick": 41
  }
}
```

- **Response** `201`:

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

- **Behavior**:
  - V1 only supports branching from `simulationId=base`.
  - The request body is optional. Clients can still send no body.
  - If `sourceEvent` is provided, `id`, `type`, and `tick` are all required. Empty or partial `sourceEvent` payloads fail with `400`.
  - The response is the new branch `SimulationInfo`, including lineage metadata.
  - `splitTick` and `splitTimestamp` are the true branch fork coordinates copied from the base snapshot.
  - `sourceEvent.id`, `sourceEvent.type`, and `sourceEvent.tick` are clicked-anchor metadata only. They do not change branch snapshot semantics.
  - When accepted, `sourceEvent` is persisted in runtime branch state and is returned by branch-create, `GET /simulations`, `GET /simulations/:simulationId`, and `branch_created` for the current backend process lifetime.
  - `sourceEvent` is omitted from JSON when absent.
- **Response** `400`: malformed or partial `sourceEvent`
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
- **Behavior**:
  - Removes the simulation immediately.
  - Also emits a terminal `simulation_closed` websocket event for that `simulationId`.
  - `reason` is required on `simulation_closed`, `reset` when the base simulation is removed and `cancel` when a branch is removed.
  - `simulation_closed.summary` uses the same direct servicing summary shape as `simulation_ended.summary`: `completedVisitCount`, `totalDurationMs`, `averageDurationMs`.
  - `totalDurationMs` and `averageDurationMs` are milliseconds. `averageDurationMs` is `null` until at least one servicing visit completes.
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
      "capabilities": {
        "fuel": {"recoveryMultiplierPermille": 1300},
        "mission_configuration": {"recoveryMultiplierPermille": 1050}
      },
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
  - The override is allowed only while the aircraft is still in the inbound pre-commit window.
  - If the aircraft has no assignment yet, the service may first register the algorithmic landing assignment and then apply the human override.
  - The response returns both the aircraft read model and the assignment metadata.

- **Response** `200`:

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
  - Current event types: `simulation_step`, `simulation_ended`, `simulation_closed`, `aircraft_state_change`, `landing_assignment`, `all_aircraft_positions`, `threat_spawned`, `threat_targeted`, `threat_despawned`, `branch_created`.
  - `simulation_ended` is the natural terminal event. It includes a direct `summary` object with `completedVisitCount`, `totalDurationMs`, and nullable `averageDurationMs`.
  - `simulation_closed` is the non-natural terminal event emitted when the simulation is removed. It includes the same `summary` object plus required `reason` (`reset` for base reset, `cancel` for branch reset).
  - `totalDurationMs` and `averageDurationMs` are milliseconds. `averageDurationMs` stays `null` until at least one servicing visit completes.
  - `branch_created` is emitted on `/ws/simulations/base/events` when a new V1 branch is created from the base simulation. It carries branch lineage summary fields for the new branch, may include optional `sourceEvent`, and stays on the base stream because the event `simulationId` is `base`.
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
  "timestamp": "2026-03-11T18:00:15Z",
  "summary": {
    "completedVisitCount": 0,
    "totalDurationMs": 0,
    "averageDurationMs": null
  }
}
```

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
  "type": "threat_targeted",
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

## Notes

- `simulationId` is part of the path for base and branch simulations.
- Current implementation supports `simulationId=base` plus service-generated branch IDs created from the base simulation.
- Branch creation is available over HTTP via `POST /simulations/:simulationId/branch`; successful base branch creation also emits `branch_created` on `/ws/simulations/base/events`.
- Branch lineage metadata is available from the branch creation response, `GET /simulations`, `GET /simulations/:simulationId`, and the base-stream `branch_created` event.
- `splitTick` and `splitTimestamp` remain the canonical fork coordinates. Optional `sourceEvent` fields are clicked-anchor metadata only and are omitted when absent.
- `sourceEvent` persistence is runtime-only. Restarting the backend loses that metadata along with in-memory branch state.
- Terminal websocket events use a direct `summary` object, not nested `summary.servicing`.
- `simulation_ended` is natural completion; `simulation_closed` is non-natural removal and adds required `reason`.
- First branch support is base simulation only; checkpoint-based branch creation and branch-from-branch workflows are not implemented.
- Determinism guarantee: branching copies the current simulation state and RNG state, so if base and branch advance equivalently after branching they produce the same future behavior.
- The local tester now auto-creates the base simulation at startup, shows `Base` as the initial tab, and adds separate tabs for created branches.
- Airbase IDs and tail numbers are serialized as lowercase hex strings.
