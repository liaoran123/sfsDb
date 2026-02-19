// 导入导出组件
export default {
    name: 'ImportExportComponent',
    template: `
        <div class="card">
            <div class="card-body">
                <h5 class="card-title mb-4">数据导入导出</h5>
                
                <!-- 导出部分 -->
                <div class="mb-6">
                    <h6 class="card-subtitle mb-3">导出数据</h6>
                    <div class="form-group mb-4">
                        <label for="exportTableName" class="form-label">表名</label>
                        <input 
                            type="text" 
                            id="exportTableName" 
                            v-model="exportForm.tableName" 
                            class="form-control"
                            placeholder="输入要导出的表名"
                        >
                    </div>
                    <div class="form-group mb-4">
                        <label for="exportFormat" class="form-label">导出格式</label>
                        <select id="exportFormat" v-model="exportForm.format" class="form-select">
                            <option value="json">JSON</option>
                            <option value="csv">CSV</option>
                        </select>
                    </div>
                    <button 
                        class="btn btn-primary" 
                        @click="exportTable"
                        :disabled="!exportForm.tableName"
                    >
                        导出数据
                    </button>
                </div>
                
                <!-- 导入部分 -->
                <div class="mb-6">
                    <h6 class="card-subtitle mb-3">导入数据</h6>
                    <div class="form-group mb-4">
                        <label for="importTableName" class="form-label">表名</label>
                        <input 
                            type="text" 
                            id="importTableName" 
                            v-model="importForm.tableName" 
                            class="form-control"
                            placeholder="输入要导入的表名"
                        >
                    </div>
                    <div class="form-group mb-4">
                        <label for="importPath" class="form-label">导入文件路径</label>
                        <input 
                            type="text" 
                            id="importPath" 
                            v-model="importForm.importPath" 
                            class="form-control"
                            placeholder="输入导入文件的路径"
                        >
                    </div>
                    <div class="form-group mb-4">
                        <label for="importFormat" class="form-label">导入格式</label>
                        <select id="importFormat" v-model="importForm.format" class="form-select">
                            <option value="json">JSON</option>
                            <option value="csv">CSV</option>
                        </select>
                    </div>
                    <div class="form-check mb-4">
                        <input 
                            type="checkbox" 
                            id="ignoreErrors" 
                            v-model="importForm.ignoreErrors"
                            class="form-check-input"
                        >
                        <label for="ignoreErrors" class="form-check-label">忽略错误</label>
                    </div>
                    <div class="form-check mb-4">
                        <input 
                            type="checkbox" 
                            id="truncate" 
                            v-model="importForm.truncate"
                            class="form-check-input"
                        >
                        <label for="truncate" class="form-check-label">导入前清空表</label>
                    </div>
                    <button 
                        class="btn btn-primary" 
                        @click="importTable"
                        :disabled="!importForm.tableName || !importForm.importPath"
                    >
                        导入数据
                    </button>
                </div>
                
                <!-- 操作结果 -->
                <div v-if="message" class="mt-4 alert" :class="messageType === 'success' ? 'alert-success' : 'alert-danger'">
                    {{ message }}
                </div>
            </div>
        </div>
    `,
    data() {
        return {
            exportForm: {
                tableName: '',
                format: 'json'
            },
            importForm: {
                tableName: '',
                importPath: '',
                format: 'json',
                ignoreErrors: false,
                truncate: false
            },
            message: '',
            messageType: 'success'
        }
    },
    methods: {
        async exportTable() {
            if (!this.exportForm.tableName) {
                this.showMessage('请输入表名', 'danger');
                return;
            }
            
            try {
                const data = await window.api.exportTable(
                    this.exportForm.tableName,
                    this.exportForm.format
                );
                
                if (data.error) {
                    this.showMessage(data.message || data.error, 'danger');
                } else {
                    this.showMessage(`导出成功，文件路径: ${data.exportPath}`, 'success');
                    // 清空表单
                    this.exportForm.tableName = '';
                }
            } catch (err) {
                this.showMessage('导出失败: ' + err.message, 'danger');
                console.error('导出失败:', err);
            }
        },
        
        async importTable() {
            if (!this.importForm.tableName) {
                this.showMessage('请输入表名', 'danger');
                return;
            }
            
            if (!this.importForm.importPath) {
                this.showMessage('请输入导入文件路径', 'danger');
                return;
            }
            
            try {
                const data = await window.api.importTable(
                    this.importForm.tableName,
                    this.importForm.importPath,
                    this.importForm.format,
                    this.importForm.ignoreErrors,
                    this.importForm.truncate
                );
                
                if (data.error) {
                    this.showMessage(data.message || data.error, 'danger');
                } else {
                    this.showMessage(`导入成功，导入了 ${data.importedCount} 条记录`, 'success');
                    // 清空表单
                    this.importForm.tableName = '';
                    this.importForm.importPath = '';
                }
            } catch (err) {
                this.showMessage('导入失败: ' + err.message, 'danger');
                console.error('导入失败:', err);
            }
        },
        
        showMessage(message, type) {
            this.message = message;
            this.messageType = type;
            
            // 3秒后自动清除消息
            setTimeout(() => {
                this.message = '';
            }, 3000);
        }
    }
};
