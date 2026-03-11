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

## Notes

- `simulationId` is already part of the path to support future branch simulations.
- Current implementation supports only `simulationId=base`.
- Airbase IDs and tail numbers are serialized as lowercase hex strings.
