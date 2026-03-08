/**
 * Shared HTML utilities for ui-default views and components.
 */

/**
 * Escape a string for safe inclusion in HTML.
 * @param {string} str
 * @returns {string}
 */
export function esc(str) {
  const el = document.createElement('span');
  el.textContent = str;
  return el.innerHTML;
}
