// 主题切换功能
const themeToggle = document.getElementById('theme-toggle');
const themeIcon = document.getElementById('theme-icon');
const html = document.documentElement;

// 初始化主题：先检查本地存储，再检查设备偏好，最后使用默认值
const savedTheme = localStorage.getItem('theme');
const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
const initialTheme = savedTheme || (prefersDark ? 'dark' : 'light');
html.setAttribute('data-theme', initialTheme);
localStorage.setItem('theme', initialTheme);
updateThemeIcon(initialTheme);

// 切换主题
themeToggle.addEventListener('click', () => {
    const currentTheme = html.getAttribute('data-theme');
    const newTheme = currentTheme === 'light' ? 'dark' : 'light';
    
    html.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    updateThemeIcon(newTheme);
});

// 更新主题图标
function updateThemeIcon(theme) {
    if (theme === 'dark') {
        themeIcon.innerHTML = '<path d="M12 3a9 9 0 0 0-9 9 9.75 9.75 0 0 0 6.74 9A9.75 9.75 0 0 0 12 21h.75a.75.75 0 0 0 .75-.75v-1.5a.75.75 0 0 0-.75-.75H12a8.25 8.25 0 0 1-8.25-8.25A8.25 8.25 0 0 1 12 4.5h.75a.75.75 0 0 0 .75-.75V2.25a.75.75 0 0 0-.75-.75H12Zm-9 9a9 9 0 0 0 4.5 7.74V18a.75.75 0 0 0-.75-.75H12A9 9 0 0 0 3 12Zm9 0a9 9 0 0 0-4.5-7.74V6a.75.75 0 0 0 .75-.75H12a9 9 0 0 0 9 9Z" fill="currentColor"></path>';
    } else {
        themeIcon.innerHTML = '<path d="M12 2.5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 12 2.5zm0 17a7.5 7.5 0 1 1 7.5-7.5 7.5 7.5 0 0 1-7.5 7.5z" fill="currentColor"></path>';
    }
}

// 表单提交处理
function toSubmit(e) {
    e.preventDefault();
    const url = document.getElementById('url-input').value.trim();
    if (url) {
        window.open(location.href.substr(0, location.href.lastIndexOf('/') + 1) + url);
    }
    return false;
}

// 获取表单元素
const form = document.querySelector('form');
const urlInput = document.getElementById('url-input');

// 添加表单提交事件监听器
if (form) {
    form.addEventListener('submit', toSubmit);
}

// 回车键提交
if (urlInput) {
    urlInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            toSubmit(e);
        }
    });
}
