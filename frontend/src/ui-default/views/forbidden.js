/**
 * ForbiddenPage view.
 */

export function render(container) {
  container.innerHTML =
    '<div class="error-page">' +
    '<h1>403</h1>' +
    '<p>You do not have permission to view this content.</p>' +
    '<a href="#/" class="btn">Back to gallery</a>' +
    '</div>';
}

export function destroy() {}
