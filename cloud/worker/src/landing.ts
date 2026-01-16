/**
 * Minimal landing page - login/signup + board picker
 */

export const landingPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Tickboard</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: #0a0a0a;
      color: #e5e5e5;
      min-height: 100vh;
    }
    .container { max-width: 400px; margin: 0 auto; padding: 2rem 1rem; }
    .card {
      background: #141414;
      border: 1px solid #262626;
      border-radius: 0.5rem;
      padding: 1.5rem;
    }
    h1 { text-align: center; margin-bottom: 1.5rem; font-size: 1.5rem; }
    .form-group { margin-bottom: 1rem; }
    label { display: block; margin-bottom: 0.375rem; font-size: 0.875rem; color: #a3a3a3; }
    input {
      width: 100%;
      padding: 0.625rem 0.75rem;
      background: #1a1a1a;
      border: 1px solid #262626;
      border-radius: 0.375rem;
      color: #e5e5e5;
      font-size: 0.875rem;
    }
    input:focus { outline: none; border-color: #3b82f6; }
    .btn {
      width: 100%;
      padding: 0.625rem 1rem;
      border: none;
      border-radius: 0.375rem;
      font-size: 0.875rem;
      font-weight: 500;
      cursor: pointer;
      margin-top: 0.5rem;
    }
    .btn-primary { background: #3b82f6; color: white; }
    .btn-primary:hover { background: #2563eb; }
    .btn-secondary { background: #262626; color: #e5e5e5; }
    .btn-secondary:hover { background: #333; }
    .error {
      padding: 0.75rem;
      margin-bottom: 1rem;
      background: rgba(239, 68, 68, 0.1);
      border: 1px solid #ef4444;
      border-radius: 0.375rem;
      color: #ef4444;
      font-size: 0.875rem;
    }
    .toggle { text-align: center; margin-top: 1rem; font-size: 0.875rem; color: #a3a3a3; }
    .toggle a { color: #3b82f6; cursor: pointer; }
    .hidden { display: none; }
    .header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.5rem;
    }
    .header h1 { margin: 0; }
    .logout { font-size: 0.875rem; color: #a3a3a3; cursor: pointer; }
    .logout:hover { color: #e5e5e5; }
    .board-list { list-style: none; }
    .board-item {
      display: flex;
      align-items: center;
      padding: 0.75rem;
      border: 1px solid #262626;
      border-radius: 0.375rem;
      margin-bottom: 0.5rem;
      cursor: pointer;
      transition: background 0.15s;
    }
    .board-item:hover { background: #1a1a1a; }
    .status {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      margin-right: 0.75rem;
    }
    .status.online { background: #22c55e; }
    .status.offline { background: #525252; }
    .board-name { flex: 1; }
    .board-offline { color: #737373; }
    .empty { text-align: center; color: #737373; padding: 2rem; }
    .token-section { margin-top: 1.5rem; padding-top: 1.5rem; border-top: 1px solid #262626; }
    .token-section h2 { font-size: 1rem; margin-bottom: 1rem; }
    .token-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0.5rem 0;
      border-bottom: 1px solid #262626;
      font-size: 0.875rem;
    }
    .token-name { color: #e5e5e5; }
    .token-revoke { color: #ef4444; cursor: pointer; font-size: 0.75rem; }
    .new-token { margin-top: 1rem; display: flex; gap: 0.5rem; }
    .new-token input { flex: 1; }
    .new-token .btn { width: auto; margin-top: 0; }
    .token-value {
      background: #1a1a1a;
      padding: 0.75rem;
      border-radius: 0.375rem;
      font-family: monospace;
      font-size: 0.75rem;
      word-break: break-all;
      margin-top: 0.5rem;
    }
  </style>
</head>
<body>
  <div class="container">
    <!-- Auth Forms -->
    <div id="auth" class="card">
      <div id="login-form">
        <h1>Sign In</h1>
        <div id="login-error" class="error hidden"></div>
        <div class="form-group">
          <label>Email</label>
          <input type="email" id="login-email" required>
        </div>
        <div class="form-group">
          <label>Password</label>
          <input type="password" id="login-password" required>
        </div>
        <button class="btn btn-primary" onclick="login()">Sign In</button>
        <div class="toggle">Don't have an account? <a onclick="showSignup()">Sign up</a></div>
      </div>

      <div id="signup-form" class="hidden">
        <h1>Create Account</h1>
        <div id="signup-error" class="error hidden"></div>
        <div class="form-group">
          <label>Email</label>
          <input type="email" id="signup-email" required>
        </div>
        <div class="form-group">
          <label>Password</label>
          <input type="password" id="signup-password" required minlength="8">
        </div>
        <div class="form-group">
          <label>Confirm Password</label>
          <input type="password" id="signup-confirm" required>
        </div>
        <button class="btn btn-primary" onclick="signup()">Create Account</button>
        <div class="toggle">Already have an account? <a onclick="showLogin()">Sign in</a></div>
      </div>
    </div>

    <!-- Dashboard -->
    <div id="dashboard" class="card hidden">
      <div class="header">
        <h1>Boards</h1>
        <span class="logout" onclick="logout()">Logout</span>
      </div>
      <ul id="board-list" class="board-list"></ul>
      <div id="no-boards" class="empty hidden">
        No boards connected yet.<br><br>
        Create a token below, then run:<br>
        <code style="font-size: 0.75rem">TICKBOARD_TOKEN=xxx tickboard</code>
      </div>

      <div class="token-section">
        <h2>API Tokens</h2>
        <div id="token-list"></div>
        <div id="new-token-display" class="hidden">
          <div style="color: #22c55e; font-size: 0.875rem; margin-bottom: 0.5rem;">Token created! Copy it now (won't be shown again):</div>
          <div class="token-value" id="new-token-value"></div>
        </div>
        <div class="new-token">
          <input type="text" id="token-name" placeholder="Token name (e.g. laptop)">
          <button class="btn btn-secondary" onclick="createToken()">Create</button>
        </div>
      </div>
    </div>
  </div>

  <script>
    let token = localStorage.getItem('token');

    // Check auth on load
    if (token) {
      checkAuth();
    }

    async function checkAuth() {
      try {
        const res = await fetch('/api/boards', {
          headers: { Authorization: 'Bearer ' + token }
        });
        if (res.ok) {
          showDashboard();
          loadBoards();
          loadTokens();
        } else {
          localStorage.removeItem('token');
          token = null;
        }
      } catch (e) {
        console.error(e);
      }
    }

    function showLogin() {
      document.getElementById('login-form').classList.remove('hidden');
      document.getElementById('signup-form').classList.add('hidden');
    }

    function showSignup() {
      document.getElementById('login-form').classList.add('hidden');
      document.getElementById('signup-form').classList.remove('hidden');
    }

    function showDashboard() {
      document.getElementById('auth').classList.add('hidden');
      document.getElementById('dashboard').classList.remove('hidden');
    }

    function showAuth() {
      document.getElementById('auth').classList.remove('hidden');
      document.getElementById('dashboard').classList.add('hidden');
    }

    async function login() {
      const email = document.getElementById('login-email').value;
      const password = document.getElementById('login-password').value;
      const errorEl = document.getElementById('login-error');
      errorEl.classList.add('hidden');

      try {
        const res = await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Login failed');

        token = data.token;
        localStorage.setItem('token', token);
        showDashboard();
        loadBoards();
        loadTokens();
      } catch (e) {
        errorEl.textContent = e.message;
        errorEl.classList.remove('hidden');
      }
    }

    async function signup() {
      const email = document.getElementById('signup-email').value;
      const password = document.getElementById('signup-password').value;
      const confirm = document.getElementById('signup-confirm').value;
      const errorEl = document.getElementById('signup-error');
      errorEl.classList.add('hidden');

      if (password !== confirm) {
        errorEl.textContent = 'Passwords do not match';
        errorEl.classList.remove('hidden');
        return;
      }

      try {
        const res = await fetch('/api/auth/signup', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Signup failed');

        // Auto-login
        document.getElementById('login-email').value = email;
        document.getElementById('login-password').value = password;
        await login();
      } catch (e) {
        errorEl.textContent = e.message;
        errorEl.classList.remove('hidden');
      }
    }

    function logout() {
      localStorage.removeItem('token');
      token = null;
      fetch('/api/auth/logout', { method: 'POST' });
      showAuth();
      showLogin();
    }

    async function loadBoards() {
      try {
        const res = await fetch('/api/boards', {
          headers: { Authorization: 'Bearer ' + token }
        });
        const data = await res.json();
        const list = document.getElementById('board-list');
        const empty = document.getElementById('no-boards');

        if (!data.boards || data.boards.length === 0) {
          list.innerHTML = '';
          empty.classList.remove('hidden');
          return;
        }

        empty.classList.add('hidden');
        list.innerHTML = data.boards.map(b => \`
          <li class="board-item \${b.online ? '' : 'board-offline'}" onclick="openBoard('\${b.name}', \${b.online})">
            <span class="status \${b.online ? 'online' : 'offline'}"></span>
            <span class="board-name">\${b.name}</span>
            <span style="font-size: 0.75rem; color: #737373">\${b.online ? 'online' : 'offline'}</span>
          </li>
        \`).join('');
      } catch (e) {
        console.error(e);
      }
    }

    function openBoard(name, online) {
      if (!online) {
        alert('Board is offline. Start tickboard locally to connect.');
        return;
      }
      window.location.href = '/b/' + encodeURIComponent(name) + '/';
    }

    async function loadTokens() {
      try {
        const res = await fetch('/api/tokens', {
          headers: { Authorization: 'Bearer ' + token }
        });
        const data = await res.json();
        const list = document.getElementById('token-list');

        if (!data.tokens || data.tokens.length === 0) {
          list.innerHTML = '<div style="color: #737373; font-size: 0.875rem">No tokens yet</div>';
          return;
        }

        list.innerHTML = data.tokens
          .filter(t => !t.revoked && t.name !== 'session')
          .map(t => \`
            <div class="token-item">
              <span class="token-name">\${t.name}</span>
              <span class="token-revoke" onclick="revokeToken('\${t.id}')">revoke</span>
            </div>
          \`).join('') || '<div style="color: #737373; font-size: 0.875rem">No tokens yet</div>';
      } catch (e) {
        console.error(e);
      }
    }

    async function createToken() {
      const nameInput = document.getElementById('token-name');
      const name = nameInput.value.trim();
      if (!name) return;

      try {
        const res = await fetch('/api/tokens', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: 'Bearer ' + token
          },
          body: JSON.stringify({ name })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Failed to create token');

        nameInput.value = '';
        document.getElementById('new-token-value').textContent = data.token;
        document.getElementById('new-token-display').classList.remove('hidden');
        loadTokens();
      } catch (e) {
        alert(e.message);
      }
    }

    async function revokeToken(id) {
      if (!confirm('Revoke this token?')) return;

      try {
        await fetch('/api/tokens/' + id, {
          method: 'DELETE',
          headers: { Authorization: 'Bearer ' + token }
        });
        loadTokens();
      } catch (e) {
        console.error(e);
      }
    }

    // Handle enter key on forms
    document.getElementById('login-password').addEventListener('keypress', e => {
      if (e.key === 'Enter') login();
    });
    document.getElementById('signup-confirm').addEventListener('keypress', e => {
      if (e.key === 'Enter') signup();
    });
    document.getElementById('token-name').addEventListener('keypress', e => {
      if (e.key === 'Enter') createToken();
    });
  </script>
</body>
</html>`;
