# Turnaround Planner Context

## Objective Function
- Primary objective: minimize **Landing -> Next Sortie time** per aircraft and across scenario.
- Secondary objective: reduce resource idle time and contention hotspots.

## Constraints
- Base/runway/pad capacity limits.
- Resource availability windows and exclusivity constraints.
- Task ordering and dependency constraints.
- Operational policy constraints (safety, sequencing, priority classes).

## Heuristic Approach (Hackathon Scope)
- Use deterministic heuristics for allocation and ordering.
- Evaluate candidate plans with simulation rather than static estimates alone.
- Favor fast, explainable strategies over globally optimal but opaque methods.

## What "Optimal" Means Here
- "Optimal" in hackathon scope means best among evaluated deterministic candidates under current constraints and runtime budget.
- Not a proof of global optimum.

## Future Extensibility
- Plug-in scoring weights and objective terms.
- Policy modules per mission type/base profile.
- Hybrid approach with search/optimization methods while keeping deterministic replay.
