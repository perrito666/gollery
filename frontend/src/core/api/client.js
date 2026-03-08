/**
 * API client — communicates with the backend REST API.
 *
 * All backend interaction goes through this client.
 * Views must never call fetch() directly.
 */

export class ApiClient {
  constructor(baseURL = '/api/v1') {
    this.baseURL = baseURL;
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

  async login(username, password) {
    return this._post('/auth/login', { username, password });
  }

  async getMe() {
    return this._get('/auth/me');
  }

  async logout() {
    return this._post('/auth/logout', {});
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
    const resp = await fetch(this.baseURL + path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!resp.ok) {
      const data = await resp.json().catch(() => ({}));
      throw new ApiError(resp.status, data.message || resp.statusText);
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
