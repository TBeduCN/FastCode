// Waline评论系统初始化
import { init } from 'https://unpkg.com/@waline/client@v3/dist/waline.js';

window.WalineInstance = init({
    el: '#waline',
    serverURL: 'https://waline.tbedu.top',
    lang: 'zh-CN',
    dark: 'html[data-theme="dark"]',
    path: '/github',
    pageSize: 10,
    placeholder: '分享你的使用体验...',
    pageview: true,
    emoji: [
        'https://cdn.jsdelivr.net/gh/walinejs/emojis/weibo',
        'https://cdn.jsdelivr.net/gh/walinejs/emojis/bilibili'
    ]
});