# Prompt 32 — Backend analytics event recording middleware

Wire analytics event recording into the API layer.

Implement:
- an analytics middleware or post-handler hook that records events for album views, asset views, and original hits
- respect `album.json` analytics.enabled flag — skip recording when disabled for a subtree
- implement dedup window: skip duplicate events from the same visitor within `dedup_window_seconds`
- use `HashVisitorID` for visitor identification
- make the middleware a no-op when analytics store is nil
- tests

Do not change the analytics store interface.
