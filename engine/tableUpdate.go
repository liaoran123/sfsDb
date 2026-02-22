package engine

import (
	"github.com/liaoran123/sfsDb/storage"
)

// 更新记录，不支持修改主键字段
// fields *map[string]any 主键值，可能是组合主键

func (t *Table) Update(fields *map[string]any, batchs ...storage.Batch) error {
	// 使用 UpdateImpl
	updateImpl := NewUpdateImpl(t, fields)
	if err := updateImpl.CheckParams(); err != nil {
		return err
	}
	updateImpl.PrepareBatch(batchs...)
	if err := updateImpl.ReadRecord(); err != nil {
		return err
	}
	if err := updateImpl.PrepareUpdateFields(); err != nil {
		return err
	}
	if err := updateImpl.GetFieldsBytes(); err != nil {
		return err
	}
	updateImpl.DeleteOldRecord()
	updateImpl.AddNewRecord()
	if err := updateImpl.Commit(); err != nil {
		return err
	}
	return nil
}
