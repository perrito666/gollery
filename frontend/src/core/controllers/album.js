/**
 * Album controller — fetches album data and produces view models.
 *
 * Views never call the API directly. Controllers do.
 */

import { handleApiError } from './errors.js';

export class AlbumController {
  constructor(api, store) {
    this.api = api;
    this.store = store;
  }

  /** Load the root album and push its view model to the store. */
  async showRoot() {
    this.store.set({ loading: true, error: null });
    try {
      const album = await this.api.getAlbumsRoot();
      const viewModel = this._toViewModel(album);
      this.store.set({ currentView: 'home', viewModel, loading: false });
    } catch (err) {
      handleApiError(this.store, err);
    }
  }

  /** Load a specific album by ID. */
  async showAlbum(id) {
    this.store.set({ loading: true, error: null });
    try {
      const album = await this.api.getAlbum(id);
      const viewModel = this._toViewModel(album);
      this.store.set({ currentView: 'album', viewModel, loading: false });
    } catch (err) {
      handleApiError(this.store, err);
    }
  }

  _toViewModel(album) {
    const assets = (album.assets || []).map(a => this._assetToViewModel(a));
    return {
      id: album.id,
      title: album.title,
      description: album.description || '',
      path: album.path,
      children: (album.children || []).map(c => ({ id: c.id, path: c.path, title: c.title })),
      assets,
      // totalAssets is the true count after ACL filtering (see backend
      // AlbumResponse.total_assets). The view uses it to know when to
      // stop lazily loading more pages.
      totalAssets: typeof album.total_assets === 'number' ? album.total_assets : assets.length,
    };
  }

  _assetToViewModel(a) {
    return {
      id: a.id,
      filename: a.filename,
      title: a.title || '',
      description: a.description || '',
      thumbnailURL: this.api.thumbnailURL(a.id),
    };
  }

  /**
   * Fetch a page of assets for an album and return them as view models.
   * Does not touch the store — the view appends them incrementally to
   * avoid re-rendering the full grid (which would reset scroll).
   */
  async loadAssetsPage(id, { offset, limit }) {
    const album = await this.api.getAlbum(id, { offset, limit });
    const assets = (album.assets || []).map(a => this._assetToViewModel(a));
    const total = typeof album.total_assets === 'number' ? album.total_assets : (offset + assets.length);
    return { assets, total };
  }
}
