// Patbin - Main JavaScript

// ===== Theme Management =====
const ThemeManager = {
    init() {
        const savedTheme = localStorage.getItem('theme');
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        const theme = savedTheme || (prefersDark ? 'dark' : 'light');
        this.setTheme(theme);

        // Listen for system theme changes
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            if (!localStorage.getItem('theme')) {
                this.setTheme(e.matches ? 'dark' : 'light');
            }
        });
    },

    setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);
    },

    toggle() {
        const current = document.documentElement.getAttribute('data-theme');
        const newTheme = current === 'dark' ? 'light' : 'dark';
        this.setTheme(newTheme);
    }
};

// ===== Toast Notifications =====
const Toast = {
    container: null,

    init() {
        this.container = document.createElement('div');
        this.container.className = 'toast-container';
        document.body.appendChild(this.container);
    },

    show(message, type = 'success', duration = 3000) {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.innerHTML = `
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        ${type === 'success'
                ? '<path d="M20 6L9 17l-5-5"/>'
                : type === 'error'
                    ? '<circle cx="12" cy="12" r="10"/><path d="M15 9l-6 6M9 9l6 6"/>'
                    : '<path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><path d="M12 9v4M12 17h.01"/>'
            }
      </svg>
      <span>${message}</span>
    `;
        this.container.appendChild(toast);

        setTimeout(() => {
            toast.style.animation = 'slideIn 0.3s ease reverse';
            setTimeout(() => toast.remove(), 300);
        }, duration);
    }
};

// ===== Clipboard =====
async function copyToClipboard(text, button) {
    try {
        await navigator.clipboard.writeText(text);
        Toast.show('Copied to clipboard!', 'success');

        // Update button temporarily
        if (button) {
            const originalHTML = button.innerHTML;
            button.innerHTML = `
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M20 6L9 17l-5-5"/>
        </svg>
        Copied!
      `;
            setTimeout(() => {
                button.innerHTML = originalHTML;
            }, 2000);
        }
    } catch (err) {
        Toast.show('Failed to copy', 'error');
    }
}

