/**
 * LoginPage view — login form.
 *
 * Submits credentials via the session object from ctx.
 */

export function render(container, viewModel, ctx) {
  const { session, router } = ctx;

  container.innerHTML =
    '<div class="login-page">' +
    '<h1>Log in</h1>' +
    '<form class="login-form">' +
    '<label>Username<input type="text" name="username" autocomplete="username" required></label>' +
    '<label>Password<input type="password" name="password" autocomplete="current-password" required></label>' +
    '<div class="login-error" hidden></div>' +
    '<button type="submit" class="btn">Log in</button>' +
    '</form>' +
    '</div>';

  const form = container.querySelector('.login-form');
  const errorEl = container.querySelector('.login-error');

  const btn = form.querySelector('button[type="submit"]');
  let submitting = false;

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    if (submitting) return;

    errorEl.hidden = true;
    submitting = true;
    btn.disabled = true;
    btn.textContent = 'Logging in\u2026';

    const username = form.elements.username.value.trim();
    const password = form.elements.password.value;

    try {
      await session.login(username, password);
      router.navigate('/');
    } catch (err) {
      errorEl.textContent = err.message || 'Login failed';
      errorEl.hidden = false;
    } finally {
      submitting = false;
      btn.disabled = false;
      btn.textContent = 'Log in';
    }
  });
}

export function destroy() {}
