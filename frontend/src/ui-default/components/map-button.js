/**
 * Split "View on map" button with provider selection.
 *
 * Main click opens the default (or saved) provider.
 * Dropdown arrow reveals all providers; selecting one opens it
 * and saves the choice in a cookie.
 */

import { esc } from '../util/html.js';

const COOKIE_NAME = 'gollery_map_provider';
const COOKIE_MAX_AGE = 365 * 24 * 60 * 60; // 1 year

/** Map provider definitions. */
const providers = [
  {
    id: 'osm',
    label: 'OpenStreetMap',
    url: (lat, lon) =>
      `https://www.openstreetmap.org/?mlat=${lat}&mlon=${lon}#map=15/${lat}/${lon}`,
  },
  {
    id: 'google',
    label: 'Google Maps',
    url: (lat, lon) =>
      `https://www.google.com/maps?q=${lat},${lon}`,
  },
  {
    id: 'apple',
    label: 'Apple Maps',
    url: (lat, lon) =>
      `https://maps.apple.com/?ll=${lat},${lon}&q=${lat},${lon}`,
  },
  {
    id: 'geo',
    label: 'geo: URI',
    url: (lat, lon) =>
      `geo:${lat},${lon}`,
  },
];

/** Detect a sensible default provider based on platform. */
function detectDefault() {
  const ua = navigator.userAgent || '';
  const platform = navigator.platform || '';

  // iOS — Apple Maps works best
  if (/iPhone|iPad|iPod/.test(ua)) return 'apple';

  // macOS — Apple Maps handles https links
  if (/Mac/.test(platform) && !/Android/.test(ua)) return 'apple';

  // Android — geo: URI opens the native chooser
  if (/Android/.test(ua)) return 'geo';

  // Everything else — OSM is the safe web default
  return 'osm';
}

function getSavedProvider() {
  const match = document.cookie.match(
    new RegExp('(?:^|; )' + COOKIE_NAME + '=([^;]*)')
  );
  return match ? decodeURIComponent(match[1]) : null;
}

function saveProvider(id) {
  document.cookie =
    COOKIE_NAME + '=' + encodeURIComponent(id) +
    '; max-age=' + COOKIE_MAX_AGE +
    '; path=/; SameSite=Lax';
}

function getProvider() {
  const saved = getSavedProvider();
  if (saved && providers.find(p => p.id === saved)) return saved;
  return detectDefault();
}

/**
 * Render the split map button HTML.
 * @param {number} lat
 * @param {number} lon
 * @returns {string}
 */
export function renderMapButton(lat, lon) {
  const current = getProvider();
  const prov = providers.find(p => p.id === current) || providers[0];

  let html = '<span class="map-split-btn">';
  html += `<a href="${esc(prov.url(lat, lon))}" class="btn map-btn-main" target="_blank" rel="noopener">${esc(prov.label)}</a>`;
  html += '<button class="btn map-btn-dropdown" type="button" aria-label="Choose map provider">\u25BE</button>';
  html += '<div class="map-dropdown-menu" style="display:none">';
  for (const p of providers) {
    html += `<a href="${esc(p.url(lat, lon))}" class="map-dropdown-item" data-provider="${esc(p.id)}" target="_blank" rel="noopener">${esc(p.label)}</a>`;
  }
  html += '</div>';
  html += '</span>';
  return html;
}

/**
 * Wire up event listeners for the split map button.
 * Call after inserting renderMapButton HTML into the DOM.
 * @param {HTMLElement} container — element containing the split button
 */
export function setupMapButton(container) {
  const wrapper = container.querySelector('.map-split-btn');
  if (!wrapper) return;

  const dropdownBtn = wrapper.querySelector('.map-btn-dropdown');
  const menu = wrapper.querySelector('.map-dropdown-menu');

  dropdownBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    menu.style.display = menu.style.display === 'none' ? 'block' : 'none';
  });

  // Close on outside click.
  const closeMenu = (e) => {
    if (!wrapper.contains(e.target)) {
      menu.style.display = 'none';
    }
  };
  document.addEventListener('click', closeMenu);

  // Store a cleanup ref so destroy() can remove the listener.
  wrapper._closeMenu = closeMenu;

  // On item click: save preference, close menu (link opens via default <a> behavior).
  for (const item of menu.querySelectorAll('.map-dropdown-item')) {
    item.addEventListener('click', () => {
      saveProvider(item.dataset.provider);
      menu.style.display = 'none';
    });
  }
}

/**
 * Clean up document-level listeners added by setupMapButton.
 * @param {HTMLElement} container
 */
export function destroyMapButton(container) {
  const wrapper = container.querySelector('.map-split-btn');
  if (wrapper && wrapper._closeMenu) {
    document.removeEventListener('click', wrapper._closeMenu);
  }
}
