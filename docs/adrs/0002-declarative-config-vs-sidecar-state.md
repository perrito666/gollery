# ADR 0002: Declarative Config vs Sidecar State

- Status: Accepted

## Context

The system needs both declarative publication config and mutable editorial state.

## Decision

Use:
- `album.json` for declarative configuration
- `.gallery/*.state.json` for mutable editorial/operational state
