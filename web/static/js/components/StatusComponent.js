/**
 * 系统状态组件
 * 显示系统的当前状态信息
 */
export default {
    name: 'StatusComponent',
    data() {
        return {
            loading: false,
            error: null,
            statusData: {
                memory: {
                    alloc: 0,
                    totalAlloc: 0,
                    sys: 0,
                    numGC: 0
                },
                storage: {
                    storeType: 'unknown'
                },
                status: 'running'
            }
        };
    },
    methods: {
        async loadStatusData() {
            this.loading = true;
            this.error = null;
            try {
                const data = await window.api.getSystemStatus();
                this.statusData = {
                    memory: {
                        alloc: data.Memory?.Alloc || 0,
                        totalAlloc: data.Memory?.TotalAlloc || 0,
                        sys: data.Memory?.Sys || 0,
                        numGC: data.Memory?.NumGC || 0
                    },
                    storage: {
                        storeType: data.Storage?.StoreType || 'unknown'
                    },
                    status: 'running'
                };
            } catch (err) {
                this.error = '加载系统状态失败: ' + err.message;
                console.error('加载系统状态失败:', err);
            } finally {
                this.loading = false;
            }
        },
        formatMemory(bytes) {
            if (typeof bytes !== 'number' || isNaN(bytes)) {
                return '0 GB';
            }
            if (bytes < 1024) return bytes + ' B';
            if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
            if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
            return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
        }
    },
    mounted() {
        this.loadStatusData();
        console.log('StatusComponent 已加载');
    },
    template: `
        <div class="card">
            <div class="card-body">
                <h5 class="card-title mb-4">系统状态</h5>
                
                <div v-if="loading" class="text-center">
                    <div class="spinner-border" role="status">
                        <span class="visually-hidden">加载中...</span>
                    </div>
                    <p class="mt-2 text-muted">正在加载系统状态...</p>
                </div>
                
                <div v-else-if="error" class="alert alert-danger">
                    <i class="bi bi-exclamation-triangle me-2"></i>
                    {{ error }}
                    <button class="btn btn-sm btn-outline-danger mt-2" @click="loadStatusData">
                        重试
                    </button>
                </div>
                
                <div v-else>
                    <div class="row">
                        <div class="col-md-6">
                            <div class="card mb-4">
                                <div class="card-body">
                                    <h6 class="card-subtitle mb-3 text-muted">基本信息</h6>
                                    <div class="row mb-2">
                                        <div class="col-4 text-muted">状态:</div>
                                        <div class="col-8">
                                            <span class="badge" :class="{
                                                'bg-success': statusData.status === 'running',
                                                'bg-warning': statusData.status === 'warning',
                                                'bg-danger': statusData.status === 'error',
                                                'bg-secondary': statusData.status === 'unknown'
                                            }">
                                                {{ statusData.status === 'running' ? '运行中' : 
                                                   statusData.status === 'warning' ? '警告' : 
                                                   statusData.status === 'error' ? '错误' : '未知' }}
                                            </span>
                                        </div>
                                    </div>
                                    <div class="row mb-2">
                                        <div class="col-4 text-muted">存储类型:</div>
                                        <div class="col-8">{{ statusData.storage.storeType }}</div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <div class="col-md-6">
                            <div class="card mb-4">
                                <div class="card-body">
                                    <h6 class="card-subtitle mb-3 text-muted">内存使用</h6>
                                    <div class="row mb-2">
                                        <div class="col-4 text-muted">当前分配:</div>
                                        <div class="col-8">{{ formatMemory(statusData.memory.alloc) }}</div>
                                    </div>
                                    <div class="row mb-2">
                                        <div class="col-4 text-muted">总分配:</div>
                                        <div class="col-8">{{ formatMemory(statusData.memory.totalAlloc) }}</div>
                                    </div>
                                    <div class="row mb-2">
                                        <div class="col-4 text-muted">系统内存:</div>
                                        <div class="col-8">{{ formatMemory(statusData.memory.sys) }}</div>
                                    </div>
                                    <div class="row mb-2">
                                        <div class="col-4 text-muted">GC次数:</div>
                                        <div class="col-8">{{ statusData.memory.numGC }}</div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <div class="text-center mt-4">
                        <button class="btn btn-outline-primary" @click="loadStatusData" :disabled="loading">
                            <i class="bi bi-refresh me-1"></i> 刷新状态
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `
};
