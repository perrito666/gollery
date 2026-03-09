/**
 * HomePage view — root album listing.
 *
 * Displays the root album's children and assets as a grid.
 * Admins can edit the root album title and description.
 * Consumes only the view model from the album controller.
 */

import { esc } from '../util/html.js';
import { renderNav } from '../util/nav.js';

export function render(container, viewModel, ctx) {
  if (!viewModel) {
    container.innerHTML = '<div class="loading">Loading\u2026</div>';
    return;
  }

  const nav = renderNav(ctx);
  const state = ctx.store.get();
  const isAdmin = state.principal && state.principal.is_admin;

  let html = nav.html;

  html += `<header class="page-header"><h1>${esc(viewModel.title || 'Gallery')}</h1>`;
  if (viewModel.description) {
    html += `<p class="album-description">${esc(viewModel.description)}</p>`;
  }
  if (isAdmin && viewModel.id) {
    html += '<button class="btn btn-small album-edit-meta" type="button">Edit album</button>';
  }
  html += '</header>';

  // Edit form (hidden by default)
  if (isAdmin && viewModel.id) {
    html += '<div class="album-edit-form" style="display:none">';
    html += `<label>Title<br><input type="text" class="edit-title" value="${esc(viewModel.title || '')}"></label>`;
    html += `<label>Description<br><textarea class="edit-description" rows="3">${esc(viewModel.description)}</textarea></label>`;
    html += '<div class="edit-actions">';
    html += '<button class="btn btn-small album-save-meta" type="button">Save</button> ';
    html += '<button class="btn btn-small album-cancel-meta" type="button">Cancel</button>';
    html += '</div></div>';
  }

  // Child albums
  if (viewModel.children && viewModel.children.length > 0) {
    html += '<section class="album-children"><h2>Albums</h2><ul class="album-list">';
    for (const child of viewModel.children) {
      html += `<li class="album-list-item"><a href="#/albums/${esc(child.id)}" class="album-link">${esc(child.title || child.path)}</a></li>`;
    }
    html += '</ul></section>';
  }

  // Assets grid
  if (viewModel.assets && viewModel.assets.length > 0) {
    html += '<section class="asset-grid">';
    for (const asset of viewModel.assets) {
      html += `<a href="#/assets/${esc(asset.id)}" class="asset-thumb">` +
        `<img src="${esc(asset.thumbnailURL)}" alt="${esc(asset.title || asset.filename)}" loading="lazy">` +
        '</a>';
    }
    html += '</section>';
  }

  if ((!viewModel.children || viewModel.children.length === 0) &&
      (!viewModel.assets || viewModel.assets.length === 0)) {
    html += '<p class="empty-state">This gallery is empty.</p>';
  }

  container.innerHTML = html;
  nav.setup(container);

  // Wire up edit form
  const editBtn = container.querySelector('.album-edit-meta');
  const editForm = container.querySelector('.album-edit-form');
  if (editBtn && editForm) {
    editBtn.addEventListener('click', () => {
      editForm.style.display = editForm.style.display === 'none' ? 'block' : 'none';
    });
    container.querySelector('.album-cancel-meta').addEventListener('click', () => {
      editForm.style.display = 'none';
    });
    container.querySelector('.album-save-meta').addEventListener('click', async () => {
      const title = container.querySelector('.edit-title').value;
      const description = container.querySelector('.edit-description').value;
      const api = ctx.session.api;
      try {
        await api.patchAlbumMetadata(viewModel.id, { title, description });
        ctx.router.navigate('/');
      } catch (err) {
        alert('Failed to save: ' + (err.message || err));
      }
    });
  }
}

export function destroy() {}
