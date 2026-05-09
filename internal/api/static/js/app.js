const API_BASE = '';
let authToken = localStorage.getItem('auth_token');
let currentLang = localStorage.getItem('lang') || (navigator.language.startsWith('zh') ? 'zh' : 'en');

// Elements
const langSelect = document.getElementById('lang-select');
const loginOverlay = document.getElementById('login-overlay');
const loginForm = document.getElementById('login-form');
const loginError = document.getElementById('login-error');
const logoutBtn = document.getElementById('logout-btn');
const navItems = document.querySelectorAll('.nav-item');
const views = document.querySelectorAll('.view');
const changePasswordForm = document.getElementById('change-password-form');
const passwordMsg = document.getElementById('password-msg');

// State
let dashboardData = {
    channels: [],
    tokens: [],
    logs: []
};

// Initialization
document.addEventListener('DOMContentLoaded', () => {
    initLang();
    if (!authToken) {
        showLogin();
    } else {
        hideLogin();
        loadDashboard();
    }
});

// i18n Logic
function initLang() {
    langSelect.value = currentLang;
    applyTranslations();

    langSelect.addEventListener('change', (e) => {
        currentLang = e.target.value;
        localStorage.setItem('lang', currentLang);
        applyTranslations();
    });
}

function applyTranslations() {
    const dict = locales[currentLang];

    // Text elements
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        if (dict[key]) {
            el.textContent = dict[key];
        }
    });

    // Placeholders
    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
        const key = el.getAttribute('data-i18n-placeholder');
        if (dict[key]) {
            el.placeholder = dict[key];
        }
    });
}

// Auth Logic
async function handleLogin(e) {
    e.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();

        if (response.ok) {
            authToken = data.token;
            localStorage.setItem('auth_token', authToken);
            hideLogin();
            loadDashboard();
        } else {
            loginError.textContent = data.error || 'Login failed';
        }
    } catch (err) {
        loginError.textContent = 'Connection error';
    }
}

function showLogin() {
    loginOverlay.classList.remove('hidden');
}

function hideLogin() {
    loginOverlay.classList.add('hidden');
}

function handleLogout() {
    localStorage.removeItem('auth_token');
    authToken = null;
    showLogin();
}

// Navigation
navItems.forEach(item => {
    item.addEventListener('click', (e) => {
        e.preventDefault();
        const targetView = item.getAttribute('data-view');

        navItems.forEach(i => i.classList.remove('active'));
        item.classList.add('active');

        views.forEach(v => {
            v.classList.remove('active');
            if (v.id === `view-${targetView}`) {
                v.classList.add('active');
            }
        });
    });
});

// Data Loading
async function loadDashboard() {
    await Promise.all([
        fetchChannels(),
        fetchTokens(),
        fetchLogs()
    ]);
    updateStats();
    renderRecentLogs();
    renderChannelsTable();
    renderTokensTable();
}

async function fetchWithAuth(url, options = {}) {
    options.headers = {
        ...options.headers,
        'Authorization': `Bearer ${authToken}`
    };

    const response = await fetch(url, options);
    if (response.status === 401) {
        handleLogout();
        return null;
    }
    return response.json();
}

async function fetchChannels() {
    const data = await fetchWithAuth(`${API_BASE}/admin/channels`);
    if (data) dashboardData.channels = data;
}

async function fetchTokens() {
    const data = await fetchWithAuth(`${API_BASE}/admin/tokens`);
    if (data) dashboardData.tokens = data;
}

async function fetchLogs() {
    const filters = {
        model: document.getElementById('filter-model').value,
        provider: document.getElementById('filter-provider').value,
        ip: document.getElementById('filter-ip').value,
        status_code: document.getElementById('filter-status').value,
        start_date: document.getElementById('filter-start').value,
        end_date: document.getElementById('filter-end').value
    };

    const params = new URLSearchParams(filters);
    const data = await fetchWithAuth(`${API_BASE}/admin/logs?${params.toString()}`);
    if (data) dashboardData.logs = data;
}

