/**
 * Gollery frontend entry point.
 *
 * Wires together core, ui-contract, ui-default/site layers.
 * The _resolved/registry.js is generated at build time by resolve-theme.js,
 * choosing site overrides when they exist.
 */

import { init as initCore } from './core/index.js';
import { ComponentRegistry } from './ui-contract/interfaces.js';
import { registerAll, siteConfig } from './_resolved/registry.js';

document.addEventListener('DOMContentLoaded', () => {
  const registry = new ComponentRegistry();
  registerAll(registry);

  const container = document.getElementById('app');
  initCore(registry, container, siteConfig);
});
