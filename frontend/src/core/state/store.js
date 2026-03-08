/**
 * Simple reactive state store.
 *
 * Controllers update state, views subscribe to changes.
 */

export class Store {
  constructor() {
    this.state = {
      currentView: null,
      viewModel: null,
      principal: null,
      loading: false,
      error: null,
    };
    this._listeners = [];
  }

  /** Get current state. */
  get() {
    return this.state;
  }

  /** Update state and notify listeners. */
  set(partial) {
    this.state = { ...this.state, ...partial };
    this._notify();
  }

  /** Subscribe to state changes. Returns an unsubscribe function. */
  subscribe(fn) {
    this._listeners.push(fn);
    return () => {
      this._listeners = this._listeners.filter(l => l !== fn);
    };
  }

  _notify() {
    for (const fn of this._listeners) {
      fn(this.state);
    }
  }
}
