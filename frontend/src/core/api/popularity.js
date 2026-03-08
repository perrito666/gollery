/**
 * Popularity API client extension.
 *
 * Fetches popularity data from the backend if analytics endpoints are available.
 * Returns null gracefully when analytics are not enabled.
 */

export class PopularityClient {
  constructor(apiClient) {
    this.api = apiClient;
    this._available = null; // tri-state: null=unknown, true, false
  }

  /**
   * Probe whether the popularity API is available.
   * Caches the result after the first call.
   */
  async isAvailable() {
    if (this._available !== null) return this._available;
    try {
      const resp = await fetch(`${this.api.baseURL}/popularity/status`);
      this._available = resp.ok;
    } catch {
      this._available = false;
    }
    return this._available;
  }

  /**
   * Get popularity summary for an object (album or asset).
   * Returns null if analytics are unavailable or the request fails.
   *
   * @param {string} objectId
   * @returns {Promise<{totalViews: number, views7d: number, views30d: number, originalHits: number, discussionClicks: number}|null>}
   */
  async getPopularity(objectId) {
    if (!(await this.isAvailable())) return null;
    try {
      const resp = await fetch(
        `${this.api.baseURL}/popularity/${encodeURIComponent(objectId)}`
      );
      if (!resp.ok) return null;
      return resp.json();
    } catch {
      return null;
    }
  }

  /**
   * Get popular assets within an album.
   * Returns empty array if unavailable.
   *
   * @param {string} albumId
   * @param {number} limit
   * @returns {Promise<Array<{id: string, filename: string, totalViews: number}>>}
   */
  async getPopularInAlbum(albumId, limit = 10) {
    if (!(await this.isAvailable())) return [];
    try {
      const resp = await fetch(
        `${this.api.baseURL}/popularity/albums/${encodeURIComponent(albumId)}/popular?limit=${limit}`
      );
      if (!resp.ok) return [];
      return resp.json();
    } catch {
      return [];
    }
  }
}
