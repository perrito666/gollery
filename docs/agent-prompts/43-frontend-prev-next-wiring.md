# Prompt 43 — Frontend previous/next asset wiring

Wire the backend's prev/next asset data into the frontend.

Implement:
- update `AssetController` to read `prev_asset_id` and `next_asset_id` from the API response and set them in the view model
- the asset view already renders navigation arrows based on these fields — verify they work
- tests where practical (mock API response with prev/next)

Do not change view rendering logic.
