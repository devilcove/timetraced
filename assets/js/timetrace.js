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

function showMenu() {
    var x = document.getElementById("tab");
    if (x.className === "tab") {
        x.className += " responsive";
    } else {
        x.className = "tab";
    }
}

function valPass() {
    var x = document.forms["editUser"]["password"].value;
    var y = document.forms["editUser"]["verify"].value;
    if (x != y) {
        alert("password are not the same");
        return false;
    }
}

applyTheme(current);