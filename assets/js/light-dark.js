let current = localStorage.getItem('theme') || 'system';

function changeTheme() {
    var x = document.getElementById("theme").value;
    applyTheme(x);
}

function applyTheme(mode) {
    if (mode === 'os') {
        document.documentElement.removeAttribute('data-theme');
    } else {
        document.documentElement.setAttribute('data-theme', mode);
    }
    localStorage.setItem('theme', mode);
}

function currentTheme() {
    var current = localStorage.getItem('theme') || 'system';
    var x = document.getElementById("theme");
    x.value = current;
}

applyTheme(current);