// Waline评论系统初始化
import { init } from 'https://unpkg.com/@waline/client@v3/dist/waline.js';

// 通过API获取UUID，然后初始化Waline
fetch('/api/uuid')
    .then(response => response.json())
    .then(data => {
        const uuid = data.uuid || 'unknown';
        window.WalineInstance = init({
            el: '#waline',
            serverURL: 'https://walinejs.comment.lithub.cc',
            lang: 'zh-CN',
            login: 'force',
            dark: 'html[data-theme="dark"]',
            path: `/fastcode-${uuid}`,
            pageSize: 10,
            placeholder: '分享你的使用体验...',
            pageview: true,
            emoji: [
                'https://cdn.jsdelivr.net/gh/walinejs/emojis/weibo',
                'https://cdn.jsdelivr.net/gh/walinejs/emojis/bilibili'
            ]
        });
    })
    .catch(error => {
        console.error('获取UUID失败:', error);
        // 如果获取UUID失败，使用默认路径
        window.WalineInstance = init({
            el: '#waline',
            serverURL: 'https://walinejs.comment.lithub.cc',
            lang: 'zh-CN',
            login: 'force',
            dark: 'html[data-theme="dark"]',
            path: '/fastcode-unknown',
            pageSize: 10,
            placeholder: '分享你的使用体验...',
            pageview: true,
            emoji: [
                'https://cdn.jsdelivr.net/gh/walinejs/emojis/weibo',
                'https://cdn.jsdelivr.net/gh/walinejs/emojis/bilibili'
            ]
        });
    });