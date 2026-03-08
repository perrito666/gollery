/**
 * Asset controller — fetches asset data and produces view models.
 */

import { handleApiError } from './errors.js';

export class AssetController {
  constructor(api, store) {
    this.api = api;
    this.store = store;
  }

  /** Load a specific asset by ID. */
  async showAsset(id) {
    this.store.set({ loading: true, error: null });
    try {
      const asset = await this.api.getAsset(id);
      const viewModel = {
        id: asset.id,
        filename: asset.filename,
        albumPath: asset.album_path,
        albumId: asset.album_id,
        previewURL: this.api.previewURL(asset.id),
        originalURL: this.api.originalURL(asset.id),
        prevAssetId: asset.prev_asset_id || null,
        nextAssetId: asset.next_asset_id || null,
      };
      this.store.set({ currentView: 'asset', viewModel, loading: false });
    } catch (err) {
      handleApiError(this.store, err);
    }
  }
}
