package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
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

// compressBackup 压缩备份文件
// 参数:
//   backupPath: 备份文件路径
// 返回:
//   string: 压缩后的文件路径
//   error: 错误信息

func compressBackup(backupPath string) (string, error) {
	// 生成压缩文件名
	compressedPath := backupPath + ".zip"

	// 创建压缩文件
	zipFile, err := os.Create(compressedPath)
	if err != nil {
		return "", fmt.Errorf("failed to create compressed file: %v", err)
	}
	defer zipFile.Close()

	// 创建 zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历备份目录中的所有文件
	var errWalk error
	filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errWalk = err
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(backupPath, path)
		if err != nil {
			errWalk = err
			return err
		}

		// 创建 zip 文件中的条目
		zipEntry, err := zipWriter.Create(relPath)
		if err != nil {
			errWalk = err
			return err
		}

		// 打开源文件
		srcFile, err := os.Open(path)
		if err != nil {
			errWalk = err
			return err
		}
		defer srcFile.Close()

		// 复制文件内容到 zip 条目
		if _, err := io.Copy(zipEntry, srcFile); err != nil {
			errWalk = err
			return err
		}

		return nil
	})

	if errWalk != nil {
		// 清理压缩文件
		os.Remove(compressedPath)
		return "", fmt.Errorf("failed to walk backup directory: %v", errWalk)
	}

	return compressedPath, nil
}

// decompressBackup 解压备份文件
// 参数:
//   compressedPath: 压缩备份文件路径
//   targetPath: 解压目标路径
// 返回:
//   error: 错误信息

func decompressBackup(compressedPath, targetPath string) error {
	// 打开压缩文件
	zipFile, err := os.Open(compressedPath)
	if err != nil {
		return fmt.Errorf("failed to open compressed file: %v", err)
	}
	defer zipFile.Close()

	// 获取文件大小
	fileInfo, err := zipFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// 创建 zip reader
	zipReader, err := zip.NewReader(zipFile, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %v", err)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// 解压所有文件
	for _, file := range zipReader.File {
		// 计算目标文件路径
		targetFile := filepath.Join(targetPath, file.Name)

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
			return fmt.Errorf("failed to create directory for file: %v", err)
		}

		// 打开 zip 文件中的条目
		zipEntry, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip entry: %v", err)
		}

		// 创建目标文件
		dstFile, err := os.Create(targetFile)
		if err != nil {
			zipEntry.Close()
			return fmt.Errorf("failed to create target file: %v", err)
		}

		// 复制文件内容
		if _, err := io.Copy(dstFile, zipEntry); err != nil {
			zipEntry.Close()
			dstFile.Close()
			return fmt.Errorf("failed to copy file content: %v", err)
		}

		// 关闭文件
		zipEntry.Close()
		dstFile.Close()
	}

	return nil
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
	log.Println("Starting backup operation...")
	log.Printf("Backup directory: %s", path)

	// 确保备份目录存在
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("Failed to create backup directory: %v", err)
		return "", fmt.Errorf("failed to create backup directory: %v", err)
	}
	log.Println("Backup directory created successfully")

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(path, "backup_"+timestamp)
	log.Printf("Backup file: %s", backupFile)

	// 执行批量备份
	log.Println("Executing batch backup...")
	if err := storage.BatchBackupDb(backupFile); err != nil {
		log.Printf("Failed to batch backup database: %v", err)
		return "", fmt.Errorf("failed to batch backup database: %v", err)
	}
	log.Println("Batch backup executed successfully")

	log.Printf("Backup completed successfully. Backup file: %s", backupFile)
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
	log.Println("Starting backup operation with options...")
	log.Printf("Backup directory: %s", path)
	log.Printf("Compress option: %v", options.Compress)

	// 确保备份目录存在
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("Failed to create backup directory: %v", err)
		return "", fmt.Errorf("failed to create backup directory: %v", err)
	}
	log.Println("Backup directory created successfully")

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(path, "backup_"+timestamp)
	log.Printf("Backup file: %s", backupFile)

	// 执行批量备份
	log.Println("Executing batch backup...")
	if err := storage.BatchBackupDb(backupFile); err != nil {
		log.Printf("Failed to batch backup database: %v", err)
		return "", fmt.Errorf("failed to batch backup database: %v", err)
	}
	log.Println("Batch backup executed successfully")

	// 根据 options 执行额外的操作，如压缩
	if options.Compress {
		log.Println("Compressing backup file...")
		// 压缩备份文件
		compressedFile, err := compressBackup(backupFile)
		if err != nil {
			// 压缩失败，删除未压缩的备份文件
			log.Printf("Failed to compress backup: %v", err)
			os.RemoveAll(backupFile)
			return "", fmt.Errorf("failed to compress backup: %v", err)
		}
		// 删除未压缩的备份文件
		os.RemoveAll(backupFile)
		backupFile = compressedFile
		log.Printf("Backup compressed successfully. Compressed file: %s", backupFile)
	}

	log.Printf("Backup with options completed successfully. Backup file: %s", backupFile)
	return backupFile, nil
}

// Restore 恢复数据库
// 参数:
//   backupPath: 备份文件路径
// 返回:
//   error: 错误信息

