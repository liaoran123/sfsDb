package wal

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

// LogType 日志类型
type LogType uint8

const (
	// LogTypeBegin 事务开始
	LogTypeBegin LogType = iota
	// LogTypePut 写入操作
	LogTypePut
	// LogTypeDelete 删除操作
	LogTypeDelete
	// LogTypeCommit 事务提交
	LogTypeCommit
	// LogTypeRollback 事务回滚
	LogTypeRollback
	// LogTypeCheckpoint 检查点
	LogTypeCheckpoint
)

// LogRecord 日志记录
type LogRecord struct {
	Type      LogType
	TxID      uint64
	Key       []byte
	Value     []byte
	Timestamp int64
}

// WAL 事务日志管理器
type WAL struct {
	db            *leveldb.DB
	logDir        string
	currentLogID  uint64
	logFile       *os.File
	logFileMutex  sync.Mutex
	checkpointMtx sync.Mutex
}

// NewWAL 创建新的WAL实例
func NewWAL(db *leveldb.DB, logDir string) (*WAL, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// 找到当前最大的日志文件ID
	currentLogID, err := findCurrentLogID(logDir)
	if err != nil {
		return nil, err
	}

	// 创建或打开当前日志文件
	logFile, err := openLogFile(logDir, currentLogID)
	if err != nil {
		return nil, err
	}

	return &WAL{
		db:           db,
		logDir:       logDir,
		currentLogID: currentLogID,
		logFile:      logFile,
	}, nil
}

// findCurrentLogID 找到当前最大的日志文件ID
func findCurrentLogID(logDir string) (uint64, error) {
	files, err := os.ReadDir(logDir)
	if err != nil {
		return 0, err
	}

	var maxID uint64
	for _, file := range files {
		if !file.IsDir() {
			var id uint64
			if _, err := fmt.Sscanf(file.Name(), "wal_%016x.log", &id); err == nil {
				if id > maxID {
					maxID = id
				}
			}
		}
	}

	return maxID, nil
}

// openLogFile 打开或创建日志文件
func openLogFile(logDir string, logID uint64) (*os.File, error) {
	logPath := filepath.Join(logDir, fmt.Sprintf("wal_%016x.log", logID))
	return os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

// WriteLog 写入日志记录
func (w *WAL) WriteLog(record LogRecord) error {
	w.logFileMutex.Lock()
	defer w.logFileMutex.Unlock()

	// 序列化日志记录
	buffer := make([]byte, 0, 1024)

	// 写入类型
	buffer = append(buffer, byte(record.Type))

	// 写入事务ID
	txIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(txIDBytes, record.TxID)
	buffer = append(buffer, txIDBytes...)

	// 写入时间戳
	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(record.Timestamp))
	buffer = append(buffer, timestampBytes...)

	// 写入键长度和键
	keyLen := len(record.Key)
	keyLenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(keyLenBytes, uint32(keyLen))
	buffer = append(buffer, keyLenBytes...)
	buffer = append(buffer, record.Key...)

	// 写入值长度和值
	valueLen := len(record.Value)
	valueLenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueLenBytes, uint32(valueLen))
	buffer = append(buffer, valueLenBytes...)
	buffer = append(buffer, record.Value...)

	// 写入记录长度
	recordLen := len(buffer)
	recordLenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(recordLenBytes, uint32(recordLen))

	// 写入文件
	if _, err := w.logFile.Write(recordLenBytes); err != nil {
		return err
	}
	if _, err := w.logFile.Write(buffer); err != nil {
		return err
	}

	// 刷新到磁盘
	return w.logFile.Sync()
}

// RotateLog 滚动日志文件
func (w *WAL) RotateLog() error {
	w.logFileMutex.Lock()
	defer w.logFileMutex.Unlock()

	// 关闭当前日志文件
	if err := w.logFile.Close(); err != nil {
		return err
	}

	// 递增日志ID
	w.currentLogID++

	// 打开新的日志文件
	var err error
	w.logFile, err = openLogFile(w.logDir, w.currentLogID)
	return err
}