function updateStats() {
    document.getElementById('stat-requests').textContent = dashboardData.logs.length;

    const totalTokens = dashboardData.logs.reduce((acc, l) => acc + (l.total_tokens || 0), 0);
    document.getElementById('stat-tokens').textContent = totalTokens.toLocaleString();

    const avgLatency = dashboardData.logs.length > 0
        ? Math.round(dashboardData.logs.reduce((acc, l) => acc + l.latency_ms, 0) / dashboardData.logs.length)
        : 0;
    document.getElementById('stat-latency').textContent = `${avgLatency}ms`;

    const errors = dashboardData.logs.filter(l => l.status_code >= 400).length;
    document.getElementById('stat-errors').textContent = errors;
}

function renderChannelsTable() {
    const body = document.getElementById('channels-table-body');
    if (!body) return;

    const chSearch = document.getElementById('channel-search')?.value.toLowerCase().trim() || '';
    const globalSearch = document.getElementById('global-search')?.value.toLowerCase().trim() || '';
    const searchTerm = chSearch || globalSearch;
    
    // Filter channels based on search term
    const filteredChannels = dashboardData.channels.filter(ch => {
        const name = (ch.name || '').toLowerCase();
        const type = (ch.type || '').toLowerCase();
        const models = (ch.allowed_models || '').toLowerCase();
        const baseUrl = (ch.base_url || '').toLowerCase();

        return name.includes(searchTerm) || 
               type.includes(searchTerm) || 
               models.includes(searchTerm) ||
               baseUrl.includes(searchTerm);
    });

    if (filteredChannels.length === 0) {
        body.innerHTML = `<tr><td colspan="5" style="text-align:center; padding: 40px; color: var(--text-secondary);">${locales[currentLang].no_logs || 'No results found'}</td></tr>`;
        return;
    }

    const t = locales[currentLang];
    body.innerHTML = filteredChannels.map(ch => {
        let models = ch.allowed_models ? ch.allowed_models.split(',').map(m => m.trim()).filter(m => m) : ['*'];
        
        // If searching, sort models to show matches first
        if (searchTerm) {
            const matching = models.filter(m => m.toLowerCase().includes(searchTerm));
            const nonMatching = models.filter(m => !m.toLowerCase().includes(searchTerm));
            models = [...matching, ...nonMatching];
        }

        const modelTags = models.map(m => {
            const isMatch = searchTerm && m.toLowerCase().includes(searchTerm);
            const matchClass = isMatch ? 'active' : '';
            const matchStyle = isMatch ? 'style="border-color: var(--accent-cyan); background: rgba(34, 211, 238, 0.1); color: var(--accent-cyan);"' : '';
            return `<span class="model-tag ${matchClass}" ${matchStyle}>${m}</span>`;
        }).join('');

        return `
        <tr>
            <td>
                <strong>${ch.name}</strong><br>
                <small class="type-tag">${ch.type}</small>
            </td>
            <td><code>${ch.base_url}</code></td>
            <td><div class="models-column">${modelTags}</div></td>
            <td><span class="status-badge ${ch.is_active ? 'active' : 'inactive'}">
                ${ch.is_active ? t.active : (t.inactive || 'Off')}
            </span></td>
            <td>
                <button class="icon-btn" onclick="editChannel(${ch.id})"><i data-lucide="edit"></i></button>
                <button class="icon-btn" onclick="deleteChannel(${ch.id})"><i data-lucide="trash-2"></i></button>
            </td>
        </tr>
    `}).join('');
    lucide.createIcons();
}

