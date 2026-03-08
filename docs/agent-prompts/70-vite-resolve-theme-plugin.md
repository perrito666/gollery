# Prompt 70 — Frontend: Vite plugin for theme resolution

Replace the pre-build `resolve-theme.js` script with a Vite plugin that resolves theme overrides during the build.

## Current state

`resolve-theme.js` runs before the bundler and generates `src/_resolved/registry.js`. This works but requires a separate build step and generates a file that must be gitignored.

## Implement

1. Create `frontend/plugins/vite-plugin-theme-resolve.js`:
   - Export a Vite plugin that uses the `resolveId` and `load` hooks.
   - When `src/_resolved/registry.js` is imported, intercept and generate the module content dynamically (same logic as resolve-theme.js).
   - Check `src/site/views/` for overrides, fall back to `src/ui-default/views/`.
   - Read `src/site/site.config.json` for site config.

2. Register the plugin in `vite.config.js`.

3. Remove the `node scripts/resolve-theme.js` step from `package.json` build script (Vite handles it now).

4. Keep `resolve-theme.js` as a standalone tool (useful for debugging with `--print` flag) but it's no longer required for builds.

5. Delete `src/_resolved/` from `.gitignore` if it was gitignored, or remove the directory entirely since it's no longer generated on disk.

6. Verify the build still produces a working bundle with correct view resolution.

## Verify

```bash
cd frontend && npm run build && npm test
```

## Do not

- Change view implementations
- Add HMR for theme overrides (future enhancement)
- Change the site.config.json format
