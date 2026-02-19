/**
 * 索引统计组件
 * 显示系统的索引使用统计信息
 */
export default {
    name: 'IndexStatsComponent',
    data() {
        return {
            loading: false,
            error: null,
            monitorData: {
                indexStats: null
            }
        };
    },
    methods: {
        async refreshData() {
            this.loading = true;
            this.error = null;
            try {
                const data = await window.api.getMonitor();
                this.monitorData.indexStats = data.indexStats;
                console.log('索引统计数据刷新成功:', data);
                window.utils.showNotification('索引统计数据刷新成功', 'success');
            } catch (err) {
                this.error = '获取索引统计失败: ' + err.message;
                console.error('获取索引统计失败:', err);
                window.utils.showNotification('获取索引统计失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        formatTime(nanoseconds) {
            return (nanoseconds / 1000000).toFixed(2) + 'ms';
        }
    },
    mounted() {
        this.refreshData();
        console.log('IndexStatsComponent 已加载');
    },
    template: `
        <div class="card">
            <div class="card-body">
                <div class="d-flex justify-content-between align-items-center mb-4">
                    <h5 class="card-title mb-0">索引统计</h5>
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
                    <p class="mt-2 text-muted">正在加载索引统计数据...</p>
                </div>
                
                <div v-else>
                    <!-- 索引统计信息 -->
                    <div class="mb-6">
                        <h6 class="mb-3">索引使用统计</h6>
                        <div v-if="monitorData.indexStats" class="table-responsive">
                            <table class="table table-striped table-hover">
                                <thead>
                                    <tr>
                                        <th>表名</th>
                                        <th>索引名</th>
                                        <th>搜索类型</th>
                                        <th>使用次数</th>
                                        <th>平均搜索耗时</th>
                                        <th>总搜索耗时</th>
                                        <th>最大搜索耗时</th>
                                        <th>最小搜索耗时</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(stats, key) in monitorData.indexStats" :key="key">
                                        <td>{{ stats.tblName }}</td>
                                        <td>{{ stats.indxName }}</td>
                                        <td>{{ stats.searchType }}</td>
                                        <td>{{ stats.count }}</td>
                                        <td>{{ formatTime(stats.avgTime) }}</td>
                                        <td>{{ formatTime(stats.totalTime) }}</td>
                                        <td>{{ formatTime(stats.maxTime) }}</td>
                                        <td>{{ formatTime(stats.minTime) }}</td>
                                    </tr>
                                </tbody>
                                <tfoot>
                                    <tr>
                                        <th colspan="3">总计</th>
                                        <th>{{ Object.values(monitorData.indexStats).reduce((sum, stats) => sum + stats.count, 0) }}</th>
                                        <th>
                                            {{ formatTime(Object.values(monitorData.indexStats).reduce((sum, stats) => sum + stats.avgTime, 0) / Object.values(monitorData.indexStats).length) }}
                                        </th>
                                        <th>
                                            {{ formatTime(Object.values(monitorData.indexStats).reduce((sum, stats) => sum + stats.totalTime, 0)) }}
                                        </th>
                                        <th>
                                            {{ formatTime(Math.max(...Object.values(monitorData.indexStats).map(stats => stats.maxTime))) }}
                                        </th>
                                        <th>
                                            {{ formatTime(Math.min(...Object.values(monitorData.indexStats).map(stats => stats.minTime))) }}
                                        </th>
                                    </tr>
                                </tfoot>
                            </table>
                        </div>
                        <div v-else class="alert alert-info">
                            <i class="bi bi-info-circle me-2"></i>
                            索引统计数据未获取，请刷新数据
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `
};
