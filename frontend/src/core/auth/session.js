/**
 * Auth/session management.
 *
 * Wraps the API client's auth endpoints and manages the
 * current principal in the store.
 */

export class Session {
  constructor(api, store) {
    this.api = api;
    this.store = store;
  }

  /** Try to restore session from server. */
  async restore() {
    try {
      const me = await this.api.getMe();
      this.store.set({ principal: me });
      // Fetch CSRF token for POST requests.
      await this.api.fetchCSRFToken();
    } catch {
      this.store.set({ principal: null });
    }
  }

  /** Log in with credentials. */
  async login(username, password) {
    const result = await this.api.login(username, password);
    this.store.set({ principal: result });
    return result;
  }

  /** Log out. */
  async logout() {
    await this.api.logout();
    this.store.set({ principal: null });
  }

  /** Check if the user is currently authenticated. */
  isAuthenticated() {
    return this.store.get().principal !== null;
  }
}
