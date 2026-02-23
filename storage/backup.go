package storage

// 备份数据库
func BackupDb(Path string) error {
	//保存当前的源数据库引用
	sourceDb := dbManager.GetDB()
	if sourceDb == nil {
		return NewError("源数据库未打开")
	}

	//打开备份目标数据库，不修改全局KVDb
	backupDb, err := NewLevelDBStore(Path, nil)
	if err != nil {
		return err
	}
	defer backupDb.Close()

	//创建源数据库的全库遍历迭代器，使用Iterator方法传入nil作为start和limit
	iter := sourceDb.Iterator(nil, nil)
	defer iter.Release()

	//遍历所有记录
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()
		//将记录写入备份数据库
		if err := backupDb.Put(key, value); err != nil {
			return err
		}
	}
	return nil
}

// BatchBackupDb 批量备份数据库
// Path: 备份路径
// batchSize: 批量操作的大小阈值
func BatchBackupDb(Path string, batchSize ...int) error {
	if len(batchSize) == 0 || batchSize[0] <= 0 {
		batchSize = []int{1000}
	}
	// 使用批量大小
	bs := batchSize[0]
	//保存当前的源数据库引用
	sourceDb := dbManager.GetDB()
	if sourceDb == nil {
		return NewError("源数据库未打开")
	}

	//打开备份目标数据库，不修改全局KVDb
	backupDb, err := NewLevelDBStore(Path, nil)
	if err != nil {
		return err
	}
	defer backupDb.Close()

	//创建源数据库的全库遍历迭代器，使用Iterator方法传入nil作为start和limit
	iter := sourceDb.Iterator(nil, nil)
	defer iter.Release()

	//创建批量操作对象
	batch := backupDb.GetBatch()
	count := 0

	//遍历所有记录
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()
		//将记录添加到批量操作对象中
		batch.Put(key, value)
		count++

		//当批量操作对象的大小达到阈值时，执行批量操作
		if count >= bs {
			if err := backupDb.WriteBatch(batch); err != nil {
				return err
			}
			//重置批量操作对象
			batch.Reset()
			count = 0
		}
	}

	//执行剩余的批量操作
	if count > 0 {
		if err := backupDb.WriteBatch(batch); err != nil {
			return err
		}
	}

	return nil
}
