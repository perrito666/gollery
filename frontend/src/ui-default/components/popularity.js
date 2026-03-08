/**
 * Popularity component — optional.
 *
 * Renders popularity data for an asset or album.
 * Only renders when the popularity feature flag is enabled.
 */

import { esc } from '../util/html.js';

/**
 * Render a popularity badge/summary into a container.
 *
 * @param {HTMLElement} el - Target element
 * @param {{totalViews: number, views7d: number}} data - Popularity summary
 */
export function renderPopularityBadge(el, data) {
  if (!data || !data.totalViews) {
    el.innerHTML = '';
    return;
  }
  el.innerHTML =
    `<span class="popularity-badge" title="${esc(String(data.totalViews))} total views">` +
    `${esc(formatCount(data.views7d))} views this week` +
    '</span>';
}

/**
 * Render a list of popular assets within an album.
 *
 * @param {HTMLElement} el - Target element
 * @param {Array<{id: string, filename: string, totalViews: number}>} assets
 */
export function renderPopularAssets(el, assets) {
  if (!assets || assets.length === 0) {
    el.innerHTML = '';
    return;
  }
  let html = '<div class="popular-assets"><h3>Popular</h3><ul>';
  for (const a of assets) {
    html += `<li><a href="#/assets/${esc(a.id)}">${esc(a.filename)}</a>` +
      ` <span class="view-count">${formatCount(a.totalViews)}</span></li>`;
  }
  html += '</ul></div>';
  el.innerHTML = html;
}

/**
 * Render an admin-only popularity panel.
 *
 * @param {HTMLElement} el
 * @param {{totalViews: number, views7d: number, views30d: number, originalHits: number, discussionClicks: number}} data
 */
export function renderAdminPopularityPanel(el, data) {
  if (!data) {
    el.innerHTML = '';
    return;
  }
  el.innerHTML =
    '<div class="admin-popularity-panel">' +
    '<h3>Popularity (admin)</h3>' +
    '<table>' +
    `<tr><td>Total views</td><td>${esc(String(data.totalViews))}</td></tr>` +
    `<tr><td>Last 7 days</td><td>${esc(String(data.views7d))}</td></tr>` +
    `<tr><td>Last 30 days</td><td>${esc(String(data.views30d))}</td></tr>` +
    `<tr><td>Original downloads</td><td>${esc(String(data.originalHits))}</td></tr>` +
    `<tr><td>Discussion clicks</td><td>${esc(String(data.discussionClicks))}</td></tr>` +
    '</table></div>';
}

function formatCount(n) {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
  return String(n);
}