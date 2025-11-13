package db

/*
ACID 特性	LevelDB 实现方式
原子性	WriteBatch 批量提交
一致性	应用层逻辑约束 + 原子写
隔离性	单线程写入，无并发修改
持久性	写前日志（WAL）+ SSTable 持久化

通过ApproximateSize()方法监控批次大小，防止内存溢出：sDB
WriteBatch.SetMaxBatchSize(size int)

在合并两个 WriteBatch 的时候，也会累计两部分的 count 的值，如下 WriteBatchInternal::Append 方法：

void WriteBatchInternal::Append(WriteBatch* dst, const WriteBatch* src) {
  SetCount(dst, Count(dst) + Count(src));
  assert(src->rep_.size() >= kHeader);
  dst->rep_.append(src->rep_.data() + kHeader, src->rep_.size() - kHeader);
}

*/
