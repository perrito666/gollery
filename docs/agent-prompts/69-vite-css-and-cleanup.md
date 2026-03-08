# Prompt 69 — Frontend: Vite CSS handling and esbuild removal

After prompt 68 set up Vite, migrate CSS handling and remove esbuild.

## Implement

1. Import CSS from JavaScript instead of manual concatenation:
   - In `src/main.js`, add: `import './ui-default/styles/main.css'`
   - If site CSS exists, import it after: `import './site/styles/site.css'` (conditionally, or via resolve-theme.js)

2. Update `resolve-theme.js` to also generate CSS imports in the registry file, so site CSS overrides are bundled automatically.

3. Remove the Makefile CSS copy/concatenation steps.

4. Remove esbuild from `package.json` devDependencies: `npm uninstall esbuild`

5. Update root `Makefile` `frontend-build` target if it references esbuild commands.

6. Run frontend tests to confirm nothing broke.

## Verify

```bash
cd frontend && npm run build && npm test
# Check that dist/ contains CSS (either inline or as a separate file)
```

## Do not

- Add CSS modules or PostCSS — keep plain CSS
- Change any view implementations
- Add new Vite plugins
