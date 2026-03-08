/**
 * Permission interpretation service.
 *
 * Interprets the principal from the store to determine
 * what the current user can do. Views use this to show/hide
 * actions — actual enforcement happens server-side.
 */

export class PermissionService {
  constructor(store) {
    this.store = store;
  }

  /** Whether the current user is logged in. */
  isAuthenticated() {
    return this.store.get().principal !== null;
  }

  /** Whether the current user is an admin. */
  isAdmin() {
    const p = this.store.get().principal;
    return p !== null && p.is_admin === true;
  }

  /** Whether the current user can view a resource given its ACL mode. */
  canView(accessMode) {
    if (accessMode === 'public') return true;
    if (accessMode === 'authenticated') return this.isAuthenticated();
    // 'restricted' — server enforces, but we optimistically allow if authenticated
    return this.isAuthenticated();
  }
}
