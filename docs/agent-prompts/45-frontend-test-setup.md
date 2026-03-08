# Prompt 45 — Frontend test setup

Set up a lightweight test runner for the frontend core modules.

Implement:
- configure Node's built-in test runner (`node --test`) in `package.json`
- add `make test` target to the frontend Makefile
- write tests for:
  - `Store`: set, get, subscribe, unsubscribe
  - `Router`: pattern matching, param extraction, onNotFound
  - `FeatureFlags`: default flags, siteConfig overrides, enable/disable
  - `ComponentRegistry`: register, get, override
- mock `window` and `fetch` where needed using minimal stubs

Do not add a test framework dependency.
Use only Node built-in test runner and assert module.
