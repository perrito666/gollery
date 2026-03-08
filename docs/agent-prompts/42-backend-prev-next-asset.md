# Prompt 42 — Backend previous/next asset navigation

Extend the asset API response with navigation context.

Implement:
- add `prev_asset_id` and `next_asset_id` fields to `AssetResponse`
- determine previous/next by the asset's position within its album's asset list (sorted by filename)
- return null at the boundaries (first asset has no prev, last has no next)
- tests for middle, first, last, and single-asset albums

Do not change the album asset listing.
