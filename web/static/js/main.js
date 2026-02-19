/**
 * 主应用逻辑模块
 * 负责处理应用的核心逻辑和状态管理
 */

// 使用全局的Vue对象
const { createApp } = window.Vue;
// 使用全局的api对象
const api = window.api;
// 使用全局的utils对象
const utils = window.utils;

// 动态导入组件
const { defineAsyncComponent } = window.Vue;

const components = {
    StatusComponent: defineAsyncComponent(() => import('./components/StatusComponent.js')),
    SystemComponent: defineAsyncComponent(() => import('./components/SystemComponent.js')),
    ConfigComponent: defineAsyncComponent(() => import('./components/ConfigComponent.js')),
    BackupComponent: defineAsyncComponent(() => import('./components/BackupComponent.js')),
    MemoryComponent: defineAsyncComponent(() => import('./components/MemoryComponent.js')),
    KeyComponent: defineAsyncComponent(() => import('./components/KeyComponent.js')),
    IndexStatsComponent: defineAsyncComponent(() => import('./components/IndexStatsComponent.js')),
    TransactionStatsComponent: defineAsyncComponent(() => import('./components/TransactionStatsComponent.js')),
    CrudComponent: defineAsyncComponent(() => import('./components/CrudComponent.js')),
    DataQueryComponent: defineAsyncComponent(() => import('./components/DataQueryComponent.js')),
    ImportExportComponent: defineAsyncComponent(() => import('./components/ImportExportComponent.js'))
};

