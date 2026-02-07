package storage

// 备份数据库
func BackupDb(Path string) error {
	//保存当前的源数据库引用
	sourceDb := KVDb
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