// Global Search & Channel Search Listener
document.addEventListener('DOMContentLoaded', () => {
    const globalSearch = document.getElementById('global-search');
    const channelSearch = document.getElementById('channel-search');
    const suggestionsList = document.getElementById('search-suggestions');

    const handleSearch = () => {
        const activeView = document.querySelector('.view.active')?.id;
        if (activeView === 'view-channels') {
            renderChannelsTable();
        } else if (activeView === 'view-tokens') {
            renderTokensTable();
        } else if (activeView === 'view-logs') {
            renderLogsTable();
        }
    };

    globalSearch?.addEventListener('input', handleSearch);
    
    channelSearch?.addEventListener('input', (e) => {
        handleSearch();
        const term = e.target.value.toLowerCase().trim();
        if (!term || term.length < 1) {
            suggestionsList.style.display = 'none';
            return;
        }

        const suggestions = [];
        const seen = new Set();

        dashboardData.channels.forEach(ch => {
            // Match channel name
            if (ch.name.toLowerCase().includes(term)) {
                suggestions.push({ id: ch.id, text: ch.name, type: 'Channel' });
                seen.add(`ch-${ch.id}-${ch.name}`);
            }
            // Match models
            if (ch.allowed_models) {
                ch.allowed_models.split(',').forEach(m => {
                    const modelName = m.trim();
                    if (modelName.toLowerCase().includes(term) && !seen.has(modelName)) {
                        suggestions.push({ id: ch.id, text: modelName, type: ch.name });
                        seen.add(modelName);
                    }
                });
            }
        });

        const unique = suggestions.slice(0, 10);
        if (unique.length > 0) {
            suggestionsList.innerHTML = unique.map(s => `
                <div class="suggestion-item" onclick="jumpToChannel(${s.id}, '${s.text.replace(/'/g, "\\'")}')">
                    <span class="match-text">${s.text}</span>
                    <span class="match-type">${s.type}</span>
                </div>
            `).join('');
            suggestionsList.style.display = 'block';
        } else {
            suggestionsList.style.display = 'none';
        }
    });

    // Hide suggestions when clicking outside
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.search-box')) {
            suggestionsList.style.display = 'none';
        }
    });
});

window.jumpToChannel = (id, highlightText) => {
    const tableBody = document.getElementById('channels-table-body');
    const rows = tableBody.querySelectorAll('tr');
    let targetRow = null;

    rows.forEach(row => {
        const editBtn = row.querySelector('button[onclick*="editChannel"]');
        if (editBtn && editBtn.getAttribute('onclick').includes(id)) {
            targetRow = row;
        }
    });

    if (targetRow) {
        targetRow.scrollIntoView({ behavior: 'smooth', block: 'center' });
        targetRow.classList.add('highlight-row');
        setTimeout(() => targetRow.classList.remove('highlight-row'), 2000);
        
        document.getElementById('channel-search').value = highlightText;
        renderChannelsTable();
    }
    document.getElementById('search-suggestions').style.display = 'none';
};

function renderRecentLogs() {
    const dashboardList = document.getElementById('recent-logs-list');
    const fullList = document.getElementById('full-logs-list');

    // Dashboard Card Style
    const dashboardHtml = (logs) => logs.map(log => `
        <div class="activity-item">
            <div class="activity-status ${log.status_code < 400 ? 'status-success' : 'status-error'}"></div>
            <div class="activity-detail">
                <div style="display: flex; justify-content: space-between; align-items: center;">
                    <h4>${log.model} <small class="provider-badge">${log.provider || locales[currentLang].unknown}</small></h4>
                    <span class="tokens-badge">${log.total_tokens || 0} ${locales[currentLang].tokens}</span>
                </div>
                <p>
                    <span class="activity-meta">${log.ip}</span> • 
                    ${log.latency_ms}ms • ${new Date(log.timestamp).toLocaleString()}
                </p>
            </div>
        </div>
    `).join('');

    // Full Table Row Style
    const tableHtml = (logs) => logs.map(log => `
        <tr>
            <td><span class="status-pill ${log.status_code < 400 ? 'success' : 'error'}">${log.status_code}</span></td>
            <td>${new Date(log.timestamp).toLocaleString()}</td>
            <td><strong>${log.model}</strong></td>
            <td><span class="provider-badge">${log.provider || locales[currentLang].unknown}</span></td>
            <td>${log.total_tokens || 0}</td>
            <td>${log.latency_ms}ms</td>
            <td class="mono-font">${log.ip}</td>
            <td class="mono-font">${log.request_id}</td>
        </tr>
    `).join('');

    if (dashboardList) {
        dashboardList.innerHTML = dashboardHtml(dashboardData.logs.slice(0, 10));
    }
    if (fullList) {
        fullList.innerHTML = tableHtml(dashboardData.logs);
    }
}

