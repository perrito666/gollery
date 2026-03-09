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

      // Fetch discussions if available (non-fatal).
      let discussions = [];
      try {
        discussions = await this.api.getAssetDiscussions(id);
      } catch {
        // Discussions may not be configured — ignore.
      }

      const viewModel = {
        id: asset.id,
        filename: asset.filename,
        title: asset.title || '',
        description: asset.description || '',
        albumPath: asset.album_path,
        albumId: asset.album_id,
        previewURL: this.api.previewURL(asset.id),
        originalURL: this.api.originalURL(asset.id),
        prevAssetId: asset.prev_asset_id || null,
        nextAssetId: asset.next_asset_id || null,
        discussions,
      };
      this.store.set({ currentView: 'asset', viewModel, loading: false });
    } catch (err) {
      handleApiError(this.store, err);
    }
  }
}
