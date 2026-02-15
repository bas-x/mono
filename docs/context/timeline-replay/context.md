# Timeline Replay Context

## Replay Model
- Replay is driven from immutable event history.
- UI state at time T is reconstructed from events up to T.

## Branching Rules
- A branch is created from a selected checkpoint/time index.
- Branch inherits prior history up to checkpoint and diverges only from modified decisions onward.
- Parent and child runs remain separately addressable for comparison.

## State Reconstruction
- Reconstruction must be deterministic and side-effect free.
- Derived views (charts, overlays, map annotations) must be computed from reconstructed state, not cached mutable shortcuts.

## Frontend Scrubbing
- Timeline scrubber navigates event index/time position.
- Scrubbing must not mutate backend truth; it only changes viewed reconstruction point.

## Compare Runs
- Compare mode aligns runs by consistent coordinate (time/event index/phase).
- Show metric deltas, bottleneck shifts, and critical-path differences.
