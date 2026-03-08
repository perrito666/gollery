/**
 * Discussion links component.
 *
 * Renders a list of external discussion links (Mastodon, Bluesky, etc.)
 * for an asset or album. Receives discussion bindings from the view model.
 */

import { esc } from '../util/html.js';

/**
 * Render discussion links into a container element.
 * @param {HTMLElement} el - Target element
 * @param {Array<{provider: string, url: string}>} discussions
 */
export function renderDiscussionLinks(el, discussions) {
  if (!discussions || discussions.length === 0) {
    el.innerHTML = '';
    return;
  }

  let html = '<div class="discussion-links"><h3>Discussions</h3><ul>';
  for (const d of discussions) {
    html += `<li><a href="${esc(d.url)}" target="_blank" rel="noopener" class="discussion-link">` +
      `${esc(d.provider)}</a></li>`;
  }
  html += '</ul></div>';
  el.innerHTML = html;
}