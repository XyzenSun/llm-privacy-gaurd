document.addEventListener('DOMContentLoaded', function() {
    var logsBody = document.getElementById('logs-body');
    var filterBtn = document.getElementById('filter-btn');
    var filterSession = document.getElementById('filter-session');
    var logoutLink = document.getElementById('logout-link');
    var modal = document.getElementById('detail-modal');
    var closeBtn = document.querySelector('.close-btn');
    var detailContent = document.getElementById('detail-content');

    loadLogs();

    filterBtn.addEventListener('click', loadLogs);
    logoutLink.addEventListener('click', logout);
    closeBtn.addEventListener('click', closeModal);
    modal.addEventListener('click', function(e) {
        if (e.target === modal) closeModal();
    });

    function loadLogs() {
        var sid = filterSession.value.trim();
        var url = '/api/v1/logs?limit=50';
        if (sid) url += '&sessionId=' + encodeURIComponent(sid);

        fetch(url)
        .then(function(res) { return res.json(); })
        .then(function(data) {
            logsBody.innerHTML = '';
            (data.logs || []).forEach(function(log) {
                var tr = document.createElement('tr');
                tr.innerHTML = '<td>' + formatTime(log.createdAt) + '</td>' +
                    '<td>' + (log.sessionId || '').substring(0, 8) + '</td>' +
                    '<td>' + log.turnId + '</td>' +
                    '<td>' + truncate(log.promptBeforeMask, 50) + '</td>' +
                    '<td>' + truncate(log.promptAfterMask, 50) + '</td>' +
                    '<td><button data-id="' + log.id + '">详情</button></td>';
                logsBody.appendChild(tr);
            });
            logsBody.querySelectorAll('button').forEach(function(btn) {
                btn.addEventListener('click', function() {
                    showDetail(btn.dataset.id);
                });
            });
        })
        .catch(function(err) { console.error('load logs:', err); });
    }

    function showDetail(id) {
        fetch('/api/v1/logs/' + id)
        .then(function(res) { return res.json(); })
        .then(function(log) {
            detailContent.innerHTML = '<div class="detail-section"><h4>脱敏前 Prompt</h4><pre>' + escapeHtml(log.promptBeforeMask) + '</pre></div>' +
                '<div class="detail-section"><h4>脱敏后 Prompt</h4><pre>' + escapeHtml(log.promptAfterMask) + '</pre></div>' +
                '<div class="detail-section"><h4>映射表</h4><pre>' + escapeHtml(log.byPlaceholderJson) + '</pre></div>' +
                '<div class="detail-section"><h4>云端响应 (Masked)</h4><pre>' + escapeHtml(log.cloudResponseMasked) + '</pre></div>' +
                '<div class="detail-section"><h4>最终响应 (Unmasked)</h4><pre>' + escapeHtml(log.cloudResponseUnmasked) + '</pre></div>';
            modal.classList.remove('hidden');
        })
        .catch(function(err) { console.error('show detail:', err); });
    }

    function closeModal() {
        modal.classList.add('hidden');
    }

    function logout(e) {
        e.preventDefault();
        fetch('/api/v1/auth/logout', {method: 'POST'})
        .then(function() { window.location.href = '/login.html'; })
        .catch(function() { window.location.href = '/login.html'; });
    }

    function formatTime(t) {
        if (!t) return '-';
        return new Date(t).toLocaleString('zh-CN');
    }

    function truncate(s, len) {
        if (!s) return '-';
        return s.length > len ? s.substring(0, len) + '...' : s;
    }

    function escapeHtml(s) {
        if (!s) return '';
        return s.replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;');
    }
});