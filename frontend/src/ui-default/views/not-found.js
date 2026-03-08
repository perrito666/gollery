/**
 * NotFoundPage view.
 */

export function render(container) {
  container.innerHTML =
    '<div class="error-page">' +
    '<h1>404</h1>' +
    '<p>The page you are looking for does not exist.</p>' +
    '<a href="#/" class="btn">Back to gallery</a>' +
    '</div>';
}

export function destroy() {}
