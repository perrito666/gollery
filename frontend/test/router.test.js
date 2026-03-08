import { describe, it, beforeEach } from 'node:test';
import assert from 'node:assert/strict';

// Minimal window/location stub for Router.
globalThis.window = globalThis.window || {};
window.location = window.location || {};
window.location.hash = '';
window.addEventListener = window.addEventListener || (() => {});
window.removeEventListener = window.removeEventListener || (() => {});

import { Router } from '../src/core/router/router.js';

describe('Router', () => {
  let router;

  beforeEach(() => {
    router = new Router();
  });

  it('matches exact path', () => {
    let matched = false;
    router.on('/', () => { matched = true; });

    window.location.hash = '#/';
    router._onHashChange();
    assert.equal(matched, true);
  });

  it('extracts named params', () => {
    let params = null;
    router.on('/albums/:id', (p) => { params = p; });

    window.location.hash = '#/albums/alb_abc123';
    router._onHashChange();
    assert.deepEqual(params, { id: 'alb_abc123' });
  });

  it('extracts multiple params', () => {
    let params = null;
    router.on('/albums/:albumId/assets/:assetId', (p) => { params = p; });

    window.location.hash = '#/albums/alb_1/assets/ast_2';
    router._onHashChange();
    assert.deepEqual(params, { albumId: 'alb_1', assetId: 'ast_2' });
  });

  it('calls onNotFound for unmatched routes', () => {
    let notFoundCalled = false;
    router.on('/', () => {});
    router.onNotFound(() => { notFoundCalled = true; });

    window.location.hash = '#/unknown/path';
    router._onHashChange();
    assert.equal(notFoundCalled, true);
  });

  it('first matching route wins', () => {
    const order = [];
    router.on('/test', () => { order.push('first'); });
    router.on('/test', () => { order.push('second'); });

    window.location.hash = '#/test';
    router._onHashChange();
    assert.deepEqual(order, ['first']);
  });

  it('decodes URI-encoded params', () => {
    let params = null;
    router.on('/search/:query', (p) => { params = p; });

    window.location.hash = '#/search/hello%20world';
    router._onHashChange();
    assert.equal(params.query, 'hello world');
  });
});