const app = createApp({
    components: {
        StatusComponent: components.StatusComponent,
        SystemComponent: components.SystemComponent,
        ConfigComponent: components.ConfigComponent,
        BackupComponent: components.BackupComponent,
        MemoryComponent: components.MemoryComponent,
        KeyComponent: components.KeyComponent,
        IndexStatsComponent: components.IndexStatsComponent,
        TransactionStatsComponent: components.TransactionStatsComponent,
        CrudComponent: components.CrudComponent,
        DataQueryComponent: components.DataQueryComponent,
        ImportExportComponent: components.ImportExportComponent
    },
    data() {
        return {
            activePage: 'status',
            loading: false,
            error: null,
            componentErrors: {},
            monitorData: {},
            monitorConfig: {
                interval: '5s',
                memoryThreshold: 100,
                gcThreshold: 100
            },
            monitorStatus: {},
            crudForm: {
                tableName: '',
                operation: 'insert',
                fields: '',
                confirmDeleteAll: false
            },
            crudResult: '',
            queryForm: {
                tableName: '',
                queryType: 'timeAxis',
                timeField: 'timestamp',
                startTime: '',
                endTime: '',
                indexFields: [],
                indexFieldValues: {},
                availableDevices: ['设备A', '设备B', '设备C'],
                selectedDevices: [],
                availableSensors: ['温度', '湿度', '压力', '流量'],
                selectedSensors: []
            },
            queryLoading: false,
            queryError: null,
            queryResult: null
        }
    },
    computed: {
        pageTitle() {
            const titles = {
                status: '系统状态',
                system: '系统信息',
                config: '配置管理',
                backup: '备份管理',
                memory: '内存监控',
                key: '键值监控',
                indexStats: '索引统计',

                crud: '表操作',
                dataQuery: '数据查询',
                importExport: '数据导入导出'
            };
            return titles[this.activePage] || 'sfsDb管理界面';
        }
    },
    methods: {
        setActivePage(page) {
            this.activePage = page;
            this.error = null;
            this.componentErrors[page] = null;
        },
        async loadMonitorData() {
            this.loading = true;
            this.error = null;
            
            try {
                const data = await api.getMonitor();
                this.monitorData = data;
                if (this.activePage === 'memory') {
                    await this.refreshMonitorStatus();
                }
            } catch (err) {
                this.error = '加载监控数据失败: ' + err.message;
                console.error('监控数据加载失败:', err);
                utils.showNotification('加载监控数据失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        async executeCRUD() {
            if (!this.crudForm.tableName) {
                this.crudResult = '错误: 表名不能为空';
                return;
            }

            try {
                let data;

                switch (this.crudForm.operation) {
                    case 'insert': {
                        const fields = JSON.parse(this.crudForm.fields);
                        data = await api.insertRecord(this.crudForm.tableName, fields);
                        break;
                    }
                    case 'search': {
                        const fields = JSON.parse(this.crudForm.fields);
                        data = await api.getRecords(this.crudForm.tableName, fields);
                        break;
                    }
                    case 'update': {
                        const fields = JSON.parse(this.crudForm.fields);
                        data = await api.updateRecord(this.crudForm.tableName, fields);
                        break;
                    }
                    case 'delete': {
                        const fields = JSON.parse(this.crudForm.fields);
                        data = await api.deleteRecord(this.crudForm.tableName, fields);
                        break;
                    }
                    case 'deleteAll': {
                        data = await api.deleteTable(this.crudForm.tableName);
                        break;
                    }
                    default:
                        this.crudResult = '错误: 不支持的操作类型';
                        return;
                }

                this.crudResult = JSON.stringify(data, null, 2);

                if (this.crudForm.operation === 'deleteAll') {
                    this.crudForm.confirmDeleteAll = false;
                }
                
                utils.showNotification('操作成功', 'success');
            } catch (err) {
                this.crudResult = '错误: ' + err.message;
                console.error('CRUD操作失败:', err);
                utils.showNotification('操作失败: ' + err.message, 'error');
            }
        },
        async getIndexFields() {
            if (!this.queryForm.tableName) {
                this.queryError = '错误: 表名不能为空';
                return;
            }

            this.queryLoading = true;
            this.queryError = null;

            try {
                const systemData = await api.getSystemInfo();
                
                let tableFound = false;
                if (systemData.tableDetails) {
                    for (const [tableID, tableDetails] of Object.entries(systemData.tableDetails)) {
                        if (tableDetails.name === this.queryForm.tableName) {
                            if (tableDetails.indexes && tableDetails.indexes.length > 0) {
                                const indexFields = [];
                                tableDetails.indexes.forEach(index => {
                                    let fieldName = index.Name;
                                    if (fieldName.startsWith('index_')) {
                                        fieldName = fieldName.substring(6);
                                    }
                                    if (fieldName.includes('_')) {
                                        const parts = fieldName.split('_');
                                        if (parts.length === 2) {
                                            if (parts[1] === 'id' || parts[1] === 'type' || parts[1] === 'name') {
                                                indexFields.push(parts[0] + '_' + parts[1]);
                                            } else {
                                                parts.forEach(part => {
                                                    if (part.length > 2) {
                                                        indexFields.push(part);
                                                    }
                                                });
                                            }
                                        } else if (parts.length > 2) {
                                            let currentField = '';
                                            for (let i = 0; i < parts.length; i++) {
                                                const part = parts[i];
                                                if (part === 'id' || part === 'type' || part === 'name' || part === 'time') {
                                                    if (currentField) {
                                                        indexFields.push(currentField + '_' + part);
                                                        currentField = '';
                                                    } else {
                                                        indexFields.push(part);
                                                    }
                                                } else {
                                                    if (currentField) {
                                                        currentField += '_' + part;
                                                    } else {
                                                        currentField = part;
                                                    }
                                                }
                                            }
                                            if (currentField && currentField.length > 2) {
                                                indexFields.push(currentField);
                                            }
                                        }
                                    } else {
                                        indexFields.push(fieldName);
                                    }
                                });
                                this.queryForm.indexFields = indexFields;
                                this.queryForm.indexFieldValues = {};
                            } else {
                                if (tableDetails.fields && tableDetails.fields.length > 0) {
                                    this.queryForm.indexFields = tableDetails.fields.map(field => field.Name);
                                    this.queryForm.indexFieldValues = {};
                                } else {
                                    this.queryForm.indexFields = [];
                                    this.queryError = '错误: 表中没有字段信息';
                                }
                            }
                            tableFound = true;
                            break;
                        }
                    }
                }

                if (!tableFound) {
                    const tableData = await api.getTable(this.queryForm.tableName);
                    
                    if (tableData.fields && tableData.fields.length > 0) {
                        this.queryForm.indexFields = tableData.fields;
                        this.queryForm.indexFieldValues = {};
                    } else {
                        this.queryForm.indexFields = [];
                        this.queryError = '错误: 表中没有字段信息';
                    }
                }
                
                utils.showNotification('索引字段获取成功', 'success');
            } catch (err) {
                this.queryError = '错误: ' + err.message;
                console.error('获取索引字段失败:', err);
                utils.showNotification('获取索引字段失败: ' + err.message, 'error');
            } finally {
                this.queryLoading = false;
            }
        },
        getFieldTypeHint(field) {
            return utils.getFieldTypeHint(field);
        },
        async executeQuery() {
            if (!this.queryForm.tableName) {
                this.queryError = '错误: 表名不能为空';
                return;
            }

            this.queryLoading = true;
            this.queryError = null;
            this.queryResult = null;

            try {
                let data;

                if (this.queryForm.queryType === 'timeAxis') {
                    if (!this.queryForm.timeField) {
                        this.queryError = '错误: 时间字段不能为空';
                        this.queryLoading = false;
                        return;
                    }

                    let startTime = this.queryForm.startTime;
                    let endTime = this.queryForm.endTime;

                    if (startTime) {
                        startTime = new Date(startTime).getTime() / 1000;
                    }
                    if (endTime) {
                        endTime = new Date(endTime).getTime() / 1000;
                    }

                    data = await api.searchRange(this.queryForm.tableName, this.queryForm.timeField, startTime, endTime);
                } else if (this.queryForm.queryType === 'tagFilter') {
                    const params = {};

                    for (const field of this.queryForm.indexFields) {
                        const value = this.queryForm.indexFieldValues[field];
                        if (value !== undefined && value !== null && value !== '') {
                            params[field] = value;
                        }
                    }

                    if (this.queryForm.selectedDevices.length > 0) {
                        this.queryForm.selectedDevices.forEach(device => {
                            params.device_id = device;
                        });
                    }

                    if (this.queryForm.selectedSensors.length > 0) {
                        this.queryForm.selectedSensors.forEach(sensor => {
                            params.sensor_type = sensor;
                        });
                    }

                    data = await api.getRecords(this.queryForm.tableName, params);
                }

                if (!data) {
                    this.queryError = '错误: 查询失败';
                    this.queryLoading = false;
                    return;
                }

                if (data.error) {
                    this.queryError = data.message || data.error;
                } else {
                    if (this.queryForm.queryType === 'timeAxis') {
                        this.queryResult = {
                            columns: data.records.length > 0 ? Object.keys(data.records[0]) : [],
                            data: data.records,
                            total: data.count || data.records.length,
                            queryTime: 0
                        };
                    } else if (this.queryForm.queryType === 'tagFilter') {
                        this.queryResult = {
                            columns: data.records.length > 0 ? Object.keys(data.records[0]) : [],
                            data: data.records,
                            total: data.count || data.records.length,
                            queryTime: 0
                        };
                    }
                }
                
                utils.showNotification('查询成功', 'success');
            } catch (err) {
                this.queryError = '错误: ' + err.message;
                console.error('查询失败:', err);
                utils.showNotification('查询失败: ' + err.message, 'error');
            } finally {
                this.queryLoading = false;
            }
        }, 
        async startMonitor() {
            try {
                this.loading = true;

                const data = await api.getMonitor({
                    op: 'start',
                    interval: this.monitorConfig.interval,
                    memoryThreshold: this.monitorConfig.memoryThreshold,
                    gcThreshold: this.monitorConfig.gcThreshold
                });

                this.error = null;
                await this.loadMonitorData();
                await this.refreshMonitorStatus();
                
                utils.showNotification('监控启动成功', 'success');
            } catch (err) {
                this.error = '启动监控失败: ' + err.message;
                console.error('启动监控失败:', err);
                utils.showNotification('启动监控失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        async stopMonitor() {
            try {
                this.loading = true;

                const data = await api.getMonitor({ op: 'stop' });

                this.error = null;
                this.monitorData.monitorRunning = false;
                await this.loadMonitorData();
                await this.refreshMonitorStatus();
                
                utils.showNotification('监控停止成功', 'success');
            } catch (err) {
                this.error = '停止监控失败: ' + err.message;
                console.error('停止监控失败:', err);
                utils.showNotification('停止监控失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        async refreshMonitorStatus() {
            try {
                this.loading = true;

                const data = await api.getMonitor({ op: 'status' });

                this.error = null;
                this.monitorStatus = data;
            } catch (err) {
                this.error = '获取监控状态失败: ' + err.message;
                console.error('获取监控状态失败:', err);
                utils.showNotification('获取监控状态失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        async refreshData() {
            await this.loadMonitorData();
        },
        handleComponentError(page, error) {
            this.componentErrors[page] = error;
            console.error(`${page} 组件错误:`, error);
        }
    },
    mounted() {
        this.loadMonitorData();
        console.log('sfsDb管理界面已加载');
    }
});

app.mount('#app');
