/**
 * 事务统计组件
 * 显示系统的事务使用统计信息
 */
export default {
    name: 'TransactionStatsComponent',
    data() {
        return {
            loading: false,
            error: null,
            monitorData: {
                transactionStats: null
            }
        };
    },
    methods: {
        async refreshData() {
            this.loading = true;
            this.error = null;
            try {
                const data = await window.api.getMonitor();
                this.monitorData.transactionStats = data.transactionStats;
                console.log('事务统计数据刷新成功:', data);
                window.utils.showNotification('事务统计数据刷新成功', 'success');
            } catch (err) {
                this.error = '获取事务统计失败: ' + err.message;
                console.error('获取事务统计失败:', err);
                window.utils.showNotification('获取事务统计失败: ' + err.message, 'error');
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
        console.log('TransactionStatsComponent 已加载');
    },
    template: `
        <div class="card">
            <div class="card-body">
                <div class="d-flex justify-content-between align-items-center mb-4">
                    <h5 class="card-title mb-0">事务统计</h5>
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
                    <p class="mt-2 text-muted">正在加载事务统计数据...</p>
                </div>
                
                <div v-else>
                    <!-- 事务统计信息 -->
                    <div class="mb-6">
                        <h6 class="mb-3">事务使用统计</h6>
                        <div v-if="monitorData.transactionStats" class="table-responsive">
                            <table class="table table-striped table-hover">
                                <thead>
                                    <tr>
                                        <th>事务ID</th>
                                        <th>表名</th>
                                        <th>隔离级别</th>
                                        <th>是否提交</th>
                                        <th>事务次数</th>
                                        <th>平均耗时</th>
                                        <th>总耗时</th>
                                        <th>最大耗时</th>
                                        <th>最小耗时</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(stats, key) in monitorData.transactionStats" :key="key">
                                        <td>{{ stats.txID }}</td>
                                        <td>{{ stats.tableName }}</td>
                                        <td>{{ stats.isolationLevel }}</td>
                                        <td>
                                            <span class="badge" :class="{
                                                'bg-success': stats.isCommitted,
                                                'bg-danger': !stats.isCommitted
                                            }">
                                                {{ stats.isCommitted ? '是' : '否' }}
                                            </span>
                                        </td>
                                        <td>{{ stats.totalCount }}</td>
                                        <td>{{ formatTime(stats.avgDuration) }}</td>
                                        <td>{{ formatTime(stats.duration) }}</td>
                                        <td>{{ formatTime(stats.maxDuration) }}</td>
                                        <td>{{ formatTime(stats.minDuration) }}</td>
                                    </tr>
                                </tbody>
                                <tfoot>
                                    <tr>
                                        <th colspan="4">总计</th>
                                        <th>{{ Object.values(monitorData.transactionStats).reduce((sum, stats) => sum + stats.totalCount, 0) }}</th>
                                        <th>
                                            {{ formatTime(Object.values(monitorData.transactionStats).reduce((sum, stats) => sum + stats.avgDuration, 0) / Object.values(monitorData.transactionStats).length) }}
                                        </th>
                                        <th>
                                            {{ formatTime(Object.values(monitorData.transactionStats).reduce((sum, stats) => sum + stats.duration, 0)) }}
                                        </th>
                                        <th>
                                            {{ formatTime(Math.max(...Object.values(monitorData.transactionStats).map(stats => stats.maxDuration))) }}
                                        </th>
                                        <th>
                                            {{ formatTime(Math.min(...Object.values(monitorData.transactionStats).map(stats => stats.minDuration))) }}
                                        </th>
                                    </tr>
                                </tfoot>
                            </table>
                        </div>
                        <div v-else class="alert alert-info">
                            <i class="bi bi-info-circle me-2"></i>
                            事务统计数据未获取，请刷新数据
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `
};
