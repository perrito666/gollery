/**
 * AlbumPage view — displays a single album with child albums and asset grid.
 * Admins can edit album title and description.
 *
 * Consumes only the AlbumViewModel from the album controller.
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

  html += '<nav class="breadcrumb"><a href="#/">Home</a></nav>';

  html += `<header class="page-header"><h1>${esc(viewModel.title)}</h1>`;
  if (viewModel.description) {
    html += `<p class="album-description">${esc(viewModel.description)}</p>`;
  }
  if (isAdmin) {
    html += '<button class="btn btn-small album-edit-meta" type="button">Edit album</button>';
  }
  html += '</header>';

  // Edit form (hidden by default)
  if (isAdmin) {
    html += '<div class="album-edit-form" style="display:none">';
    html += `<label>Title<br><input type="text" class="edit-title" value="${esc(viewModel.title)}"></label>`;
    html += `<label>Description<br><textarea class="edit-description" rows="3">${esc(viewModel.description)}</textarea></label>`;
    html += '<div class="edit-actions">';
    html += '<button class="btn btn-small album-save-meta" type="button">Save</button> ';
    html += '<button class="btn btn-small album-cancel-meta" type="button">Cancel</button>';
    html += '</div></div>';
  }

  // Child albums
  if (viewModel.children && viewModel.children.length > 0) {
    html += '<section class="album-children"><h2>Sub-albums</h2><ul class="album-list">';
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
    html += '<p class="empty-state">This album is empty.</p>';
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
      const titleInput = container.querySelector('.edit-title');
      const descInput = container.querySelector('.edit-description');
      const saveBtn = container.querySelector('.album-save-meta');
      const title = titleInput.value;
      const description = descInput.value;
      const api = ctx.session.api;
      saveBtn.textContent = 'Saving\u2026';
      saveBtn.disabled = true;
      try {
        await api.patchAlbumMetadata(viewModel.id, { title, description });
        // Update displayed text in-place.
        const header = container.querySelector('.page-header');
        const h1 = header.querySelector('h1');
        if (h1) h1.textContent = title || viewModel.title;
        const descEl = header.querySelector('.album-description');
        if (description) {
          if (descEl) {
            descEl.textContent = description;
          } else {
            const p = document.createElement('p');
            p.className = 'album-description';
            p.textContent = description;
            h1.after(p);
          }
        } else if (descEl) {
          descEl.remove();
        }
        editForm.style.display = 'none';
      } catch (err) {
        alert('Failed to save: ' + (err.message || err));
      } finally {
        saveBtn.textContent = 'Save';
        saveBtn.disabled = false;
      }
    });
  }
}

export function destroy() {}
