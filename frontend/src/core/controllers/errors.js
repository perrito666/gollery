/**
 * Shared controller error handling.
 *
 * Maps API errors to view transitions via the store.
 */

/**
 * Handle an API error by setting the appropriate view in the store.
 * @param {import('../../core/state/store.js').Store} store
 * @param {Error & {status?: number}} err
 */
export function handleApiError(store, err) {
  if (err.status === 401) {
    store.set({ currentView: 'login', viewModel: null, loading: false, error: null });
  } else if (err.status === 403) {
    store.set({ currentView: 'forbidden', viewModel: null, loading: false, error: null });
  } else if (err.status === 404) {
    store.set({ currentView: 'not-found', viewModel: null, loading: false, error: null });
  } else {
    store.set({ currentView: null, viewModel: null, loading: false, error: err.message });
  }
}
