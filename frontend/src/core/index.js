/**
 * Core layer entry point.
 *
 * Owns: API access, auth/session, router, state, controllers,
 * action execution, permission interpretation.
 */

import { ApiClient } from './api/client.js';
import { Router } from './router/router.js';
import { Store } from './state/store.js';
import { Session } from './auth/session.js';
import { AlbumController } from './controllers/album.js';
import { AssetController } from './controllers/asset.js';
import { PermissionService } from './services/permissions.js';
import { PopularityClient } from './api/popularity.js';
import { FeatureFlags } from './services/features.js';

/**
 * Initialize the core layer.
 *
 * @param {import('../ui-contract/interfaces.js').ComponentRegistry} registry
 * @param {HTMLElement} container - The root DOM element for rendering
 * @param {Object} [siteConfig] - Site configuration from resolve-theme
 * @returns {Object} The app context (store, router, session, permissions, features, popularity)
 */
export function init(registry, container, siteConfig = {}) {
  const api = new ApiClient();
  const store = new Store();
  const session = new Session(api, store);
  const permissions = new PermissionService(store);
  const features = new FeatureFlags(siteConfig);
  const popularity = new PopularityClient(api);
  const albumCtrl = new AlbumController(api, store);
  const assetCtrl = new AssetController(api, store);
  const router = new Router();

  // Track the active renderer so we can destroy it on view change.
  let activeRenderer = null;

  // Subscribe to state changes — re-render when view changes.
  store.subscribe((state) => {
    let viewName = state.currentView;

    // If no view is set but there's an error, show the error view.
    if (!viewName && state.error) {
      viewName = 'error';
    }
    if (!viewName) return;

    const renderer = registry.get(viewName);
    if (!renderer) {
      console.warn(`gollery: no renderer registered for view "${viewName}"`);
      return;
    }

    if (activeRenderer && activeRenderer !== renderer) {
      activeRenderer.destroy();
    }
    activeRenderer = renderer;
    renderer.render(container, state.viewModel, { store, router, session, permissions, features, popularity });
  });

  // Routes
  router
    .on('/', () => albumCtrl.showRoot())
    .on('/albums/:id', (params) => albumCtrl.showAlbum(params.id))
    .on('/assets/:id', (params) => assetCtrl.showAsset(params.id));

  // Catch-all for unmatched routes.
  router.onNotFound(() => {
    store.set({ currentView: 'not-found', viewModel: null, loading: false, error: null });
  });

  // Restore session, then start router.
  session.restore().then(() => {
    router.start();
  });

  return { store, router, session, permissions, features, popularity };
}
