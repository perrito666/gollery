# Prompt 11 — Backend popularity analytics with PostgreSQL

Implement the optional analytics subsystem.

Requirements:
- PostgreSQL-backed analytics store
- event model for album views, asset views, original hits, discussion clicks
- privacy-safe design
- hashed identifiers if needed
- retention-aware schema or maintenance hooks
- popularity queries (`views_7d`, `views_30d`, etc.)
- migrations and tests

Do not make analytics required for gallery correctness.
