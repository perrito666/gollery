# Prompt 40 — Backend cache eviction

Add cleanup for orphaned derivative cache files.

Implement:
- `cache.PurgeOrphans(layout *Layout, knownAssetIDs map[string]bool) (int, error)` that removes cached files for asset IDs no longer in the snapshot
- call `PurgeOrphans` during snapshot rebuild (after the new snapshot is built, before replacing the old one)
- log the number of evicted files
- tests with a temp directory containing orphaned and valid cache files

Do not add LRU eviction or size-based limits.
