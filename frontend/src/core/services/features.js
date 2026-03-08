/**
 * Feature flags service.
 *
 * Controls optional features like popularity UI.
 * Flags can be set from site config or at runtime.
 */

const DEFAULT_FLAGS = {
  popularity: false,
};

export class FeatureFlags {
  constructor(siteConfig = {}) {
    this._flags = { ...DEFAULT_FLAGS };
    // Site config can enable features
    if (siteConfig.features) {
      for (const [key, val] of Object.entries(siteConfig.features)) {
        if (key in this._flags) {
          this._flags[key] = Boolean(val);
        }
      }
    }
  }

  /** Check if a feature is enabled. */
  isEnabled(flag) {
    return this._flags[flag] === true;
  }

  /** Enable a feature at runtime. */
  enable(flag) {
    this._flags[flag] = true;
  }

  /** Disable a feature at runtime. */
  disable(flag) {
    this._flags[flag] = false;
  }
}
