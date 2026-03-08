import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { Store } from '../src/core/state/store.js';
import { AlbumController } from '../src/core/controllers/album.js';
import { AssetController } from '../src/core/controllers/asset.js';
import { handleApiError } from '../src/core/controllers/errors.js';
import { Session } from '../src/core/auth/session.js';

/** Create a fake API client with canned responses. */
function fakeApi(overrides = {}) {
  return {
    getAlbumsRoot: async () => ({
      id: 'alb_root', title: 'Root', path: '', children: ['photos'], assets: [
        { id: 'ast_1', filename: 'pic.jpg' },
      ],
    }),
    getAlbum: async (id) => ({
      id, title: 'Vacation', path: 'vacation', children: [], assets: [],
    }),
    getAsset: async (id) => ({
      id, filename: 'beach.jpg', album_path: 'vacation', album_id: 'alb_vac',
      prev_asset_id: 'ast_prev', next_asset_id: 'ast_next',
    }),
    thumbnailURL: (id) => `/thumb/${id}`,
    previewURL: (id) => `/preview/${id}`,
    originalURL: (id) => `/original/${id}`,
    getMe: async () => ({ username: 'alice', groups: [], is_admin: false }),
    login: async () => ({ username: 'alice', groups: [], is_admin: false }),
    logout: async () => {},
    ...overrides,
  };
}

function apiError(status, message) {
  const err = new Error(message);
  err.status = status;
  return err;
}

// --- handleApiError ---

describe('handleApiError', () => {
  it('401 navigates to login', () => {
    const store = new Store();
    handleApiError(store, apiError(401, 'unauthorized'));
    assert.equal(store.get().currentView, 'login');
    assert.equal(store.get().loading, false);
  });

  it('403 navigates to forbidden', () => {
    const store = new Store();
    handleApiError(store, apiError(403, 'forbidden'));
    assert.equal(store.get().currentView, 'forbidden');
  });

  it('404 navigates to not-found', () => {
    const store = new Store();
    handleApiError(store, apiError(404, 'not found'));
    assert.equal(store.get().currentView, 'not-found');
  });

  it('500 sets error message', () => {
    const store = new Store();
    handleApiError(store, apiError(500, 'server error'));
    assert.equal(store.get().currentView, null);
    assert.equal(store.get().error, 'server error');
  });
});

// --- AlbumController ---

describe('AlbumController', () => {
  it('showRoot sets home view', async () => {
    const store = new Store();
    const ctrl = new AlbumController(fakeApi(), store);
    await ctrl.showRoot();
    assert.equal(store.get().currentView, 'home');
    assert.equal(store.get().loading, false);
    assert.equal(store.get().viewModel.title, 'Root');
    assert.equal(store.get().viewModel.assets.length, 1);
  });

  it('showAlbum sets album view', async () => {
    const store = new Store();
    const ctrl = new AlbumController(fakeApi(), store);
    await ctrl.showAlbum('alb_vac');
    assert.equal(store.get().currentView, 'album');
    assert.equal(store.get().viewModel.id, 'alb_vac');
  });

  it('showRoot handles 401', async () => {
    const store = new Store();
    const api = fakeApi({
      getAlbumsRoot: async () => { throw apiError(401, 'unauthorized'); },
    });
    const ctrl = new AlbumController(api, store);
    await ctrl.showRoot();
    assert.equal(store.get().currentView, 'login');
  });

  it('showAlbum handles 403', async () => {
    const store = new Store();
    const api = fakeApi({
      getAlbum: async () => { throw apiError(403, 'forbidden'); },
    });
    const ctrl = new AlbumController(api, store);
    await ctrl.showAlbum('alb_priv');
    assert.equal(store.get().currentView, 'forbidden');
  });

  it('showAlbum handles 404', async () => {
    const store = new Store();
    const api = fakeApi({
      getAlbum: async () => { throw apiError(404, 'not found'); },
    });
    const ctrl = new AlbumController(api, store);
    await ctrl.showAlbum('alb_missing');
    assert.equal(store.get().currentView, 'not-found');
  });
});

// --- AssetController ---

describe('AssetController', () => {
  it('showAsset sets asset view with prev/next', async () => {
    const store = new Store();
    const ctrl = new AssetController(fakeApi(), store);
    await ctrl.showAsset('ast_1');
    assert.equal(store.get().currentView, 'asset');
    const vm = store.get().viewModel;
    assert.equal(vm.id, 'ast_1');
    assert.equal(vm.prevAssetId, 'ast_prev');
    assert.equal(vm.nextAssetId, 'ast_next');
    assert.equal(vm.previewURL, '/preview/ast_1');
  });

  it('showAsset handles error', async () => {
    const store = new Store();
    const api = fakeApi({
      getAsset: async () => { throw apiError(404, 'not found'); },
    });
    const ctrl = new AssetController(api, store);
    await ctrl.showAsset('ast_missing');
    assert.equal(store.get().currentView, 'not-found');
  });

  it('showAsset nulls prev/next when missing', async () => {
    const store = new Store();
    const api = fakeApi({
      getAsset: async (id) => ({
        id, filename: 'only.jpg', album_path: 'single', album_id: 'alb_s',
      }),
    });
    const ctrl = new AssetController(api, store);
    await ctrl.showAsset('ast_only');
    const vm = store.get().viewModel;
    assert.equal(vm.prevAssetId, null);
    assert.equal(vm.nextAssetId, null);
  });
});

// --- Session ---

describe('Session', () => {
  it('restore sets principal on success', async () => {
    const store = new Store();
    const session = new Session(fakeApi(), store);
    await session.restore();
    assert.equal(store.get().principal.username, 'alice');
  });

  it('restore clears principal on failure', async () => {
    const store = new Store();
    store.set({ principal: { username: 'old' } });
    const api = fakeApi({
      getMe: async () => { throw new Error('not authenticated'); },
    });
    const session = new Session(api, store);
    await session.restore();
    assert.equal(store.get().principal, null);
  });

  it('login sets principal', async () => {
    const store = new Store();
    const session = new Session(fakeApi(), store);
    await session.login('alice', 'pass');
    assert.equal(store.get().principal.username, 'alice');
  });

  it('logout clears principal', async () => {
    const store = new Store();
    store.set({ principal: { username: 'alice' } });
    const session = new Session(fakeApi(), store);
    await session.logout();
    assert.equal(store.get().principal, null);
  });

  it('isAuthenticated reflects state', () => {
    const store = new Store();
    const session = new Session(fakeApi(), store);
    assert.equal(session.isAuthenticated(), false);
    store.set({ principal: { username: 'alice' } });
    assert.equal(session.isAuthenticated(), true);
  });
});
