/**
 * API调用模块
 * 负责处理与后端API的所有交互
 */

// API基础URL
const API_BASE_URL = '/api';

// API密钥配置
const API_KEY = 'test_api_key_for_development';

/**
 * 通用API请求函数
 * @param {string} endpoint - API端点
 * @param {object} options - 请求选项
 * @returns {Promise} - 返回Promise对象
 */
async function apiRequest(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;
    
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            'X-API-Key': API_KEY
        }
    };
    
    const mergedOptions = {
        ...defaultOptions,
        ...options
    };
    
    try {
        const response = await fetch(url, mergedOptions);
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.message || data.error || 'API请求失败');
        }
        
        return data;
    } catch (error) {
        console.error('API请求错误:', error);
        throw error;
    }
}

// 全局API对象
const api = {
    /**
     * 导出诊断包
     * @returns {Promise} 诊断包下载
     */
    async exportDiagnostic() {
        const response = await apiRequest('/api/diagnostic', {
            method: 'GET'
        });
        return response;
    },

    /**
     * 获取系统状态
     * @returns {Promise} - 系统状态数据
     */
    async getSystemStatus() {
        return apiRequest('/status');
    },

    /**
     * 获取系统信息
     * @returns {Promise} - 系统信息数据
     */
    async getSystemInfo() {
        return apiRequest('/system');
    },

    /**
     * 获取配置信息
     * @returns {Promise} - 配置信息数据
     */
    async getConfig() {
        return apiRequest('/config');
    },

    /**
     * 获取监控信息
     * @param {object} params - 查询参数
     * @returns {Promise} - 监控信息数据
     */
    async getMonitor(params = {}) {
        const queryParams = new URLSearchParams(params).toString();
        const endpoint = `/monitor${queryParams ? `?${queryParams}` : ''}`;
        return apiRequest(endpoint);
    },

    /**
     * 创建备份
     * @param {string} path - 备份路径
     * @returns {Promise} - 备份结果
     */
    async createBackup(path = '') {
        const queryParams = new URLSearchParams({ op: 'create' });
        if (path) {
            queryParams.append('path', path);
        }
        return apiRequest(`/backup?${queryParams.toString()}`);
    },

    /**
     * 恢复备份
     * @param {string} file - 备份文件路径
     * @returns {Promise} - 恢复结果
     */
    async restoreBackup(file) {
        return apiRequest(`/backup?op=restore&file=${encodeURIComponent(file)}`);
    },

    /**
     * 生成API密钥
     * @param {string} role - 角色
     * @returns {Promise} - API密钥信息
     */
    async generateAPIKey(role) {
        return apiRequest('/auth/generate', {
            method: 'POST',
            body: JSON.stringify({ role })
        });
    },

    /**
     * 撤销API密钥
     * @param {string} key - API密钥
     * @returns {Promise} - 撤销结果
     */
    async revokeAPIKey(key) {
        return apiRequest('/auth/revoke', {
            method: 'POST',
            body: JSON.stringify({ key })
        });
    },

    /**
     * 获取所有表
     * @returns {Promise} - 表列表
     */
    async getTables() {
        return apiRequest('/tables');
    },

    /**
     * 获取表信息
     * @param {string} tableName - 表名
     * @returns {Promise} - 表信息
     */
    async getTable(tableName) {
        return apiRequest(`/tables/${tableName}`);
    },

    /**
     * 创建表
     * @param {string} tableName - 表名
     * @param {object} fields - 表字段
     * @returns {Promise} - 创建结果
     */
    async createTable(tableName, fields) {
        return apiRequest('/tables', {
            method: 'POST',
            body: JSON.stringify({ name: tableName, fields })
        });
    },

    /**
     * 更新表
     * @param {string} tableName - 表名
     * @param {object} fields - 表字段
     * @returns {Promise} - 更新结果
     */
    async updateTable(tableName, fields) {
        return apiRequest(`/tables/${tableName}`, {
            method: 'PUT',
            body: JSON.stringify({ fields })
        });
    },





    /**
     * 获取指标信息
     * @returns {Promise} - 指标信息
     */
    async getMetrics() {
        return apiRequest('/metrics');
    },

    /**
     * 获取告警信息
     * @returns {Promise} - 告警信息
     */
    async getAlerts() {
        return apiRequest('/alerts');
    },

    /**
     * 设置阈值
     * @param {object} thresholds - 阈值配置
     * @returns {Promise} - 设置结果
     */
    async setThresholds(thresholds) {
        return apiRequest('/thresholds', {
            method: 'POST',
            body: JSON.stringify(thresholds)
        });
    },

    /**
     * 获取阈值配置
     * @returns {Promise} - 阈值配置
     */
    async getThresholds() {
        return apiRequest('/thresholds');
    },

    /**
     * 收集指标
     * @returns {Promise} - 收集结果
     */
    async collectMetrics() {
        return apiRequest('/metrics/collect', {
            method: 'POST'
        });
    },

    /**
     * 设置配置
     * @param {string} key - 配置键
     * @param {string} value - 配置值
     * @returns {Promise} - 设置结果
     */
    async setConfig(key, value) {
        return apiRequest('/config', {
            method: 'POST',
            body: JSON.stringify({ key, value })
        });
    },

    /**
     * 设置场景配置
     * @param {string} scenario - 场景名称
     * @returns {Promise} - 设置结果
     */
    async setScenarioConfig(scenario) {
        return apiRequest('/config/scenario', {
            method: 'POST',
            body: JSON.stringify({ scenario })
        });
    },

    /**
     * 重置配置
     * @returns {Promise} - 重置结果
     */
    async resetConfig() {
        return apiRequest('/config/reset', {
            method: 'POST'
        });
    },

    /**
     * 导出表数据
     * @param {string} tableName - 表名
     * @param {string} format - 导出格式
     * @param {array} fields - 要导出的字段
     * @param {string} whereClause - 导出条件
     * @returns {Promise} - 导出结果
     */
    async exportTable(tableName, format = 'json', fields = [], whereClause = '') {
        return apiRequest('/tables/export', {
            method: 'POST',
            body: JSON.stringify({ tableName, format, fields, whereClause })
        });
    },

    /**
     * 导入表数据
     * @param {string} tableName - 表名
     * @param {string} importPath - 导入文件路径
     * @param {string} format - 导入格式
     * @param {boolean} ignoreErrors - 是否忽略错误
     * @param {boolean} truncate - 是否在导入前清空表
     * @returns {Promise} - 导入结果
     */
    async importTable(tableName, importPath, format = '', ignoreErrors = false, truncate = false) {
        return apiRequest('/tables/import', {
            method: 'POST',
            body: JSON.stringify({ tableName, importPath, format, ignoreErrors, truncate })
        });
    }
};

// 将api对象暴露到全局作用域
window.api = api;