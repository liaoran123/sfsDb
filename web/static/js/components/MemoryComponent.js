/**
 * 内存监控组件
 * 显示系统内存使用情况和监控状态
 */
export default {
    name: 'MemoryComponent',
    data() {
        return {
            loading: false,
            error: null,
            monitorConfig: {
                interval: '5s',
                memoryThreshold: 100,
                gcThreshold: 100
            },
            monitorData: {
                monitorRunning: false
            },
            monitorStatus: {
                currentStatus: null
            }
        };
    },
    methods: {
        async startMonitor() {
            this.loading = true;
            this.error = null;
            try {
                const data = await window.api.getMonitor({
                    op: 'start',
                    interval: this.monitorConfig.interval,
                    memoryThreshold: this.monitorConfig.memoryThreshold,
                    gcThreshold: this.monitorConfig.gcThreshold
                });
                this.monitorData.monitorRunning = true;
                await this.refreshMonitorStatus();
                console.log('监控启动成功:', data);
                window.utils.showNotification('监控启动成功', 'success');
            } catch (err) {
                this.error = '启动监控失败: ' + err.message;
                console.error('启动监控失败:', err);
                window.utils.showNotification('启动监控失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        async stopMonitor() {
            this.loading = true;
            this.error = null;
            try {
                const data = await window.api.getMonitor({ op: 'stop' });
                this.monitorData.monitorRunning = false;
                console.log('监控停止成功:', data);
                window.utils.showNotification('监控停止成功', 'success');
            } catch (err) {
                this.error = '停止监控失败: ' + err.message;
                console.error('停止监控失败:', err);
                window.utils.showNotification('停止监控失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        async refreshMonitorStatus() {
            this.loading = true;
            this.error = null;
            try {
                const data = await window.api.getMonitor({ op: 'status' });
                this.monitorStatus = data;
                console.log('监控状态刷新成功:', data);
            } catch (err) {
                this.error = '获取监控状态失败: ' + err.message;
                console.error('获取监控状态失败:', err);
                window.utils.showNotification('获取监控状态失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        }
    },
    mounted() {
        this.refreshMonitorStatus();
        console.log('MemoryComponent 已加载');
    },
    template: `
        <div class="card">
            <div class="card-body">
                <h5 class="card-title mb-4">内存监控</h5>
                
                <div v-if="error" class="alert alert-danger mb-4">
                    <i class="bi bi-exclamation-triangle me-2"></i>
                    {{ error }}
                </div>
                
                <!-- 监控控制 -->
                <div class="mb-6">
                    <h6 class="mb-3">监控控制</h6>
                    <div class="row">
                        <div class="col-md-6">
                            <div class="card">
                                <div class="card-body">
                                    <div class="form-group mb-4">
                                        <label for="interval" class="form-label">监控间隔</label>
                                        <input 
                                            type="text" 
                                            id="interval" 
                                            v-model="monitorConfig.interval" 
                                            class="form-control" 
                                            placeholder="5s"
                                        >
                                        <div class="form-text text-muted">例如: 5s, 10s, 1m</div>
                                    </div>
                                    <div class="form-group mb-4">
                                        <label for="memoryThreshold" class="form-label">内存使用阈值 (MB)</label>
                                        <input 
                                            type="number" 
                                            id="memoryThreshold" 
                                            v-model.number="monitorConfig.memoryThreshold" 
                                            class="form-control" 
                                            placeholder="100"
                                        >
                                    </div>
                                    <div class="form-group mb-4">
                                        <label for="gcThreshold" class="form-label">GC次数阈值</label>
                                        <input 
                                            type="number" 
                                            id="gcThreshold" 
                                            v-model.number="monitorConfig.gcThreshold" 
                                            class="form-control" 
                                            placeholder="100"
                                        >
                                    </div>
                                    <div class="mb-4">
                                        <button 
                                            v-if="!monitorData.monitorRunning" 
                                            class="btn btn-primary me-2" 
                                            @click="startMonitor" 
                                            :disabled="loading"
                                        >
                                            <i class="bi bi-play me-1"></i>
                                            <span v-if="loading">启动中...</span>
                                            <span v-else>启动监控</span>
                                        </button>
                                        <button 
                                            v-else 
                                            class="btn btn-danger me-2" 
                                            @click="stopMonitor" 
                                            :disabled="loading"
                                        >
                                            <i class="bi bi-stop me-1"></i>
                                            <span v-if="loading">停止中...</span>
                                            <span v-else>停止监控</span>
                                        </button>
                                        <button 
                                            class="btn btn-info" 
                                            @click="refreshMonitorStatus" 
                                            :disabled="loading"
                                        >
                                            <i class="bi bi-refresh me-1"></i>
                                            <span v-if="loading">刷新中...</span>
                                            <span v-else>刷新状态</span>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="col-md-6">
                            <div class="card">
                                <div class="card-body">
                                    <h6 class="card-subtitle mb-3 text-muted">监控状态</h6>
                                    <div v-if="monitorData.monitorRunning" class="alert alert-success mb-4">
                                        <i class="bi bi-check-circle me-2"></i>
                                        监控已启动
                                    </div>
                                    <div v-else class="alert alert-warning mb-4">
                                        <i class="bi bi-exclamation-circle me-2"></i>
                                        监控未启动
                                    </div>
                                    <div v-if="monitorStatus.currentStatus" class="mt-4">
                                        <h6 class="mb-3">当前状态</h6>
                                        <div class="row mb-2">
                                            <div class="col-4 text-muted">内存使用:</div>
                                            <div class="col-8">
                                                <span class="font-weight-bold">
                                                    {{ monitorStatus.currentStatus.memoryUsage.toFixed(2) }} MB
                                                </span>
                                            </div>
                                        </div>
                                        <div class="row mb-2">
                                            <div class="col-4 text-muted">GC次数:</div>
                                            <div class="col-8">
                                                <span class="font-weight-bold">
                                                    {{ monitorStatus.currentStatus.gcCount }}
                                                </span>
                                            </div>
                                        </div>
                                        <div class="row mb-2">
                                            <div class="col-4 text-muted">存储类型:</div>
                                            <div class="col-8">{{ monitorStatus.currentStatus.storeType }}</div>
                                        </div>
                                    </div>
                                    <div v-else class="text-muted">
                                        监控状态未获取，请刷新状态
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- 内存监控数据 -->
                <div class="mb-6">
                    <h6 class="mb-3">内存监控数据</h6>
                    <div v-if="monitorStatus.currentStatus" class="card">
                        <div class="card-body">
                            <div class="row">
                                <div class="col-md-6">
                                    <div class="mb-3">
                                        <label class="form-label">内存使用</label>
                                        <div class="progress">
                                            <div 
                                                class="progress-bar" 
                                                role="progressbar" 
                                                :style="{
                                                    width: Math.min(100, (monitorStatus.currentStatus.memoryUsage / monitorConfig.memoryThreshold) * 100) + '%',
                                                    backgroundColor: monitorStatus.currentStatus.memoryUsage > monitorConfig.memoryThreshold ? '#dc3545' : '#0d6efd'
                                                }"
                                                :aria-valuenow="monitorStatus.currentStatus.memoryUsage"
                                                aria-valuemin="0"
                                                :aria-valuemax="monitorConfig.memoryThreshold"
                                            >
                                                {{ monitorStatus.currentStatus.memoryUsage.toFixed(2) }} MB
                                            </div>
                                        </div>
                                    </div>
                                    <div class="mb-3">
                                        <label class="form-label">GC次数</label>
                                        <div class="progress">
                                            <div 
                                                class="progress-bar bg-warning" 
                                                role="progressbar" 
                                                :style="{
                                                    width: Math.min(100, (monitorStatus.currentStatus.gcCount / monitorConfig.gcThreshold) * 100) + '%'
                                                }"
                                                :aria-valuenow="monitorStatus.currentStatus.gcCount"
                                                aria-valuemin="0"
                                                :aria-valuemax="monitorConfig.gcThreshold"
                                            >
                                                {{ monitorStatus.currentStatus.gcCount }}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <div class="col-md-6">
                                    <div class="card bg-light">
                                        <div class="card-body">
                                            <h6 class="card-subtitle mb-3 text-muted">详细信息</h6>
                                            <div v-if="monitorStatus.currentStatus">
                                                <div class="row mb-2">
                                                    <div class="col-4 text-muted">内存使用:</div>
                                                    <div class="col-8">{{ monitorStatus.currentStatus.memoryUsage.toFixed(2) }} MB</div>
                                                </div>
                                                <div class="row mb-2">
                                                    <div class="col-4 text-muted">GC次数:</div>
                                                    <div class="col-8">{{ monitorStatus.currentStatus.gcCount }}</div>
                                                </div>
                                                <div class="row mb-2">
                                                    <div class="col-4 text-muted">存储类型:</div>
                                                    <div class="col-8">{{ monitorStatus.currentStatus.storeType }}</div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div v-else class="alert alert-info">
                        <i class="bi bi-info-circle me-2"></i>
                        内存监控数据未获取，请启动监控或刷新状态
                    </div>
                </div>
            </div>
        </div>
    `
};
