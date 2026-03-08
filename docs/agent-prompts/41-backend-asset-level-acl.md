# Prompt 41 — Backend asset-level ACL overrides

Support per-asset access control from sidecar state.

Implement:
- extend `state.AssetState` with an optional `Access *config.AccessConfig` field
- extend `domain.Asset` with an optional `Access *config.AccessConfig` field
- populate asset-level access during snapshot building from sidecar state
- in API handlers, check asset-level ACL first (if set), fall back to album ACL
- update `access.CheckView` or add a helper that resolves the effective ACL for an asset
- tests for asset-level override, album-level fallback, and admin override

Do not change the ACL evaluation modes.
