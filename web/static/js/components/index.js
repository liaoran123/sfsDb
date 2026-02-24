/**
 * 组件索引文件
 * 导出所有Vue组件
 */

import StatusComponent from './StatusComponent.js';
import SystemComponent from './SystemComponent.js';
import ConfigComponent from './ConfigComponent.js';
import BackupComponent from './BackupComponent.js';
import MemoryComponent from './MemoryComponent.js';
import KeyComponent from './KeyComponent.js';
import IndexStatsComponent from './IndexStatsComponent.js';


import DataQueryComponent from './DataQueryComponent.js';
import ImportExportComponent from './ImportExportComponent.js';

// 导出所有组件
export {
    StatusComponent,
    SystemComponent,
    ConfigComponent,
    BackupComponent,
    MemoryComponent,
    KeyComponent,
    IndexStatsComponent,
    TransactionStatsComponent,
    DataQueryComponent,
    ImportExportComponent
};

// 组件映射，用于动态加载
const components = {
    status: StatusComponent,
    system: SystemComponent,
    config: ConfigComponent,
    backup: BackupComponent,
    memory: MemoryComponent,
    key: KeyComponent,
    indexStats: IndexStatsComponent,

    dataQuery: DataQueryComponent,
    importExport: ImportExportComponent
};

export default components;
