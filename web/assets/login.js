document.addEventListener('DOMContentLoaded', function() {
    var form = document.getElementById('login-form');
    var errorMsg = document.getElementById('error-msg');

    checkAuthStatus(function(authenticated) {
        if (authenticated) {
            window.location.href = '/index.html';
        }
    });

    form.addEventListener('submit', function(e) {
        e.preventDefault();
        var token = document.getElementById('authtoken').value;

        fetch('/api/v1/auth/login', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({authtoken: token})
        })
        .then(function(res) { return res.json(); })
        .then(function(data) {
            if (data.error) {
                errorMsg.textContent = data.error;
            } else {
                window.location.href = '/index.html';
            }
        })
        .catch(function(err) {
            errorMsg.textContent = '登录失败: ' + err.message;
        });
    });
});

function checkAuthStatus(callback) {
    fetch('/api/v1/auth/status')
    .then(function(res) { return res.json(); })
    .then(function(data) { callback(data.authenticated); })
    .catch(function() { callback(false); });
}