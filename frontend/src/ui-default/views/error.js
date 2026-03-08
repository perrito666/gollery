/**
 * Error view — displays generic error messages.
 */

import { esc } from '../util/html.js';

export function render(container, viewModel, ctx) {
  const error = ctx.store.get().error;
  container.innerHTML =
    '<div class="error-page">' +
    '<h1>Error</h1>' +
    `<p>${esc(error || 'An unexpected error occurred.')}</p>` +
    '<a href="#/" class="btn">Back to gallery</a>' +
    '</div>';
}

export function destroy() {}