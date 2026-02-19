/**
 * 键值监控组件
 * 显示系统的键值变更统计信息
 */
export default {
    name: 'KeyComponent',
    data() {
        return {
            loading: false,
            error: null,
            monitorData: {
                keyChangeStats: null
            }
        };
    },
    methods: {
        async refreshData() {
            this.loading = true;
            this.error = null;
            try {
                const data = await window.api.getMonitor();
                this.monitorData.keyChangeStats = data.keyChangeStats;
                console.log('键值变更统计数据刷新成功:', data);
                window.utils.showNotification('键值变更统计数据刷新成功', 'success');
            } catch (err) {
                this.error = '获取键值变更统计失败: ' + err.message;
                console.error('获取键值变更统计失败:', err);
                window.utils.showNotification('获取键值变更统计失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        }
    },
    mounted() {
        this.refreshData();
        console.log('KeyComponent 已加载');
    },
    template: `
        <div class="card">
            <div class="card-body">
                <div class="d-flex justify-content-between align-items-center mb-4">
                    <h5 class="card-title mb-0">键值监控</h5>
                    <button 
                        class="btn btn-outline-primary" 
                        @click="refreshData" 
                        :disabled="loading"
                    >
                        <i class="bi bi-refresh me-1"></i>
                        <span v-if="loading">刷新中...</span>
                        <span v-else>刷新数据</span>
                    </button>
                </div>
                
                <div v-if="error" class="alert alert-danger mb-4">
                    <i class="bi bi-exclamation-triangle me-2"></i>
                    {{ error }}
                </div>
                
                <div v-if="loading" class="text-center">
                    <div class="spinner-border" role="status">
                        <span class="visually-hidden">加载中...</span>
                    </div>
                    <p class="mt-2 text-muted">正在加载键值变更统计数据...</p>
                </div>
                
                <div v-else>
                    <!-- 键值变更统计 -->
                    <div class="mb-6">
                        <h6 class="mb-3">键值变更统计</h6>
                        <div v-if="monitorData.keyChangeStats" class="table-responsive">
                            <table class="table table-striped table-hover">
                                <thead>
                                    <tr>
                                        <th>表ID</th>
                                        <th>索引名</th>
                                        <th>Put操作次数</th>
                                        <th>Delete操作次数</th>
                                        <th>净增次数</th>
                                        <th>操作比率</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(stats, key) in monitorData.keyChangeStats" :key="key">
                                        <td>{{ stats.TbId }}</td>
                                        <td>{{ stats.IndxName }}</td>
                                        <td>{{ stats.PutCount }}</td>
                                        <td>{{ stats.DeleteCount }}</td>
                                        <td>
                                            <span :class="{
                                                'text-success': stats.PutCount > stats.DeleteCount,
                                                'text-danger': stats.PutCount < stats.DeleteCount,
                                                'text-muted': stats.PutCount === stats.DeleteCount
                                            }">
                                                {{ stats.PutCount - stats.DeleteCount }}
                                            </span>
                                        </td>
                                        <td>
                                            <div class="progress" style="height: 8px;">
                                                <div 
                                                    class="progress-bar bg-success" 
                                                    role="progressbar" 
                                                    :style="{
                                                        width: (stats.PutCount / (stats.PutCount + stats.DeleteCount) * 100) + '%'
                                                    }"
                                                    aria-valuenow="{{ stats.PutCount }}"
                                                    aria-valuemin="0"
                                                    :aria-valuemax="stats.PutCount + stats.DeleteCount"
                                                >
                                                </div>
                                                <div 
                                                    class="progress-bar bg-danger" 
                                                    role="progressbar" 
                                                    :style="{
                                                        width: (stats.DeleteCount / (stats.PutCount + stats.DeleteCount) * 100) + '%'
                                                    }"
                                                    aria-valuenow="{{ stats.DeleteCount }}"
                                                    aria-valuemin="0"
                                                    :aria-valuemax="stats.PutCount + stats.DeleteCount"
                                                >
                                                </div>
                                            </div>
                                            <small class="text-muted">
                                                {{ Math.round(stats.PutCount / (stats.PutCount + stats.DeleteCount) * 100) }}% / {{ Math.round(stats.DeleteCount / (stats.PutCount + stats.DeleteCount) * 100) }}%
                                            </small>
                                        </td>
                                    </tr>
                                </tbody>
                                <tfoot>
                                    <tr>
                                        <th colspan="2">总计</th>
                                        <th>{{ Object.values(monitorData.keyChangeStats).reduce((sum, stats) => sum + stats.PutCount, 0) }}</th>
                                        <th>{{ Object.values(monitorData.keyChangeStats).reduce((sum, stats) => sum + stats.DeleteCount, 0) }}</th>
                                        <th>
                                            <span :class="{
                                                'text-success': Object.values(monitorData.keyChangeStats).reduce((sum, stats) => sum + (stats.PutCount - stats.DeleteCount), 0) > 0,
                                                'text-danger': Object.values(monitorData.keyChangeStats).reduce((sum, stats) => sum + (stats.PutCount - stats.DeleteCount), 0) < 0,
                                                'text-muted': Object.values(monitorData.keyChangeStats).reduce((sum, stats) => sum + (stats.PutCount - stats.DeleteCount), 0) === 0
                                            }">
                                                {{ Object.values(monitorData.keyChangeStats).reduce((sum, stats) => sum + (stats.PutCount - stats.DeleteCount), 0) }}
                                            </span>
                                        </th>
                                        <th></th>
                                    </tr>
                                </tfoot>
                            </table>
                        </div>
                        <div v-else class="alert alert-info">
                            <i class="bi bi-info-circle me-2"></i>
                            键值变更统计数据未获取，请刷新数据
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `
};
