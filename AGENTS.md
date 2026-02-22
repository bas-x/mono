# AGENTS.md

## Product Overview

Smart Airbase (Bas X) is a decision-support system for sortie turnaround planning.

The system simulates post-landing operations and helps operators choose plans that reduce **Landing -> Next Sortie time**. It combines deterministic simulation, auto base suggestions, resource-aware scheduling, replay, and branch comparison.

What differentiates it:

- Deterministic event-driven simulation (same inputs + seed => same outputs)
- Replayable runs for debugging and trust
- Branching futures for plan comparison before execution

Hackathon demo goal:

- Show measurable turnaround-time improvement on a constrained multi-aircraft scenario
- Explain _why_ a plan is better using timeline and bottleneck visibility

## Domain Model (Authoritative Glossary)

This glossary is the source of truth for terms used in code, docs, and UI.

- **Base**: Operational location where aircraft are assigned after landing and prepared for next sortie.
- **Runway**: Arrival/departure surface with sequencing and occupancy constraints.
- **Pad**: Parking/service position where turnaround tasks occur.
- **Aircraft**: Unit moving through landing, servicing, and re-launch workflow.
- **Resource**: Limited support capability (for example fuel truck, weapons crew, repair bay) required by tasks.
- **Task**: Work item in turnaround flow (refuel, rearm, repair, load, taxi).
- **Event**: Immutable record of a state transition at a specific simulation time.
- **Simulation Run**: Complete execution of a scenario from initial conditions + policy + seed.
- **Seed**: Deterministic input controlling any pseudo-random decision path.
- **Timeline**: Ordered event history for a run; foundation for replay and analysis.
- **Plan**: Candidate policy/schedule/allocation strategy evaluated by simulation.
- **Branch**: Alternative run derived from a prior timeline checkpoint with modified decisions.
- **Critical Path**: Sequence of dependent tasks/events that determines turnaround completion time.
- **Bottleneck**: Resource or dependency that most constrains throughput and increases turnaround time.

## System Architecture Overview

- **Backend (`/backend`)**: Deterministic event-driven simulation engine and planning logic.
- **Simulation model**: State transitions are represented as immutable events; state is reconstructed/advanced from events.
- **Interface layer**: Transport-neutral API abstraction. Frontend uses HTTP endpoints for request/response and WebSocket streams for realtime updates.
- **Frontend (`/frontend`)**: UI rendering layer for map, timeline, branch controls, and plan comparison.
- **Replay + branching**: Timeline checkpoints allow branch creation and side-by-side run comparison.

## Determinism Rules (Critical)

- No randomness without an explicit seed.
- No `Date.now()` (or wall-clock reads) inside simulation logic.
- No hidden side effects in simulation steps.
- Simulation state must be derived from event history only.
- Same inputs + seed must produce identical outputs (events, metrics, and terminal state).

## Demo Narrative

Target demo sequence:

- Multiple aircraft arrive within a short window.
- Shared resources become contested (fueling, rearming, repair capacity).
- System proposes base allocations and schedule decisions.
- Operator applies an override to one decision.
- Operator creates a branch and compares outcomes.
- Demo highlights improved Landing -> Next Sortie metric and explains bottlenecks.

## Development Guidelines for AI Agents

- Always update relevant context files when adding or changing features.
- Never introduce non-deterministic behavior into simulation.
- Keep domain language aligned with this glossary.
- Frontend must not reimplement simulation logic.
- Keep API/realtime integration behind the `src/lib/api/*` abstraction.
- New features must function in replay mode.

## AI Update Rule

When implementing or modifying a feature:

- Update relevant `docs/context/*` files.
- If introducing new domain terms, update the glossary.
- If modifying simulation behavior, document determinism impact.
- If adding or changing HTTP/WebSocket endpoints, document in `api.md`.
- If affecting demo flow, update demo documentation.
