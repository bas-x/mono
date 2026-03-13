# Simulation Engine Context

## Execution Model
- Core model is event-driven, not fixed-tick simulation.
- Time advances to the next scheduled event timestamp.
- Optional "tick-like" UI updates are derived views and must not drive simulation state.

## Time Advancement
- Simulation clock moves monotonically to each event time.
- No wall-clock dependencies in simulation decisions.
- Deterministic ordering is required for same-time events (stable tie-break rules).

## Current Aircraft State Progression
- `simulation/state.go` applies deterministic per-step aircraft movement using `ctx.Clock.Resolution` as the movement interval.
- Ready aircraft cache the claimed threat and compute a deterministic target point from the centroid of the first polygon in the named threat region.
- Outbound aircraft move toward the cached threat centroid and may transition to `Engaged` either by elapsed duration or by reaching `EngagementRange` proximity.
- Engaged aircraft orbit the cached threat centroid with fixed `OrbitRadius` and `OrbitAngleDeltaPerTick`; no randomness is used in orbit updates.
- Inbound and Committed aircraft move toward the assigned airbase location; Committed transitions to `Servicing` primarily by `LandingRange` proximity with the existing duration-based fallback retained.
- Servicing snaps the aircraft position to the assigned base on entry and does not continue moving the aircraft afterward.

## Resource Locking
- Tasks require explicit resource reservation before start.
- Resource lock lifetime is defined by task start/end events.
- Locks are released only via corresponding completion/cancel events.

## Conflict Resolution
- Conflicts are resolved by deterministic policy order (for example priority, then FIFO, then stable ID tie-break).
- No implicit race behavior; conflict outcomes must be reproducible from inputs + seed.

## Current Frontend Configuration Surface
- The simulation setup UI still exposes the same operator-facing controls and checkboxes, but `POST /simulations/base` now submits them under `simulationOptions.constellationOpts` and `simulationOptions.fleetOpts` instead of flattening everything to a seed-only request.
- User-editable airbase fields are region filters, per-region counts, max total, and region probability.
- User-editable fleet fields are aircraft count range, need count range, need pool, severity range, and blocking chance.
- `simulationOptions.threatOpts` and `simulationOptions.lifecycleOpts` are part of the request contract but are not yet surfaced as editable controls in the current UI.
- Internal factories (`MetadataFactory`, `StateFactory`) and low-level generation controls remain backend-owned and are not exposed in the UI.

## Performance Expectations
- Must handle demo scenarios with interactive replay responsiveness.
- Event processing should scale with event count and avoid per-step full-state recomputation.
- Replay/branching should reuse event history efficiently where possible.
