const ThemeManager = {
    init() {
        const saved = localStorage.getItem('theme');
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        this.setTheme(saved || (prefersDark ? 'dark' : 'light'));
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
            if (!localStorage.getItem('theme')) this.setTheme(e.matches ? 'dark' : 'light');
        });
    },
    setTheme(t) { document.documentElement.setAttribute('data-theme', t); localStorage.setItem('theme', t); },
    toggle() { this.setTheme(document.documentElement.getAttribute('data-theme') === 'dark' ? 'light' : 'dark'); }
};

const Toast = {
    container: null,
    init() { this.container = document.createElement('div'); this.container.className = 'toast-container'; document.body.appendChild(this.container); },
    show(msg, type = 'success', dur = 2500) {
        const t = document.createElement('div');
        t.className = `toast ${type}`;
        t.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">${type === 'success' ? '<path d="M20 6L9 17l-5-5"/>' : '<circle cx="12" cy="12" r="10"/><path d="M15 9l-6 6M9 9l6 6"/>'}</svg><span>${msg}</span>`;
        this.container.appendChild(t);
        setTimeout(() => { t.style.animation = 'slideIn 0.3s reverse'; setTimeout(() => t.remove(), 300); }, dur);
    }
};

async function copyToClipboard(text, btn) {
    try {
        await navigator.clipboard.writeText(text);
        Toast.show('Copied!');
        if (btn) { const o = btn.innerHTML; btn.innerHTML = '✓ Copied'; setTimeout(() => btn.innerHTML = o, 1500); }
    } catch { Toast.show('Copy failed', 'error'); }
}

function toggleWrap() {
    const body = document.querySelector('.code-body');
    const btn = document.getElementById('wrap-toggle');
    if (!body) return;
    body.classList.toggle('wrap-text');
    const isWrapped = body.classList.contains('wrap-text');
    localStorage.setItem('wrapText', isWrapped);
    if (btn) btn.classList.toggle('btn-primary', isWrapped);
}

const API = {
    async request(url, opts = {}) {
        const r = await fetch(url, { headers: { 'Content-Type': 'application/json' }, credentials: 'same-origin', ...opts });
        const d = await r.json();
        if (!r.ok) throw new Error(d.error || 'Error');
        return d;
    },
    createPaste: (d) => API.request('/api/paste', { method: 'POST', body: JSON.stringify(d) }),
    updatePaste: (id, d) => API.request(`/api/paste/${id}`, { method: 'PUT', body: JSON.stringify(d) }),
    deletePaste: (id) => API.request(`/api/paste/${id}`, { method: 'DELETE' }),
    forkPaste: (id) => API.request(`/api/paste/${id}/fork`, { method: 'POST' }),
    login: (u, p) => API.request('/api/auth/login', { method: 'POST', body: JSON.stringify({ username: u, password: p }) }),
    register: (u, p) => API.request('/api/auth/register', { method: 'POST', body: JSON.stringify({ username: u, password: p }) }),
    logout: () => API.request('/api/auth/logout', { method: 'POST' })
};

function setupPasteForm() {
    const f = document.getElementById('paste-form');
    if (!f) return;
    f.addEventListener('submit', async e => {
        e.preventDefault();
        const btn = f.querySelector('button[type="submit"]');
        try {
            btn.disabled = true; btn.textContent = 'Creating...';
            const paste = await API.createPaste({
                title: f.title.value || 'Untitled', content: f.content.value, language: f.language.value,
                is_public: f.is_public.checked, expires_in: f.expires_in?.value || 'never', burn_after_read: f.burn_after_read?.checked || false
            });
            window.location.href = `/${paste.id}`;
        } catch (err) { Toast.show(err.message, 'error'); btn.disabled = false; btn.textContent = 'Create Paste'; }
    });
}

function setupEditForm() {
    const f = document.getElementById('edit-form');
    if (!f) return;
    f.addEventListener('submit', async e => {
        e.preventDefault();
        const btn = f.querySelector('button[type="submit"]');
        try {
            btn.disabled = true; btn.textContent = 'Saving...';
            await API.updatePaste(f.dataset.pasteId, { title: f.title.value, content: f.content.value, language: f.language.value, is_public: f.is_public.checked });
            Toast.show('Saved!'); setTimeout(() => window.location.href = `/${f.dataset.pasteId}`, 800);
        } catch (err) { Toast.show(err.message, 'error'); btn.disabled = false; btn.textContent = 'Save'; }
    });
}

function setupAuthForms() {
    const login = document.getElementById('login-form');
    const reg = document.getElementById('register-form');
    if (login) login.addEventListener('submit', async e => {
        e.preventDefault();
        const btn = login.querySelector('button[type="submit"]');
        try { btn.disabled = true; await API.login(login.username.value, login.password.value); window.location.href = '/dashboard'; }
        catch (err) { Toast.show(err.message, 'error'); btn.disabled = false; }
    });
    if (reg) reg.addEventListener('submit', async e => {
        e.preventDefault();
        const btn = reg.querySelector('button[type="submit"]');
        try { btn.disabled = true; await API.register(reg.username.value, reg.password.value); window.location.href = '/dashboard'; }
        catch (err) { Toast.show(err.message, 'error'); btn.disabled = false; }
    });
}

function setupDeleteButton() {
    const btn = document.getElementById('delete-paste');
    if (!btn) return;
    btn.addEventListener('click', async () => {
        if (!confirm('Delete this paste?')) return;
        try { await API.deletePaste(btn.dataset.pasteId); Toast.show('Deleted!'); setTimeout(() => window.location.href = '/dashboard', 800); }
        catch (err) { Toast.show(err.message, 'error'); }
    });
}

function setupForkButton() {
    const btn = document.getElementById('fork-paste');
    if (!btn) return;
    btn.addEventListener('click', async () => {
        try { const f = await API.forkPaste(btn.dataset.pasteId); Toast.show('Forked!'); setTimeout(() => window.location.href = `/${f.id}`, 800); }
        catch (err) { Toast.show(err.message, 'error'); }
    });
}

function setupLogout() {
    const btn = document.getElementById('logout-btn');
    if (!btn) return;
    btn.addEventListener('click', async e => { e.preventDefault(); try { await API.logout(); window.location.href = '/'; } catch { } });
}

function setupKeyboardShortcuts() {
    document.addEventListener('keydown', e => {
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            const f = document.querySelector('#paste-form, #edit-form');
            if (f) { e.preventDefault(); f.dispatchEvent(new Event('submit')); }
        }
        if ((e.ctrlKey || e.metaKey) && e.key === 's') {
            const f = document.getElementById('edit-form');
            if (f) { e.preventDefault(); f.dispatchEvent(new Event('submit')); }
        }
    });
}

function restoreWrapState() {
    if (localStorage.getItem('wrapText') === 'true') {
        const body = document.querySelector('.code-body');
        const btn = document.getElementById('wrap-toggle');
        if (body) body.classList.add('wrap-text');
        if (btn) btn.classList.add('btn-primary');
    }
}

document.addEventListener('DOMContentLoaded', () => {
    ThemeManager.init();
    Toast.init();
    setupPasteForm();
    setupEditForm();
    setupAuthForms();
    setupDeleteButton();
    setupForkButton();
    setupLogout();
    setupKeyboardShortcuts();
    restoreWrapState();
});

window.toggleTheme = () => ThemeManager.toggle();
window.copyToClipboard = copyToClipboard;
window.toggleWrap = toggleWrap;