// Listeners
loginForm.addEventListener('submit', handleLogin);
logoutBtn.addEventListener('click', handleLogout);

changePasswordForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const oldPassword = document.getElementById('old-password').value;
    const newPassword = document.getElementById('new-password').value;
    const confirmPassword = document.getElementById('confirm-password').value;

    if (newPassword !== confirmPassword) {
        showMsg(passwordMsg, 'Passwords do not match', 'error');
        return;
    }

    try {
        const data = await fetchWithAuth(`${API_BASE}/admin/user/password`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                old_password: oldPassword,
                new_password: newPassword
            })
        });

        if (data && !data.error) {
            showMsg(passwordMsg, locales[currentLang].password_success, 'success');
            changePasswordForm.reset();
        } else {
            showMsg(passwordMsg, data.error || locales[currentLang].password_error, 'error');
        }
    } catch (err) {
        showMsg(passwordMsg, 'Connection error', 'error');
    }
});

function showMsg(el, text, type) {
    el.textContent = text;
    el.className = `msg-box ${type}`;
    setTimeout(() => {
        el.textContent = '';
        el.className = 'msg-box';
    }, 3000);
}

// Channel Modal Logic
const channelModal = document.getElementById('channel-modal');
const channelForm = document.getElementById('channel-form');
const addChannelBtn = document.getElementById('add-channel-btn');
const closeModalBtns = document.querySelectorAll('.close-modal');

if (addChannelBtn) {
    addChannelBtn.addEventListener('click', () => {
        channelForm.reset();
        document.getElementById('channel-id').value = '';
        document.querySelector('.modal-header h2').textContent = locales[currentLang].add_channel;
        channelModal.classList.add('active');
    });
}

closeModalBtns.forEach(btn => {
    btn.addEventListener('click', () => channelModal.classList.remove('active'));
});

channelForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = document.getElementById('channel-id').value;
    const payload = {
        name: document.getElementById('channel-name').value,
        type: document.getElementById('channel-type').value,
        base_url: document.getElementById('channel-base-url').value,
        api_keys: document.getElementById('channel-api-key').value.split('\n').filter(k => k.trim()),
        allowed_models: document.getElementById('channel-allowed-models').value,
        denied_models: document.getElementById('channel-denied-models').value,
        rpm: parseInt(document.getElementById('channel-rpm').value) || 0,
        is_active: document.getElementById('channel-status').checked
    };

    const method = id ? 'PUT' : 'POST';
    const url = id ? `${API_BASE}/admin/channels/${id}` : `${API_BASE}/admin/channels`;

    const res = await fetchWithAuth(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    });

    if (res && res.error) {
        alert("Error saving channel: " + res.error);
        return;
    }

    if (res) {
        channelModal.classList.remove('active');
        loadDashboard();
    }
});

