package util

//全文索引算法
// CombineBytes 接收多个字节数组，使用指定分隔符返回第一个数组与其他数组的所有可能拼接组合
// 返回的结果是从对象池获取的，外部调用者使用后需要通过 PutBytesArray 归还到对象池
func CombineBytes(arrays [][][]byte, sep []byte) [][]byte {
	if len(arrays) == 0 {
		return [][]byte{}
	}
	// 转义分隔符
	//escapedSep := Bytes(sep).Escape()
	// 从第一个数组开始，复制元素以避免修改原始数据
	result := GetBytesArray() //全文索引数据量大，所以使用对象池，避免频繁分配内存

	// 初始化第一个数组，转义所有元素
	for _, b := range arrays[0] {
		result = append(result, append([]byte{}, b...)) // 复制数组，避免共享底层数组
	}

	// 处理剩余数组，生成所有可能的拼接组合
	for _, array := range arrays[1:] {
		newResult := GetBytesArray() // 使用对象池获取新的切片
		for _, existing := range result {
			// existing 已经是转义过的结果，不需要再次转义
			for _, nextBytes := range array {
				//nextBytes = Bytes(nextBytes).Escape() // 转义新的数组元素
				// 创建新的组合：existing + sep + nextBytes
				combined := make([]byte, 0, len(existing)+len(sep)+len(nextBytes))
				combined = append(combined, existing...)
				combined = append(combined, sep...)
				combined = append(combined, nextBytes...)
				newResult = append(newResult, combined)
			}
		}
		// 旧的result将被覆盖，外部调用者不会使用，所以可以安全释放
		PutBytesArray(result) // 归还旧的result到对象池
		result = newResult
	}
	// 直接返回result，外部调用者使用后需要归还到对象池
	return result
}

// CombineStrings 接收多个字符串数组，使用指定分隔符返回第一个数组与其他数组的所有可能拼接组合
// 这是 CombineBytes 的字符串包装版本
func CombineStrings(arrays [][]string, sep string) []string {
	// 转换为字节数组
	byteArrays := make([][][]byte, len(arrays))
	for i, array := range arrays {
		byteArrays[i] = make([][]byte, len(array))
		for j, s := range array {
			byteArrays[i][j] = []byte(s)
		}
	}

	// 调用 CombineBytes，获取对象池中的结果
	byteResult := CombineBytes(byteArrays, []byte(sep))
	// 转换回字符串数组前，需要先复制结果，因为 byteResult 会被归还到对象池
	defer PutBytesArray(byteResult) // 确保 byteResult 被归还到对象池

	// 转换回字符串数组
	result := make([]string, len(byteResult))
	for i, b := range byteResult {
		result[i] = string(b)
	}

	return result
}
