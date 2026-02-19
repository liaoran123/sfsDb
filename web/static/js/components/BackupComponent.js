/**
 * 备份管理组件
 * 用于创建和恢复系统备份
 */
export default {
    name: 'BackupComponent',
    data() {
        return {
            loading: false,
            error: null,
            backupPath: '',
            restoreFile: '',
            backupResult: null,
            restoreResult: null
        };
    },
    methods: {
        async createBackup() {
            this.loading = true;
            this.error = null;
            this.backupResult = null;
            try {
                const data = await window.api.createBackup(this.backupPath);
                this.backupResult = {
                    success: true,
                    message: data.message || '备份创建成功'
                };
                console.log('备份创建成功:', data);
                window.utils.showNotification('备份创建成功', 'success');
            } catch (err) {
                this.backupResult = {
                    success: false,
                    message: err.message || '创建备份失败'
                };
                this.error = '创建备份失败: ' + err.message;
                console.error('创建备份失败:', err);
                window.utils.showNotification('创建备份失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        async restoreBackup() {
            if (!this.restoreFile) {
                this.error = '请输入备份文件路径';
                return;
            }
            
            this.loading = true;
            this.error = null;
            this.restoreResult = null;
            try {
                const data = await window.api.restoreBackup(this.restoreFile);
                this.restoreResult = {
                    success: true,
                    message: data.message || '备份恢复成功'
                };
                console.log('备份恢复成功:', data);
                window.utils.showNotification('备份恢复成功', 'success');
            } catch (err) {
                this.restoreResult = {
                    success: false,
                    message: err.message || '恢复备份失败'
                };
                this.error = '恢复备份失败: ' + err.message;
                console.error('恢复备份失败:', err);
                window.utils.showNotification('恢复备份失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        clearResults() {
            this.backupResult = null;
            this.restoreResult = null;
            this.error = null;
        }
    },
    mounted() {
        console.log('BackupComponent 已加载');
    },
    template: `
        <div class="card">
            <div class="card-body">
                <h5 class="card-title mb-4">备份管理</h5>
                
                <div v-if="error" class="alert alert-danger mb-4">
                    <i class="bi bi-exclamation-triangle me-2"></i>
                    {{ error }}
                </div>
                
                <div class="row">
                    <!-- 创建备份 -->
                    <div class="col-md-6 mb-4">
                        <div class="card">
                            <div class="card-body">
                                <h6 class="card-subtitle mb-3 text-muted">创建备份</h6>
                                
                                <div class="mb-3">
                                    <label for="backupPath" class="form-label">备份路径</label>
                                    <input 
                                        type="text" 
                                        id="backupPath" 
                                        v-model="backupPath" 
                                        class="form-control" 
                                        placeholder="默认为 ./backups"
                                    >
                                    <div class="form-text text-muted">
                                        留空使用默认路径
                                    </div>
                                </div>
                                
                                <button 
                                    class="btn btn-primary" 
                                    @click="createBackup" 
                                    :disabled="loading"
                                >
                                    <i class="bi bi-save me-1"></i>
                                    <span v-if="loading">创建中...</span>
                                    <span v-else>创建备份</span>
                                </button>
                                
                                <div v-if="backupResult" class="mt-3">
                                    <div class="alert" :class="backupResult.success ? 'alert-success' : 'alert-danger'">
                                        {{ backupResult.message }}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <!-- 恢复备份 -->
                    <div class="col-md-6 mb-4">
                        <div class="card">
                            <div class="card-body">
                                <h6 class="card-subtitle mb-3 text-muted">恢复备份</h6>
                                
                                <div class="mb-3">
                                    <label for="restoreFile" class="form-label">备份文件路径</label>
                                    <input 
                                        type="text" 
                                        id="restoreFile" 
                                        v-model="restoreFile" 
                                        class="form-control" 
                                        placeholder="请输入备份文件路径"
                                    >
                                    <div class="form-text text-muted">
                                        请指定要恢复的备份文件路径
                                    </div>
                                </div>
                                
                                <button 
                                    class="btn btn-primary" 
                                    @click="restoreBackup" 
                                    :disabled="loading || !restoreFile"
                                >
                                    <i class="bi bi-upload me-1"></i>
                                    <span v-if="loading">恢复中...</span>
                                    <span v-else>恢复备份</span>
                                </button>
                                
                                <div v-if="restoreResult" class="mt-3">
                                    <div class="alert" :class="restoreResult.success ? 'alert-success' : 'alert-danger'">
                                        {{ restoreResult.message }}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="text-center mt-4">
                    <button class="btn btn-outline-secondary" @click="clearResults">
                        <i class="bi bi-trash me-1"></i> 清除结果
                    </button>
                </div>
            </div>
        </div>
    `
};