window.editChannel = (id) => {
    const ch = dashboardData.channels.find(c => c.id === id);
    if (!ch) return;

    document.getElementById('channel-id').value = ch.id;
    document.getElementById('channel-name').value = ch.name;
    document.getElementById('channel-type').value = ch.type;
    document.getElementById('channel-base-url').value = ch.base_url;

    // Join API keys array back to string
    const keysStr = (ch.api_keys || []).map(k => k.key_value).join('\n');
    document.getElementById('channel-api-key').value = keysStr;

    document.getElementById('channel-allowed-models').value = ch.allowed_models || '';
    document.getElementById('channel-denied-models').value = ch.denied_models || '';
    document.getElementById('channel-rpm').value = ch.rpm || 0;
    document.getElementById('channel-status').checked = ch.is_active;

    // Initialize model picker with existing models
    const wrapper = document.getElementById('model-picker-wrapper');
    if (ch.allowed_models) {
        const models = ch.allowed_models.split(',').map(m => m.trim()).filter(m => m);
        window.allFetchedModels = models;
        wrapper.style.display = 'block';
        renderModelTags(models);
    } else {
        wrapper.style.display = 'none';
        window.allFetchedModels = [];
    }

    document.querySelector('.modal-header h2').textContent = locales[currentLang].edit_channel || 'Edit Channel';
    channelModal.classList.add('active');
};

window.deleteChannel = async (id) => {
    if (!confirm('Are you sure?')) return;
    await fetchWithAuth(`${API_BASE}/admin/channels/${id}`, { method: 'DELETE' });
    loadDashboard();
};

// Update Username Logic
const updateUsernameBtn = document.getElementById('update-username-btn');
const usernameMsg = document.getElementById('username-msg');

if (updateUsernameBtn) {
    updateUsernameBtn.addEventListener('click', async () => {
        const newUsername = document.getElementById('new-username').value;
        if (!newUsername) return;

        try {
            const data = await fetchWithAuth(`${API_BASE}/admin/user/username`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ new_username: newUsername })
            });

            if (data && !data.error) {
                showMsg(usernameMsg, data.message || 'Username updated', 'success');
                setTimeout(() => handleLogout(), 2000);
            } else {
                showMsg(usernameMsg, data.error || 'Failed to update username', 'error');
            }
        } catch (err) {
            showMsg(usernameMsg, 'Connection error', 'error');
        }
    });
}

// Model Picker Logic
function renderModelTags(models) {
    const cloud = document.getElementById('model-tag-cloud');
    if (!cloud) return;

    const input = document.getElementById('channel-allowed-models');
    const currentSelected = input.value.split(',').map(m => m.trim());

    cloud.innerHTML = models.map(m => {
        const active = currentSelected.includes(m) ? 'active' : '';
        return `<span class="model-tag picker-tag ${active}" onclick="toggleModelSelection('${m}')">${m}</span>`;
    }).join('');
}

window.toggleModelSelection = (model) => {
    const input = document.getElementById('channel-allowed-models');
    let selected = input.value.split(',').map(m => m.trim()).filter(m => m);

    if (selected.includes(model)) {
        selected = selected.filter(m => m !== model);
    } else {
        selected.push(model);
    }

    input.value = selected.join(',');
    // Re-render only if cloud is visible
    const wrapper = document.getElementById('model-picker-wrapper');
    if (wrapper.style.display !== 'none') {
        const term = document.getElementById('model-search-filter').value.toLowerCase();
        const filtered = (window.allFetchedModels || []).filter(m => m.toLowerCase().includes(term));
        renderModelTags(filtered);
    }
};

