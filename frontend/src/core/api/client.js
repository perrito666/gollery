/**
 * API client — communicates with the backend REST API.
 *
 * All backend interaction goes through this client.
 * Views must never call fetch() directly.
 */

export class ApiClient {
  constructor(baseURL = '/api/v1') {
    this.baseURL = baseURL;
    this._csrfToken = null;
  }

  async getAlbumsRoot() {
    return this._get('/albums/root');
  }

  async getAlbum(id) {
    return this._get(`/albums/${encodeURIComponent(id)}`);
  }

  async getAsset(id) {
    return this._get(`/assets/${encodeURIComponent(id)}`);
  }

  thumbnailURL(assetId, size = 400) {
    return `${this.baseURL}/assets/${encodeURIComponent(assetId)}/thumbnail?size=${size}`;
  }

  previewURL(assetId, size = 1600) {
    return `${this.baseURL}/assets/${encodeURIComponent(assetId)}/preview?size=${size}`;
  }

  originalURL(assetId) {
    return `${this.baseURL}/assets/${encodeURIComponent(assetId)}/original`;
  }

  async patchAssetMetadata(assetId, metadata) {
    return this._patch(`/assets/${encodeURIComponent(assetId)}/metadata`, metadata);
  }

  async patchAlbumMetadata(albumId, metadata) {
    return this._patch(`/albums/${encodeURIComponent(albumId)}/metadata`, metadata);
  }

  async getAssetDiscussions(assetId) {
    return this._get(`/assets/${encodeURIComponent(assetId)}/discussion-threads`);
  }

  async createAssetDiscussion(assetId, payload) {
    return this._post(`/assets/${encodeURIComponent(assetId)}/discussion-threads`, payload);
  }

  async getAlbumDiscussions(albumId) {
    return this._get(`/albums/${encodeURIComponent(albumId)}/discussion-threads`);
  }

  async login(username, password) {
    const result = await this._post('/auth/login', { username, password });
    // Fetch CSRF token after successful login.
    await this._fetchCSRFToken();
    return result;
  }

  async getMe() {
    return this._get('/auth/me');
  }

  async logout() {
    await this._post('/auth/logout', {});
    this._csrfToken = null;
  }

  /** Fetch a CSRF token from the server. Called after login and session restore. */
  async fetchCSRFToken() {
    await this._fetchCSRFToken();
  }

  async _fetchCSRFToken() {
    try {
      const data = await this._get('/auth/csrf-token');
      this._csrfToken = data.token;
    } catch {
      this._csrfToken = null;
    }
  }

  async _get(path) {
    const resp = await fetch(this.baseURL + path);
    if (!resp.ok) {
      const body = await resp.json().catch(() => ({}));
      throw new ApiError(resp.status, body.message || resp.statusText);
    }
    return resp.json();
  }

  async _post(path, body) {
    return this._mutate('POST', path, body);
  }

  async _patch(path, body) {
    return this._mutate('PATCH', path, body);
  }

  async _mutate(method, path, body) {
    const headers = { 'Content-Type': 'application/json' };
    if (this._csrfToken) {
      headers['X-CSRF-Token'] = this._csrfToken;
    }
    const resp = await fetch(this.baseURL + path, {
      method,
      headers,
      body: JSON.stringify(body),
    });
    if (!resp.ok) {
      const data = await resp.json().catch(() => ({}));
      throw new ApiError(resp.status, data.message || resp.statusText);
    }
    // Handle empty responses (e.g. 204 No Content).
    if (resp.status === 204 || resp.headers.get('content-length') === '0') {
      return {};
    }
    return resp.json();
  }
}

export class ApiError extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
    this.name = 'ApiError';
  }
}