// ===== API Helpers =====
const API = {
    async request(url, options = {}) {
        const defaults = {
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'same-origin',
        };

        const response = await fetch(url, { ...defaults, ...options });
        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'An error occurred');
        }

        return data;
    },

    async createPaste(data) {
        return this.request('/api/paste', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    },

    async updatePaste(id, data) {
        return this.request(`/api/paste/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    },

    async deletePaste(id) {
        return this.request(`/api/paste/${id}`, {
            method: 'DELETE',
        });
    },

    async forkPaste(id) {
        return this.request(`/api/paste/${id}/fork`, {
            method: 'POST',
        });
    },

    async login(username, password) {
        return this.request('/api/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password }),
        });
    },

    async register(username, password) {
        return this.request('/api/auth/register', {
            method: 'POST',
            body: JSON.stringify({ username, password }),
        });
    },

    async logout() {
        return this.request('/api/auth/logout', {
            method: 'POST',
        });
    }
};

// ===== Form Handlers =====
function setupPasteForm() {
    const form = document.getElementById('paste-form');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const submitBtn = form.querySelector('button[type="submit"]');
        const originalText = submitBtn.textContent;

        try {
            submitBtn.disabled = true;
            submitBtn.textContent = 'Creating...';

            const data = {
                title: form.title.value || 'Untitled',
                content: form.content.value,
                language: form.language.value,
                is_public: form.is_public.checked,
                expires_in: form.expires_in?.value || 'never',
                burn_after_read: form.burn_after_read?.checked || false,
            };

            const paste = await API.createPaste(data);
            window.location.href = `/${paste.id}`;
        } catch (err) {
            Toast.show(err.message, 'error');
            submitBtn.disabled = false;
            submitBtn.textContent = originalText;
        }
    });
}

function setupEditForm() {
    const form = document.getElementById('edit-form');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const submitBtn = form.querySelector('button[type="submit"]');
        const originalText = submitBtn.textContent;
        const pasteId = form.dataset.pasteId;

        try {
            submitBtn.disabled = true;
            submitBtn.textContent = 'Saving...';

            const data = {
                title: form.title.value,
                content: form.content.value,
                language: form.language.value,
                is_public: form.is_public.checked,
            };

            await API.updatePaste(pasteId, data);
            Toast.show('Paste updated!', 'success');
            setTimeout(() => {
                window.location.href = `/${pasteId}`;
            }, 1000);
        } catch (err) {
            Toast.show(err.message, 'error');
            submitBtn.disabled = false;
            submitBtn.textContent = originalText;
        }
    });
}

function setupAuthForms() {
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');

    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const submitBtn = loginForm.querySelector('button[type="submit"]');

            try {
                submitBtn.disabled = true;
                submitBtn.textContent = 'Logging in...';

                await API.login(loginForm.username.value, loginForm.password.value);
                window.location.href = '/dashboard';
            } catch (err) {
                Toast.show(err.message, 'error');
                submitBtn.disabled = false;
                submitBtn.textContent = 'Login';
            }
        });
    }

    if (registerForm) {
        registerForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const submitBtn = registerForm.querySelector('button[type="submit"]');

            try {
                submitBtn.disabled = true;
                submitBtn.textContent = 'Creating account...';

                await API.register(registerForm.username.value, registerForm.password.value);
                Toast.show('Account created! Logging you in...', 'success');
                setTimeout(() => {
                    window.location.href = '/dashboard';
                }, 1000);
            } catch (err) {
                Toast.show(err.message, 'error');
                submitBtn.disabled = false;
                submitBtn.textContent = 'Create Account';
            }
        });
    }
}

function setupDeleteButton() {
    const deleteBtn = document.getElementById('delete-paste');
    if (!deleteBtn) return;

    deleteBtn.addEventListener('click', async () => {
        if (!confirm('Are you sure you want to delete this paste?')) return;

        const pasteId = deleteBtn.dataset.pasteId;

        try {
            await API.deletePaste(pasteId);
            Toast.show('Paste deleted!', 'success');
            setTimeout(() => {
                window.location.href = '/dashboard';
            }, 1000);
        } catch (err) {
            Toast.show(err.message, 'error');
        }
    });
}

function setupForkButton() {
    const forkBtn = document.getElementById('fork-paste');
    if (!forkBtn) return;

    forkBtn.addEventListener('click', async () => {
        const pasteId = forkBtn.dataset.pasteId;

        try {
            const forked = await API.forkPaste(pasteId);
            Toast.show('Paste forked!', 'success');
            setTimeout(() => {
                window.location.href = `/${forked.id}`;
            }, 1000);
        } catch (err) {
            Toast.show(err.message, 'error');
        }
    });
}

function setupLogout() {
    const logoutBtn = document.getElementById('logout-btn');
    if (!logoutBtn) return;

    logoutBtn.addEventListener('click', async (e) => {
        e.preventDefault();
        try {
            await API.logout();
            window.location.href = '/';
        } catch (err) {
            Toast.show(err.message, 'error');
        }
    });
}

// ===== Line Numbers =====
function setupLineNumbers() {
    const lineNumbers = document.querySelector('.line-numbers');
    if (!lineNumbers) return;

    const lines = lineNumbers.querySelectorAll('span');
    lines.forEach((line, index) => {
        line.addEventListener('click', () => {
            const lineNum = index + 1;
            const newUrl = `${window.location.pathname}#L${lineNum}`;
            window.history.replaceState(null, '', newUrl);

            // Highlight the line
            const codeLines = document.querySelectorAll('.code-line');
            codeLines.forEach(l => l.classList.remove('highlighted'));
            if (codeLines[index]) {
                codeLines[index].classList.add('highlighted');
            }
        });
    });
}

// ===== Keyboard Shortcuts =====
function setupKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
        // Ctrl/Cmd + Enter to submit
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            const form = document.querySelector('#paste-form, #edit-form');
            if (form) {
                e.preventDefault();
                form.dispatchEvent(new Event('submit'));
            }
        }

        // Ctrl/Cmd + S to save
        if ((e.ctrlKey || e.metaKey) && e.key === 's') {
            const editForm = document.getElementById('edit-form');
            if (editForm) {
                e.preventDefault();
                editForm.dispatchEvent(new Event('submit'));
            }
        }
    });
}

// ===== Auto-resize Textarea =====
function setupAutoResize() {
    const textarea = document.querySelector('.form-textarea');
    if (!textarea) return;

    const resize = () => {
        textarea.style.height = 'auto';
        textarea.style.height = Math.max(300, textarea.scrollHeight) + 'px';
    };

    textarea.addEventListener('input', resize);
    resize();
}

// ===== Initialize =====
document.addEventListener('DOMContentLoaded', () => {
    ThemeManager.init();
    Toast.init();
    setupPasteForm();
    setupEditForm();
    setupAuthForms();
    setupDeleteButton();
    setupForkButton();
    setupLogout();
    setupLineNumbers();
    setupKeyboardShortcuts();
    setupAutoResize();
});

// Expose toggle for theme button
window.toggleTheme = () => ThemeManager.toggle();
window.copyToClipboard = copyToClipboard;