// Fetch Models Logic
const fetchModelsBtn = document.getElementById('fetch-models-btn');
if (fetchModelsBtn) {
    fetchModelsBtn.addEventListener('click', async () => {
        const baseUrl = document.getElementById('channel-base-url').value;
        const apiKey = document.getElementById('channel-api-key').value;
        if (!baseUrl || !apiKey) {
            alert('Please fill Base URL and API Key first');
            return;
        }

        const icon = fetchModelsBtn.querySelector('i, svg');
        if (icon) icon.classList.add('spin');
        fetchModelsBtn.disabled = true;

        try {
            let firstLine = apiKey.split(/[\n,\r]/)[0].trim();
            const firstKey = firstLine.replace(/[^\x21-\x7E]/g, '');

            const data = await fetchWithAuth(`${API_BASE}/admin/channels/fetch-models`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ base_url: baseUrl, api_key: firstKey })
            });

            if (data && data.models) {
                const modelStr = data.models;
                document.getElementById('channel-allowed-models').value = modelStr;

                const wrapper = document.getElementById('model-picker-wrapper');
                const models = modelStr.split(',').map(m => m.trim()).filter(m => m);

                wrapper.style.display = 'block';
                window.allFetchedModels = models;
                renderModelTags(models);
            } else if (data && data.error) {
                alert(data.error);
            }
        } catch (err) {
            alert('获取模型失败: ' + err.message);
        } finally {
            if (icon) icon.classList.remove('spin');
            fetchModelsBtn.disabled = false;
        }
    });
}

document.getElementById('model-search-filter')?.addEventListener('input', (e) => {
    const term = e.target.value.toLowerCase();
    const filtered = (window.allFetchedModels || []).filter(m => m.toLowerCase().includes(term));
    renderModelTags(filtered);
});

let filterTimeout;
function debounceFilter() {
    clearTimeout(filterTimeout);
    filterTimeout = setTimeout(() => {
        applyFilters();
    }, 500);
}

async function applyFilters() {
    await fetchLogs();
    renderRecentLogs();
}

function resetFilters() {
    document.getElementById('filter-model').value = '';
    document.getElementById('filter-provider').value = '';
    document.getElementById('filter-ip').value = '';
    document.getElementById('filter-status').value = '';
    document.getElementById('filter-start').value = '';
    document.getElementById('filter-end').value = '';
    applyFilters();
}

// Token Management
function renderTokensTable() {
    const body = document.getElementById('tokens-table-body');
    if (!body) return;

    body.innerHTML = dashboardData.tokens.map(tk => {
        const policyParts = [];
        const t = locales[currentLang];
        if (tk.allowed_channels) policyParts.push(t.channels + ': ' + tk.allowed_channels);
        if (tk.allowed_models) policyParts.push(t.allowed_models + ': ' + tk.allowed_models);
        if (tk.denied_models) policyParts.push(t.denied_models + ': ' + tk.denied_models);
        if (tk.allowed_ips) policyParts.push(t.allowed_ips + ': ' + tk.allowed_ips);
        if (tk.rpm > 0) policyParts.push('RPM: ' + tk.rpm);

        return `
        <tr>
            <td><strong>${tk.name}</strong></td>
            <td><span class="status-badge ${tk.is_active ? 'active' : 'inactive'}">
                ${tk.is_active ? t.active : t.inactive}
            </span></td>
            <td><small class="policy-hint">${policyParts.join(' | ') || t.no_restrictions}</small></td>
            <td>${new Date(tk.created_at).toLocaleDateString()}</td>
            <td>
                <button class="action-btn-sm blue" onclick="editToken(${tk.id})"><i data-lucide="edit"></i></button>
                <button class="action-btn-sm red" onclick="deleteToken(${tk.id})"><i data-lucide="trash"></i></button>
            </td>
        </tr>`;
    }).join('');
    lucide.createIcons();
}

