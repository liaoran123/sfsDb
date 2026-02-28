/**
 * 系统信息组件
 * 显示系统的详细信息，包括表信息和表的详细信息
 */
export default {
    name: 'SystemComponent',
    data() {
        return {
            loading: false,
            error: null,
            systemData: {
                tables: [],
                tableDetails: {}
            }
        };
    },
    methods: {
        async loadSystemData() {
            this.loading = true;
            this.error = null;
            try {
                this.systemData = await window.api.getSystemInfo();
            } catch (err) {
                this.error = '加载系统信息失败: ' + err.message;
                console.error('加载系统信息失败:', err);
            } finally {
                this.loading = false;
            }
        }
    },
    mounted() {
        this.loadSystemData();
        console.log('SystemComponent 已加载');
    },
    template: `
        <div class="card">
            <div class="card-body">
                <h5 class="card-title mb-4">系统信息</h5>
                
                <div v-if="loading" class="text-center">
                    <div class="spinner-border" role="status">
                        <span class="visually-hidden">加载中...</span>
                    </div>
                    <p class="mt-2 text-muted">正在加载系统信息...</p>
                </div>
                
                <div v-else-if="error" class="alert alert-danger">
                    <i class="bi bi-exclamation-triangle me-2"></i>
                    {{ error }}
                    <button class="btn btn-sm btn-outline-danger mt-2" @click="loadSystemData">
                        重试
                    </button>
                </div>
                
                <div v-else>
                    <div class="mb-6">
                        <h6 class="mb-3">表信息</h6>
                        
                        <div v-if="systemData.tables && systemData.tables.length > 0">
                            <div class="table-responsive">
                                <table class="table table-striped">
                                    <thead>
                                        <tr>
                                            <th>表名</th>
                                            <th>表ID</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr v-for="table in systemData.tables" :key="table.ID">
                                            <td>{{ table.Name }}</td>
                                            <td>{{ table.ID }}</td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                            
                            <!-- 表详细信息 -->
                            <div class="mt-4">
                                <h6 class="mb-3">表详细信息</h6>
                                <div 
                                    v-for="(details, tableID) in systemData.tableDetails" 
                                    :key="tableID" 
                                    :id="'tableDetails_' + tableID"
                                    class="collapse mb-4"
                                >
                                    <div class="card">
                                        <div class="card-header">
                                            <h7 class="mb-0">{{ details.name }} (ID: {{ tableID }})</h7>
                                        </div>
                                        <div class="card-body">
                                            <!-- 字段信息 -->
                                            <div class="mb-4">
                                                <h6 class="mb-2">字段信息</h6>
                                                <div v-if="details.fields && details.fields.length > 0">
                                                    <div class="table-responsive">
                                                        <table class="table table-sm table-hover">
                                                            <thead>
                                                                <tr>
                                                                    <th>字段名</th>
                                                                    <th>字段ID</th>
                                                                </tr>
                                                            </thead>
                                                            <tbody>
                                                                <tr v-for="field in details.fields" :key="field.ID">
                                                                    <td>{{ field.Name }}</td>
                                                                    <td>{{ field.ID }}</td>
                                                                </tr>
                                                            </tbody>
                                                        </table>
                                                    </div>
                                                </div>
                                                <div v-else class="text-muted">
                                                    无字段信息
                                                </div>
                                            </div>
                                            
                                            <!-- 索引信息 -->
                                            <div>
                                                <h6 class="mb-2">索引信息</h6>
                                                <div v-if="details.indexes && details.indexes.length > 0">
                                                    <div class="table-responsive">
                                                        <table class="table table-sm table-hover">
                                                            <thead>
                                                                <tr>
                                                                    <th>索引名</th>
                                                                    <th>索引ID</th>
                                                                </tr>
                                                            </thead>
                                                            <tbody>
                                                                <tr v-for="index in details.indexes" :key="index.ID">
                                                                    <td>{{ index.Name }}</td>
                                                                    <td>{{ index.ID }}</td>
                                                                </tr>
                                                            </tbody>
                                                        </table>
                                                    </div>
                                                </div>
                                                <div v-else class="text-muted">
                                                    无索引信息
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div v-else class="alert alert-info">
                            <i class="bi bi-info-circle me-2"></i>
                            无表信息
                        </div>
                    </div>
                    
                    <div class="text-center mt-4">
                        <button class="btn btn-outline-primary" @click="loadSystemData" :disabled="loading">
                            <i class="bi bi-refresh me-1"></i> 刷新信息
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `
};
