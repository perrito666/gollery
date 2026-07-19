/**
 * AlbumPage view — displays a single album with child albums and asset grid.
 * Admins can edit album title and description.
 *
 * Consumes only the AlbumViewModel from the album controller.
 */

import { esc } from '../util/html.js';
import { renderNav } from '../util/nav.js';

// Page size for lazy-loaded asset chunks. Matches the backend default.
const PAGE_SIZE = 100;

// Holds cleanup for the current render so destroy() can tear down the
// infinite-scroll observer without leaking listeners across views.
let cleanup = null;

export function render(container, viewModel, ctx) {
  // Tear down any previous render's observer before we replace innerHTML.
  destroy();

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
      html += assetThumbHTML(asset);
    }
    html += '</section>';
    const total = viewModel.totalAssets || viewModel.assets.length;
    if (viewModel.assets.length < total) {
      html += '<div class="asset-grid-sentinel" aria-hidden="true">Loading more…</div>';
    }
  }

  if ((!viewModel.children || viewModel.children.length === 0) &&
      (!viewModel.assets || viewModel.assets.length === 0)) {
    html += '<p class="empty-state">This album is empty.</p>';
  }

  container.innerHTML = html;
  nav.setup(container);

  setupInfiniteScroll(container, viewModel, ctx);

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

export function destroy() {
  if (cleanup) {
    cleanup();
    cleanup = null;
  }
}

function assetThumbHTML(asset) {
  return `<a href="#/assets/${esc(asset.id)}" class="asset-thumb">` +
    `<img src="${esc(asset.thumbnailURL)}" alt="${esc(asset.title || asset.filename)}" loading="lazy">` +
    '</a>';
}

// setupInfiniteScroll wires an IntersectionObserver on the sentinel below
// the asset grid. Each time it becomes visible, the next page of assets
// is fetched and appended to the grid in place — no full re-render, so
// scroll position and DOM state are preserved.
function setupInfiniteScroll(container, viewModel, ctx) {
  const grid = container.querySelector('.asset-grid');
  const sentinel = container.querySelector('.asset-grid-sentinel');
  if (!grid || !sentinel) return;

  const total = viewModel.totalAssets || viewModel.assets.length;
  let loaded = viewModel.assets.length;
  let loading = false;

  if (loaded >= total) {
    sentinel.remove();
    return;
  }

  // If no controller is available (e.g. a stripped ctx from a site override),
  // there is no way to fetch more — drop the sentinel silently.
  if (!ctx || !ctx.albumController || typeof ctx.albumController.loadAssetsPage !== 'function') {
    sentinel.remove();
    return;
  }

  // Fallback for environments without IntersectionObserver (older browsers,
  // jsdom): fetch pages eagerly so no assets are hidden.
  if (typeof IntersectionObserver === 'undefined') {
    void fetchAll();
    return;
  }

  const observer = new IntersectionObserver(async (entries) => {
    if (!entries.some(e => e.isIntersecting) || loading) return;
    await fetchNext();
  }, { rootMargin: '400px 0px' });

  observer.observe(sentinel);

  cleanup = () => {
    observer.disconnect();
  };

  async function fetchNext() {
    if (loading || loaded >= total) return;
    loading = true;
    try {
      const { assets, total: newTotal } = await ctx.albumController.loadAssetsPage(
        viewModel.id,
        { offset: loaded, limit: PAGE_SIZE }
      );
      appendAssets(assets);
      loaded += assets.length;
      // Server may have returned a larger total than we knew; respect it.
      const cap = Math.max(newTotal, total);
      if (loaded >= cap || assets.length === 0) {
        observer.disconnect();
        sentinel.remove();
      }
    } catch (err) {
      // Leave the sentinel visible so the user (or a retry) can try again.
      sentinel.textContent = 'Failed to load more. Scroll to retry.';
      console.warn('gollery: failed to load more assets', err);
    } finally {
      loading = false;
    }
  }

  async function fetchAll() {
    while (loaded < total) {
      // eslint-disable-next-line no-await-in-loop
      await fetchNext();
    }
  }

  function appendAssets(assets) {
    if (!assets || assets.length === 0) return;
    const frag = document.createDocumentFragment();
    const tmp = document.createElement('div');
    tmp.innerHTML = assets.map(assetThumbHTML).join('');
    while (tmp.firstChild) frag.appendChild(tmp.firstChild);
    grid.appendChild(frag);
  }
}
