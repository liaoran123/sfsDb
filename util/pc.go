package util

import (
	"encoding/binary"
	"strconv"
)

// intSize 表示 int 类型的位数，32 或 64
const intSize = strconv.IntSize

var EndianOrder = binary.BigEndian // 使用固定大端序
/*
### 实际应用场景
在 SFSDB 项目中，使用固定大端序特别重要的场景包括：

- 索引键的存储和比较 ：确保索引在不同系统上行为一致
- 数据文件的持久化 ：确保数据文件可以在不同系统间移植
- 网络客户端访问 ：简化客户端与服务器之间的通信
### 性能考虑
虽然小端序在某些情况下可能有微小的性能优势，但在现代处理器上，这种差异通常可以忽略不计。而大端序带来的跨平台兼容性和一致性优势，远超过可能的性能差异。

### 结论
对于数据库系统这样需要长期稳定运行、可能跨平台部署的应用，使用固定的大端序是一个更为稳妥和前瞻性的选择。它提供了更好的兼容性、一致性和可维护性，为系统的长期发展奠定了坚实的基础。
*/
/*
该函数不能使用，自己的电脑系统都出现大小端不一致问题。
// 判断大小端
func Endian() binary.ByteOrder {
	var i int = 0x1
	ptr := unsafe.Pointer(&i)
	b := *(*byte)(ptr)
	if b == 1 {
		return binary.LittleEndian
	}
	return binary.BigEndian
}
*/
