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
    return {
      id: album.id,
      title: album.title,
      description: album.description || '',
      path: album.path,
      children: (album.children || []).map(c => ({ id: c.id, path: c.path, title: c.title })),
      assets: (album.assets || []).map(a => ({
        id: a.id,
        filename: a.filename,
        thumbnailURL: this.api.thumbnailURL(a.id),
      })),
    };
  }

}
