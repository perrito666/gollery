/**
 * Main layout — wraps page content with header/footer.
 */

export function renderLayout(container) {
  container.innerHTML = '<div id="app-root"></div>';
  return container.querySelector('#app-root');
}
