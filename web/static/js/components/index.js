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
import TransactionStatsComponent from './TransactionStatsComponent.js';
import CrudComponent from './CrudComponent.js';
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
    CrudComponent,
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
    transactionStats: TransactionStatsComponent,
    crud: CrudComponent,
    dataQuery: DataQueryComponent,
    importExport: ImportExportComponent
};

export default components;