// CreateCheckpoint 创建检查点
func (w *WAL) CreateCheckpoint() error {
	w.checkpointMtx.Lock()
	defer w.checkpointMtx.Unlock()

	// 写入检查点记录
	checkpointRecord := LogRecord{
		Type:      LogTypeCheckpoint,
		TxID:      0, // 检查点没有事务ID
		Timestamp: time.Now().UnixNano(),
	}

	if err := w.WriteLog(checkpointRecord); err != nil {
		return err
	}

	// 滚动日志文件
	return w.RotateLog()
}

// Recover 恢复事务
func (w *WAL) Recover() error {
	// 读取所有日志文件
	logFiles, err := w.getLogFiles()
	if err != nil {
		return err
	}

	// 处理每个日志文件
	for _, logFile := range logFiles {
		if err := w.processLogFile(logFile); err != nil {
			return err
		}
	}

	return nil
}

// getLogFiles 获取所有日志文件
func (w *WAL) getLogFiles() ([]string, error) {
	files, err := os.ReadDir(w.logDir)
	if err != nil {
		return nil, err
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() {
			if _, err := fmt.Sscanf(file.Name(), "wal_%016x.log", &struct{}{}); err == nil {
				logFiles = append(logFiles, filepath.Join(w.logDir, file.Name()))
			}
		}
	}

	return logFiles, nil
}

// processLogFile 处理单个日志文件
func (w *WAL) processLogFile(logPath string) error {
	file, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 读取并处理日志记录
	for {
		// 读取记录长度
		recordLenBytes := make([]byte, 4)
		if _, err := file.Read(recordLenBytes); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		recordLen := binary.LittleEndian.Uint32(recordLenBytes)

		// 读取记录数据
		recordData := make([]byte, recordLen)
		if _, err := file.Read(recordData); err != nil {
			return err
		}

		// 解析记录
		record, err := w.parseLogRecord(recordData)
		if err != nil {
			return err
		}

		// 处理记录
		if err := w.processLogRecord(record); err != nil {
			return err
		}
	}

	return nil
}

// parseLogRecord 解析日志记录
func (w *WAL) parseLogRecord(data []byte) (LogRecord, error) {
	var record LogRecord
	pos := 0

	// 解析类型
	record.Type = LogType(data[pos])
	pos++

	// 解析事务ID
	record.TxID = binary.LittleEndian.Uint64(data[pos : pos+8])
	pos += 8

	// 解析时间戳
	record.Timestamp = int64(binary.LittleEndian.Uint64(data[pos : pos+8]))
	pos += 8

	// 解析键长度和键
	keyLen := binary.LittleEndian.Uint32(data[pos : pos+4])
	pos += 4
	record.Key = data[pos : pos+int(keyLen)]
	pos += int(keyLen)

	// 解析值长度和值
	valueLen := binary.LittleEndian.Uint32(data[pos : pos+4])
	pos += 4
	record.Value = data[pos : pos+int(valueLen)]

	return record, nil
}

// processLogRecord 处理日志记录
func (w *WAL) processLogRecord(record LogRecord) error {
	switch record.Type {
	case LogTypeBegin:
		// 事务开始，不需要特殊处理
	case LogTypePut:
		// 写入操作，在恢复时需要重新执行
		batch := new(leveldb.Batch)
		batch.Put(record.Key, record.Value)
		return w.db.Write(batch, nil)
	case LogTypeDelete:
		// 删除操作，在恢复时需要重新执行
		batch := new(leveldb.Batch)
		batch.Delete(record.Key)
		return w.db.Write(batch, nil)
	case LogTypeCommit:
		// 事务提交，不需要特殊处理
	case LogTypeRollback:
		// 事务回滚，不需要特殊处理
	case LogTypeCheckpoint:
		// 检查点，不需要特殊处理
	}

	return nil
}

// Close 关闭WAL
func (w *WAL) Close() error {
	w.logFileMutex.Lock()
	defer w.logFileMutex.Unlock()

	return w.logFile.Close()
}
