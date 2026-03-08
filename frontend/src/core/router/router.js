/**
 * Hash-based client-side router.
 *
 * Routes map URL patterns to controller actions.
 * Uses hash fragments (e.g. #/albums/alb_xxx) to avoid server config.
 */

export class Router {
  constructor() {
    this.routes = [];
    this._notFoundHandler = null;
    this._onHashChange = this._onHashChange.bind(this);
  }

  /**
   * Register a route pattern.
   * Patterns use :param for named segments (e.g. '/albums/:id').
   */
  on(pattern, handler) {
    const regex = this._patternToRegex(pattern);
    this.routes.push({ pattern, regex, handler });
    return this;
  }

  /** Start listening for hash changes. */
  start() {
    window.addEventListener('hashchange', this._onHashChange);
    this._onHashChange();
  }

  /** Stop listening. */
  stop() {
    window.removeEventListener('hashchange', this._onHashChange);
  }

  /** Register a handler for unmatched routes. */
  onNotFound(handler) {
    this._notFoundHandler = handler;
    return this;
  }

  /** Navigate to a hash path. */
  navigate(path) {
    window.location.hash = '#' + path;
  }

  /** Get the current hash path. */
  currentPath() {
    return window.location.hash.slice(1) || '/';
  }

  _onHashChange() {
    const path = this.currentPath();
    for (const route of this.routes) {
      const match = path.match(route.regex);
      if (match) {
        const params = this._extractParams(route.pattern, match);
        route.handler(params);
        return;
      }
    }
    // No match — invoke not-found handler if registered.
    if (this._notFoundHandler) {
      this._notFoundHandler();
    }
  }

  _patternToRegex(pattern) {
    const escaped = pattern.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const withParams = escaped.replace(/:(\w+)/g, '([^/]+)');
    return new RegExp('^' + withParams + '$');
  }

  _extractParams(pattern, match) {
    const params = {};
    const paramNames = [];
    const paramRegex = /:(\w+)/g;
    let m;
    while ((m = paramRegex.exec(pattern)) !== null) {
      paramNames.push(m[1]);
    }
    for (let i = 0; i < paramNames.length; i++) {
      params[paramNames[i]] = decodeURIComponent(match[i + 1]);
    }
    return params;
  }
}
