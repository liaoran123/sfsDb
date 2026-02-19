/**
 * 表操作组件
 * 用于执行表的CRUD操作，包括插入、查询、更新、删除等
 */
export default {
    name: 'CrudComponent',
    data() {
        return {
            loading: false,
            error: null,
            crudForm: {
                tableName: '',
                operation: 'insert',
                fields: '',
                confirmDeleteAll: false
            },
            crudResult: '',
            resultVisible: false
        };
    },
    methods: {
        async executeCRUD() {
            if (!this.crudForm.tableName) {
                this.error = '错误: 表名不能为空';
                return;
            }

            this.loading = true;
            this.error = null;
            this.crudResult = '';
            this.resultVisible = false;

            try {
                let data;

                switch (this.crudForm.operation) {
                    case 'insert': {
                        const fields = JSON.parse(this.crudForm.fields);
                        data = await window.api.insertRecord(this.crudForm.tableName, fields);
                        break;
                    }
                    case 'search': {
                        const fields = JSON.parse(this.crudForm.fields);
                        data = await window.api.getRecords(this.crudForm.tableName, fields);
                        break;
                    }
                    case 'update': {
                        const fields = JSON.parse(this.crudForm.fields);
                        data = await window.api.updateRecord(this.crudForm.tableName, fields);
                        break;
                    }
                    case 'delete': {
                        const fields = JSON.parse(this.crudForm.fields);
                        data = await window.api.deleteRecord(this.crudForm.tableName, fields);
                        break;
                    }
                    case 'deleteAll': {
                        if (!this.crudForm.confirmDeleteAll) {
                            this.error = '错误: 请确认删除所有数据';
                            this.loading = false;
                            return;
                        }
                        data = await window.api.deleteTable(this.crudForm.tableName);
                        break;
                    }
                    default:
                        this.error = '错误: 不支持的操作类型';
                        this.loading = false;
                        return;
                }

                this.crudResult = JSON.stringify(data, null, 2);
                this.resultVisible = true;

                if (this.crudForm.operation === 'deleteAll') {
                    this.crudForm.confirmDeleteAll = false;
                }
                
                console.log('CRUD操作成功:', data);
                window.utils.showNotification('操作成功', 'success');
            } catch (err) {
                this.error = '错误: ' + err.message;
                this.crudResult = '错误: ' + err.message;
                this.resultVisible = true;
                console.error('CRUD操作失败:', err);
                window.utils.showNotification('操作失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        resetForm() {
            this.crudForm = {
                tableName: '',
                operation: 'insert',
                fields: '',
                confirmDeleteAll: false
            };
            this.crudResult = '';
            this.error = null;
            this.resultVisible = false;
        },
        copyResult() {
            navigator.clipboard.writeText(this.crudResult)
                .then(() => {
                    window.utils.showNotification('结果已复制到剪贴板', 'success');
                })
                .catch(err => {
                    console.error('复制失败:', err);
                    window.utils.showNotification('复制失败: ' + err.message, 'error');
                });
        }
    },
    mounted() {
        console.log('CrudComponent 已加载');
    },
    template: `
        <div class="card">
            <div class="card-body">
                <h5 class="card-title mb-4">表操作</h5>
                
                <div v-if="error" class="alert alert-danger mb-4">
                    <i class="bi bi-exclamation-triangle me-2"></i>
                    {{ error }}
                </div>
                
                <div class="row">
                    <div class="col-md-6">
                        <div class="card mb-4">
                            <div class="card-body">
                                <h6 class="card-subtitle mb-3 text-muted">操作表单</h6>
                                
                                <div class="mb-3">
                                    <label for="tableName" class="form-label">表名</label>
                                    <input 
                                        type="text" 
                                        id="tableName" 
                                        v-model="crudForm.tableName" 
                                        class="form-control" 
                                        placeholder="输入表名"
                                    >
                                </div>
                                
                                <div class="mb-3">
                                    <label for="operation" class="form-label">操作类型</label>
                                    <select 
                                        id="operation" 
                                        v-model="crudForm.operation" 
                                        class="form-select"
                                    >
                                        <option value="insert">插入</option>
                                        <option value="search">查询</option>
                                        <option value="update">更新</option>
                                        <option value="delete">删除</option>
                                        <option value="deleteAll">删除所有数据</option>
                                    </select>
                                </div>
                                
                                <div class="mb-3" v-if="crudForm.operation !== 'deleteAll'">
                                    <label for="fields" class="form-label">字段 (JSON格式)</label>
                                    <textarea 
                                        id="fields" 
                                        v-model="crudForm.fields" 
                                        class="form-control" 
                                        rows="5" 
                                        placeholder="{id: 1, name: '测试'}"
                                    ></textarea>
                                    <div class="form-text text-muted">
                                        请输入有效的JSON格式字段数据
                                    </div>
                                </div>
                                
                                <div class="mb-3" v-if="crudForm.operation === 'deleteAll'">
                                    <div class="alert alert-warning" role="alert">
                                        <i class="bi bi-exclamation-triangle me-2"></i>
                                        警告：此操作将删除表中的所有数据，且无法恢复！
                                    </div>
                                    <div class="form-check">
                                        <input 
                                            type="checkbox" 
                                            id="confirmDeleteAll" 
                                            v-model="crudForm.confirmDeleteAll" 
                                            class="form-check-input"
                                        >
                                        <label for="confirmDeleteAll" class="form-check-label">
                                            我确认要删除所有数据
                                        </label>
                                    </div>
                                </div>
                                
                                <div class="d-flex gap-2">
                                    <button 
                                        class="btn btn-primary" 
                                        @click="executeCRUD" 
                                        :disabled="loading || !crudForm.tableName || (crudForm.operation === 'deleteAll' && !crudForm.confirmDeleteAll)"
                                    >
                                        <i class="bi bi-play-circle me-1"></i>
                                        <span v-if="loading">执行中...</span>
                                        <span v-else>执行操作</span>
                                    </button>
                                    <button 
                                        class="btn btn-outline-secondary" 
                                        @click="resetForm" 
                                        :disabled="loading"
                                    >
                                        <i class="bi bi-arrow-counterclockwise me-1"></i>
                                        重置
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <div class="col-md-6">
                        <div class="card mb-4">
                            <div class="card-body">
                                <h6 class="card-subtitle mb-3 text-muted">执行结果</h6>
                                
                                <div v-if="resultVisible" class="mb-3">
                                    <div class="card">
                                        <div class="card-header d-flex justify-content-between align-items-center">
                                            <h7 class="mb-0">操作结果</h7>
                                            <div>
                                                <button 
                                                    class="btn btn-sm btn-outline-secondary" 
                                                    @click="copyResult"
                                                >
                                                    <i class="bi bi-clipboard me-1"></i>
                                                    复制
                                                </button>
                                            </div>
                                        </div>
                                        <div class="card-body">
                                            <pre class="mb-0">{{ crudResult }}</pre>
                                        </div>
                                    </div>
                                </div>
                                
                                <div v-else class="text-center text-muted py-4">
                                    <i class="bi bi-file-earmark-text display-4 mb-2"></i>
                                    <p>执行操作后，结果将显示在这里</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `
};
