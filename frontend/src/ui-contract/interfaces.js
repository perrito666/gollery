/**
 * Component contracts define the interface that UI implementations must satisfy.
 *
 * ui-default implements these by default.
 * site/ may override any of them.
 */

/**
 * @typedef {Object} ViewRenderer
 * @property {function(HTMLElement, Object): void} render - Render into a container with a view model
 * @property {function(): void} destroy - Clean up the view
 */

/**
 * Component registry — maps view names to renderer modules.
 *
 * Core calls registry.get(viewName) to obtain the renderer for the current view.
 * ui-default provides the default registry; site/ can override individual entries.
 */
export class ComponentRegistry {
  constructor() {
    this._renderers = new Map();
  }

  /** Register a renderer for a view name. */
  register(viewName, renderer) {
    this._renderers.set(viewName, renderer);
  }

  /** Get the renderer for a view name, or null. */
  get(viewName) {
    return this._renderers.get(viewName) || null;
  }

  /** Override a renderer (same as register, explicit intent). */
  override(viewName, renderer) {
    this._renderers.set(viewName, renderer);
  }
}
