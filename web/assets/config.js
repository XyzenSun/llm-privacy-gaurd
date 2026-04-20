document.addEventListener('DOMContentLoaded', function() {
    var logoutLink = document.getElementById('logout-link');
    var configForm = document.getElementById('config-form');
    var saveMessage = document.getElementById('save-message');

    loadConfig();
    loadHealth();

    logoutLink.addEventListener('click', logout);
    configForm.addEventListener('submit', saveConfig);

    function loadConfig() {
        fetch('/api/v1/config')
        .then(function(res) { return res.json(); })
        .then(function(data) {
            document.getElementById('db-type').textContent = data.database ? data.database.type : '-';

            document.getElementById('trusted-provider').value = data.trustedLlm ? (data.trustedLlm.provider || '') : '';
            document.getElementById('trusted-model').value = data.trustedLlm ? (data.trustedLlm.model || '') : '';
            document.getElementById('trusted-base-url').value = data.trustedLlm ? (data.trustedLlm.baseUrl || '') : '';
            document.getElementById('trusted-system-prompt').value = data.trustedLlm ? (data.trustedLlm.systemPrompt || '') : '';
            document.getElementById('trusted-api-key').value = data.trustedLlm ? (data.trustedLlm.apiKey || '') : '';
            document.getElementById('trusted-api-key-set').textContent = data.trustedLlm && data.trustedLlm.apiKeySet ? '是' : '否';

            document.getElementById('cloud-provider').value = data.cloudLlm ? (data.cloudLlm.provider || '') : '';
            document.getElementById('cloud-model').value = data.cloudLlm ? (data.cloudLlm.model || '') : '';
            document.getElementById('cloud-base-url').value = data.cloudLlm ? (data.cloudLlm.baseUrl || '') : '';
            document.getElementById('cloud-api-key').value = data.cloudLlm ? (data.cloudLlm.apiKey || '') : '';
            document.getElementById('cloud-api-key-set').textContent = data.cloudLlm && data.cloudLlm.apiKeySet ? '是' : '否';
        })
        .catch(function(err) { console.error('load config:', err); });
    }

    function saveConfig(e) {
        e.preventDefault();
        saveMessage.className = '';
        saveMessage.textContent = '保存中…';

        var payload = {
            trustedLlm: {
                provider: document.getElementById('trusted-provider').value.trim(),
                model: document.getElementById('trusted-model').value.trim(),
                baseUrl: document.getElementById('trusted-base-url').value.trim(),
                systemPrompt: document.getElementById('trusted-system-prompt').value,
                apiKey: document.getElementById('trusted-api-key').value
            },
            cloudLlm: {
                provider: document.getElementById('cloud-provider').value.trim(),
                model: document.getElementById('cloud-model').value.trim(),
                baseUrl: document.getElementById('cloud-base-url').value.trim(),
                apiKey: document.getElementById('cloud-api-key').value
            }
        };

        fetch('/api/v1/config', {
            method: 'PUT',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(payload)
        })
        .then(function(res) { return res.json(); })
        .then(function(data) {
            if (data.error) {
                saveMessage.textContent = '保存失败: ' + data.error;
                saveMessage.className = 'is-error';
                return;
            }
            saveMessage.textContent = '保存成功';
            saveMessage.className = 'is-ok';
            loadConfig();
        })
        .catch(function(err) {
            saveMessage.textContent = '保存失败: ' + err.message;
            saveMessage.className = 'is-error';
        });
    }

    function loadHealth() {
        fetch('/api/v1/health')
        .then(function(res) { return res.json(); })
        .then(function(data) {
            var statusEl = document.getElementById('health-status');
            var status = data.status || '—';
            statusEl.textContent = status;
            statusEl.dataset.state = status === 'ok' ? 'ok' : 'error';
        })
        .catch(function() {
            var statusEl = document.getElementById('health-status');
            statusEl.textContent = 'error';
            statusEl.dataset.state = 'error';
        });
    }

    function logout(e) {
        e.preventDefault();
        fetch('/api/v1/auth/logout', {method: 'POST'})
        .then(function() { window.location.href = '/login.html'; })
        .catch(function() { window.location.href = '/login.html'; });
    }
});