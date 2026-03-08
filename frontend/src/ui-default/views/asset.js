/**
 * AssetPage view — displays a single asset with preview and navigation.
 *
 * Consumes only the AssetViewModel from the asset controller.
 */

import { esc } from '../util/html.js';

export function render(container, viewModel, ctx) {
  if (!viewModel) {
    container.innerHTML = '<div class="loading">Loading…</div>';
    return;
  }

  let html = '<nav class="breadcrumb">' +
    '<a href="#/">Home</a>';
  if (viewModel.albumId) {
    html += ` &rsaquo; <a href="#/albums/${esc(viewModel.albumId)}">Album</a>`;
  }
  html += '</nav>';

  html += '<div class="asset-viewer">';

  // Navigation
  html += '<div class="asset-nav">';
  if (viewModel.prevAssetId) {
    html += `<a href="#/assets/${esc(viewModel.prevAssetId)}" class="nav-prev">&larr; Previous</a>`;
  } else {
    html += '<span class="nav-prev disabled">&larr; Previous</span>';
  }
  if (viewModel.nextAssetId) {
    html += `<a href="#/assets/${esc(viewModel.nextAssetId)}" class="nav-next">Next &rarr;</a>`;
  } else {
    html += '<span class="nav-next disabled">Next &rarr;</span>';
  }
  html += '</div>';

  // Preview image
  html += `<figure class="asset-figure">` +
    `<img src="${esc(viewModel.previewURL)}" alt="${esc(viewModel.filename)}" class="asset-preview">` +
    `<figcaption>${esc(viewModel.filename)}</figcaption>` +
    '</figure>';

  // Original download link
  html += `<div class="asset-actions">` +
    `<a href="${esc(viewModel.originalURL)}" class="btn" target="_blank" rel="noopener">Download original</a>` +
    '</div>';

  // Discussion links placeholder
  html += '<div class="asset-discussions" data-placeholder="discussions"></div>';

  html += '</div>';

  container.innerHTML = html;
}

export function destroy() {}