function openTokenModal(tk = null) {
    const modal = document.getElementById('token-modal');
    const form = document.getElementById('token-form');
    form.reset();

    // Render Channel Checkboxes
    const container = document.getElementById('token-channels-container');
    const allowedArr = tk?.allowed_channels ? tk.allowed_channels.split(',') : [];

    container.innerHTML = dashboardData.channels.map(ch => `
        <label class="checkbox-item">
            <input type="checkbox" name="token-channel" value="${ch.id}" ${allowedArr.includes(ch.id.toString()) ? 'checked' : ''}>
            <span>${ch.name}</span>
        </label>
    `).join('');

    if (tk) {
        document.getElementById('token-id').value = tk.id;
        document.getElementById('token-name').value = tk.name;
        document.getElementById('token-status').checked = tk.is_active;
        document.getElementById('token-expiry').value = 'permanent'; // Reset to permanent on edit for simplicity
        document.getElementById('token-rpm').value = tk.rpm || 0;
        document.getElementById('token-allowed-models').value = tk.allowed_models || '';
        document.getElementById('token-denied-models').value = tk.denied_models || '';
        document.getElementById('token-allowed-ips').value = tk.allowed_ips || '';
        document.getElementById('token-denied-ips').value = tk.denied_ips || '';
    } else {
        document.getElementById('token-id').value = '';
        document.getElementById('token-rpm').value = 0;
    }
    modal.classList.add('active');
}

async function handleTokenSubmit(e) {
    e.preventDefault();
    const id = document.getElementById('token-id').value;

    // Gather checked channels
    const checkedChannels = Array.from(document.querySelectorAll('input[name="token-channel"]:checked'))
        .map(cb => cb.value)
        .join(',');

    const data = {
        name: document.getElementById('token-name').value,
        is_active: document.getElementById('token-status').checked,
        allowed_channels: checkedChannels,
        rpm: parseInt(document.getElementById('token-rpm').value) || 0,
        allowed_models: document.getElementById('token-allowed-models').value,
        denied_models: document.getElementById('token-denied-models').value,
        allowed_ips: document.getElementById('token-allowed-ips').value,
        denied_ips: document.getElementById('token-denied-ips').value,
    };

    const expiryValue = document.getElementById('token-expiry').value;
    if (expiryValue !== 'permanent') {
        const days = parseInt(expiryValue);
        const date = new Date();
        date.setDate(date.getDate() + days);
        data.expires_at = date.toISOString();
    } else {
        data.expires_at = null;
    }

    const url = id ? `${API_BASE}/admin/tokens/${id}` : `${API_BASE}/admin/tokens`;
    const method = id ? 'PUT' : 'POST';

    const result = await fetchWithAuth(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });

    if (result) {
        if (!id && result.token) {
            // Show Reveal Modal instead of alert
            document.getElementById('new-token-value').textContent = result.token;
            document.getElementById('reveal-modal').classList.add('active');
            lucide.createIcons();
        }
        document.getElementById('token-modal').classList.remove('active');
        await fetchTokens();
        renderTokensTable();
    }
}

window.copyNewToken = () => {
    const val = document.getElementById('new-token-value').textContent;
    navigator.clipboard.writeText(val).then(() => {
        const btn = document.querySelector('#reveal-modal .icon-btn');
        const originalIcon = btn.innerHTML;
        btn.innerHTML = '<i data-lucide="check"></i>';
        lucide.createIcons();
        setTimeout(() => {
            btn.innerHTML = originalIcon;
            lucide.createIcons();
        }, 2000);
    });
};

window.closeRevealModal = () => {
    document.getElementById('reveal-modal').classList.remove('active');
};

function editToken(id) {
    const tk = dashboardData.tokens.find(t => t.id === id);
    if (tk) openTokenModal(tk);
}

async function deleteToken(id) {
    if (confirm('Delete this token?')) {
        const res = await fetchWithAuth(`${API_BASE}/admin/tokens/${id}`, { method: 'DELETE' });
        if (res !== null) {
            await fetchTokens();
            renderTokensTable();
        }
    }
}

// Bind Token Events
document.getElementById('add-token-btn')?.addEventListener('click', () => openTokenModal());
document.getElementById('token-form')?.addEventListener('submit', handleTokenSubmit);

// Handle Token Modal Close
document.querySelector('.close-token-modal')?.addEventListener('click', () => {
    document.getElementById('token-modal').classList.remove('active');
});

// Also handle the X button for token modal
document.querySelector('#token-modal .close-modal')?.addEventListener('click', () => {
    document.getElementById('token-modal').classList.remove('active');
});
