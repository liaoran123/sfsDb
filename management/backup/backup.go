package backup

import (
	"os"
	"path/filepath"
	"time"

	"github.com/liaoran123/sfsDb/storage"
)

// BackupOptions 备份选项
type BackupOptions struct {
	Compress bool // 是否压缩备份
	// 其他备份选项可以根据需要扩展
}

// BackupManager 备份管理器
type BackupManager struct {
	store storage.Store
}

// NewBackupManager 创建备份管理器
// 参数:
//   store: 存储实例
// 返回:
//   *BackupManager: 备份管理器实例

func NewBackupManager(store storage.Store) *BackupManager {
	return &BackupManager{
		store: store,
	}
}

// Backup 备份数据库
// 参数:
//   path: 备份路径
// 返回:
//   string: 备份文件路径
//   error: 错误信息

func (bm *BackupManager) Backup(path string) (string, error) {
	// 确保备份目录存在
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(path, "backup_"+timestamp)

	// 执行备份
	if err := storage.BackupDb(backupFile); err != nil {
		return "", err
	}

	return backupFile, nil
}

// BackupWithOptions 带选项的备份
// 参数:
//   path: 备份路径
//   options: 备份选项
// 返回:
//   string: 备份文件路径
//   error: 错误信息

func (bm *BackupManager) BackupWithOptions(path string, options BackupOptions) (string, error) {
	// 确保备份目录存在
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(path, "backup_"+timestamp)

	// 执行备份
	if err := storage.BackupDb(backupFile); err != nil {
		return "", err
	}

	// 这里可以根据 options 执行额外的操作，如压缩
	// 实际实现时，需要根据具体的备份选项进行处理

	return backupFile, nil
}

// Restore 恢复数据库
// 参数:
//   backupPath: 备份文件路径
// 返回:
//   error: 错误信息

func (bm *BackupManager) Restore(backupPath string) error {
	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return err
	}

	// 打开备份文件作为源数据库
	backupDb, err := storage.NewLevelDBStore(backupPath, nil)
	if err != nil {
		return err
	}
	defer backupDb.Close()

	// 获取当前的目标数据库
	targetDb := storage.KVDb
	if targetDb == nil {
		return storage.NewError("目标数据库未打开")
	}

	// 创建备份数据库的全库遍历迭代器
	iter := backupDb.Iterator(nil, nil)
	defer iter.Release()

	// 遍历所有记录，将它们写入目标数据库
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()
		// 将记录写入目标数据库
		if err := targetDb.Put(key, value); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBackup 验证备份文件完整性
// 参数:
//   backupPath: 备份文件路径
// 返回:
//   bool: 是否有效
//   error: 错误信息

func (bm *BackupManager) ValidateBackup(backupPath string) (bool, error) {
	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return false, err
	}

	// 注意：这里需要实现备份文件验证逻辑
	// 实际实现时，需要检查备份文件的完整性和有效性

	// 这里返回 true 作为占位，实际实现需要根据具体情况修改
	return true, nil
}
