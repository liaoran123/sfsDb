// 配置管理组件
export default {
    name: 'ConfigComponent',
    template: `
        <div class="card">
            <div class="card-body">
                <h5 class="card-title mb-4">配置管理</h5>
                
                <!-- 配置信息 -->
                <div class="mb-6">
                    <div v-if="loading" class="text-center">
                        <div class="spinner-border" role="status">
                            <span class="visually-hidden">加载中...</span>
                        </div>
                    </div>
                    <div v-else-if="error" class="alert alert-danger">
                        {{ error }}
                    </div>
                    <div v-else>
                        <!-- 场景配置 -->
                        <div class="card mb-4">
                            <div class="card-body">
                                <h6 class="card-subtitle mb-3">场景配置</h6>
                                <div class="form-group mb-4">
                                    <label for="scenarioSelect" class="form-label">选择场景</label>
                                    <select id="scenarioSelect" v-model="selectedScenario" class="form-select">
                                        <option v-for="scenario in config.Scenarios" :key="scenario" :value="scenario">
                                            {{ scenario }}
                                        </option>
                                    </select>
                                </div>
                                <button class="btn btn-primary" @click="setScenarioConfig" :disabled="!selectedScenario">
                                    应用场景配置
                                </button>
                            </div>
                        </div>

                        <!-- 详细配置 -->
                        <div class="card mb-4">
                            <div class="card-body">
                                <h6 class="card-subtitle mb-3">详细配置</h6>
                                <div class="row">
                                    <div class="col-md-6">
                                        <div class="form-group mb-4">
                                            <label for="writeBuffer" class="form-label">Write Buffer</label>
                                            <input 
                                                type="text" 
                                                id="writeBuffer" 
                                                v-model="configForm.write_buffer" 
                                                class="form-control"
                                                placeholder="例如：67108864 或 64MB"
                                            >
                                        </div>
                                    </div>
                                    <div class="col-md-6">
                                        <div class="form-group mb-4">
                                            <label for="maxOpenFiles" class="form-label">Max Open Files</label>
                                            <input 
                                                type="number" 
                                                id="maxOpenFiles" 
                                                v-model="configForm.max_open_files" 
                                                class="form-control"
                                                placeholder="例如：200"
                                            >
                                        </div>
                                    </div>
                                    <div class="col-md-6">
                                        <div class="form-group mb-4">
                                            <label for="blockCache" class="form-label">Block Cache</label>
                                            <input 
                                                type="text" 
                                                id="blockCache" 
                                                v-model="configForm.block_cache" 
                                                class="form-control"
                                                placeholder="例如：134217728 或 128MB"
                                            >
                                        </div>
                                    </div>
                                    <div class="col-md-6">
                                        <div class="form-group mb-4">
                                            <label for="compression" class="form-label">Compression</label>
                                            <select id="compression" v-model="configForm.compression" class="form-select">
                                                <option value="true">启用</option>
                                                <option value="false">禁用</option>
                                            </select>
                                        </div>
                                    </div>
                                </div>
                                <div class="mt-4">
                                    <button class="btn btn-primary" @click="updateConfig">
                                        更新配置
                                    </button>
                                    <button class="btn btn-secondary ms-2" @click="resetConfig">
                                        重置为默认配置
                                    </button>
                                </div>
                            </div>
                        </div>

                        <!-- 优化建议 -->
                        <div class="card mb-4">
                            <div class="card-body">
                                <h6 class="card-subtitle mb-3">优化建议</h6>
                                <div v-if="suggestions.length > 0">
                                    <ul class="list-group">
                                        <li v-for="(suggestion, index) in suggestions" :key="index" class="list-group-item">
                                            {{ suggestion }}
                                        </li>
                                    </ul>
                                </div>
                                <div v-else class="alert alert-info">
                                    暂无优化建议
                                </div>
                                <button class="btn btn-outline-secondary mt-3" @click="getSuggestions">
                                    刷新优化建议
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `,
    data() {
        return {
            loading: false,
            error: null,
            config: {
                StoreType: '',
                Options: {},
                Scenarios: []
            },
            configForm: {
                write_buffer: '',
                max_open_files: '',
                block_cache: '',
                compression: ''
            },
            selectedScenario: '',
            suggestions: []
        }
    },
    mounted() {
        this.loadConfig();
    },
    methods: {
        async loadConfig() {
            this.loading = true;
            this.error = null;
            
            try {
                const data = await window.api.getConfig();
                if (data.error) {
                    this.error = data.message || data.error;
                } else {
                    // 确保 config 对象包含必要的字段
                    this.config = {
                        StoreType: data.StoreType || '',
                        Options: data.Options || {},
                        Scenarios: data.Scenarios || []
                    };
                    // 初始化表单数据
                    this.configForm.write_buffer = this.config.Options.write_buffer || '';
                    this.configForm.max_open_files = this.config.Options.max_open_files || '';
                    this.configForm.block_cache = this.config.Options.block_cache || '';
                    this.configForm.compression = this.config.Options.compression || '';
                    
                    // 默认选择第一个场景
                    if (this.config.Scenarios.length > 0) {
                        this.selectedScenario = this.config.Scenarios[0];
                    }
                }
            } catch (err) {
                this.error = '错误: ' + err.message;
                console.error('加载配置失败:', err);
            } finally {
                this.loading = false;
            }
        },
        async updateConfig() {
            this.loading = true;
            this.error = null;
            
            try {
                // 逐个更新配置项
                for (const [key, value] of Object.entries(this.configForm)) {
                    if (value !== undefined && value !== null && value !== '') {
                        await window.api.setConfig(key, value);
                    }
                }
                
                window.utils.showNotification('配置更新成功', 'success');
                // 重新加载配置
                this.loadConfig();
            } catch (err) {
                this.error = '错误: ' + err.message;
                console.error('更新配置失败:', err);
                window.utils.showNotification('更新配置失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        async setScenarioConfig() {
            if (!this.selectedScenario) {
                this.error = '请选择场景';
                return;
            }
            
            this.loading = true;
            this.error = null;
            
            try {
                await window.api.setScenarioConfig(this.selectedScenario);
                window.utils.showNotification('场景配置应用成功', 'success');
                // 重新加载配置
                this.loadConfig();
            } catch (err) {
                this.error = '错误: ' + err.message;
                console.error('应用场景配置失败:', err);
                window.utils.showNotification('应用场景配置失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        async resetConfig() {
            if (!confirm('确定要重置为默认配置吗？')) {
                return;
            }
            
            this.loading = true;
            this.error = null;
            
            try {
                await window.api.resetConfig();
                window.utils.showNotification('配置重置成功', 'success');
                // 重新加载配置
                this.loadConfig();
            } catch (err) {
                this.error = '错误: ' + err.message;
                console.error('重置配置失败:', err);
                window.utils.showNotification('重置配置失败: ' + err.message, 'error');
            } finally {
                this.loading = false;
            }
        },
        async getSuggestions() {
            this.loading = true;
            this.error = null;
            
            try {
                const data = await window.api.getConfig();
                if (data.error) {
                    this.error = data.message || data.error;
                } else {
                    // 这里简化处理，实际应该调用专门的优化建议接口
                    this.suggestions = [
                        '对于 IoT 场景，建议使用 \'iot\' 配置',
                        '对于边缘计算场景，建议使用 \'edge\' 配置',
                        '对于嵌入式场景，建议使用 \'embedded\' 配置',
                        '考虑增加 write_buffer 大小以提高写入性能，建议设置为 16MB 或更高',
                        '根据内存情况调整 max_open_files 参数，内存充足时可设置为 100 或更高',
                        '考虑增加 block_cache 大小以提高读取性能，建议设置为 32MB 或更高'
                    ];
                }
            } catch (err) {
                this.error = '错误: ' + err.message;
                console.error('获取优化建议失败:', err);
            } finally {
                this.loading = false;
            }
        }
    }
};
