package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestBackupAndRestore 测试备份和恢复功能
func TestBackupAndRestore(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_test")
	backupDir := filepath.Join(testDir, "backups")
	testDbPath := filepath.Join(testDir, "test_db")
	restoreDbPath := filepath.Join(testDir, "restore_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 写入测试数据
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	for k, v := range testData {
		if err := testDb.Put([]byte(k), []byte(v)); err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
	}

	// 创建备份管理器
	backupManager := NewBackupManager(testDb)

	// 测试备份功能
	backupPath, err := backupManager.Backup(backupDir)
	if err != nil {
		t.Fatalf("Failed to backup database: %v", err)
	}

	// 验证备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatalf("Backup file not created: %s", backupPath)
	}

	// 测试备份验证功能
	valid, err := backupManager.ValidateBackup(backupPath)
	if err != nil {
		t.Fatalf("Failed to validate backup: %v", err)
	}
	if !valid {
		t.Fatalf("Backup validation failed")
	}

	// 关闭测试数据库
	testDb.Close()
	storage.KVDb = nil

	// 创建恢复目标数据库
	restoreDb, err := storage.OpenDefaultDb(restoreDbPath)
	if err != nil {
		t.Fatalf("Failed to open restore database: %v", err)
	}
	defer restoreDb.Close()

	// 测试恢复功能
	error := backupManager.Restore(backupPath)
	if error != nil {
		t.Fatalf("Failed to restore database: %v", error)
	}

	// 验证恢复后的数据
	for k, expectedV := range testData {
		v, err := restoreDb.Get([]byte(k))
		if err != nil {
			t.Fatalf("Failed to get restored data for key %s: %v", k, err)
		}
		if string(v) != expectedV {
			t.Fatalf("Restored data mismatch for key %s: expected %s, got %s", k, expectedV, string(v))
		}
	}

	t.Logf("Backup and restore test passed successfully")
}

// TestBackupWithOptions 测试带选项的备份功能
func TestBackupWithOptions(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_test_options")
	backupDir := filepath.Join(testDir, "backups")
	testDbPath := filepath.Join(testDir, "test_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 写入测试数据
	if err := testDb.Put([]byte("key1"), []byte("value1")); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// 创建备份管理器
	backupManager := NewBackupManager(testDb)

	// 测试带选项的备份
	options := BackupOptions{
		Compress: true,
	}
	backupPath, err := backupManager.BackupWithOptions(backupDir, options)
	if err != nil {
		t.Fatalf("Failed to backup with options: %v", err)
	}

	// 验证备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatalf("Backup file not created: %s", backupPath)
	}

	t.Logf("Backup with options test passed successfully")
}

// TestBackupWithoutDatabase 测试源数据库未打开时的备份操作
func TestBackupWithoutDatabase(t *testing.T) {
	// 测试目录设置
	backupDir := filepath.Join(os.TempDir(), "sfsdb_test_nodb")

	// 清理测试目录
	defer func() {
		os.RemoveAll(backupDir)
	}()

	// 确保KVDb为nil
	storage.KVDb = nil

	// 创建备份管理器
	backupManager := NewBackupManager(nil)

	// 测试备份操作，应该失败
	_, err := backupManager.Backup(backupDir)
	if err == nil {
		t.Fatalf("Backup should fail when source database is not open")
	}

	t.Logf("Backup without database test passed successfully")
}

// TestRestoreNonExistentBackup 测试备份文件不存在时的恢复操作
func TestRestoreNonExistentBackup(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_test_nonexistent")
	testDbPath := filepath.Join(testDir, "test_db")
	nonExistentBackup := filepath.Join(testDir, "non_existent_backup")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 创建备份管理器
	backupManager := NewBackupManager(testDb)

	// 测试恢复操作，应该失败
	err = backupManager.Restore(nonExistentBackup)
	if err == nil {
		t.Fatalf("Restore should fail when backup file does not exist")
	}

	t.Logf("Restore non-existent backup test passed successfully")
}

// TestValidateNonExistentBackup 测试验证不存在的备份文件
func TestValidateNonExistentBackup(t *testing.T) {
	// 测试不存在的备份文件
	nonExistentBackup := filepath.Join(os.TempDir(), "sfsdb_test_nonexistent", "non_existent_backup")

	// 创建备份管理器
	backupManager := NewBackupManager(nil)

	// 测试验证操作，应该失败
	_, err := backupManager.ValidateBackup(nonExistentBackup)
	if err == nil {
		t.Fatalf("Validate should fail when backup file does not exist")
	}

	t.Logf("Validate non-existent backup test passed successfully")
}

// TestBackupDirectoryCreation 测试备份目录不存在时的备份操作
func TestBackupDirectoryCreation(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_test")
	// 创建一个深度嵌套的备份目录路径
	backupDir := filepath.Join(testDir, "level1", "level2", "level3", "backups")
	testDbPath := filepath.Join(testDir, "test_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 写入测试数据
	if err := testDb.Put([]byte("key1"), []byte("value1")); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// 创建备份管理器
	backupManager := NewBackupManager(testDb)

	// 测试备份操作，应该自动创建目录
	backupPath, err := backupManager.Backup(backupDir)
	if err != nil {
		t.Fatalf("Failed to backup with directory creation: %v", err)
	}

	// 验证备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatalf("Backup file not created: %s", backupPath)
	}

	t.Logf("Backup directory creation test passed successfully")
}