func (bm *BackupManager) Restore(backupPath string) error {
	log.Println("Starting restore operation...")
	log.Printf("Backup file: %s", backupPath)

	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		log.Printf("Backup file does not exist: %s", backupPath)
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}
	log.Println("Backup file exists")

	// 处理压缩备份文件
	var actualBackupPath string

	// 检查是否是压缩文件
	if filepath.Ext(backupPath) == ".zip" {
		log.Println("Detected compressed backup file")
		// 创建临时目录用于解压
		tempDir, err := os.MkdirTemp("", "sfsdb_backup_")
		if err != nil {
			log.Printf("Failed to create temporary directory: %v", err)
			return fmt.Errorf("failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		log.Printf("Created temporary directory for decompression: %s", tempDir)

		// 解压备份文件
		log.Println("Decompressing backup file...")
		if err := decompressBackup(backupPath, tempDir); err != nil {
			log.Printf("Failed to decompress backup: %v", err)
			return fmt.Errorf("failed to decompress backup: %v", err)
		}
		log.Println("Backup file decompressed successfully")

		actualBackupPath = tempDir
	} else {
		// 非压缩文件，直接使用
		log.Println("Detected uncompressed backup file")
		actualBackupPath = backupPath
	}

	// 打开备份文件作为源数据库
	log.Println("Opening backup file as source database...")
	backupDb, err := storage.NewLevelDBStore(actualBackupPath, nil)
	if err != nil {
		log.Printf("Failed to open backup file as database: %v", err)
		return fmt.Errorf("failed to open backup file as database: %v", err)
	}
	defer backupDb.Close()
	log.Println("Backup file opened successfully as source database")

	// 获取当前的目标数据库
	targetDb := storage.KVDb
	if targetDb == nil {
		log.Println("Target database is not open")
		return storage.NewError("target database is not open")
	}
	log.Println("Target database is ready")

	// 创建备份数据库的全库遍历迭代器
	iter := backupDb.Iterator(nil, nil)
	defer iter.Release()

	// 遍历所有记录，将它们写入目标数据库
	log.Println("Starting data restoration...")
	count := 0
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()
		// 将记录写入目标数据库
		if err := targetDb.Put(key, value); err != nil {
			log.Printf("Failed to write record to target database: %v", err)
			return fmt.Errorf("failed to write record to target database: %v", err)
		}
		count++
	}
	log.Printf("Data restoration completed successfully. Restored %d records", count)

	return nil
}

// ValidateBackup 验证备份文件完整性
// 参数:
//   backupPath: 备份文件路径
// 返回:
//   bool: 是否有效
//   error: 错误信息

func (bm *BackupManager) ValidateBackup(backupPath string) (bool, error) {
	log.Println("Starting backup validation operation...")
	log.Printf("Backup file: %s", backupPath)

	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		log.Printf("Backup file does not exist: %s", backupPath)
		return false, fmt.Errorf("backup file does not exist: %s", backupPath)
	}
	log.Println("Backup file exists")

	// 处理压缩备份文件
	var actualBackupPath string

	// 检查是否是压缩文件
	if filepath.Ext(backupPath) == ".zip" {
		log.Println("Detected compressed backup file")
		// 创建临时目录用于解压
		tempDir, err := os.MkdirTemp("", "sfsdb_backup_")
		if err != nil {
			log.Printf("Failed to create temporary directory: %v", err)
			return false, fmt.Errorf("failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		log.Printf("Created temporary directory for decompression: %s", tempDir)

		// 解压备份文件
		log.Println("Decompressing backup file...")
		if err := decompressBackup(backupPath, tempDir); err != nil {
			log.Printf("Failed to decompress backup: %v", err)
			return false, fmt.Errorf("failed to decompress backup: %v", err)
		}
		log.Println("Backup file decompressed successfully")

		actualBackupPath = tempDir
	} else {
		// 非压缩文件，直接使用
		log.Println("Detected uncompressed backup file")
		actualBackupPath = backupPath
	}

	// 尝试打开备份文件作为数据库
	log.Println("Opening backup file as database for validation...")
	backupDb, err := storage.NewLevelDBStore(actualBackupPath, nil)
	if err != nil {
		log.Printf("Failed to open backup file as database: %v", err)
		return false, fmt.Errorf("failed to open backup file as database: %v", err)
	}
	defer backupDb.Close()
	log.Println("Backup file opened successfully as database")

	// 检查备份文件的结构完整性
	// 尝试创建迭代器，验证数据库结构
	log.Println("Validating backup file structure...")
	iter := backupDb.Iterator(nil, nil)
	defer iter.Release()

	// 验证迭代器是否有效
	if !iter.Valid() {
		// 即使是空数据库，迭代器也应该是有效的（只是没有数据）
		// 这里不返回错误，因为空数据库也是有效的备份
		log.Println("Backup contains no data (empty database)")
	} else {
		log.Println("Backup structure is valid")
	}

	// 验证备份文件中的数据格式
	// 尝试读取一些数据，验证格式是否正确
	log.Println("Validating backup data format...")
	hasData := false
	for iter.First(); iter.Valid() && !hasData; iter.Next() {
		key := iter.Key()
		// 检查键值是否为空
		if len(key) == 0 {
			log.Println("Backup contains empty key")
			return false, fmt.Errorf("backup contains empty key")
		}
		// 检查值是否为空（允许空值，因为某些场景下可能需要存储空值）
		hasData = true
	}

	if hasData {
		log.Println("Backup data format is valid")
	} else {
		log.Println("Backup contains no data (empty database)")
	}

	// 检查备份文件大小是否合理
	// 这里可以根据实际情况添加大小检查逻辑
	// 例如：检查文件大小是否大于某个最小值，或者是否小于某个最大值

	// 验证成功
	log.Println("Backup validation completed successfully")
	return true, nil
}
