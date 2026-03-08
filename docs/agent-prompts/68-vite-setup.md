# Prompt 68 — Frontend: initialize Vite project

Replace the esbuild + resolve-theme.js build pipeline with Vite while preserving the theme override system.

## Current state

- `frontend/package.json` has only `esbuild` as a devDependency
- `scripts/resolve-theme.js` generates `src/_resolved/registry.js` at build time
- `Makefile` orchestrates: resolve-theme → esbuild bundle → copy assets
- Bundle output: `dist/app.js` (ESM, minified)

## Implement

1. Install Vite: `npm install --save-dev vite`

2. Create `frontend/vite.config.js`:
   - Set `root` to `frontend/`
   - Set `build.outDir` to `dist/`
   - Set `build.target` to `es2020`
   - Entry point: `src/main.js`
   - Configure `publicDir` to `public/`

3. Update `frontend/package.json` scripts:
   - `"dev": "vite"` (replaces esbuild watch)
   - `"build": "node scripts/resolve-theme.js && vite build"` (replaces esbuild bundle)
   - `"preview": "vite preview"`
   - Keep `"test": "node --test test/*.test.js"`

4. Create `frontend/index.html` at project root (Vite requires it):
   - Move the content from `public/index.html`
   - Add `<script type="module" src="/src/main.js"></script>`
   - Link CSS via `<link>` tag

5. Update `frontend/Makefile`:
   - `build` target: `npm run build`
   - `dev` target: `npm run dev`
   - Remove esbuild-specific flags

6. Verify the build produces a working bundle.

## Verify

```bash
cd frontend && npm run build
# Check dist/ contains index.html and JS bundle
ls -la dist/
```

## Do not

- Remove `resolve-theme.js` — it's still needed for view override resolution
- Change any view or core source files
- Add CSS modules or other Vite plugins yet
- Remove esbuild from devDependencies yet (next prompt confirms everything works)
