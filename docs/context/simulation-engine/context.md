# Simulation Engine Context

## Execution Model
- Core model is event-driven, not fixed-tick simulation.
- Time advances to the next scheduled event timestamp.
- Optional "tick-like" UI updates are derived views and must not drive simulation state.

## Time Advancement
- Simulation clock moves monotonically to each event time.
- No wall-clock dependencies in simulation decisions.
- Deterministic ordering is required for same-time events (stable tie-break rules).

## Resource Locking
- Tasks require explicit resource reservation before start.
- Resource lock lifetime is defined by task start/end events.
- Locks are released only via corresponding completion/cancel events.

## Conflict Resolution
- Conflicts are resolved by deterministic policy order (for example priority, then FIFO, then stable ID tie-break).
- No implicit race behavior; conflict outcomes must be reproducible from inputs + seed.

## Current Frontend Configuration Surface
- The current frontend simulation setup UI is limited to fields backed by `simulation.SimulationOptions`, `simulation.ConstellationOptions`, and `simulation.FleetOptions`.
- User-editable airbase fields are region filters, per-region counts, max total, and region probability.
- User-editable fleet fields are aircraft count range, need count range, need pool, severity range, and blocking chance.
- Internal factories (`MetadataFactory`, `StateFactory`) and low-level generation controls remain backend-owned and are not exposed in the UI.

## Performance Expectations
- Must handle demo scenarios with interactive replay responsiveness.
- Event processing should scale with event count and avoid per-step full-state recomputation.
- Replay/branching should reuse event history efficiently where possible.
