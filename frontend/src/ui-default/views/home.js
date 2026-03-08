/**
 * HomePage view — root album listing.
 *
 * Displays the root album's children and assets as a grid.
 * Consumes only the view model from the album controller.
 */

import { esc } from '../util/html.js';

export function render(container, viewModel, ctx) {
  if (!viewModel) {
    container.innerHTML = '<div class="loading">Loading…</div>';
    return;
  }

  const { router } = ctx;

  let html = `<header class="page-header"><h1>${esc(viewModel.title || 'Gallery')}</h1>`;
  if (viewModel.description) {
    html += `<p class="album-description">${esc(viewModel.description)}</p>`;
  }
  html += '</header>';

  // Child albums
  if (viewModel.children && viewModel.children.length > 0) {
    html += '<section class="album-children"><h2>Albums</h2><ul class="album-list">';
    for (const child of viewModel.children) {
      html += `<li class="album-list-item"><a href="#/albums/${esc(child.path)}" class="album-link">${esc(child.path)}</a></li>`;
    }
    html += '</ul></section>';
  }

  // Assets grid
  if (viewModel.assets && viewModel.assets.length > 0) {
    html += '<section class="asset-grid">';
    for (const asset of viewModel.assets) {
      html += `<a href="#/assets/${esc(asset.id)}" class="asset-thumb">` +
        `<img src="${esc(asset.thumbnailURL)}" alt="${esc(asset.filename)}" loading="lazy">` +
        '</a>';
    }
    html += '</section>';
  }

  if ((!viewModel.children || viewModel.children.length === 0) &&
      (!viewModel.assets || viewModel.assets.length === 0)) {
    html += '<p class="empty-state">This gallery is empty.</p>';
  }

  container.innerHTML = html;
}

export function destroy() {}