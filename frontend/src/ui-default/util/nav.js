/**
 * Shared navigation bar renderer.
 *
 * Renders a top nav bar with login/logout controls.
 * Returns HTML string and a post-render setup function for event listeners.
 */

import { esc } from './html.js';

/**
 * Render nav bar HTML.
 * @param {Object} ctx - The app context with store, session, router.
 * @returns {{ html: string, setup: (container: HTMLElement) => void }}
 */
export function renderNav(ctx) {
  const { store, session } = ctx;
  const state = store.get();
  const principal = state.principal;

  let html = '<nav class="site-nav">';
  html += '<a href="#/" class="nav-home">Gallery</a>';
  html += '<span class="nav-spacer"></span>';

  if (principal) {
    html += `<span class="nav-user">${esc(principal.username)}</span>`;
    html += '<button class="btn btn-small nav-logout" type="button">Log out</button>';
  } else {
    html += '<a href="#/login" class="btn btn-small nav-login">Log in</a>';
  }

  html += '</nav>';

  function setup(container) {
    const logoutBtn = container.querySelector('.nav-logout');
    if (logoutBtn) {
      logoutBtn.addEventListener('click', async () => {
        await session.logout();
        ctx.router.navigate('/');
      });
    }
  }

  return { html, setup };
}
