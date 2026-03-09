/**
 * AssetPage view — displays a single asset with preview, navigation,
 * metadata editing (admin), discussion links, and Mastodon share.
 *
 * Consumes only the AssetViewModel from the asset controller.
 */

import { esc } from '../util/html.js';
import { renderNav } from '../util/nav.js';
import { renderDiscussionLinks } from '../components/discussion-links.js';

export function render(container, viewModel, ctx) {
  if (!viewModel) {
    container.innerHTML = '<div class="loading">Loading\u2026</div>';
    return;
  }

  const nav = renderNav(ctx);
  const state = ctx.store.get();
  const isAdmin = state.principal && state.principal.is_admin;
  const displayTitle = viewModel.title || viewModel.filename;

  let html = nav.html;

  html += '<nav class="breadcrumb">' +
    '<a href="#/">Home</a>';
  if (viewModel.albumId) {
    html += ` &rsaquo; <a href="#/albums/${esc(viewModel.albumId)}">Album</a>`;
  }
  html += ` &rsaquo; <span>${esc(displayTitle)}</span>`;
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
    `<img src="${esc(viewModel.previewURL)}" alt="${esc(displayTitle)}" class="asset-preview">` +
    '</figure>';

  // Title and description
  html += '<div class="asset-meta">';
  html += `<h2 class="asset-title">${esc(displayTitle)}</h2>`;
  if (viewModel.description) {
    html += `<p class="asset-description">${esc(viewModel.description)}</p>`;
  }
  if (isAdmin) {
    html += '<button class="btn btn-small asset-edit-meta" type="button">Edit details</button>';
  }
  html += '</div>';

  // Edit form (hidden by default, shown when "Edit details" is clicked)
  if (isAdmin) {
    html += '<div class="asset-edit-form" style="display:none">';
    html += `<label>Title<br><input type="text" class="edit-title" value="${esc(viewModel.title)}" placeholder="${esc(viewModel.filename)}"></label>`;
    html += `<label>Description<br><textarea class="edit-description" rows="3" placeholder="Add a description...">${esc(viewModel.description)}</textarea></label>`;
    html += '<div class="edit-actions">';
    html += '<button class="btn btn-small asset-save-meta" type="button">Save</button> ';
    html += '<button class="btn btn-small asset-cancel-meta" type="button">Cancel</button>';
    html += '</div></div>';
  }

  // Actions
  html += '<div class="asset-actions">';
  html += `<a href="${esc(viewModel.originalURL)}" class="btn" target="_blank" rel="noopener">Download original</a>`;

  // Mastodon share button
  html += ' <button class="btn asset-share-mastodon" type="button">Share on Mastodon</button>';
  html += '</div>';

  // Discussion links
  html += '<div class="asset-discussions"></div>';

  html += '</div>';

  container.innerHTML = html;
  nav.setup(container);

  // Render discussion links
  const discEl = container.querySelector('.asset-discussions');
  if (discEl && viewModel.discussions) {
    renderDiscussionLinks(discEl, viewModel.discussions);
  }

  // Wire up edit form toggle
  const editBtn = container.querySelector('.asset-edit-meta');
  const editForm = container.querySelector('.asset-edit-form');
  if (editBtn && editForm) {
    editBtn.addEventListener('click', () => {
      editForm.style.display = editForm.style.display === 'none' ? 'block' : 'none';
    });
    container.querySelector('.asset-cancel-meta').addEventListener('click', () => {
      editForm.style.display = 'none';
    });
    container.querySelector('.asset-save-meta').addEventListener('click', async () => {
      const title = container.querySelector('.edit-title').value;
      const description = container.querySelector('.edit-description').value;
      const api = ctx.session.api;
      try {
        await api.patchAssetMetadata(viewModel.id, { title, description });
        ctx.router.navigate(`/assets/${viewModel.id}`);
      } catch (err) {
        alert('Failed to save: ' + (err.message || err));
      }
    });
  }

  // Wire up Mastodon share
  const shareBtn = container.querySelector('.asset-share-mastodon');
  if (shareBtn) {
    shareBtn.addEventListener('click', () => {
      const shareTitle = viewModel.title || viewModel.filename;
      const pageURL = window.location.href;
      const text = `${shareTitle}\n\n${pageURL}`;
      // Prompt for Mastodon instance
      const instance = prompt('Enter your Mastodon instance (e.g. mastodon.social):');
      if (instance) {
        const cleanInstance = instance.replace(/^https?:\/\//, '').replace(/\/+$/, '');
        const shareURL = `https://${encodeURIComponent(cleanInstance)}/share?text=${encodeURIComponent(text)}`;
        window.open(shareURL, '_blank', 'noopener');
      }
    });
  }
}

export function destroy() {}
