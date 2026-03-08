import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { Store } from '../src/core/state/store.js';

describe('Store', () => {
  it('has default state', () => {
    const store = new Store();
    const state = store.get();
    assert.equal(state.currentView, null);
    assert.equal(state.loading, false);
    assert.equal(state.error, null);
  });

  it('set() merges partial state', () => {
    const store = new Store();
    store.set({ loading: true });
    assert.equal(store.get().loading, true);
    assert.equal(store.get().currentView, null); // unchanged
  });

  it('subscribe() receives updates', () => {
    const store = new Store();
    const received = [];
    store.subscribe((state) => received.push(state));

    store.set({ currentView: 'home' });
    assert.equal(received.length, 1);
    assert.equal(received[0].currentView, 'home');
  });

  it('unsubscribe works', () => {
    const store = new Store();
    const received = [];
    const unsub = store.subscribe((state) => received.push(state));

    store.set({ loading: true });
    assert.equal(received.length, 1);

    unsub();
    store.set({ loading: false });
    assert.equal(received.length, 1); // no new notification
  });

  it('multiple subscribers all receive updates', () => {
    const store = new Store();
    let countA = 0;
    let countB = 0;
    store.subscribe(() => countA++);
    store.subscribe(() => countB++);

    store.set({ currentView: 'album' });
    assert.equal(countA, 1);
    assert.equal(countB, 1);
  });
});
