/**
 * 工具函数模块
 * 提供通用的工具函数
 */

// 全局工具对象
const utils = {
    /**
     * 格式化日期时间
     * @param {string|Date} date - 日期对象或日期字符串
     * @returns {string} - 格式化后的日期时间字符串
     */
    formatDateTime(date) {
        if (!date) return '';
        
        const d = typeof date === 'string' ? new Date(date) : date;
        return d.toLocaleString();
    },

    /**
     * 格式化数字
     * @param {number} num - 数字
     * @param {number} decimals - 小数位数
     * @returns {string} - 格式化后的数字字符串
     */
    formatNumber(num, decimals = 2) {
        if (typeof num !== 'number') return '';
        return num.toFixed(decimals);
    },

    /**
     * 生成唯一ID
     * @returns {string} - 唯一ID
     */
    generateUUID() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c === 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    },

    /**
     * 深拷贝对象
     * @param {object} obj - 要拷贝的对象
     * @returns {object} - 拷贝后的对象
     */
    deepClone(obj) {
        if (obj === null || typeof obj !== 'object') return obj;
        if (obj instanceof Date) return new Date(obj.getTime());
        if (obj instanceof Array) return obj.map(item => utils.deepClone(item));
        if (typeof obj === 'object') {
            const clonedObj = {};
            for (const key in obj) {
                if (obj.hasOwnProperty(key)) {
                    clonedObj[key] = utils.deepClone(obj[key]);
                }
            }
            return clonedObj;
        }
    },

    /**
     * 验证邮箱格式
     * @param {string} email - 邮箱地址
     * @returns {boolean} - 是否为有效的邮箱格式
     */
    validateEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email);
    },

    /**
     * 验证URL格式
     * @param {string} url - URL地址
     * @returns {boolean} - 是否为有效的URL格式
     */
    validateURL(url) {
        try {
            new URL(url);
            return true;
        } catch {
            return false;
        }
    },

    /**
     * 截断字符串
     * @param {string} str - 字符串
     * @param {number} length - 最大长度
     * @param {string} suffix - 后缀
     * @returns {string} - 截断后的字符串
     */
    truncateString(str, length = 50, suffix = '...') {
        if (!str || str.length <= length) return str;
        return str.substring(0, length) + suffix;
    },

    /**
     * 防抖函数
     * @param {function} func - 要执行的函数
     * @param {number} wait - 等待时间（毫秒）
     * @returns {function} - 防抖处理后的函数
     */
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    },

    /**
     * 节流函数
     * @param {function} func - 要执行的函数
     * @param {number} limit - 时间限制（毫秒）
     * @returns {function} - 节流处理后的函数
     */
    throttle(func, limit) {
        let inThrottle;
        return function(...args) {
            if (!inThrottle) {
                func.apply(this, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    },

    /**
     * 获取字段类型提示
     * @param {string} field - 字段名
     * @returns {string} - 字段类型提示
     */
    getFieldTypeHint(field) {
        // 根据字段名猜测字段类型
        const lowerField = field.toLowerCase();
        
        // 时间类型字段
        if (lowerField.includes('time') || lowerField.includes('date') || lowerField.includes('timestamp')) {
            return '时间';
        }
        
        // 数字类型字段
        if (lowerField.includes('id') || lowerField.includes('num') || lowerField.includes('count') || 
            lowerField.includes('price') || lowerField.includes('value') || lowerField.includes('amount') ||
            lowerField.includes('score') || lowerField.includes('rate') || lowerField.includes('level')) {
            return '数字';
        }
        
        // 布尔类型字段
        if (lowerField.includes('bool') || lowerField.includes('flag') || lowerField.includes('status') ||
            lowerField.includes('enable') || lowerField.includes('active') || lowerField.includes('valid')) {
            return '布尔';
        }
        
        // 设备相关字段
        if (lowerField.includes('device') || lowerField.includes('sensor') || lowerField.includes('machine') ||
            lowerField.includes('equipment')) {
            return '设备';
        }
        
        // 位置相关字段
        if (lowerField.includes('loc') || lowerField.includes('position') || lowerField.includes('place') ||
            lowerField.includes('address') || lowerField.includes('area')) {
            return '位置';
        }
        
        // 默认类型
        return '文本';
    },

    /**
     * 解析查询字符串
     * @param {string} queryString - 查询字符串
     * @returns {object} - 解析后的对象
     */
    parseQueryString(queryString) {
        const params = {};
        if (!queryString) return params;
        
        const pairs = queryString.startsWith('?') ? queryString.substring(1).split('&') : queryString.split('&');
        pairs.forEach(pair => {
            const [key, value] = pair.split('=');
            if (key) {
                params[decodeURIComponent(key)] = decodeURIComponent(value || '');
            }
        });
        
        return params;
    },

    /**
     * 构建查询字符串
     * @param {object} params - 参数对象
     * @returns {string} - 查询字符串
     */
    buildQueryString(params) {
        if (!params || typeof params !== 'object') return '';
        
        const pairs = [];
        for (const [key, value] of Object.entries(params)) {
            if (value !== undefined && value !== null && value !== '') {
                pairs.push(`${encodeURIComponent(key)}=${encodeURIComponent(value)}`);
            }
        }
        
        return pairs.length > 0 ? `?${pairs.join('&')}` : '';
    },

    /**
     * 显示通知
     * @param {string} message - 通知消息
     * @param {string} type - 通知类型 (success, error, info, warning)
     * @param {number} duration - 显示时长（毫秒）
     */
    showNotification(message, type = 'info', duration = 3000) {
        const notification = document.createElement('div');
        notification.className = `notification ${type}-message`;
        notification.textContent = message;
        notification.style.position = 'fixed';
        notification.style.top = '20px';
        notification.style.right = '20px';
        notification.style.padding = '15px';
        notification.style.borderRadius = '4px';
        notification.style.zIndex = '9999';
        notification.style.boxShadow = '0 2px 8px rgba(0, 0, 0, 0.15)';
        notification.style.transition = 'all 0.3s ease';
        notification.style.opacity = '0';
        notification.style.transform = 'translateX(100%)';
        
        // 设置不同类型的样式
        switch (type) {
            case 'success':
                notification.style.backgroundColor = '#d4edda';
                notification.style.color = '#155724';
                break;
            case 'error':
                notification.style.backgroundColor = '#f8d7da';
                notification.style.color = '#721c24';
                break;
            case 'warning':
                notification.style.backgroundColor = '#fff3cd';
                notification.style.color = '#856404';
                break;
            default:
            case 'info':
                notification.style.backgroundColor = '#d1ecf1';
                notification.style.color = '#0c5460';
        }
        
        document.body.appendChild(notification);
        
        // 显示通知
        setTimeout(() => {
            notification.style.opacity = '1';
            notification.style.transform = 'translateX(0)';
        }, 100);
        
        // 隐藏通知
        setTimeout(() => {
            notification.style.opacity = '0';
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => {
                if (document.body.contains(notification)) {
                    document.body.removeChild(notification);
                }
            }, 300);
        }, duration);
    },

    /**
     * 检查是否为空对象
     * @param {object} obj - 要检查的对象
     * @returns {boolean} - 是否为空对象
     */
    isEmptyObject(obj) {
        return Object.keys(obj).length === 0 && obj.constructor === Object;
    },

    /**
     * 检查是否为数字
     * @param {any} value - 要检查的值
     * @returns {boolean} - 是否为数字
     */
    isNumber(value) {
        return typeof value === 'number' && !isNaN(value);
    },

    /**
     * 检查是否为字符串
     * @param {any} value - 要检查的值
     * @returns {boolean} - 是否为字符串
     */
    isString(value) {
        return typeof value === 'string';
    },

    /**
     * 检查是否为布尔值
     * @param {any} value - 要检查的值
     * @returns {boolean} - 是否为布尔值
     */
    isBoolean(value) {
        return typeof value === 'boolean';
    },

    /**
     * 检查是否为数组
     * @param {any} value - 要检查的值
     * @returns {boolean} - 是否为数组
     */
    isArray(value) {
        return Array.isArray(value);
    },

    /**
     * 检查是否为对象
     * @param {any} value - 要检查的值
     * @returns {boolean} - 是否为对象
     */
    isObject(value) {
        return value !== null && typeof value === 'object' && !Array.isArray(value);
    }
};

// 将utils对象暴露到全局作用域
window.utils = utils;